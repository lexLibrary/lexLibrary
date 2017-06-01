// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"os/user"
	"path/filepath"
)

func userLocations() []string {
	usr, err := user.Current()
	if err != nil {
		return []string{}
	}

	return []string{
		filepath.Join(usr.HomeDir, ".config"),
		filepath.Join(usr.HomeDir, "Library/Application Support"),
	}
}

func systemLocations() []string {
	return []string{"/Library/Application Support", "/etc"}
}
