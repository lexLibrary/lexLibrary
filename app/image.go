// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"github.com/rs/xid"
)

// Image is a user uploaded image, either for a profile or for a document
type Image struct {
	ID          xid.ID `json:"id"`
	OwnerID     xid.ID `json:"ownerID"`
	ContentType string `json:"contentType,omitempty"`

	Data            []byte `json:"-"` // full image
	ThumbData       []byte `json:"-"` // thumbnail image
	PlaceholderData []byte `json:"-"` // Placeholder, small quick download to show while waiting for the full image
}

// func ImageNew(owner *User, contentType string, reader io.ReadCloser) (*Image, error) {
// 	defer func() {
// 		if cerr := reader.Close(); cerr != nil && err == nil {
// 			err = cerr
// 		}
// 	}()
// }
