// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/disintegration/imaging"
	"github.com/lexLibrary/lexLibrary/data"
	"github.com/pkg/errors"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	imageMaxSize      = int64(10 << 20) //10MB
	imageMaxDimension = 4096            // largest possible image dimension allowed in the system

	// thumbnails and placeholders are generated by scaling the image so that the smallest of the two dimensions
	// is set to the ThumbPx or PlaceholdPx value
	imageDefaultThumbDimension       = 300 // default thumb pixels
	imageDefaultPlaceholderDimension = 30  // default placeholder pixels
)

var (
	errImageNotFound = NotFound("Image not found")
	errImageTooLarge = NewFailureWithStatus(fmt.Sprintf("The uploaded image is too large.  The max size is %d MB",
		imageMaxSize>>20), http.StatusRequestEntityTooLarge)
	errImageInvalidType = NewFailure("Unsupported image type, please use a jpeg, png, or gif")
)

var imageValidContentTypes = []string{"image/gif", "image/jpeg", "image/png"}

var (
	sqlImageInsert = data.NewQuery(`
		insert into images (
			id, 
			name,
			version,
			content_type,
			data,
			thumb,
			placeholder,
			updated,
			created
		) values (
			{{arg "id"}}, 
			{{arg "name"}},
			{{arg "version"}},
			{{arg "content_type"}},
			{{arg "data"}},
			{{arg "thumb"}},
			{{arg "placeholder"}},
			{{arg "updated"}},
			{{arg "created"}}
		)
	`)
	sqlImageDelete = data.NewQuery(`delete from images where id = {{arg "id"}}`)

	sqlImageGetFull        = data.NewQuery(`select name, version, content_type, data, updated from images where id = {{arg "id"}}`)
	sqlImageGetThumb       = data.NewQuery(`select name, version, content_type, thumb, updated from images where id = {{arg "id"}}`)
	sqlImageGetPlaceholder = data.NewQuery(`select name, version, content_type, placeholder, updated from images where id = {{arg "id"}}`)

	sqlImageUpdate = data.NewQuery(`
		update images
		set	data = {{arg "data"}},
			thumb = {{arg "thumb"}},
			placeholder = {{arg "placeholder"}},
			updated = {{now}},
			version = version + 1
		where id = {{arg "id"}}
		and version = {{arg "version"}}
	`)
)

// Image is a user uploaded image, either for a profile or for a document presented on the web
// depending on how the image is looked up, it may be fullsize, thumbnail or a placeholder
type Image struct {
	ID          data.ID
	Name        string
	Version     int
	ContentType string
	ModTime     time.Time
}

// Full returns the full size image
func (i *Image) Full() (io.ReadSeeker, error) {
	return i.fromRow(sqlImageGetFull.QueryRow(sql.Named("id", i.ID)))
}

// Thumb returns the image thumbnail
func (i *Image) Thumb() (io.ReadSeeker, error) {
	return i.fromRow(sqlImageGetThumb.QueryRow(sql.Named("id", i.ID)))
}

// Placeholder returns the image placeholder which is shown while waiting for the
// rest of the image to load
func (i *Image) Placeholder() (io.ReadSeeker, error) {
	return i.fromRow(sqlImageGetPlaceholder.QueryRow(sql.Named("id", i.ID)))
}

func (i *Image) fromRow(row *sql.Row) (io.ReadSeeker, error) {
	var b []byte
	err := row.Scan(
		&i.Name,
		&i.Version,
		&i.ContentType,
		&b,
		&i.ModTime,
	)
	if err == sql.ErrNoRows {
		return nil, errImageNotFound
	}
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}

// Etag returns an etag specifing this unique version of the image
func (i *Image) Etag() string {
	return fmt.Sprintf("%s-%d", i.ID, i.Version)
}

func imageGet(id data.ID) *Image {
	if !id.Valid {
		return nil
	}
	return &Image{ID: id}
}

func (i *Image) raw() (*imageRaw, error) {
	raw := &imageRaw{}

	err := sqlImageGetFull.QueryRow(sql.Named("id", i.ID)).Scan(
		&raw.name,
		&raw.version,
		&raw.contentType,
		&raw.data,
		&raw.updated,
	)
	if err == sql.ErrNoRows {
		return nil, errImageNotFound
	}
	if err != nil {
		return nil, err
	}

	err = raw.decode()
	if err != nil {
		return nil, err
	}

	return raw, nil
}

