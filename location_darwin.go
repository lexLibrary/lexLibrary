// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"os/user"
	"path/filepath"
)

func userCFGLocations() []string {
	return userLocations()
}

func systemCFGLocations() []string {
	return systemLocations()
}

func userDataLocations() []string {
	return userLocations()
}

func systemDataLocations() []string {
	return systemLocations()
}

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
