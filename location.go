// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"os"
	"path/filepath"
)

func configLocations(appName string) []string {
	return getFullLocations(appName, append(userCFGLocations(), systemCFGLocations()...))
}

func dataLocations(appName string) []string {
	return getFullLocations(appName, append(userDataLocations(), systemDataLocations()...))
}

//getFullLocations builds an OS specific list of standard file locations
// for where a config/ data file should be loaded from.
// Generally follows this priority list:
// 1. User locations are used before...
// 2. System locations which are used before ...
// 3. The imediate running directory of the application
// The result set will be joined with the passed in filepath.  Passing in
// a filepath with a leading directory is encouraged to keep your config folders
// clean.
//
// For example a filepath of myApp/config.json might return the following on linux
// 	"/home/user/.config/myApp/config.json",
//	"/etc/xdg/myApp/config.json",
//	"./config.json"
// Note that parent folder paths (myApp in this example) are stripped for the first eligible file location
// so the config file will exist in the same directory as the running executable
func getFullLocations(name string, locations []string) []string {
	for i := range locations {
		if locations[i] != "" {
			locations[i] = filepath.Join(locations[i], name)
		}
	}

	//get running dir
	runningDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		runningDir = "."
	}

	locations = append(locations, runningDir)
	return locations
}
