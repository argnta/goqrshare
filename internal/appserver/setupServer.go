package appserver

import (
	"context"
	"golang-share-files/internal/apperror"
	"net"
	"net/http"
	"strconv"
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

	s.Server = &http.Server{
		// Handler: //myhandler,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := s.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
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

	return s.Server.Shutdown(shutdownCtx)
}
