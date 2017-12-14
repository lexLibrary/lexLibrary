// Copyright (c) 2017 Townsourced Inc.

package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
)

//TODO: https://github.com/shuLhan/go-bindata

type staticHandler struct {
	filepath string
	gzip     bool //whether or not to gzip the response
}

//TODO: Respond to HEAD requests for static files with version etag generated based on LL version

func newStaticHandler(filepath string, gzip bool) *staticHandler {
	s := &staticHandler{
		filepath: filepath,
		gzip:     gzip,
	}

	return s
}

func (s *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		//TODO
		// four04(w, r)
	}
	var reader io.ReadSeeker

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && s.gzip &&
		w.Header().Get("Content-Encoding") != "gzip" {
		w.Header().Set("Content-Encoding", "gzip")
		reader = bytes.NewReader(s.zipData)
	} else {
		reader = bytes.NewReader(s.fileData)
	}

	standardHeaders(w)

	http.ServeContent(w, r, s.info.Name(), s.modTime, reader)
}

// recusively setup staticHandlers for every file under the dir
func serveStaticDir(mux *httprouter.Router, pattern, dir string, gzip bool) {
	file, err := os.OpenFile(dir, os.O_RDONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("Error opening dir for static handling %s Error: %s", dir, err))
	}
	defer func() {
		ferr := file.Close()
		if ferr != nil {
			panic(fmt.Sprintf("Error closing folder %s after read: %s", dir, ferr))
		}
	}()

	children, err := file.Readdir(-1)
	if err != nil {
		panic(fmt.Sprintf("Error dir children for static handling %s Error: %s", dir, err))
	}
	for i := range children {
		cPattern := path.Join(pattern, filepath.Base(children[i].Name()))
		if children[i].IsDir() {
			serveStaticDir(mux, cPattern, filepath.Join(dir, children[i].Name()), gzip)
		} else {
			mux.Handler("GET", cPattern, newStaticHandler(filepath.Join(dir, children[i].Name()), gzip))
		}
	}
}
