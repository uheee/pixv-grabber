package job

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/database"
	"github.com/uheee/pixiv-grabber/request"
	"github.com/urfave/cli/v2"
)

func ProcessHttp(db *sqlx.DB, cc *cli.Context, mCh chan<- request.BookmarkWorkItem, dCh chan<- DownloadTask, wg *sync.WaitGroup) {
	offset := 0
	for {
		total, err := getBookmark(db, cc, mCh, dCh, &offset, wg)
		if err != nil {
			slog.Error("get bookmark", "error", err)
		}
		if offset >= total {
			break
		}
	}
}

func getBookmark(db *sqlx.DB, cc *cli.Context, mCh chan<- request.BookmarkWorkItem, dCh chan<- DownloadTask, offset *int, wg *sync.WaitGroup) (int, error) {
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

	checkMode := cc.Bool("check")
	downloadTaskWorkIds, err := database.GetAllDownloadTaskWorkIds(db)
	for _, work := range bookmark.Works {
		if checkMode && !slices.Contains(downloadTaskWorkIds, strconv.FormatUint(work.GetId(), 10)) {
			continue
		}
		slog.Log(context.Background(), -3, "start work", "id", work.Id, "title", work.Title, "pages", work.PageCount)
		go getBookmarkContent(cc, mCh, dCh, work, wg)
	}
	*offset += limit
	return bookmark.Total, nil
}

func getBookmarkContent(cc *cli.Context, mCh chan<- request.BookmarkWorkItem, dCh chan<- DownloadTask, work request.BookmarkWorkItem, wg *sync.WaitGroup) {
	output := viper.GetString("job.output")

	id := work.GetId()
	idStr := strconv.FormatUint(id, 10)
	if wg != nil {
		wg.Add(1)
	}
	mCh <- work

	ut, err := time.Parse("2006-01-02T15:04:05-07:00", work.UpdateDate)
	if err != nil {
		slog.Error("parse update time", "error", err)
		return
	}
	cp := path.Join(output, idStr, ut.UTC().Format("20060102150405"))
	if work.IsMasked {
		return
	}

	if _, err := os.Stat(cp); !cc.Bool("check") && !os.IsNotExist(err) {
		slog.Log(context.Background(), -7, "target is latest, skip", "id", id)
		return
	}

	err = os.MkdirAll(cp, os.ModePerm)
	if err != nil {
		slog.Error("create latest dir", "error", err)
	}

	switch work.IllustType {
	case 0:
		err := getImages(dCh, idStr, cp, wg)
		if err != nil {
			slog.Error("get images", "error", err)
		}
	case 1:
		err := getImages(dCh, idStr, cp, wg)
		if err != nil {
			slog.Error("get images", "error", err)
		}
	case 2:
		err := getVideos(dCh, idStr, cp, wg)
		if err != nil {
			slog.Error("get videos", "error", err)
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
