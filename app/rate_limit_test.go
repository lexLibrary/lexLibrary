// Copyright (c) 2017-2018 Townsourced Inc.

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
	for i := 1; i <= int(rType.Limit); i++ {
		left, err := rType.Attempt("testID")
		if err != nil {
			t.Fatalf("Error attempting rate limit withing limits. Attempt #%d rate left %d: %s",
				i, left.Remaining, err)
		}
		if left.Remaining != rType.Limit-int32(i) {
			t.Fatalf("Incorrect rate left. Expected %d, got %d", rType.Limit-int32(i), left.Remaining)
		}
	}

	left, err := rType.Attempt("testID")
	if err != app.ErrTooManyRequests {
		t.Fatalf("Rate limited request didn't return an error")
	}
	if left.Remaining != -1 {
		t.Fatalf("Rate limit remaining is incorrect. Expected %d got %d", -1, left.Remaining)
	}

	if testing.Short() {
		t.SkipNow()
	}

	// attempt limit should be freed after range expires
	time.Sleep(rType.Period)

	// rate limits should reset
	for i := 1; i <= int(rType.Limit); i++ {
		left, err := rType.Attempt("testID")
		if err != nil {
			t.Fatalf("Rate limit did not expire. Attempt #%d rate left %d: %s",
				i, left.Remaining, err)

		}
		if left.Remaining != rType.Limit-int32(i) {
			t.Fatalf("Incorrect rate left. Expected %d, got %d", rType.Limit-int32(i), left.Remaining)
		}
	}

	left, err = rType.Attempt("testID")
	if err != app.ErrTooManyRequests {
		t.Fatalf("Rate limited request didn't return an error")
	}
	if left.Remaining != -1 {
		t.Fatalf("Rate limit remaining is incorrect. Expected %d got %d", -1, left.Remaining)
	}

}

func TestRateDelay(t *testing.T) {

	rType := app.RateDelay{
		Type:   "TestRateDelay",
		Limit:  10,
		Period: 5 * time.Second,
		Delay:  2 * time.Second,
		Max:    10 * time.Second,
	}

	// free attempts
	for i := 1; i <= int(rType.Limit); i++ {
		left, err := rType.Attempt("testID")
		if err != nil {
			t.Fatalf("Error attempting rate limit withing limits. Attempt #%d rate left %d: %s",
				i, left.Remaining, err)

		}
	}

	if testing.Short() {
		t.SkipNow()
	}

	c := make(chan bool)
	go func() {
		_, _ = rType.Attempt("testID")
		c <- true
	}()

	select {
	case <-c:
		t.Fatalf("Rate was not delayed")
	case <-time.After(1 * time.Second):
	}

	max := int(rType.Max/rType.Delay) - 1 //one spent already

	// these requests will be delayed, but shouldn't error
	for i := 0; i < max; i++ {
		go func() {
			_, _ = rType.Attempt("testID")
		}()
	}

	// sleep a bit to ensure that this attempt happens last
	time.Sleep(1 * time.Millisecond)
	_, err := rType.Attempt("testID")
	if err != app.ErrTooManyRequests {
		t.Fatalf("Rate delayed request past it's max delay didn't return too many requests error: %s", err)
	}

}
