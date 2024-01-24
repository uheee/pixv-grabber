package main

import (
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"time"
)

func main() {
	err := initConfig()
	if err != nil {
		panic(err)
	}
	initLog()

	c := cron.New()
	ce := viper.GetString("job.cron")
	_, err = c.AddFunc(ce, func() {
		ct := time.Now()
		log.Info().Time("current", ct).Msg("start cron job")
		err = getAll()
		if err != nil {
			log.Error().Err(err).Msg("get all")
		}
		log.Info().Time("current", ct).Msg("finish cron job")
	})
	if err != nil {
		panic(err)
	}
	c.Start()
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
