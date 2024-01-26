package job

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/request"
	"net/url"
	"os"
	"path"
	"strings"
)

func Download(ch <-chan DownloadTask) {
	host := viper.GetString("job.host")
	for task := range ch {
		log.Info().Str("id", task.Id).Msg("start download")
		taskUrl, err := url.Parse(task.Url)
		if err != nil {
			log.Error().Err(err).Msg("download task")
			continue
		}
		items := strings.Split(taskUrl.Path, "/")
		filename := items[len(items)-1]
		raw, err := request.GetRawFromHttpReq(task.Url, map[string]string{
			"User-Agent": "Mozilla/5.0",
			"Referer":    host,
		})
		if err != nil {
			log.Error().Err(err).Msg("download task")
			continue
		}
		file, err := os.OpenFile(path.Join(task.Path, filename), os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Error().Err(err).Msg("download task")
			continue
		}
		_, err = file.Write(raw)
		if err != nil {
			log.Error().Err(err).Msg("download task")
			continue
		}
	}
}
