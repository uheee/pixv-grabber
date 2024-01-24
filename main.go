package main

import (
	"github.com/spf13/viper"
	"log"
)

func main() {
	initLog()
	err := initConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = getAll()
	if err != nil {
		log.Fatal(err)
	}
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.SetDefault("job.host", "https://www.pixiv.net")
	viper.SetDefault("job.lang", "zh")
	viper.SetDefault("job.limit", 100)
	viper.SetDefault("job.output", "output")

	return viper.ReadInConfig()
}
