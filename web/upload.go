// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

const lastModifiedHeader = "LL-LastModified"

// func imageGet(w http.ResponseWriter, r *http.Request, c ctx) {
// 	id, err := xid.FromString(c.params.ByName("image"))
// 	if err != nil {
// 		notFound(w, r)
// 		return
// 	}

// 	// ?thumb
// 	// ?placeholder
// 	values := r.URL.Query()

// 	var i *app.Image

// 	if _, ok := values["placeholder"]; ok {
// 		i, err = app.ImageGetPlaceholder(id)
// 		if errHandled(err, w, r) {
// 			return
// 		}
// 	} else if _, ok := values["thumb"]; ok {
// 		i, err = app.ImageGetThumb(id)
// 		if errHandled(err, w, r) {
// 			return
// 		}
// 	} else {
// 		i, err = app.ImageGet(id)
// 		if errHandled(err, w, r) {
// 			return
// 		}
// 	}

// 	serveImage(w, r, i)
// }

// func imagePost(w http.ResponseWriter, r *http.Request, c ctx) {
// 	// new image
// 	if c.session == nil {
// 		unauthorized(w, r)
// 		return
// 	}

// 	u, err := c.session.User()
// 	if errHandled(err, w, r) {
// 		return
// 	}

// 	images, err := imagesFromForm(u, r)
// 	if errHandled(err, w, r) {
// 		return
// 	}

// 	respond(w, created(images))
// }

func filesFromForm(r *http.Request) ([]app.Upload, error) {
	var uploads []app.Upload

	err := r.ParseMultipartForm(maxUploadMemory)
	if err != nil {
		return nil, err
	}

	for _, header := range r.MultipartForm.File {
		for i := range header {
			file, err := header[i].Open()
			if err != nil {
				return nil, err
			}

			lastModified, err := time.Parse(time.RFC3339, header[i].Header.Get(lastModifiedHeader))
			if err != nil {
				lastModified = time.Time{}
			}

			uploads = append(uploads, app.Upload{
				Name:         header[i].Filename,
				ContentType:  header[i].Header.Get("Content-Type"),
				ReadCloser:   file,
				LastModified: lastModified,
			})
		}
	}

	return uploads, nil
}
