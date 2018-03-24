// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"io"
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
)

func serveImage(w http.ResponseWriter, r *http.Request, image *app.Image) {
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

	http.ServeContent(w, r, image.Name, image.ModTime, rs)
}
