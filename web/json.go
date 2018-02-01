// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/lexLibrary/lexLibrary/app"
	"github.com/pkg/errors"
)

const maxJSONSize = 1 << 20 //10MB

var errInputTooLarge = app.NewFailureWithStatus("Input size is too large, please check your input and try again",
	http.StatusRequestEntityTooLarge)

// etagger allows any data type to set an etag version on their response
type etagger interface {
	Etag() string
}

type response struct {
	data   interface{}
	status int
}

func success(data interface{}) response {
	return response{
		data:   data,
		status: http.StatusOK,
	}
}

func created(data interface{}) response {
	return response{
		data:   data,
		status: http.StatusCreated,
	}
}

func respond(w http.ResponseWriter, response response) {
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/json")

	var result []byte
	var err error
	if devMode {
		result, err = json.MarshalIndent(response.data, "", "    ")
	} else {
		result, err = json.Marshal(response.data)
	}
	if err != nil {
		app.LogError(errors.Errorf("Error marshalling response: %s", err))
		result = []byte("An internal error occurred, and we'll look into it.")
	}

	w.WriteHeader(response.status)

	// if data has an etag use it
	switch t := response.data.(type) {
	case etagger:
		w.Header().Set("ETag", t.Etag())
	}

	_, err = w.Write(result)
	if err != nil {
		app.LogError(errors.Errorf("Error writing response: %s", err))
	}
}

func parseInput(r *http.Request, result interface{}) error {
	//TODO: use sync.Pool of buffers

	lr := &io.LimitedReader{R: r.Body, N: maxJSONSize + 1}
	buff, err := ioutil.ReadAll(lr)
	if err != nil {
		return err
	}

	if lr.N == 0 {
		return errInputTooLarge
	}

	if len(buff) == 0 {
		return nil
	}

	err = json.Unmarshal(buff, result)
	return err
}
