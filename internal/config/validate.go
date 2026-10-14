package config

import (
	"fmt"
	"os"

	aperror "golang-share-files/internal/apperror"
)

func (c Config) Validate() error {
	var err error

	err = c.validateTarget()
	if err != nil {
		return err
	}

	err = c.validatePort()
	if err != nil {
		return err
	}

	err = c.validateDir()
	if err != nil {
		return err
	}

	err = c.validateFiles()
	if err != nil {
		return err
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
	}

	return nil
}