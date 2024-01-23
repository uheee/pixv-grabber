package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strconv"
	"time"
)

const bookmark string = "https://pixiv.net/ajax/user/%s/illusts/bookmarks?tag=&offset=%s&limit=%s&rest=show&version=%s"

func getAll() error {
	ch := make(chan DownloadTask)

	go func() {
		offset := 0
		for {
			total, err := getBookmark(ch, &offset)
			if err != nil {
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

	hostUrl, err := url.Parse(host)
	if err != nil {
		return -1, err
	}

	hostUrl = hostUrl.JoinPath("ajax", "user", user, "illusts", "bookmarks")
	query := hostUrl.Query()
	query.Set("tag", "")
	query.Set("rest", "show")
	query.Set("version", version)
	query.Set("lang", lang)
	query.Set("offset", strconv.Itoa(*offset))
	query.Set("limit", strconv.Itoa(limit))
	hostUrl.RawQuery = query.Encode()
	reqUrl := hostUrl.String()
	client := http.Client{}
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return -1, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Cookie", cookie)
	res, err := client.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	var result Response[BookmarkBody]
	err = json.Unmarshal(body, &result)
	if err != nil {
		return -1, err
	}
	if result.Error {
		return -1, errors.New(result.Message)
	}
	for _, work := range result.Body.Works {
		go getBookmarkContent(ch, work)
	}
	*offset += limit
	return result.Body.Total, nil
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
	attachLog(work, id)
	if work.IsMasked {
		return
	}
	if work.IllustType == 0 {
		getImages(ch, id)
	} else if work.IllustType == 2 {
		getVideos(ch, id)
	}
}

func attachLog(work BookmarkWorkItem, id string) {
	output := viper.GetString("output")
	workPath := path.Join(output, id)
	err := os.MkdirAll(workPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(path.Join(workPath, "log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	_, err = logFile.WriteString(fmt.Sprintf(`
################ %s ################
Title: %s
IsMasked: %v
#####################################################
`, time.Now().Format("2006-01-02 15:04:05"), work.Title, work.IsMasked))
	if err != nil {
		panic(err)
	}
}

func getImages(ch chan<- DownloadTask, id string) {
	host := viper.GetString("host")
	version := viper.GetString("version")
	cookie := viper.GetString("cookie")
	lang := viper.GetString("lang")

	hostUrl, err := url.Parse(host)
	if err != nil {
		panic(err)
	}

	hostUrl = hostUrl.JoinPath("ajax", "illust", id, "pages")
	query := hostUrl.Query()
	query.Set("version", version)
	query.Set("lang", lang)
	hostUrl.RawQuery = query.Encode()
	reqUrl := hostUrl.String()
	client := http.Client{}
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Cookie", cookie)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var result Response[[]ImageItem]
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(err)
	}
	if result.Error {
		println(result.Message)
	}
	for _, item := range result.Body {
		ch <- DownloadTask{
			Id:  id,
			Url: item.Urls.Original,
		}
	}
}

func getVideos(ch chan<- DownloadTask, id string) {
	host := viper.GetString("host")
	version := viper.GetString("version")
	cookie := viper.GetString("cookie")
	lang := viper.GetString("lang")

	hostUrl, err := url.Parse(host)
	if err != nil {
		panic(err)
	}

	hostUrl = hostUrl.JoinPath("ajax", "illust", id, "ugoira_meta")
	query := hostUrl.Query()
	query.Set("version", version)
	query.Set("lang", lang)
	hostUrl.RawQuery = query.Encode()
	reqUrl := hostUrl.String()
	client := http.Client{}
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Cookie", cookie)
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var result Response[VideoItem]
	err = json.Unmarshal(body, &result)
	if err != nil {
		panic(err)
	}
	if result.Error {
		panic(err)
	}
	ch <- DownloadTask{
		Id:  id,
		Url: result.Body.OriginalSrc,
	}
}
