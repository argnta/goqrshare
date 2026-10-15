package appserver

import (
	"context"
	"golang-share-files/internal/apperror"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func NewServer(opts Options) *MainServer {
	dataServer := &MainServer{
		Options: opts,
		Done:    make(chan struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/"+opts.Token+"/", dataServer.handleDownload)
	if opts.IsUpload {
		mux.HandleFunc("/"+opts.Token+"/upload", dataServer.handleUpload)
	}

	var handler http.Handler = mux
	if opts.Verbose {
		handler = withLogging(handler)
	}

	dataServer.Server = &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second, 
	}

	return dataServer
}

func (s *MainServer) Run(ctx context.Context) error {
	addr := net.JoinHostPort(s.Options.ListenerIP, strconv.Itoa(s.Options.Port))
	listener, err := net.Listen("tcp", addr)

	runError := &apperror.Error{
		Kind: apperror.KindRunError,
		Msg:  "Error run server",
		Err:  err,
	}

	if err != nil {
		return runError
	}

	s.Options.Port = listener.Addr().(*net.TCPAddr).Port

	if s.Options.Duration > 0 {
		var once sync.Once
		closeDone := func() {
			once.Do(func() {
				close(s.Done)
			})
		}

		timer := time.AfterFunc(s.Options.Duration, closeDone)
		defer timer.Stop()
	}

	errChan := make(chan error, 1)
	go func() {
		err := s.Server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
	case <-s.Done:
	case err := <-errChan:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}

func (s *MainServer) Shutdown(ctx context.Context) error {
	if s.Server == nil {
		errNil := &apperror.Error{
			Kind: apperror.KindServerNotInitiated,
			Msg:  "Server not initiated yet",
			Err:  nil,
		}
		return errNil
	}

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
