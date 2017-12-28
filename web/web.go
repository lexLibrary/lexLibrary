// Copyright (c) 2017 Townsourced Inc.

// Package web contains all the handling for the web server.  It should handle cookies and routing, but
// all application logic and access rules should happen in the app layer.
package web

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lexLibrary/lexLibrary/files"

	"github.com/lexLibrary/lexLibrary/app"
)

// Config is the configurable properties of the LL web server
type Config struct {
	ReadTimeout       string
	WriteTimeout      string
	MaxHeaderBytes    int
	MinTLSVersion     uint16
	CertFile          string
	KeyFile           string
	MaxUploadMemoryMB int
	Port              int
}

// DefaultConfig returns the default configuration for the web layer
func DefaultConfig() Config {
	return Config{
		MinTLSVersion:     tls.VersionTLS10,
		MaxUploadMemoryMB: 10, //10MB default
		Port:              8080,
	}
}

const (
	strictTransportSecurity = "max-age=86400"
)

var (
	zipPool         sync.Pool
	maxUploadMemory = int64(10 << 20)
	isSSL           = false
	devMode         = false
	version         = ""
)

func init() {
	zipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
}

var server *http.Server

// StartServer starts the Lex Library webserver with the passed in
// configuration
func StartServer(cfg Config, developMode bool) error {
	devMode = developMode

	err := setVersion()
	if err != nil {
		return err
	}
	if cfg.MaxUploadMemoryMB <= 0 {
		cfg.MaxUploadMemoryMB = DefaultConfig().MaxUploadMemoryMB
	}
	maxUploadMemory = int64(cfg.MaxUploadMemoryMB) << 20

	var readTimeout time.Duration
	var writeTimeout time.Duration

	if cfg.ReadTimeout != "" {
		readTimeout, err = time.ParseDuration(cfg.ReadTimeout)
		if err != nil {
			log.Printf("Invalid ReadTimeout duration format (%s), using default", cfg.ReadTimeout)
		}
	}
	if cfg.WriteTimeout != "" {
		writeTimeout, err = time.ParseDuration(cfg.WriteTimeout)
		if err != nil {
			log.Printf("Invalid WriteTimeout duration format (%s), using default", cfg.WriteTimeout)
		}
	}

	tlsCFG := &tls.Config{MinVersion: cfg.MinTLSVersion}

	server = &http.Server{
		Handler:        setupRoutes(),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
		ErrorLog:       app.Logger("Lex Library Web Server: "),
	}

	if cfg.CertFile == "" || cfg.KeyFile == "" {
		if cfg.Port <= 0 {
			cfg.Port = 80
		}
		log.Printf("Lex Library is running on port %d", cfg.Port)
		server.Addr = ":" + strconv.Itoa(cfg.Port)
		err = server.ListenAndServe()
	} else {
		if cfg.Port <= 0 {
			cfg.Port = 443
		}

		isSSL = true
		server.TLSConfig = tlsCFG

		log.Printf("Lex Library is running on port %d", cfg.Port)
		server.Addr = ":" + strconv.Itoa(cfg.Port)
		err = server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}

	return err
}

// Teardown gracefully tears down the webserver
func Teardown() error {
	if server != nil {
		return server.Shutdown(context.TODO())
	}
	return nil
}

func setVersion() error {
	b, err := files.Asset("version")
	if err != nil {
		version = "unset"
		return nil
	}

	version = strings.TrimSpace(string(b))

	return nil
}

func standardHeaders(w http.ResponseWriter) {
	if isSSL {
		w.Header().Set("Strict-Transport-Security", strictTransportSecurity)
	}
}
