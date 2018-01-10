// Copyright (c) 2017 Townsourced Inc.
package web

import (
	"net/http"
)

func rootTemplate(w http.ResponseWriter, r *http.Request, c ctx) {
	// if errHandled(errors.New("test"), w, r) {
	// 	return
	// }
	errHandled(w.(*templateWriter).execute(struct {
		Test string
	}{
		Test: "test string",
	}), w, r)
}
