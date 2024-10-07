package main

import (
	"context"
	"errors"
	"github.com/lmittmann/tint"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/job"
	"github.com/uheee/pixiv-grabber/internal/manifest"
	"github.com/uheee/pixiv-grabber/internal/request"
	"github.com/uheee/pixiv-grabber/internal/utils"
	"github.com/urfave/cli/v2"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

func main() {
	app := &cli.App{
		Name:  "PIXIV Grabber",
		Usage: "Fetch and download PIXIV content",
		Commands: []*cli.Command{
			{
				Name:   "sync",
				Usage:  "synchronize content from PIXIV website",
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
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Log(context.Background(), -10, "fatal error", "error", err)
	}
}

func syncAction(context *cli.Context) error {
	err := utils.InitConfig()
	if err != nil {
		return err
	}
	configLog()
	err = configProxy()
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

func configProxy() error {
	hp := viper.GetString("proxy.http")
	if hp != "" {
		pu, err := url.Parse(hp)
		if err != nil {
			return errors.Join(errors.New("unable to use proxy"), err)
		} else {
			http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(pu)}
		}
	}
	return nil
}

func configLog() {
	level := slog.Level(viper.GetInt("log.level"))
	w := os.Stdout
	logger := slog.New(tint.NewHandler(w, &tint.Options{
		Level:      level,
		TimeFormat: time.RFC3339,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if err, ok := attr.Value.Any().(error); ok {
				err := tint.Err(err)
				err.Key = attr.Key
				return err
			}
			return attr
		},
	}))
	slog.SetDefault(logger)
}
