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
	KindAmbiguousIface
	KindFailGenerateToken
	KindRunError
	KindServerNotInitiated
	KindServerFailShutdown
	KindFailOpenFile
	KindFailStatFile
	KindFailSaveFile
	KindFailCreateFile
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
