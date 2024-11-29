package subcommand

import (
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/database"
	"github.com/uheee/pixiv-grabber/job"
	"github.com/uheee/pixiv-grabber/request"
	"github.com/uheee/pixiv-grabber/utils"
	"github.com/urfave/cli/v2"
)

func SyncCommand() *cli.Command {
	return &cli.Command{
		Name:   "sync",
		Usage:  "synchronize assets from PIXIV website",
		Action: syncAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "detach",
				Usage:    "running synchronizer as a background service",
				Required: false,
				Value:    false,
				Aliases:  []string{"d"},
			},
			&cli.BoolFlag{
				Name:     "check",
				Usage:    "check local asset integrity and download all missing files",
				Required: false,
				Value:    false,
				Aliases:  []string{"c"},
			},
		},
	}
}

func syncAction(cc *cli.Context) error {
	err := utils.InitConfig()
	if err != nil {
		return err
	}
	utils.InitLogger()
	err = utils.InitProxy()
	if err != nil {
		return err
	}

	db, err := database.Instance()
	if err != nil {
		slog.Error("unable to init database")
		return err
	}

	mCh := make(chan request.BookmarkWorkItem)
	dCh := make(chan job.DownloadTask)

	detach := cc.Bool("detach")
	if !detach {
		slog.Warn("once mode")
		return onceTask(db, cc, mCh, dCh)
	} else {
		slog.Warn("cron mode")
		return cronTask(db, cc, mCh, dCh)
	}
}

func onceTask(db *sqlx.DB, cc *cli.Context, mCh chan request.BookmarkWorkItem, dCh chan job.DownloadTask) error {
	mwts := viper.GetString("job.max-wait-time")
	mwt, err := time.ParseDuration(mwts)
	if err != nil {
		mwt = 5 * time.Second
	}
	wg := &sync.WaitGroup{}
	go job.ProcessHttp(db, cc, mCh, dCh, wg)
	go database.StartRecord(db, mCh, wg)
	go job.StartDownload(dCh, wg)
	time.Sleep(mwt)
	wg.Wait()
	slog.Warn("done")
	return nil
}

func cronTask(db *sqlx.DB, cc *cli.Context, mCh chan request.BookmarkWorkItem, dCh chan job.DownloadTask) error {
	c := cron.New()
	ce := viper.GetString("job.cron")
	_, err := c.AddFunc(ce, func() {
		ct := time.Now()
		slog.Info("start cron job", "current", ct)
		job.ProcessHttp(db, cc, mCh, dCh, nil)
		slog.Info("finish cron job", "current", ct)
	})
	if err != nil {
		return err
	}
	c.Start()
	go database.StartRecord(db, mCh, nil)
	go job.StartDownload(dCh, nil)
	select {}
}
