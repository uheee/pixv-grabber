package job

import (
	"context"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/request"
	"log/slog"
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
	slog.Debug("downloading", "id", task.Id)
	if wg != nil {
		defer wg.Done()
	}
	taskUrl, err := url.Parse(task.Url)
	if err != nil {
		slog.Error("download task", "error", err)
		return
	}
	items := strings.Split(taskUrl.Path, "/")
	filename := items[len(items)-1]
	raw, err := request.GetRawFromHttpReq(task.Url, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Referer":    host,
	})
	if err != nil {
		slog.Error("download task", "error", err)
		return
	}
	file, err := os.OpenFile(path.Join(task.Path, filename), os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		slog.Error("download task", "error", err)
		return
	}
	_, err = file.Write(raw)
	if err != nil {
		slog.Error("download task", "error", err)
		return
	}
	slog.Log(context.Background(), -1, "download", "id", task.Id, "url", task.Url)
}
