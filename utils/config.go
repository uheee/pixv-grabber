package utils

import (
	"errors"
	"github.com/lmittmann/tint"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"
)

func InitConfig() error {
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

func InitProxy() error {
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

func InitLogger() {
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
