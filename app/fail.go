// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"fmt"
	"net/http"
)

//TODO: handle multiple languages here, or do it all in the client?

// Fail is an error whos contents can be exposed to the client and is usually the result
// of incorrect client input
type Fail struct {
	Message    string `json:"message,omitempty"`
	HTTPStatus int    `json:"-"` //gets set in the error response
}

func (f *Fail) Error() string {
	return f.Message
}

// NewFailure creates a new failure with a default status of 400
func NewFailure(message string, args ...interface{}) *Fail {
	return &Fail{
		Message:    fmt.Sprintf(message, args...),
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewFailureWithStatus creates a new failure with the passed in http status code
func NewFailureWithStatus(message string, httpStatus int) *Fail {
	return &Fail{
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// FailureFromErr returns a new failure based on the passed in error
// if passed in error is nil, then nil is returned
func FailureFromErr(err error, httpStatus int) *Fail {
	if err == nil {
		return nil
	}

	return NewFailure(err.Error(), httpStatus)
}

// IsFail tests whether the passed in error is a failure
func IsFail(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *Fail:
		return true
	default:
		return false
	}
}

// NotFound creates a NotFound failure that returns to the user as a 404
func NotFound(message string, args ...interface{}) *Fail {
	return NewFailureWithStatus(fmt.Sprintf(message, args...), http.StatusNotFound)
}

// Unauthorized returns an Unauthorized error for when a user doesn't have access to a resource
func Unauthorized(message string, args ...interface{}) *Fail {
	return NewFailureWithStatus(fmt.Sprintf(message, args...), http.StatusUnauthorized)
}

// Conflict is the error retured when a record is being updated, but it's not the most current version
// of the record (409)
func Conflict(message string, args ...interface{}) *Fail {
	return NewFailureWithStatus(fmt.Sprintf(message, args...), http.StatusConflict)
}
