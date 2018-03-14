// Copyright (c) 2017-2018 Townsourced Inc.

package web

import (
	"io"
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
)

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

// func imagesFromForm(u *app.User, r *http.Request) ([]*app.Image, error) {

// 	var images []*app.Image

// 	err := r.ParseMultipartForm(maxUploadMemory)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, files := range r.MultipartForm.File {
// 		for i := range files {
// 			file, err := files[i].Open()
// 			if err != nil {
// 				return nil, err
// 			}

// 			i, err := app.ImageNew(u, files[i].Header.Get("Content-Type"), file)
// 			if err != nil {
// 				return nil, err
// 			}
// 			images = append(images, i)
// 			// TODO: file.Close?
// 		}
// 	}

// 	return images, nil
// }
