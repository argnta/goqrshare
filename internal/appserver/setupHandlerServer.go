package appserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang-share-files/internal/apperror"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type countingWriter struct {
	http.ResponseWriter
	n int
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.ResponseWriter.Write(p)
	c.n += n
	return n, err
}

func sanitizeFilename(raw string) string {
	raw = path.Base(strings.ReplaceAll(raw, "\\", "/"))

	var b strings.Builder
	for _, r := range raw {
		switch {
		case r < 32 || r == 127:
			continue
		case strings.ContainsRune(`<>:"/\|?*`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}

	name := strings.Trim(b.String(), " .")
	if name == "" {
		return "upload.bin"
	}
	return name
}

func (s *MainServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	if len(s.Options.Files) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(apperror.ErrorResponse{
			Code: "NO_FILE",
			Msg:  "Tidak ada file yang dibagikan",
		})
		return
	}

	fileDir := s.Options.Files[0]

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
	if err != nil || fileInfo.IsDir() {
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
	disp := mime.FormatMediaType("attachment", map[string]string{"filename": fileName})
	if disp == "" {
		disp = "attachment"
	}
	w.Header().Set("Content-Disposition", disp)

	cw := &countingWriter{
		ResponseWriter: w,
	}

	http.ServeContent(cw, r, fileName, fileInfo.ModTime(), file)

	if r.Method == http.MethodGet && cw.n >= int(fileInfo.Size()) {
		downloadCount := s.DownloadCount.Add(1)
		log.Printf("selesai diunduh oleh %s: %s, total %d kali", r.RemoteAddr, fileName, downloadCount)

		if s.Options.OneTime {
			s.Stop("transfer selesai (--once)")
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

	r.Body = http.MaxBytesReader(w, r.Body, s.Options.MaxUploadBytes)
	err := r.ParseMultipartForm(10 << 20)
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

	safeFileName := sanitizeFilename(header.Filename)
	ext := filepath.Ext(safeFileName)
	nameWithoutExt := strings.TrimSuffix(safeFileName, ext)

	var dstFile *os.File
	for counter := 1; ; counter++ {
		dstFile, err = os.OpenFile(filepath.Join(s.Options.UploadDir, safeFileName), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if !errors.Is(err, fs.ErrExist) {
			break
		}
		safeFileName = fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
	}

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
	_, err = io.Copy(dstFile, file)
	if cerr := dstFile.Close(); err == nil {
		err = cerr
	}

	if err != nil {
		os.Remove(filepath.Join(s.Options.UploadDir, safeFileName))
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

	if s.Options.OneTime {
		s.Stop("transfer selesai (--once)")
	}
}

func (s *MainServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := pageData{MaxMB: s.Options.MaxUploadBytes >> 20}

	if len(s.Options.Files) > 0 {
		if info, err := os.Stat(s.Options.Files[0]); err == nil {
			data.HasFile = true
			data.FileName = filepath.Base(s.Options.Files[0])
			data.FileSize = fmt.Sprintf("%.1f MB", float64(info.Size())/(1<<20))
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	pageTmpl.Execute(w, data)
}