type imageRaw struct {
	id          data.ID
	name        string
	version     int
	contentType string
	data        []byte // full image
	thumb       []byte // thumbnail image
	placeholder []byte // Placeholder, small quick download to show while waiting for the full image
	updated     time.Time
	created     time.Time

	thumbMinDimension       int
	placeholderMinDimension int
	decoded                 image.Image // decoded image
}

// imageNew creates a new imageRaw object for manipulating an image before inserting it into the database
// you must call insert separately to actually insert the data
func imageNew(upload Upload) (*imageRaw, error) {
	lr := &io.LimitedReader{R: upload, N: (imageMaxSize + 1)}
	buff, err := ioutil.ReadAll(lr)
	defer func() {
		if cerr := upload.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if err != nil {
		return nil, err
	}
	if lr.N == 0 {
		return nil, errImageTooLarge
	}

	i := &imageRaw{
		id:                      data.NewID(),
		name:                    upload.Name,
		version:                 0,
		contentType:             upload.ContentType,
		data:                    buff,
		created:                 time.Now(),
		thumbMinDimension:       imageDefaultThumbDimension,
		placeholderMinDimension: imageDefaultPlaceholderDimension,
	}

	err = i.validate()
	if err != nil {
		return nil, err
	}

	err = i.prep()
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (i *imageRaw) insert(tx *sql.Tx) error {
	err := i.encode()
	if err != nil {
		return err
	}
	_, err = sqlImageInsert.Tx(tx).Exec(
		sql.Named("id", i.id),
		sql.Named("name", i.name),
		sql.Named("version", i.version),
		sql.Named("content_type", i.contentType),
		sql.Named("data", i.data),
		sql.Named("thumb", i.thumb),
		sql.Named("placeholder", i.placeholder),
		sql.Named("created", i.created),
	)
	return err
}

func (i *imageRaw) update(tx *sql.Tx, version int) error {
	err := i.encode()
	if err != nil {
		return err
	}
	r, err := sqlImageUpdate.Exec(
		sql.Named("data", i.data),
		sql.Named("thumb", i.thumb),
		sql.Named("placeholder", i.placeholder),
		sql.Named("id", i.id),
		sql.Named("version", version),
	)

	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return Conflict("You are not editing the most current version of this image. Please refresh and try again")
	}

	i.version++
	return nil
}

func (i *imageRaw) validate() error {
	if len(i.data) == 0 {
		return errors.New("Image not loaded properly")
	}
	//check content type
	found := false
	for j := range imageValidContentTypes {
		if i.contentType == imageValidContentTypes[j] {
			found = true
			break
		}
	}
	if !found {
		return errImageInvalidType
	}

	return nil
}

func imageDelete(tx *sql.Tx, id data.ID) error {
	result, err := sqlImageDelete.Exec(sql.Named("id", id))
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errImageNotFound
	}
	return nil
}

// prep prepares the image for manipulating
func (i *imageRaw) prep() error {
	if i.decoded != nil {
		//image is already prepped
		return nil
	}
	orientation := 1
	if i.contentType == "image/jpeg" {
		// rotate based on exif data if applicable
		ex, err := exif.Decode(bytes.NewReader(i.data))
		// if at any point we can't parse exif data properly, then we stick to default orientation
		if err == nil {
			//lookup orientation
			tag, err := ex.Get(exif.Orientation)
			if err == nil {
				orientation, _ = tag.Int(0)
			}
		}

	}

	//resize to max allowable image size
	err := i.decode()
	if err != nil {
		return err
	}

	//0 will preserve the ratio
	newWidth := 0
	newHeight := 0

	if i.decoded.Bounds().Dx() > imageMaxDimension {
		newWidth = imageMaxDimension
	} else if i.decoded.Bounds().Dy() > imageMaxDimension {
		newHeight = imageMaxDimension
	}

	if newWidth != 0 || newHeight != 0 {
		err = i.resize(newWidth, newHeight)
		if err != nil {
			return err
		}
	}

	switch orientation {
	case 2:
		i.decoded = imaging.FlipH(i.decoded)
	case 3:
		i.decoded = imaging.Rotate180(i.decoded)
	case 4:
		i.decoded = imaging.FlipV(i.decoded)
	case 5:
		i.decoded = imaging.Rotate90(imaging.FlipH(i.decoded))
	case 6:
		i.decoded = imaging.Rotate270(i.decoded)
	case 7:
		i.decoded = imaging.FlipH(imaging.Rotate90(i.decoded))
	case 8:
		i.decoded = imaging.Rotate90(i.decoded)
	}

	return nil
}

