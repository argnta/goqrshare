package apperror

import "errors"

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
)

type Error struct {
	Kind KindError
	Msg  string
	Err  error
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

var (
	ErrFileNotFound     = errors.New("file tidak ditemukan")
	ErrFileNotReadable  = errors.New("file tidak bisa dibaca")
	ErrIsDirectory      = errors.New("folder belum didukung")
	ErrNoTarget         = errors.New("tidak ada file yang dibagikan dan mode upload tidak aktif")
	ErrDirWithoutUpload = errors.New("--dir hanya berlaku bersama --upload")
	ErrPortInvalid      = errors.New("port di luar rentang 1-65535")
	ErrMultipleFiles    = errors.New("satu file saja untuk saat ini")
	ErrNoNetwork        = errors.New("tidak ada jaringan aktif")
	ErrNoPrivateIP      = errors.New("tidak ada IP privat yang cocok")
	ErrIfaceNotFound    = errors.New("interface tidak ditemukan")
	ErrPortInUse        = errors.New("port sudah dipakai")
	ErrTokenInvalid     = errors.New("token tidak valid")
	ErrUploadTooLarge   = errors.New("upload melebihi batas")
	ErrParse            = errors.New("error parsing")
)
