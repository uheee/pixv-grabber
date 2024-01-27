package main

import (
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/job"
	"github.com/uheee/pixiv-grabber/internal/logger"
	"github.com/uheee/pixiv-grabber/internal/manifest"
	"github.com/uheee/pixiv-grabber/internal/request"
	"net/http"
	"net/url"
	"time"
)

func main() {
	err := initConfig()
	if err != nil {
		panic(err)
	}
	logger.InitLog()

	hp := viper.GetString("proxy.http")
	if hp != "" {
		pu, err := url.Parse(hp)
		if err != nil {
			log.Error().Err(err).Str("proxy", hp).Msg("unable to use proxy")
		} else {
			http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(pu)}
		}
	}

	mCh := make(chan request.BookmarkWorkItem)
	dCh := make(chan job.DownloadTask)
	c := cron.New()
	ce := viper.GetString("job.cron")
	_, err = c.AddFunc(ce, func() {
		ct := time.Now()
		log.Info().Time("current", ct).Msg("start cron job")
		job.ProcessHttp(mCh, dCh)
		log.Info().Time("current", ct).Msg("finish cron job")
	})
	if err != nil {
		panic(err)
	}
	c.Start()
	go manifest.StartRecord(mCh)
	go job.StartDownload(dCh)
	select {}
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("job.host", "https://www.pixiv.net")
	viper.SetDefault("job.lang", "zh")
	viper.SetDefault("job.limit", 100)
	viper.SetDefault("job.output", "output")

	return viper.ReadInConfig()
}
