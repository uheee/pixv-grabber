package database

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
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

create table if not exists download_task
(
    id			BLOB		not null	constraint download_task_pk				primary key,
    work_id		integer		not null	constraint download_task_work_id_fk		references work,
    status		integer		default 0	not null,
    create_date	integer		default		current_timestamp not null,
    update_date	integer
);

create index if not exists download_task_work_id_index on download_task (work_id);

create table if not exists download_task_item
(
    id			BLOB		not null	constraint download_task_item_pk				primary key,
    task_id		BLOB		not null	constraint download_task_item_task_id_fk		references download_task,
    remote		text		not null,
    local		text		not null,
    status		integer		default 0	not null,
    err_msg		text,
    create_date	integer		default		current_timestamp not null,
    update_date	integer
);

create index if not exists download_task_item_task_id_index on download_task_item (task_id);
`

func Instance() (*sqlx.DB, error) {
	output := viper.GetString("job.output")
	dbFile := filepath.Join(output, "manifest")
	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		slog.Error("create output dir", "error", err)
	}
	db, err := sqlx.Open("sqlite3", dbFile)
	if err != nil {
		slog.Error("unable to connect to database", "error", err, "database", dbFile)
		return nil, err
	}
	db.MustExec(workSchema)
	return db, nil
}

func prepareStmt(db *sqlx.DB, sql string, errMsg string) *sqlx.NamedStmt {
	stmt, err := db.PrepareNamed(sql)
	if err != nil {
		slog.Error(errMsg, "error", err)
		panic(err)
	}
	return stmt
}
