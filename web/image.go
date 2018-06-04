// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"io"
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

func serveImage(w http.ResponseWriter, r *http.Request, image *app.Image, cache bool) {
	if image == nil {
		notFound(w, r)
		return
	}
	w.Header().Set("Content-Type", image.ContentType)
	w.Header().Set("ETag", image.Etag())

	var rs io.ReadSeeker
	var err error

	// ?thumb
	// ?placeholder
	values := r.URL.Query()

	if _, ok := values["placeholder"]; ok {
		rs, err = image.Placeholder()
		if errHandled(err, w, r) {
			return
		}
	} else if _, ok := values["thumb"]; ok {
		rs, err = image.Thumb()
		if errHandled(err, w, r) {
			return
		}
	} else {
		rs, err = image.Full()
		if errHandled(err, w, r) {
			return
		}
	}

	var modTime time.Time
	if cache {
		// allow client to cache image based on last modified time
		modTime = image.ModTime
	}

	http.ServeContent(w, r, image.Name, modTime, rs)
}
