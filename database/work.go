package database

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uheee/pixiv-grabber/request"
	"log/slog"
	"sync"
	"time"
)

var sqlQueryWork = `select * from work where id = :id`

var sqlInsertWork = `
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

var sqlUpdateWork1 = `
update work
set title       = :title,
    illust_type = :illust_type,
    tags        = :tags,
    page_count  = :page_count,
    update_date = :update_date,
    masked_date = :masked_date
where id = :id;`

var sqlUpdateWork2 = `
update work
set masked_date = :masked_date
where id = :id;`

func StartRecord(db *sqlx.DB, mCh <-chan request.BookmarkWorkItem, wg *sync.WaitGroup) {
	sts := WorkStmts{
		Query:   prepareStmt(db, sqlQueryWork, "unable to prepare query"),
		Insert:  prepareStmt(db, sqlInsertWork, "unable to prepare insert"),
		Update1: prepareStmt(db, sqlUpdateWork1, "unable to prepare update 1"),
		Update2: prepareStmt(db, sqlUpdateWork2, "unable to prepare update 2"),
	}

	for wi := range mCh {
		_, err := onceRecord(sts, wi, wg)
		if err != nil {
			slog.Error("error occurs on recording", "error", err)
			continue
		}
	}
}

func onceRecord(sts WorkStmts, wi request.BookmarkWorkItem, wg *sync.WaitGroup) (*WorkModel, error) {
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
			slog.Error("unable to insert work to database", "error", err)
			return nil, err
		}
	} else if w.MaskedDate != -1 {
		if wi.IsMasked {
			return nil, nil
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
				slog.Error("unable to update work to database", "error", err)
				return nil, err
			}
		}
	} else {
		if wi.IsMasked {
			w.MaskedDate = time.Now().UTC().Unix()

			slog.Warn("unmasked -> masked", "id", w.Id, "title", w.Title)
			_, err = sts.Update2.Exec(w)
			if err != nil {
				slog.Error("unable to update work to database", "error", err)
				return nil, err
			}
		} else {
			ut, err := time.Parse("2006-01-02T15:04:05-07:00", wi.UpdateDate)
			if err != nil {
				ut = time.Now()
			}
			uts := ut.UTC().Unix()
			if err == nil && uts == w.UpdateDate {
				return nil, nil
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
				slog.Error("unable to update work to database", "error", err)
				return nil, err
			}
		}
	}

	return &w, nil
}

type WorkStmts struct {
	Query   *sqlx.NamedStmt
	Insert  *sqlx.NamedStmt
	Update1 *sqlx.NamedStmt
	Update2 *sqlx.NamedStmt
}

func StartRead(db *sqlx.DB, ch chan<- WorkModel, wg *sync.WaitGroup) {
	var works []WorkModel
	err := db.Select(&works, `select * from work`)
	if err != nil {
		slog.Error("unable to query manifest", "error", err)
		return
	}
	for _, work := range works {
		slog.Debug("get work model", "id", work.Id, "title", work.Title, "type", work.IllustType, "count", work.PageCount)
		wg.Add(1)
		ch <- work
	}

	//rows, err := db.Queryx(`select * from work`)
	//if err != nil {
	//	slog.Error("unable to query manifest", "error", err)
	//	return
	//}
	//defer func(rows *sqlx.Rows) {
	//	err := rows.Close()
	//	if err != nil {
	//		slog.Error("unable to close rows", "error", err)
	//	}
	//}(rows)
	//slog.Info("start walk database assets")
	//for rows.Next() {
	//	var w WorkModel
	//	err := rows.StructScan(&w)
	//	if err != nil {
	//		slog.Error("unable to scan", "error", err)
	//		continue
	//	}
	//	slog.Debug("get work model", "id", w.Id, "title", w.Title, "type", w.IllustType, "count", w.PageCount)
	//	wg.Add(1)
	//	ch <- w
	//}
}
