package apperror

import (
	"errors"
	"net/http"
)

type KindError int

const (
	KindUnknown KindError = iota
	KindFileNotFound
	KindFileNotReadable
	KindIsDirectory
	KindNoTarget
	KindDirWithoutUpload
	KindPortInvalid
	KindMultipleFiles
	KindNoNetwork
	KindNoPrivateIP
	KindIfaceNotFound
	KindPortInUse
	KindTokenInvalid
	KindUploadTooLarge
	KindParse
	KindAmbiguousIface
	KindFailGenerateToken
	KindRunError
	KindServerNotInitiated
	KindServerFailShutdown
	KindFailOpenFile
	KindFailStatFile
	KindFailSaveFile
	KindFailCreateFile
	KindDirNotUsable
	KindTTLInvalid
	KindSizeInvalid
)

type Error struct {
	Kind KindError
	Msg  string
	Err  error
}

type ErrorResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

type UploadResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Filename string `json:"filename"`
}

type StatusResponseWriter struct {
	http.ResponseWriter
	Status int
}

func (w *StatusResponseWriter) WriteHeader(code int) {
	w.Status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StatusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

func (e *Error) Unwrap() error {
	return e.Err
}

func IsKind(err error, kind KindError) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind == kind
	}
	return false
}
