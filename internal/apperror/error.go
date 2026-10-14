package apperror

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

func IsKind(err error, kind KindError) bool {
	for err != nil {
		if e, ok := err.(*Error); ok {
			return e.Kind == kind
		}
		u, ok := err.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		err = u.Unwrap()
	}
	return false
}