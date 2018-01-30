// Copyright (c) 2017-2018 Townsourced Inc.
package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

func init() {
	app.SettingTrigger("RateLimit", func(value interface{}) {
		requestLimit.Limit = int32(value.(int))
	})
}

var requestLimit = &app.RateLimit{
	Type:   "General",
	Limit:  2000,
	Period: 1 * time.Minute,
}

func rateLimitHeader(w http.ResponseWriter, left app.RateLeft) {
	w.Header().Add("X-RateLimit-Limit", strconv.Itoa(int(left.Limit)))
	w.Header().Add("X-RateLimit-Remaining", strconv.Itoa(int(left.Remaining)))
	w.Header().Add("X-RateLimit-Reset", strconv.Itoa(int(left.Reset.Unix())))
}
