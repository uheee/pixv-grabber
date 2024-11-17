package database

import (
	"github.com/jmoiron/sqlx"
	"log/slog"
	"sync"
)

var sqlQueryDownloadTask = `
select work_id from download_task where status = 0`

var sqlInsertDownloadTask = `
insert into download_task
(id, work_id)
values
(:id, :work_id)`

var sqlUpdateDownloadTask = `
update download_task
set status = :status,
    update_date = current_timestamp
where id = :id`

var sqlInsertDownloadTaskItem = `
insert into download_task_item
(id, task_id, remote, local)
values
(:id, :task_id, :remote, :local)`

func GetAllDownloadTaskWorkIds(db *sqlx.DB) ([]string, error) {
	var workIds []string
	err := db.Select(&workIds, sqlQueryDownloadTask)
	if err != nil {
		slog.Error("unable to query download task", "error", err)
		return nil, err
	}
	return workIds, nil
}

func StartRecordDownloadTask(db *sqlx.DB, ch <-chan DownloadTask, wg *sync.WaitGroup) {
	st := prepareStmt(db, sqlInsertDownloadTask, "unable to prepare insert download task")
	for task := range ch {
		onceRecordDownloadTask(st, task, wg)
	}
}

func onceRecordDownloadTask(st *sqlx.NamedStmt, task DownloadTask, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := st.Exec(task)
	if err != nil {
		slog.Error("unable to insert download task to db", "error", err)
		return
	}
}
