package main

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/uheee/pixiv-grabber/internal/logger"
	"github.com/uheee/pixiv-grabber/internal/manifest"
	"github.com/uheee/pixiv-grabber/internal/request"
	"github.com/uheee/pixiv-grabber/internal/utils"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func main() {
	err := utils.InitConfig()
	if err != nil {
		panic(err)
	}
	logger.InitLog()

	output := viper.GetString("job.output")
	mCh := make(chan request.BookmarkWorkItem)
	err = filepath.WalkDir(output, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && strings.Count(path, string(os.PathSeparator)) > 1 {
			return fs.SkipDir
		}
		if !d.IsDir() && d.Name() == "log" {
			go processLog(mCh, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal().Err(err).Str("path", output).Msg("unable to walk dir")
	}
	manifest.StartRecord(mCh, nil)
}

var headRegex = regexp.MustCompile(`################\s([^#]+)\s################`)

const tail string = "#####################################################"

func processLog(mCh chan request.BookmarkWorkItem, path string) {
	id := strings.Split(path, "/")[1]

	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Error().Err(err).Str("path", path).Msg("unable to process log file")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var histories []map[string]string
	var record map[string]string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if record == nil {
			if headRegex.MatchString(line) {
				record = make(map[string]string)
			}
			continue
		}
		if line == tail {
			histories = append(histories, record)
			record = nil
			continue
		}
		kv := strings.SplitN(line, ":", 2)
		record[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	if err := scanner.Err(); err != nil {
		log.Error().Err(err).Str("path", path).Msg("unable to scan log file")
	}

	slices.Reverse(histories)
	p := false
	for _, history := range histories {
		if history["IsMasked"] == "false" {
			w := prepareModel(id, history)
			mCh <- w
			p = true
			break
		}
	}
	if !p {
		w := prepareModel(id, histories[0])
		mCh <- w
	}
}

func prepareModel(id string, history map[string]string) request.BookmarkWorkItem {
	w := request.BookmarkWorkItem{
		Id:    id,
		Title: history["Title"],
	}
	if im, err := strconv.ParseBool(history["IsMasked"]); err == nil {
		w.IsMasked = im
	}
	if ct, ok := history["CreateTime"]; ok {
		w.CreateDate = ct
	}
	if ut, ok := history["UpdateTime"]; ok {
		w.UpdateDate = ut
	}
	if t, ok := history["Tags"]; ok {
		ts := strings.Split(t, "||")
		w.Tags = ts
	}
	return w
}
