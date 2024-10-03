package main

import (
	"errors"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/job"
	"github.com/uheee/pixiv-grabber/internal/logger"
	"github.com/uheee/pixiv-grabber/internal/manifest"
	"github.com/uheee/pixiv-grabber/internal/request"
	"github.com/uheee/pixiv-grabber/internal/utils"
	"github.com/urfave/cli/v2"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

func main() {
	app := &cli.App{
		Name:   "PIXIV Grabber",
		Usage:  "Fetch and download PIXIV content",
		Action: cliAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "detach",
				Usage:    "Run pixiv-grabber as a background service",
				Required: false,
				Value:    false,
				Aliases:  []string{"d"},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err)
	}
}

func cliAction(context *cli.Context) error {
	err := utils.InitConfig()
	if err != nil {
		return err
	}
	logger.InitLog()
	err = configProxy()
	if err != nil {
		return err
	}

	mCh := make(chan request.BookmarkWorkItem)
	dCh := make(chan job.DownloadTask)

	detach := context.Bool("detach")
	if !detach {
		log.Warn().Msg("once mode")
		return onceTask(mCh, dCh)
	} else {
		log.Warn().Msg("cron mode")
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
	return nil
}

func cronTask(mCh chan request.BookmarkWorkItem, dCh chan job.DownloadTask) error {
	c := cron.New()
	ce := viper.GetString("job.cron")
	_, err := c.AddFunc(ce, func() {
		ct := time.Now()
		log.Info().Time("current", ct).Msg("start cron job")
		job.ProcessHttp(mCh, dCh, nil)
		log.Info().Time("current", ct).Msg("finish cron job")
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
