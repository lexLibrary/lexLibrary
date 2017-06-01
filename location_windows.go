// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"os"
	"os/user"
)

func userLocations() []string {
	usr, err := user.Current()
	if err != nil {
		return []string{}
	}

	return []string{
		os.Getenv("APPDATA"),
		usr.HomeDir,
	}
}

func systemLocations() []string {
	return []string{os.Getenv("APPDATA")}
}
