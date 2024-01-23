package main

import (
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

func download(ch <-chan DownloadTask) {
	host := viper.GetString("host")
	output := viper.GetString("output")
	for task := range ch {
		u, err := url.Parse(task.Url)
		if err != nil {
			continue
		}
		items := strings.Split(u.Path, "/")
		filename := items[len(items)-1]
		client := http.Client{}
		req, err := http.NewRequest("GET", task.Url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Referer", host)
		res, err := client.Do(req)
		if err != nil {
			continue
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		file, err := os.OpenFile(path.Join(output, task.Id, filename), os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			continue
		}
		_, err = file.Write(body)
		if err != nil {
			continue
		}
	}
}
