package database

import "github.com/google/uuid"

type WorkModel struct {
	Id         uint64 `db:"id"`
	Title      string `db:"title"`
	IllustType int    `db:"illust_type"`
	Tags       []byte `db:"tags"`
	PageCount  int    `db:"page_count"`
	CreateDate int64  `db:"create_date"`
	UpdateDate int64  `db:"update_date"`
	MaskedDate int64  `db:"masked_date"`
}

type DownloadTask struct {
	Id         uuid.UUID `db:"id"`
	WorkId     uint64    `db:"work_id"`
	Status     int       `db:"status"`
	CreateDate int64     `db:"create_date"`
	UpdateDate int64     `db:"update_date"`
}

type DownloadTaskItem struct {
	Id         uuid.UUID `db:"id"`
	TaskId     uuid.UUID `db:"task_id"`
	Remote     string    `db:"remote"`
	Local      string    `db:"local"`
	Status     int       `db:"status"`
	ErrMsg     string    `db:"err_msg"`
	CreateDate int64     `db:"create_date"`
	UpdateDate int64     `db:"update_date"`
}

const (
	TASK_WAITING     = 0
	TASK_FETCHING    = 1
	TASK_DOWNLOADING = 2
	TASK_COMPLETED   = 3
)
