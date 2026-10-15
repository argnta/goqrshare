package appserver

import (
	"encoding/json"
	"golang-share-files/internal/apperror"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func (s *MainServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	fileDir := s.Options.UploadDir

	file, err := os.Open(fileDir)
	if err != nil {
		errOpen := &apperror.Error{
			Kind: apperror.KindFailOpenFile,
			Msg:  "Fail open file",
			Err:  err,
		}

		log.Println(errOpen)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "FILE_NOT_FOUND",
			Msg:  "File tidak ada atau sudah dihapus",
		})

		return
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		errStat := &apperror.Error{
			Kind: apperror.KindFailStatFile,
			Msg:  "Fail stat file",
			Err:  err,
		}

		log.Println(errStat)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "INTERNAL_ERROR",
			Msg:  "Gagal membaca informasi file",
		})

		return
	}

	fileName := filepath.Base(fileDir)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	http.ServeContent(w, r, fileName, fileInfo.ModTime(), file)

	downloadCount := s.DownloadCount.Add(1)

	if s.Options.OneTime && downloadCount == 1 {
		select {
		case <-s.Done:
		default:
			close(s.Done)
		}
	}
}
