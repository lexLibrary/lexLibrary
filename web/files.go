// Copyright (c) 2017 Townsourced Inc.

package web

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/files"

	"github.com/julienschmidt/httprouter"
)

// serveStatic serves a static file or directory.
// assumes one param for directories
//	Directory: rootHandler.GET("/images/*image", serveStatic("/images", false))
//	file: rootHandler.GET("/images/image.png", serveStatic("/images/image.png", false))
func serveStatic(fileOrDir string, compress bool) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if !devMode {
			w.Header().Set("ETag", version)
		}
		if r.Method != "GET" {
			//TODO: use LL 404 handler
			http.NotFound(w, r)
			return
		}

		file := ""
		if len(params) != 1 {
			// assume direct file
			file = fileOrDir
		} else {
			file = filepath.Join(fileOrDir, params[0].Value)
		}

		var reader *bytes.Reader
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && compress &&
			w.Header().Get("Content-Encoding") != "gzip" {
			// w.Header().Set("Content-Encoding", "gzip")
			// TODO: return already compressed data https://github.com/shuLhan/go-bindata/issues/25
			data, err := files.Asset(file)
			if err != nil {
				//TODO: use LL 404 handler
				http.NotFound(w, r)
				return
			}
			reader = bytes.NewReader(data)
		} else {
			data, err := files.Asset(file)
			if err != nil {
				//TODO: use LL 404 handler
				fmt.Println("error: ", err)
				http.NotFound(w, r)
				return
			}
			reader = bytes.NewReader(data)
		}

		standardHeaders(w)

		http.ServeContent(w, r, file, time.Time{}, reader)

	}
}
