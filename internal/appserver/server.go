package appserver

import (
	"net/http"
	"sync/atomic"
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
	Server        *http.Server
	Options       Options
	DownloadCount atomic.Int64
	Done          chan struct{}
}
