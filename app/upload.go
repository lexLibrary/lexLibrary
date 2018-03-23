// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"io"
	"time"
)

// Upload is a common struct used for working with uploaded files
type Upload struct {
	Name        string
	ContentType string
	io.ReadCloser
	LastModified time.Time
}
