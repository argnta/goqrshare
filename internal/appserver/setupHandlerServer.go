package appserver

import (
	"encoding/json"
	"fmt"
	"golang-share-files/internal/apperror"
	"io"
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

func (s *MainServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)

		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "METHOD_NOT_ALLOWED",
			Msg:  "HTTP Method tidak diizinkan",
		})
		return
	}

	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)

		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "FILE_TOO_LARGE",
			Msg:  "File terlalu besar atau gagal diproses",
		})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "BAD_REQUEST",
			Msg:  "File dengan key 'file' tidak ditemukan dalam request",
		})
		return
	}
	defer file.Close()

	safeFileName := filepath.Base(header.Filename)

	destPath := filepath.Join(s.Options.UploadDir, safeFileName)
	ext := filepath.Ext(safeFileName)
	nameWithoutExt := safeFileName[:len(safeFileName)-len(ext)]

	counter := 1
	for {
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			break
		}

		safeFileName = fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
		destPath = filepath.Join(s.Options.UploadDir, safeFileName)
		counter++
	}

	dstFile, err := os.Create(destPath)
	if err != nil {
		errCreate := &apperror.Error{
			Kind: apperror.KindFailCreateFile,
			Msg:  "Gagal membuat file di direktori tujuan",
			Err:  err,
		}
		log.Println(errCreate)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "INTERNAL_ERROR",
			Msg:  "Gagal menyimpan file di server",
		})
		return
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, file)
	if err != nil {
		errCopy := &apperror.Error{
			Kind: apperror.KindFailSaveFile,
			Msg:  "Gagal menyalin stream file",
			Err:  err,
		}
		log.Println(errCopy)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "INTERNAL_ERROR",
			Msg:  "Gagal memproses unggahan file",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(apperror.UploadResponse{
		Status:   "success",
		Message:  "File berhasil diunggah",
		Filename: safeFileName,
	})
}
