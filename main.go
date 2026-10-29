package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"golang-share-files/internal/apperror"
	"golang-share-files/internal/appserver"
	"golang-share-files/internal/config"
	"golang-share-files/internal/netutil"
	"golang-share-files/internal/qr"
)

var version = "1.0.0"

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			config.Usage(os.Stdout)
			return 0
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		if apperror.IsKind(err, apperror.KindNoTarget) || apperror.IsKind(err, apperror.KindParse) {
			fmt.Fprintln(os.Stderr)
			config.Usage(os.Stderr)
		}
		return 2
	}

	if cfg.ShowVersion {
		fmt.Println("goshare", version)
		return 0
	}

	cand, err := pickInterface(cfg.Iface)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	if !cand.IsPrivate {
		fmt.Fprintf(os.Stderr, "Peringatan: %s (%s) bukan alamat privat; pastikan ini memang jaringan yang dimaksud.\n", cand.Name, cand.IP)
	}

	token := cfg.Token
	if token == "" {
		token, err = appserver.GenerateToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			return 1
		}
	}

	files := make([]string, 0, len(cfg.Files))
	for _, f := range cfg.Files {
		abs, absErr := filepath.Abs(f)
		if absErr != nil {
			abs = f
		}
		files = append(files, abs)
	}

	srv := appserver.NewServer(appserver.Options{
		Port:           cfg.Port,
		Files:          files,
		Token:          token,
		IsUpload:       cfg.IsUpload,
		UploadDir:      cfg.UploadDir,
		MaxUploadBytes: int64(cfg.MaxUploadMB) << 20,
		OneTime:        cfg.IsOneTime,
		Duration:       cfg.TTL,
		Verbose:        cfg.Verbose,
		ListenerIP:     cand.IP.String(), // hanya di interface LAN ini, bukan 0.0.0.0
	})

	if err := srv.Listen(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	url := appserver.BuildURL(cand.IP.String(), srv.Options.Port, token)
	printBanner(cfg, files, cand, url)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Serve(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	reason := srv.Reason()
	if reason == "" {
		reason = "dihentikan pengguna"
	}
	fmt.Println("\nServer berhenti:", reason)
	return 0
}

func pickInterface(name string) (netutil.IPCandidate, error) {
	cands, err := netutil.FindCandidates()
	if err != nil {
		return netutil.IPCandidate{}, err
	}
	if name != "" {
		return netutil.PickByName(cands, name)
	}
	return netutil.BestCandidate(cands)
}

func printBanner(cfg config.Config, files []string, cand netutil.IPCandidate, url string) {
	fmt.Println()
	if len(files) > 0 {
		name := filepath.Base(files[0])
		if info, err := os.Stat(files[0]); err == nil {
			fmt.Printf("Membagikan : %s (%d byte)\n", name, info.Size())
		} else {
			fmt.Printf("Membagikan : %s\n", name)
		}
	}
	if cfg.IsUpload {
		fmt.Printf("Upload ke  : %s (maks %d MB)\n", cfg.UploadDir, cfg.MaxUploadMB)
	}
	fmt.Printf("Jaringan   : %s (%s)\n", cand.Name, cand.IP)
	if cfg.TTL > 0 {
		fmt.Printf("Kedaluwarsa: %s lagi\n", cfg.TTL.Round(time.Second))
	}
	if cfg.IsOneTime {
		fmt.Println("Mode       : sekali pakai (berhenti setelah 1 transfer sukses)")
	}
	fmt.Println()
	fmt.Println("Buka di HP (Wi-Fi yang sama):")
	fmt.Println("  " + url)
	fmt.Println()

	if !cfg.NoQR {
		if err := qr.Print(url); err != nil {
			fmt.Fprintln(os.Stderr, "Gagal membuat QR:", err)
		}
	}

	fmt.Println("Tekan Ctrl+C untuk berhenti. Koneksi dienkripsi (HTTPS, sertifikat self-signed): browser akan menampilkan peringatan, pilih Lanjutkan.")
}
