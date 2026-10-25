package appserver

import (
	"context"
	"errors"
	"golang-share-files/internal/apperror"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func NewServer(opts Options) *MainServer {
	dataServer := &MainServer{
		Options: opts,
		Done:    make(chan struct{}),
	}

	base := "/" + opts.Token
	mux := http.NewServeMux()
	mux.HandleFunc("/"+opts.Token+"/", dataServer.handleDownload)

	if opts.IsUpload {
		// mux.HandleFunc("GET "+base+"/{$}", .handleIndex)
		mux.HandleFunc("POST "+base+"/upload", dataServer.handleUpload)
		if len(opts.Files) > 0 {
			mux.HandleFunc("GET "+base+"/download", dataServer.handleDownload)
		}
	} else {
		mux.HandleFunc("GET "+base+"/{$}", dataServer.handleDownload)
	}

	var handler http.Handler = mux
	handler = withSecurityHeaders(handler)
	if opts.Verbose {
		handler = withLogging(handler)
	}

	dataServer.Server = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	return dataServer
}

func (s *MainServer) Listen() error {
	addr := net.JoinHostPort(s.Options.ListenerIP, strconv.Itoa(s.Options.Port))
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		kind := apperror.KindRunError
		msg := "gagal membuka port"
		if isAddrInUse(err) {
			kind = apperror.KindPortInUse
			msg = "port sudah dipakai program lain (coba --port 0)"
		}
		return &apperror.Error{Kind: kind, Msg: msg, Err: err}
	}

	s.listener = listen
	s.Options.Port = listen.Addr().(*net.TCPAddr).Port
	return nil
}

func isAddrInUse(err error) bool {
	m := strings.ToLower(err.Error())
	return strings.Contains(m, "address already in use") || strings.Contains(m, "only one usage of each socket address")
}

func (s *MainServer) Serve(ctx context.Context) error {
	if s.listener == nil {
		return &apperror.Error{
			Kind: apperror.KindServerNotInitiated,
			Msg:  "Listen() belum dipanggil",
		}
	}

	if s.Options.Duration > 0 {
		timer := time.AfterFunc(s.Options.Duration, func() { s.Stop("waktu habis (ttl)") })
		defer timer.Stop()
	}

	errCh := make(chan error, 1)
	go func() {
		err := s.Server.Serve(s.listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	grace := 3 * time.Second
	select {
	case <-ctx.Done():
	case <-s.Done:
		grace = 5 * time.Second // beri waktu koneksi yang masih menulis
	case err := <-errCh:
		return &apperror.Error{
			Kind: apperror.KindRunError,
			Msg:  "server berhenti karena error",
			Err:  err,
		}
	}

	return s.Shutdown(grace)
}

func (s *MainServer) Shutdown(grace time.Duration) error {
	if s.Server == nil {
		errNil := &apperror.Error{
			Kind: apperror.KindServerNotInitiated,
			Msg:  "Server not initiated yet",
			Err:  nil,
		}
		return errNil
	}

	ctx, cancel := context.WithTimeout(context.Background(), grace)
	defer cancel()

	err := s.Server.Shutdown(ctx)
	if err != nil {
		errShutdown := &apperror.Error{
			Kind: apperror.KindServerFailShutdown,
			Msg:  "Shutdown failed",
			Err:  err,
		}
		return errShutdown
	}

	return nil
}
