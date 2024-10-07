package subcommand

import (
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/manifest"
	"github.com/uheee/pixiv-grabber/utils"
	"github.com/urfave/cli/v2"
	"log/slog"
	"sync"
	"time"
)

func CheckCommand() *cli.Command {
	return &cli.Command{
		Name:   "check",
		Usage:  "check local asset integrity",
		Action: checkAction,
	}
}

func checkAction(c *cli.Context) error {
	err := utils.InitConfig()
	if err != nil {
		return err
	}
	utils.InitLogger()

	mwts := viper.GetString("job.max-wait-time")
	mwt, err := time.ParseDuration(mwts)
	if err != nil {
		mwt = 5 * time.Second
	}
	ch := make(chan manifest.WorkModel)
	wg := &sync.WaitGroup{}
	go manifest.StartRead(ch, wg)
	go manifest.StartCheck(ch, wg)
	time.Sleep(mwt)
	wg.Wait()
	slog.Warn("done")
	return nil
}