// decode decodes the image data into the go Image format for processing
func (i *imageRaw) decode() error {
	if i.decoded != nil {
		//image is already decoded
		return nil
	}

	if len(i.data) == 0 {
		return errors.New("Image not loaded properly.")
	}

	var err error
	buffer := bytes.NewBuffer(i.data)

	switch i.contentType {
	case "image/gif":
		i.decoded, err = gif.Decode(buffer)
	case "image/jpeg":
		i.decoded, err = jpeg.Decode(buffer)
	case "image/png":
		i.decoded, err = png.Decode(buffer)
	default:
		return errImageInvalidType
	}

	if err != nil {
		return err
	}

	i.data = nil

	return nil
}

// encode encodes the image format data into image data
func (i *imageRaw) encode() error {
	if i.decoded == nil {
		// image is already encoded
		return nil
	}

	buffer := bytes.NewBuffer(i.data)

	err := imageEncode(i.decoded, i.contentType, buffer)

	if err != nil {
		return err
	}

	i.data = buffer.Bytes()
	err = i.buildThumbAndPlaceholder()
	if err != nil {
		return err
	}
	i.decoded = nil

	return nil
}

func imageEncode(image image.Image, contentType string, result *bytes.Buffer) error {
	switch contentType {
	case "image/gif":
		return gif.Encode(result, image, nil)
	case "image/jpeg":
		return jpeg.Encode(result, image, nil)
	case "image/png":
		return png.Encode(result, image)
	default:
		return errImageInvalidType
	}
}

func (i *imageRaw) buildThumbAndPlaceholder() error {
	//0 will preserve the ratio
	newWidth := 0
	newHeight := 0
	var err error
	// create thumb and placeholder seqeuentially to benefit from
	// increasingly smaller images

	//first thumb

	//thumbnail rules are a bit different, resize so that the smallest
	// dimension is resized instead of the largest, this is to prevent
	// cases where really tall or really wide images create poor
	// thumbnails
	if i.decoded.Bounds().Dx() > i.decoded.Bounds().Dy() {
		//wider than tall
		newHeight = i.thumbMinDimension
	} else {
		//taller than wide or same
		newWidth = i.thumbMinDimension
	}

	if newWidth != 0 || newHeight != 0 {
		err = i.resize(newWidth, newHeight)
		if err != nil {
			return err
		}
	}

	buffer := bytes.NewBuffer(i.thumb)
	err = imageEncode(i.decoded, i.contentType, buffer)
	if err != nil {
		return err
	}

	i.thumb = buffer.Bytes()

	// then placeholder
	newWidth = 0
	newHeight = 0

	if i.decoded.Bounds().Dx() > i.placeholderMinDimension {
		newWidth = i.placeholderMinDimension
	} else if i.decoded.Bounds().Dy() > i.placeholderMinDimension {
		newHeight = i.placeholderMinDimension
	}

	if newWidth != 0 || newHeight != 0 {
		err = i.resize(newWidth, newHeight)
		if err != nil {
			return err
		}
	}

	buffer = bytes.NewBuffer(i.placeholder)
	err = imageEncode(i.decoded, i.contentType, buffer)
	if err != nil {
		return err
	}

	i.placeholder = buffer.Bytes()

	return nil
}

func (i *imageRaw) resize(width, height int) error {
	err := i.decode()
	if err != nil {
		return err
	}

	i.decoded = imaging.Resize(i.decoded, width, height, imaging.Linear)
	return nil
}

func (i *imageRaw) crop(x0, y0, x1, y1 int) error {
	err := i.decode()
	if err != nil {
		return err
	}

	i.decoded = imaging.Crop(i.decoded, image.Rect(x0, y0, x1, y1))
	return nil
}

func (i *imageRaw) cropCenter(width, height int) error {
	err := i.decode()
	if err != nil {
		return err
	}

	i.decoded = imaging.CropCenter(i.decoded, width, height)
	return nil
}
