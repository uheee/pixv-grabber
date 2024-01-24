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
	viper.SetDefault("host", "https://www.pixiv.net")
	viper.SetDefault("lang", "zh")
	viper.SetDefault("limit", 100)
	viper.SetDefault("output", "output")

	return viper.ReadInConfig()
}
