package main

import (
	"database/sql"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/hashicorp/go-retryablehttp"
)

type Config struct {
	AutoFetchMeetingData bool
	FetchOtherMedia      bool
	CreatePlaylist       bool
	PurgeSaveDir         bool
	Resolution           string
	SaveLocation         string
	Language             string
	Songs                []string
	Pictures             []file
	Videos               []video
	Progress             *progress
	HttpClient           *retryablehttp.Client
	Date                 time.Time
}

type video struct {
	Name           string
	IssueTagNumber int
	MepsDocumentID sql.NullInt64
	Track          sql.NullInt64
	KeySymbol      sql.NullString
}

type mediaInfo struct {
	Files map[string]LanguageFiles
}

type LanguageFiles struct {
	JWPUB []JWPubItem `json:"JWPUB"`
	MP4   []MP4       `json:"MP4"`
}

type progress struct {
	Total       int64 // Total # of bytes written
	Title       string
	ProgressBar *widget.ProgressBar
}

type file struct {
	Name    string
	Payload []byte
}

type Document struct {
	ID   int
	Date time.Time
}

type MeetingData struct {
	DateString string
	Songs      []string
	Pictures   []file
	Videos     []video
}

type JWPubItem struct {
	File struct {
		URL string `json:"url"`
	} `json:"file"`
	Filesize int `json:"filesize"`
}

type MP4 struct {
	Title string `json:"title"`
	Track int    `json:"track"`
	File  struct {
		URL string `json:"url"`
	} `json:"file"`
	Filesize int `json:"filesize"`
}

type PubVideo struct {
	Media []Media `json:"media"`
}

type Media struct {
	Files []Files `json:"files"`
}

type Files struct {
	Progressivedownloadurl string `json:"progressiveDownloadURL"`
	Filesize               int    `json:"filesize"`
}

type Multimedia struct {
	Track    string
	FilePath string
}