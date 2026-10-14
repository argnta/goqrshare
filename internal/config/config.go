package config

import "time"

type Config struct {
    Files       []string      // file yang dibagikan (dari fs.Args())
    Port        int           // --port
    IsUpload    bool          // --upload
    UploadDir   string        // --dir
    Iface       string        // --iface
    IsOneTime   bool          // --once
    TTL         time.Duration // --ttl
    Token       string        // --token
    NoQR        bool          // --no-qr
    Verbose     bool          // --verbose
    ShowVersion bool          // --version
}
