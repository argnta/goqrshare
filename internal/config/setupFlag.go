package config

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"

	"golang-share-files/internal/apperror"
)

func Parse(args []string) (Config, error) {
	fs := flag.NewFlagSet("goshare", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var cfg Config

	fs.IntVar(&cfg.Port, "port", 0, "port HTTP (0 = pilih otomatis)")
	fs.BoolVar(&cfg.IsUpload, "upload", false, "aktifkan mode terima upload dari HP")
	fs.StringVar(&cfg.UploadDir, "dir", "", "folder tujuan upload (default ~/Downloads)")
	fs.StringVar(&cfg.Iface, "iface", "", "paksa interface jaringan tertentu")
	fs.BoolVar(&cfg.IsOneTime, "once", false, "matikan server setelah satu transfer sukses")
	fs.DurationVar(&cfg.TTL, "ttl", 0, "kedaluwarsa otomatis (mis. 10m, 1h; 0 = tanpa batas)")
	fs.StringVar(&cfg.Token, "token", "", "token tetap (kosong = acak; untuk pengujian)")
	fs.BoolVar(&cfg.NoQR, "no-qr", false, "hanya cetak URL, tanpa QR")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "tampilkan log akses detail")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "cetak versi lalu keluar")

	err := fs.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Config{}, err
		}
		return Config{}, &apperror.Error{
			Kind: apperror.KindParse,
			Msg:  "gagal membaca argumen",
			Err:  err,
		}
	}

	cfg.Files = fs.Args()

	if cfg.IsUpload && cfg.UploadDir == "" {
		cfg.UploadDir = resolveDefaultUploadDir()
	}

	err = cfg.Validate()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func resolveDefaultUploadDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	downloads := filepath.Join(home, "Downloads")
	info, err := os.Stat(downloads)
	if err != nil || !info.IsDir() {
		return "."
	}

	return downloads
}
