package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"time"
)

func getAll() error {
	ch := make(chan DownloadTask)

	go func() {
		offset := 0
		for {
			total, err := getBookmark(ch, &offset)
			if err != nil {
				log.Error().Err(err).Msg("get bookmark")
				return
			}
			if offset >= total {
				break
			}
		}
	}()

	download(ch)
	return nil
}

func getBookmark(ch chan<- DownloadTask, offset *int) (int, error) {
	host := viper.GetString("host")
	user := viper.GetString("user")
	version := viper.GetString("version")
	cookie := viper.GetString("cookie")
	lang := viper.GetString("lang")
	limit := viper.GetInt("limit")

	u, err := url.Parse(host)
	if err != nil {
		return -1, err
	}

	u = u.JoinPath("ajax", "user", user, "illusts", "bookmarks")
	query := u.Query()
	query.Set("tag", "")
	query.Set("rest", "show")
	query.Set("version", version)
	query.Set("lang", lang)
	query.Set("offset", strconv.Itoa(*offset))
	query.Set("limit", strconv.Itoa(limit))
	u.RawQuery = query.Encode()
	reqUrl := u.String()
	bookmark, err := getJsonFromHttpReq[BookmarkBody](reqUrl, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Cookie":     cookie,
	})
	if err != nil {
		return -1, err
	}

	for _, work := range bookmark.Works {
		go getBookmarkContent(ch, work)
	}
	*offset += limit
	return bookmark.Total, nil
}

func getBookmarkContent(ch chan<- DownloadTask, work BookmarkWorkItem) {
	var id string
	switch reflect.TypeOf(work.Id).Kind() {
	case reflect.String:
		id = work.Id.(string)
	case reflect.Float64:
		id = strconv.FormatFloat(work.Id.(float64), 'f', -1, 64)
	default:
		panic("unhandled default case")
	}
	err := attachLog(work, id)
	if err != nil {
		log.Error().Err(err).Msg("attach log")
	}
	if work.IsMasked {
		return
	}

	if work.IllustType == 0 {
		err := getImages(ch, id)
		if err != nil {
			log.Error().Err(err).Msg("get images")
		}
	} else if work.IllustType == 2 {
		err := getVideos(ch, id)
		if err != nil {
			log.Error().Err(err).Msg("get videos")
		}
	}

}

func attachLog(work BookmarkWorkItem, id string) error {
	output := viper.GetString("output")
	workPath := path.Join(output, id)
	err := os.MkdirAll(workPath, os.ModePerm)
	if err != nil {
		return err
	}
	logFile, err := os.OpenFile(path.Join(workPath, "log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	_, err = logFile.WriteString(fmt.Sprintf(`
################ %s ################
Title: %s
IsMasked: %v
#####################################################
`, time.Now().Format("2006-01-02 15:04:05"), work.Title, work.IsMasked))
	if err != nil {
		return err
	}
	return nil
}

func getImages(ch chan<- DownloadTask, id string) error {
	host := viper.GetString("host")
	version := viper.GetString("version")
	cookie := viper.GetString("cookie")
	lang := viper.GetString("lang")

	u, err := url.Parse(host)
	if err != nil {
		return err
	}

	u = u.JoinPath("ajax", "illust", id, "pages")
	query := u.Query()
	query.Set("version", version)
	query.Set("lang", lang)
	u.RawQuery = query.Encode()
	reqUrl := u.String()
	items, err := getJsonFromHttpReq[[]ImageItem](reqUrl, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Cookie":     cookie,
	})
	if err != nil {
		return err
	}

	for _, item := range *items {
		ch <- DownloadTask{
			Id:  id,
			Url: item.Urls.Original,
		}
	}
	return nil
}

func getVideos(ch chan<- DownloadTask, id string) error {
	host := viper.GetString("host")
	version := viper.GetString("version")
	cookie := viper.GetString("cookie")
	lang := viper.GetString("lang")

	u, err := url.Parse(host)
	if err != nil {
		return err
	}

	u = u.JoinPath("ajax", "illust", id, "ugoira_meta")
	query := u.Query()
	query.Set("version", version)
	query.Set("lang", lang)
	u.RawQuery = query.Encode()
	reqUrl := u.String()
	item, err := getJsonFromHttpReq[VideoItem](reqUrl, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Cookie":     cookie,
	})
	if err != nil {
		return err
	}

	ch <- DownloadTask{
		Id:  id,
		Url: item.OriginalSrc,
	}
	return nil
}
