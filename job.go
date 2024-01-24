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

func getAll(ch chan<- DownloadTask) {
	offset := 0
	for {
		total, err := getBookmark(ch, &offset)
		if err != nil {
			log.Error().Err(err).Msg("get bookmark")
		}
		if offset >= total {
			break
		}
	}
}

func getBookmark(ch chan<- DownloadTask, offset *int) (int, error) {
	host := viper.GetString("job.host")
	user := viper.GetString("job.user")
	version := viper.GetString("job.version")
	cookie := viper.GetString("job.cookie")
	lang := viper.GetString("job.lang")
	limit := viper.GetInt("job.limit")

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
		log.Info().Str("title", work.Title).Int("pages", work.PageCount).Msg("start work")
		go getBookmarkContent(ch, work)
	}
	*offset += limit
	return bookmark.Total, nil
}

func getBookmarkContent(ch chan<- DownloadTask, work BookmarkWorkItem) {
	output := viper.GetString("job.output")
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
		return
	}
	ut, err := time.Parse("2006-01-02T15:04:05-07:00", work.UpdateDate)
	if err != nil {
		log.Error().Err(err).Msg("parse update time")
		return
	}
	cp := path.Join(output, id, ut.UTC().Format("20060102150405"))
	if work.IsMasked {
		mfp := path.Join(output, id, "MASKED")
		if _, err := os.Stat(mfp); os.IsNotExist(err) {
			log.Warn().Str("id", id).Str("title", work.Title).Msg("new masked")
			mf, err := os.OpenFile(mfp, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				log.Error().Err(err).Msg("add masked file")
				return
			}
			defer mf.Close()
			_, err = mf.WriteString(time.Now().Format("2006-01-02 15:04:05"))
			if err != nil {
				log.Error().Err(err).Msg("write masked time")
				return
			}
		}
		return
	}

	if _, err := os.Stat(cp); !os.IsNotExist(err) {
		log.Debug().Str("id", id).Msg("target is latest, skip")
		return
	}
	err = os.MkdirAll(cp, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("create latest dir")
	}

	if work.IllustType == 0 {
		err := getImages(ch, id, cp)
		if err != nil {
			log.Error().Err(err).Msg("get images")
		}
	} else if work.IllustType == 2 {
		err := getVideos(ch, id, cp)
		if err != nil {
			log.Error().Err(err).Msg("get videos")
		}
	}
}

func attachLog(work BookmarkWorkItem, id string) error {
	output := viper.GetString("job.output")
	workPath := path.Join(output, id)
	err := os.MkdirAll(workPath, os.ModePerm)
	if err != nil {
		return err
	}
	lf, err := os.OpenFile(path.Join(workPath, "log"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer lf.Close()
	_, err = lf.WriteString(fmt.Sprintf(`
################ %s ################
Title: %s
IsMasked: %v
UpdateTime: %s
#####################################################
`, time.Now().Format("2006-01-02 15:04:05"), work.Title, work.IsMasked, work.UpdateDate))
	if err != nil {
		return err
	}
	return nil
}

func getImages(ch chan<- DownloadTask, id string, cp string) error {
	host := viper.GetString("job.host")
	version := viper.GetString("job.version")
	cookie := viper.GetString("job.cookie")
	lang := viper.GetString("job.lang")

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
			Id:   id,
			Url:  item.Urls.Original,
			Path: cp,
		}
	}
	return nil
}

func getVideos(ch chan<- DownloadTask, id string, cp string) error {
	host := viper.GetString("job.host")
	version := viper.GetString("job.version")
	cookie := viper.GetString("job.cookie")
	lang := viper.GetString("job.lang")

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
		Id:   id,
		Url:  item.OriginalSrc,
		Path: cp,
	}
	return nil
}
