package job

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/request"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
)

func StartDownload(ch <-chan DownloadTask, wg *sync.WaitGroup) {
	host := viper.GetString("job.host")
	for task := range ch {
		onceDownload(task, host, wg)
	}
}

func onceDownload(task DownloadTask, host string, wg *sync.WaitGroup) {
	log.Trace().Str("id", task.Id).Msg("downloading")
	if wg != nil {
		defer wg.Done()
	}
	taskUrl, err := url.Parse(task.Url)
	if err != nil {
		log.Error().Err(err).Msg("download task")
		return
	}
	items := strings.Split(taskUrl.Path, "/")
	filename := items[len(items)-1]
	raw, err := request.GetRawFromHttpReq(task.Url, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Referer":    host,
	})
	if err != nil {
		log.Error().Err(err).Msg("download task")
		return
	}
	file, err := os.OpenFile(path.Join(task.Path, filename), os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("download task")
		return
	}
	_, err = file.Write(raw)
	if err != nil {
		log.Error().Err(err).Msg("download task")
		return
	}
	log.Info().Str("id", task.Id).Str("url", task.Url).Msg("download")
}
