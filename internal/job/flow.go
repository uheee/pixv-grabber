package job

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/request"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"sync"
	"time"
)

func ProcessHttp(mCh chan<- request.BookmarkWorkItem, dCh chan<- DownloadTask, wg *sync.WaitGroup) {
	offset := 0
	for {
		total, err := getBookmark(mCh, dCh, &offset, wg)
		if err != nil {
			log.Error().Err(err).Msg("get bookmark")
		}
		if offset >= total {
			break
		}
	}
}

func getBookmark(mCh chan<- request.BookmarkWorkItem, dCh chan<- DownloadTask, offset *int, wg *sync.WaitGroup) (int, error) {
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
	bookmark, err := request.GetJsonFromHttpReq[request.BookmarkBody](reqUrl, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Cookie":     cookie,
	})
	if err != nil {
		return -1, err
	}

	for _, work := range bookmark.Works {
		log.Trace().Any("id", work.Id).Str("title", work.Title).Int("pages", work.PageCount).Msg("start work")
		go getBookmarkContent(mCh, dCh, work, wg)
	}
	*offset += limit
	return bookmark.Total, nil
}

func getBookmarkContent(mCh chan<- request.BookmarkWorkItem, ch chan<- DownloadTask, work request.BookmarkWorkItem, wg *sync.WaitGroup) {
	idRange := viper.GetStringSlice("patch.id-range")
	output := viper.GetString("job.output")

	id := work.GetId()
	idStr := strconv.FormatUint(id, 10)
	if idRange != nil && !slices.Contains(idRange, idStr) {
		return
	}
	if wg != nil {
		wg.Add(1)
	}
	mCh <- work

	ut, err := time.Parse("2006-01-02T15:04:05-07:00", work.UpdateDate)
	if err != nil {
		log.Error().Err(err).Msg("parse update time")
		return
	}
	cp := path.Join(output, idStr, ut.UTC().Format("20060102150405"))
	if work.IsMasked {
		return
	}

	if _, err := os.Stat(cp); !os.IsNotExist(err) {
		if idRange == nil || !slices.Contains(idRange, idStr) {
			log.Trace().Uint64("id", id).Msg("target is latest, skip")
			return
		}
	}

	err = os.MkdirAll(cp, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Msg("create latest dir")
	}

	switch work.IllustType {
	case 0:
		err := getImages(ch, idStr, cp, wg)
		if err != nil {
			log.Error().Err(err).Msg("get images")
		}
	case 1:
		err := getImages(ch, idStr, cp, wg)
		if err != nil {
			log.Error().Err(err).Msg("get images")
		}
	case 2:
		err := getVideos(ch, idStr, cp, wg)
		if err != nil {
			log.Error().Err(err).Msg("get videos")
		}
	}
}

func getImages(ch chan<- DownloadTask, id string, cp string, wg *sync.WaitGroup) error {
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
	items, err := request.GetJsonFromHttpReq[[]request.ImageItem](reqUrl, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Cookie":     cookie,
	})
	if err != nil {
		return err
	}

	if wg != nil {
		wg.Add(len(*items))
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

func getVideos(ch chan<- DownloadTask, id string, cp string, wg *sync.WaitGroup) error {
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
	item, err := request.GetJsonFromHttpReq[request.VideoItem](reqUrl, map[string]string{
		"User-Agent": "Mozilla/5.0",
		"Cookie":     cookie,
	})
	if err != nil {
		return err
	}

	if wg != nil {
		wg.Add(1)
	}
	ch <- DownloadTask{
		Id:   id,
		Url:  item.OriginalSrc,
		Path: cp,
	}
	return nil
}

type DownloadTask struct {
	Id   string
	Url  string
	Path string
}
