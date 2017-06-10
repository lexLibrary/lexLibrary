package web

import (
	"crypto/tls"
	"fmt"
)

type Config struct {
	ReadTimeout       string
	WriteTimeout      string
	MaxHeaderBytes    int
	MinTLSVersion     uint16
	CertFile          string
	KeyFile           string
	MaxUploadMemoryMB int
}

// DefaultConfig returns the default configuration for the web layer
func DefaultConfig() Config {
	return Config{
		MinTLSVersion:     tls.VersionTLS10,
		ReadTimeout:       "60s",
		WriteTimeout:      "60s",
		MaxUploadMemoryMB: 10, //10MB default
	}
}

func Init(cfg Config) error {
	return fmt.Errorf("TODO")
}
