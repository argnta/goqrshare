package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang-share-files/internal/apperror"
)

func newFlagSet(cfg *Config) *flag.FlagSet {
	fs := flag.NewFlagSet("goshare", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.IntVar(&cfg.Port, "port", 0, "port HTTP (0 = pilih otomatis)")
	fs.BoolVar(&cfg.IsUpload, "upload", false, "aktifkan mode terima upload dari HP")
	fs.StringVar(&cfg.UploadDir, "dir", "", "folder tujuan upload (default ~/Downloads)")
	fs.IntVar(&cfg.MaxUploadMB, "max-size", 1024, "batas total ukuran upload per kiriman, dalam MB")
	fs.StringVar(&cfg.Iface, "iface", "", "paksa interface jaringan tertentu")
	fs.BoolVar(&cfg.IsOneTime, "once", false, "matikan server setelah satu transfer sukses")
	fs.DurationVar(&cfg.TTL, "ttl", 0, "kedaluwarsa otomatis (mis. 10m, 1h; 0 = tanpa batas)")
	fs.StringVar(&cfg.Token, "token", "", "token tetap, 8-64 karakter A-Z a-z 0-9 _ - (kosong = acak)")
	fs.BoolVar(&cfg.NoQR, "no-qr", false, "hanya cetak URL, tanpa QR")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "tampilkan log akses detail")
	fs.BoolVar(&cfg.ShowVersion, "version", false, "cetak versi lalu keluar")
	return fs
}

func Usage(w io.Writer) {
	fmt.Fprintln(w, "goshare - bagikan file lewat Wi-Fi lokal dengan QR code")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Pemakaian:")
	fmt.Fprintln(w, "  goshare [flag] <file>     bagikan satu file")
	fmt.Fprintln(w, "  goshare --upload [flag]   terima file dari HP")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flag:")
	fs := newFlagSet(&Config{})
	fs.SetOutput(w)
	fs.PrintDefaults()
}

func Parse(args []string) (Config, error) {
	var cfg Config
	fs := newFlagSet(&cfg)

	var files []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return Config{}, err
			}
			return Config{}, &apperror.Error{
				Kind: apperror.KindParse,
				Msg:  "gagal membaca argumen",
				Err:  err,
			}
		}
		rest = fs.Args()
		if len(rest) == 0 {
			break
		}
		files = append(files, rest[0])
		rest = rest[1:]
	}
	cfg.Files = files

	if cfg.ShowVersion {
		return cfg, nil
	}

	if cfg.IsUpload && cfg.UploadDir == "" {
		cfg.UploadDir = resolveDefaultUploadDir()
	}

	err := cfg.Validate()
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
