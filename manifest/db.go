package manifest

import (
	"context"
	"encoding/json"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/request"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

var workSchema = `
create table if not exists work
(
    id			integer		not null	constraint work_pk	primary key,
    title		text,
    illust_type	integer,
    tags		jsonb,
    page_count	integer,
    create_date	integer,
    update_date	integer,
    masked_date	integer
);

create index if not exists work_title_index on work (title);

create index if not exists work_update_date_index on work (update_date desc);
`

var queryPreparation = `select * from work where id = :id`

var queryAllPreparation = `select * from work`

var insertPreparation = `
insert into work
(id,
 title,
 illust_type,
 tags,
 page_count,
 create_date,
 update_date,
 masked_date)
values (:id,
        :title,
        :illust_type,
        :tags,
        :page_count,
        :create_date,
        :update_date,
        :masked_date);`

var updatePreparation1 = `
update work
set title       = :title,
    illust_type = :illust_type,
    tags        = :tags,
    page_count  = :page_count,
    update_date = :update_date,
    masked_date = :masked_date
where id = :id;`

var updatePreparation2 = `
update work
set masked_date = :masked_date
where id = :id;`

func StartRecord(mCh <-chan request.BookmarkWorkItem, wg *sync.WaitGroup) {
	output := viper.GetString("job.output")
	dbFile := filepath.Join(output, "manifest")
	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		slog.Error("create output dir", "error", err)
	}
	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		slog.Error("unable to connect to db", "error", err, "db", dbFile)
		return
	}
	db.MustExec(workSchema)
	sts := SqlxStmts{
		Query:   prepareStmt(db, queryPreparation, "unable to prepare query"),
		Insert:  prepareStmt(db, insertPreparation, "unable to prepare insert"),
		Update1: prepareStmt(db, updatePreparation1, "unable to prepare update 1"),
		Update2: prepareStmt(db, updatePreparation2, "unable to prepare update 2"),
	}
	for wi := range mCh {
		onceRecord(sts, wi, wg)
	}
}

func onceRecord(sts SqlxStmts, wi request.BookmarkWorkItem, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	var w WorkModel
	err := sts.Query.Get(&w, wi)
	if err != nil {
		w.Id = wi.GetId()
		w.Title = wi.Title
		w.IllustType = wi.IllustType
		bs, err := json.Marshal(wi.Tags)
		if err == nil {
			w.Tags = bs
		}
		w.PageCount = wi.PageCount
		ct, err := time.Parse("2006-01-02T15:04:05-07:00", wi.CreateDate)
		if err == nil {
			cts := ct.UTC().Unix()
			w.CreateDate = cts
		}
		ut, err := time.Parse("2006-01-02T15:04:05-07:00", wi.UpdateDate)
		if err == nil {
			uts := ut.UTC().Unix()
			w.UpdateDate = uts
		}
		if wi.IsMasked {
			w.MaskedDate = time.Now().UTC().Unix()
		} else {
			w.MaskedDate = -1
		}

		slog.Warn("new work", "id", w.Id, "title", w.Title)
		_, err = sts.Insert.Exec(w)
		if err != nil {
			slog.Error("unable to insert work to db", "error", err)
		}
	} else if w.MaskedDate != -1 {
		if wi.IsMasked {
			return
		} else {
			w.Title = wi.Title
			w.IllustType = wi.IllustType
			bs, err := json.Marshal(wi.Tags)
			if err == nil {
				w.Tags = bs
			}
			w.PageCount = wi.PageCount
			ut, err := time.Parse("2006-01-02T15:04:05-07:00", wi.UpdateDate)
			if err == nil {
				uts := ut.UTC().Unix()
				w.UpdateDate = uts
			}
			w.MaskedDate = -1

			slog.Warn("masked -> unmasked", "id", w.Id, "title", w.Title)
			_, err = sts.Update1.Exec(w)
			if err != nil {
				slog.Error("unable to update work to db", "error", err)
			}
		}
	} else {
		if wi.IsMasked {
			w.MaskedDate = time.Now().UTC().Unix()

			slog.Warn("unmasked -> masked", "id", w.Id, "title", w.Title)
			_, err = sts.Update2.Exec(w)
			if err != nil {
				slog.Error("unable to update work to db", "error", err)
			}
		} else {
			ut, err := time.Parse("2006-01-02T15:04:05-07:00", wi.UpdateDate)
			if err != nil {
				ut = time.Now()
			}
			uts := ut.UTC().Unix()
			if err == nil && uts == w.UpdateDate {
				return
			}
			w.Title = wi.Title
			w.IllustType = wi.IllustType
			bs, err := json.Marshal(wi.Tags)
			if err == nil {
				w.Tags = bs
			}
			w.PageCount = wi.PageCount
			w.UpdateDate = uts

			slog.Warn("update", "id", w.Id, "title", w.Title)
			_, err = sts.Update1.Exec(w)
			if err != nil {
				slog.Error("unable to update work to db", "error", err)
			}
		}
	}
}

type SqlxStmts struct {
	Query   *sqlx.NamedStmt
	Insert  *sqlx.NamedStmt
	Update1 *sqlx.NamedStmt
	Update2 *sqlx.NamedStmt
}

func prepareStmt(db *sqlx.DB, sql string, errMsg string) *sqlx.NamedStmt {
	stmt, err := db.PrepareNamed(sql)
	if err != nil {
		slog.Error(errMsg, "error", err)
		panic(err)
	}
	return stmt
}

func StartRead(ch chan<- WorkModel, wg *sync.WaitGroup) {
	output := viper.GetString("job.output")
	dbFile := filepath.Join(output, "manifest")
	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		slog.Error("unable to connect to db", "error", err, "db", dbFile)
		return
	}
	rows, err := db.Queryx(queryAllPreparation)
	if err != nil {
		slog.Error("unable to query manifest", "error", err)
		return
	}
	slog.Info("start walk db assets")
	for rows.Next() {
		var w WorkModel
		err := rows.StructScan(&w)
		if err != nil {
			slog.Error("unable to scan", "error", err)
			continue
		}
		slog.Debug("get work model", "id", w.Id, "title", w.Title, "type", w.IllustType, "count", w.PageCount)
		wg.Add(1)
		ch <- w
	}
}

func StartCheck(ch <-chan WorkModel, wg *sync.WaitGroup) {
	output := viper.GetString("job.output")
	for w := range ch {
		onceCheck(w, output, wg)
	}
}

func onceCheck(w WorkModel, output string, wg *sync.WaitGroup) {
	defer wg.Done()
	ut := time.Unix(w.UpdateDate, 0)
	path := filepath.Join(output, strconv.FormatUint(w.Id, 10), ut.UTC().Format("20060102150405"))
	dir, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Log(context.Background(), -2, "dir does not exist", "path", path)
		} else {
			slog.Error("unable to open dir", "path", path, "error", err)
		}
		return
	}
	files, err := dir.Readdir(0)
	if err != nil {
		slog.Error("unable to read dir", "path", path, "error", err)
		return
	}
	ac := len(files)
	ec := w.PageCount
	if ac < ec {
		slog.Warn("count mismatch", "id", w.Id, "title", w.Title, "type", w.IllustType, "db-count", ec, "dir-count", ac)
	}
}
