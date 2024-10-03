package manifest

import (
	"encoding/json"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/request"
	"os"
	"path/filepath"
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

var queryPreparation = `
select * from work where id = :id`

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
		log.Error().Err(err).Msg("create output dir")
	}
	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		log.Error().Err(err).Str("db", dbFile).Msg("unable to connect to db")
		return
	}
	db.MustExec(workSchema)
	sts := SqlxStmts{
		Query:   prepareStmts(db, queryPreparation, "unable to prepare query"),
		Insert:  prepareStmts(db, insertPreparation, "unable to prepare insert"),
		Update1: prepareStmts(db, updatePreparation1, "unable to prepare update 1"),
		Update2: prepareStmts(db, updatePreparation2, "unable to prepare update 2"),
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

		log.Warn().Uint64("id", w.Id).Str("title", w.Title).Msg("new work")
		_, err = sts.Insert.Exec(w)
		if err != nil {
			log.Error().Err(err).Msg("unable to insert work to db")
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

			log.Warn().Uint64("id", w.Id).Str("title", w.Title).Msg("masked -> unmasked")
			_, err = sts.Update1.Exec(w)
			if err != nil {
				log.Error().Err(err).Msg("unable to update work to db")
			}
		}
	} else {
		if wi.IsMasked {
			w.MaskedDate = time.Now().UTC().Unix()

			log.Warn().Uint64("id", w.Id).Str("title", w.Title).Msg("unmasked -> masked")
			_, err = sts.Update2.Exec(w)
			if err != nil {
				log.Error().Err(err).Msg("unable to update work to db")
			}
		} else {
			ut, err := time.Parse("2006-01-02T15:04:05-07:00", wi.UpdateDate)
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

			log.Warn().Uint64("id", w.Id).Str("title", w.Title).Msg("update")
			_, err = sts.Update1.Exec(w)
			if err != nil {
				log.Error().Err(err).Msg("unable to update work to db")
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

func prepareStmts(db *sqlx.DB, sql string, errMsg string) *sqlx.NamedStmt {
	stmt, err := db.PrepareNamed(sql)
	if err != nil {
		log.Error().Err(err).Msg(errMsg)
		panic(err)
	}
	return stmt
}
