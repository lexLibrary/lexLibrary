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

func rateHeader(w http.ResponseWriter, left app.RateLeft) {
	w.Header().Add("X-RateLimit-Limit", strconv.Itoa(int(left.Limit)))
	w.Header().Add("X-RateLimit-Remaining", strconv.Itoa(int(left.Remaining)))
	w.Header().Add("X-RateLimit-Reset", strconv.Itoa(int(left.Reset.Unix())))
}

// all rate limits should be defined here in one spot so they are easy to find and tweak as necessary
var requestLimit = &app.RateLimit{
	Type:   "General",
	Limit:  2000,
	Period: 1 * time.Minute,
}

var publicUserNewRateDelay = &app.RateDelay{
	Type:   "userNew",
	Limit:  2,
	Delay:  15 * time.Second,
	Period: 15 * time.Minute,
	Max:    1 * time.Minute,
}

var logonRateDelay = &app.RateDelay{
	Type:   "login",
	Limit:  10,
	Delay:  5 * time.Second,
	Period: 5 * time.Minute,
	Max:    1 * time.Minute,
}
