// Copyright (c) 2017-2018 Townsourced Inc.
package web_test

import (
	"testing"
)

func TestProfile(t *testing.T) {
	uri := *llURL
	uri.Path = "profile"

	err := signupUser("testUser", "testpasswordThatisLongEnough", false)
	if err != nil {
		t.Fatalf("Error setting up user for testing: %s", err)
	}

	//TODO:
}
