// Copyright (c) 2017-2018 Townsourced Inc.

package app_test

import (
	"bytes"
	"image"
	"image/png"
	"io/ioutil"
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

func getImageUpload(t *testing.T, height, width int) app.Upload {
	var b bytes.Buffer

	err := png.Encode(&b, image.Rect(0, 0, width, height))
	if err != nil {
		t.Fatalf("Error generating testing image: %s", err)
	}

	return app.Upload{
		Name:         "TestImage.png",
		ContentType:  "image/png",
		ReadCloser:   ioutil.NopCloser(&b),
		LastModified: time.Now(),
	}
}
