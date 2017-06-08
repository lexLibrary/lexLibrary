// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"os"
	"os/user"
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
		os.Getenv("APPDATA"),
		usr.HomeDir,
	}
}

func systemLocations() []string {
	return []string{os.Getenv("APPDATA")}
}
