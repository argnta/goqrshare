package appserver

import (
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Options struct {
	Port           int
	Files          []string
	Token          string
	IsUpload       bool
	UploadDir      string
	OneTime        bool
	Duration       time.Duration
	Verbose        bool
	ListenerIP     string
	MaxUploadBytes int64
}

type MainServer struct {
	Server        *http.Server
	Options       Options
	DownloadCount atomic.Int64
	Done          chan struct{}
	listener      net.Listener
	stopOnce      sync.Once
	reason        atomic.Value
}

func (s *MainServer) Stop(reason string) {
	s.stopOnce.Do(func() {
		s.reason.Store(reason)
		close(s.Done)
	})
}

func (s *MainServer) Reason() string {
	v, _ := s.reason.Load().(string)
	return v
}
