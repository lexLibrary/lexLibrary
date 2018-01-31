// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"sync"
)

type interrupt struct {
	name string
	fn   http.HandlerFunc
}

var interrupts = struct {
	sync.Mutex
	ir []*interrupt
}{
	ir: make([]*interrupt, 0, 5),
}

// interrupts block the normal web server flow to display a page or change behavior
// Things like a maintenance page or the first time setup where we want to prevent the rest of the site from
// being used until something is completed

func interrupted(w http.ResponseWriter, r *http.Request) bool {
	if len(interrupts.ir) == 0 {
		return false
	}

	interrupts.ir[0].fn(w, r)
	return true
}

func addInterrupt(ir *interrupt) {
	interrupts.Lock()
	defer interrupts.Unlock()

	interrupts.ir = append(interrupts.ir, ir)
}

func removeInterrupt(ir *interrupt) {
	interrupts.Lock()
	defer interrupts.Unlock()

	for i := range interrupts.ir {
		if interrupts.ir[i].name == ir.name {
			interrupts.ir = append(interrupts.ir[:i], interrupts.ir[i+1:]...)
			return
		}
	}
}
