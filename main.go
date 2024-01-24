package main

import (
	"github.com/spf13/viper"
	"log"
)

func main() {
	err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	initLog()
	err = getAll()
	if err != nil {
		log.Fatal(err)
	}
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
