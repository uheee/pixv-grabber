package job

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/request"
)

func StartDownload(ch <-chan DownloadTask, wg *sync.WaitGroup) {
	host := viper.GetString("job.host")
	for task := range ch {
		onceDownload(task, host, wg)
	}
}

func onceDownload(task DownloadTask, host string, wg *sync.WaitGroup) {
	slog.Debug("downloading", "id", task.Id)
	if wg != nil {
		defer wg.Done()
	}
	taskUrl, err := url.Parse(task.Url)
	if err != nil {
		slog.Error("unable to parse download url", "error", err, "id", task.Id, "url", task.Url, "path", task.Path)
		return
	}
	items := strings.Split(taskUrl.Path, "/")
	filename := items[len(items)-1]
	ff := path.Join(task.Path, filename)
	if _, err = os.Stat(ff); !os.IsNotExist(err) {
		slog.Debug("download file already exists", "file", ff)
		return
	}
	var raw []byte
	maxRetry := viper.GetUint("download.max-retry")
	err = retry.Do(func() error {
		raw, err = request.GetRawFromHttpReq(task.Url, map[string]string{
			"User-Agent": "Mozilla/5.0",
			"Referer":    host,
		})
		return err
	},
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(3*time.Second),
		retry.MaxDelay(time.Minute),
		retry.Attempts(maxRetry),
		retry.OnRetry(func(n uint, err error) {
			slog.Error("failed to download", "retry", n, "error", err)
		}))
	if err != nil {
		slog.Error("unable to get raw from http req", "error", err, "id", task.Id, "url", task.Url, "path", task.Path)
		return
	}
	file, err := os.OpenFile(ff, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		slog.Error("unable to open local file", "error", err, "id", task.Id, "url", task.Url, "path", task.Path)
		return
	}
	_, err = file.Write(raw)
	if err != nil {
		slog.Error("unable to write local file", "error", err, "id", task.Id, "url", task.Url, "path", task.Path)
		return
	}
	slog.Log(context.Background(), -1, "download", "id", task.Id, "url", task.Url)
}
