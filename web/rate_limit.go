// Copyright (c) 2018 Townsourced Inc.
package web

import (
	"net/http"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

// X-RateLimit-Limit: 60
// X-RateLimit-Remaining: 56
// X-RateLimit-Reset: 1372700873

var requestLimit = &app.RateLimit{
	Type:   "General",
	Limit:  int32(app.SettingMust("RateLimit").Int()),
	Period: 1 * time.Minute,
}

func rateLimitHeader(w http.ResponseWriter, left app.RateLeft) {

}
