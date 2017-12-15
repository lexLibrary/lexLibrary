// Copyright (c) 2017 Townsourced Inc.

package web

import (
	"net/http"
	"strings"
)

type staticHandler struct {
	filepath string
	gzip     bool //whether or not to gzip the response
}

//registerStaticHandler creates a handler that returns a static file from the embedded file data
// filepath can be a specific file, or a directory
func registerStaticHandler(filepath string, gzip bool) *staticHandler {
	s := &staticHandler{
		filepath: filepath,
		gzip:     gzip,
	}

	return s
}

func (s *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		//TODO: use LL 404 handler
		http.NotFound(w, r)
		return
	}

	//TODO: Respond to HEAD requests for static files with version etag generated based on LL version
	// or does ServeContent handle this automatically
	w.Header().Set("ETag", gitSha)

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && s.gzip &&
		w.Header().Get("Content-Encoding") != "gzip" {
		w.Header().Set("Content-Encoding", "gzip")
		// return already compressed data
	} else {
		// decompress the data
	}

	standardHeaders(w)

	http.ServeContent(w, r, s.info.Name(), s.modTime, reader)
}
