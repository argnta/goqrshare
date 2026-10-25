package config

import (
	"fmt"
	"os"

	aperror "golang-share-files/internal/apperror"
)

func (c Config) Validate() error {
	checks := []func() error{
		c.validateTarget,
		c.validatePort,
		c.validateTTL,
		c.validateToken,
		c.validateSize,
		c.validateDir,
		c.validateFiles,
	}
	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}
	return nil
}

func (c Config) validateTarget() error {
	if len(c.Files) == 0 && !c.IsUpload {
		return &aperror.Error{
			Kind: aperror.KindNoTarget,
			Msg:  "tidak ada file yang dibagikan dan mode upload tidak aktif",
		}
	}
	return nil
}

func (c Config) validatePort() error {
	if c.Port == 0 {
		return nil
	}
	if c.Port < 1 || c.Port > 65535 {
		return &aperror.Error{
			Kind: aperror.KindPortInvalid,
			Msg:  fmt.Sprintf("port di luar rentang 1-65535: %d", c.Port),
		}
	}
	return nil
}

func (c Config) validateTTL() error {
	if c.TTL < 0 {
		return &aperror.Error{
			Kind: aperror.KindTTLInvalid,
			Msg:  "--ttl tidak boleh negatif",
		}
	}
	return nil
}

func (c Config) validateToken() error {
	if c.Token == "" {
		return nil
	}
	if len(c.Token) < 8 || len(c.Token) > 64 {
		return &aperror.Error{
			Kind: aperror.KindTokenInvalid,
			Msg:  "--token harus 8-64 karakter",
		}
	}
	for _, r := range c.Token {
		ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_'
		if !ok {
			return &aperror.Error{
				Kind: aperror.KindTokenInvalid,
				Msg:  "--token hanya boleh berisi A-Z a-z 0-9 - _",
			}
		}
	}
	return nil
}

func (c Config) validateSize() error {
	if c.MaxUploadMB < 1 {
		return &aperror.Error{
			Kind: aperror.KindSizeInvalid,
			Msg:  "--max-size minimal 1 (MB)",
		}
	}
	return nil
}

func (c Config) validateDir() error {
	if c.UploadDir == "" {
		return nil
	}
	if !c.IsUpload {
		return &aperror.Error{
			Kind: aperror.KindDirWithoutUpload,
			Msg:  "--dir hanya berlaku bersama --upload",
		}
	}

	info, err := os.Stat(c.UploadDir)
	if err != nil || !info.IsDir() {
		return &aperror.Error{
			Kind: aperror.KindDirNotUsable,
			Msg:  "folder upload tidak ditemukan atau bukan folder: " + c.UploadDir,
			Err:  err,
		}
	}

	probe, err := os.CreateTemp(c.UploadDir, ".goshare-check-*")
	if err != nil {
		return &aperror.Error{
			Kind: aperror.KindDirNotUsable,
			Msg:  "folder upload tidak bisa ditulis: " + c.UploadDir,
			Err:  err,
		}
	}
	probe.Close()
	os.Remove(probe.Name())

	return nil
}

func (c Config) validateFiles() error {
	if len(c.Files) > 1 {
		return &aperror.Error{
			Kind: aperror.KindMultipleFiles,
			Msg:  "satu file saja untuk saat ini",
		}
	}

	for _, path := range c.Files {
		info, statErr := os.Stat(path)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return &aperror.Error{
					Kind: aperror.KindFileNotFound,
					Msg:  "file tidak ditemukan: " + path,
					Err:  statErr,
				}
			}
			return &aperror.Error{
				Kind: aperror.KindFileNotReadable,
				Msg:  "file tidak bisa dibaca: " + path,
				Err:  statErr,
			}
		}
		if info.IsDir() {
			return &aperror.Error{
				Kind: aperror.KindIsDirectory,
				Msg:  "folder belum didukung: " + path,
			}
		}

		f, openErr := os.Open(path)
		if openErr != nil {
			return &aperror.Error{
				Kind: aperror.KindFileNotReadable,
				Msg:  "file tidak bisa dibaca: " + path,
				Err:  openErr,
			}
		}
		f.Close()
	}

	return nil
}
