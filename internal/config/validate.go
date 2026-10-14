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

	return err
}

func (c Config) validateTarget() error {
	if len(c.Files) == 0 && !c.IsUpload {
		return aperror.ErrNoTarget
	}
	return nil
}

func (c Config) validatePort() error {
	if c.Port == 0 {
		return nil
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("%w: %d", aperror.ErrPortInvalid, c.Port)
	}
	return nil
}

func (c Config) validateDir() error {
	if c.UploadDir == "" {
		return nil
	}
	if !c.IsUpload {
		return aperror.ErrDirWithoutUpload
	}
	return nil
}

func (c Config) validateFiles() error {
	if len(c.Files) > 1 {
		return aperror.ErrMultipleFiles
	}
	for _, path := range c.Files {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%w: %s", aperror.ErrFileNotFound, path)
			}
			return fmt.Errorf("%w: %s", aperror.ErrFileNotReadable, path)
		}
		if info.IsDir() {
			return fmt.Errorf("%w: %s", aperror.ErrIsDirectory, path)
		}
	}
	return nil
}
