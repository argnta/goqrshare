package appserver

import (
	"net/http"
	"time"
)

type Options struct {
	Port       int
	Files      []string
	Token      string
	IsUpload   bool
	UploadDir  string
	OneTime    bool
	Duration   time.Duration
	Verbose    bool
	ListenerIP string
}

type MainServer struct {
	Server        http.Server
	Options       Options
	Token         string
	DownloadCount int
	Done          chan struct{}
}
