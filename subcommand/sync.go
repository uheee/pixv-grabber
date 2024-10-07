package subcommand

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/job"
	"github.com/uheee/pixiv-grabber/manifest"
	"github.com/uheee/pixiv-grabber/request"
	"github.com/uheee/pixiv-grabber/utils"
	"github.com/urfave/cli/v2"
	"log/slog"
	"sync"
	"time"
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
		},
	}
}

func syncAction(context *cli.Context) error {
	err := utils.InitConfig()
	if err != nil {
		return err
	}
	utils.InitLogger()
	err = utils.InitProxy()
	if err != nil {
		return err
	}

	mCh := make(chan request.BookmarkWorkItem)
	dCh := make(chan job.DownloadTask)

	detach := context.Bool("detach")
	if !detach {
		slog.Warn("once mode")
		return onceTask(mCh, dCh)
	} else {
		slog.Warn("cron mode")
		return cronTask(mCh, dCh)
	}
}

func onceTask(mCh chan request.BookmarkWorkItem, dCh chan job.DownloadTask) error {
	mwts := viper.GetString("job.max-wait-time")
	mwt, err := time.ParseDuration(mwts)
	if err != nil {
		mwt = 5 * time.Second
	}
	wg := &sync.WaitGroup{}
	go job.ProcessHttp(mCh, dCh, wg)
	go manifest.StartRecord(mCh, wg)
	go job.StartDownload(dCh, wg)
	time.Sleep(mwt)
	wg.Wait()
	slog.Warn("done")
	return nil
}

func cronTask(mCh chan request.BookmarkWorkItem, dCh chan job.DownloadTask) error {
	c := cron.New()
	ce := viper.GetString("job.cron")
	_, err := c.AddFunc(ce, func() {
		ct := time.Now()
		slog.Info("start cron job", "current", ct)
		job.ProcessHttp(mCh, dCh, nil)
		slog.Info("finish cron job", "current", ct)
	})
	if err != nil {
		return err
	}
	c.Start()
	go manifest.StartRecord(mCh, nil)
	go job.StartDownload(dCh, nil)
	select {}
}
