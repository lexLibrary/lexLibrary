// Copyright (c) 2017 Townsourced Inc.

package app_test

import (
	"testing"
	"time"

	"github.com/lexLibrary/lexLibrary/app"
)

func TestRateLimit(t *testing.T) {

	rType := app.RateLimit{
		Type:   "TestRateLimit",
		Limit:  10,
		Period: 5 * time.Second,
	}

	// free attempts
	for i := 0; i < int(rType.Limit); i++ {
		left, err := rType.Attempt("testID")
		if err != nil {
			t.Fatalf("Error attempting rate limit withing limits: %s", err)
		}
		if left.Remaining != rType.Limit-int32(i) {
			t.Fatalf("Incorrect rate left. Expected %d, got %d", rType.Limit-int32(i), left.Remaining)
		}
	}

	left, err := rType.Attempt("testID")
	if err != app.ErrTooManyRequests {
		t.Fatalf("Rate limited request didn't return an error")
	}
	if left.Remaining != 0 {
		t.Fatalf("Rate limit remaining is incorrect. Expected %d got %d", 0, left.Remaining)
	}

	if testing.Short() {
		t.SkipNow()
	}

	// attempt limit should be freed after range expires
	time.Sleep(rType.Period)
	_, err = rType.Attempt("testID")
	if err != nil {
		t.Fatalf("Rate limit did not expire")
	}

}
