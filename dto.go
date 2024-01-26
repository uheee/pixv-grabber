package main

type Response[T any] struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Body    T      `json:"body"`
}

type BookmarkBody struct {
	Works []BookmarkWorkItem `json:"works"`
	Total int                `json:"total"`
}

type BookmarkWorkItem struct {
	Id         any      `json:"id"`
	Title      string   `json:"title"`
	IllustType int      `json:"illustType"`
	Url        string   `json:"url"`
	Tags       []string `json:"tags"`
	PageCount  int      `json:"pageCount"`
	CreateDate string   `json:"createDate"`
	UpdateDate string   `json:"updateDate"`
	IsMasked   bool     `json:"isMasked"`
}

type ImageItem struct {
	Urls struct {
		ThumbMini string `json:"thumb_mini"`
		Small     string `json:"small"`
		Regular   string `json:"regular"`
		Original  string `json:"original"`
	} `json:"urls"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type VideoItem struct {
	OriginalSrc string `json:"originalSrc"`
}

type DownloadTask struct {
	Id   string
	Url  string
	Path string
}
