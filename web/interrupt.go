// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"sync"
)

var interrupts = struct {
	sync.Mutex
	f []http.HandlerFunc
}{
	f: make([]http.HandlerFunc),
}

// interrupts block the normal web server flow to display a page or change behavior
// Things like a maintenance page or the first time setup where we want to prevent the rest of the site from
// being used until something is completed

func interrupted(w http.ResponseWriter, r *http.Request) bool {
	if len(interrupts.f) == 0 {
		return false
	}

	interrupts.f[0](w, r)
	return true
}

func addInterrupt(fn http.HandlerFunc) {
	interrupts.Lock()
	defer interrupts.Unlock()

	interrupts.f = append(interrupts.f, fn)
}

func removeInterrupt(fn http.HandlerFunc) {
	interrupts.Lock()
	defer interrupts.Unlock()

	for i := range interrupts.f {
		if interrupts.f[i] == fn {
			interrupts.f = append(interrupts.f[:i], interrupts.f[i+1:]...)
			return
		}
	}
}
