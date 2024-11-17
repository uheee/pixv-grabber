package subcommand

import (
	"context"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/database"
	"github.com/uheee/pixiv-grabber/utils"
	"github.com/urfave/cli/v2"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

func CheckCommand() *cli.Command {
	return &cli.Command{
		Name:   "check",
		Usage:  "check local asset integrity",
		Action: checkAction,
	}
}

func checkAction(cc *cli.Context) error {
	err := utils.InitConfig()
	if err != nil {
		return err
	}
	utils.InitLogger()

	db, err := database.Instance()
	if err != nil {
		slog.Error("unable to init database")
		return err
	}

	mwts := viper.GetString("job.max-wait-time")
	mwt, err := time.ParseDuration(mwts)
	if err != nil {
		mwt = 5 * time.Second
	}
	wCh := make(chan database.WorkModel)
	tCh := make(chan database.DownloadTask)
	wg := &sync.WaitGroup{}
	go database.StartRecordDownloadTask(db, tCh, wg)
	go database.StartRead(db, wCh, wg)
	go startCheck(wCh, tCh, wg)
	time.Sleep(mwt)
	wg.Wait()
	slog.Warn("done")
	return nil
}

func startCheck(wCh <-chan database.WorkModel, tCh chan<- database.DownloadTask, wg *sync.WaitGroup) {
	output := viper.GetString("job.output")
	for w := range wCh {
		onceCheck(tCh, w, output, wg)
	}
}

func onceCheck(tCh chan<- database.DownloadTask, w database.WorkModel, output string, wg *sync.WaitGroup) {
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
		wg.Add(1)
		tCh <- database.DownloadTask{Id: uuid.New(), WorkId: w.Id}
		slog.Warn("count mismatch", "id", w.Id, "title", w.Title, "type", w.IllustType, "database-count", ec, "dir-count", ac)
	}
}
