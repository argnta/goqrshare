package config

import "time"

type Config struct {
	Path       string
	Port       int
	IsUpload   bool
	IsOneTime  bool
	DestFolder string
	Date       time.Time
}
