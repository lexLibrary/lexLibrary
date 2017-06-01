// Copyright (c) 2017 Townsourced Inc.

package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

//Tries to adhere to the xdg base directory specification http://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html

func userLocations() []string {
	location := os.Getenv("XDG_CONFIG_HOME")
	if location != "" {
		return []string{location}
	}
	usr, err := user.Current()
	if err != nil {
		return []string{}
	}

	return []string{filepath.Join(usr.HomeDir, ".config")}
}

func systemLocations() []string {
	defaults := []string{"/etc/xdg", "/etc"}
	envLocations := os.Getenv("XDG_CONFIG_DIRS")
	if envLocations == "" {
		return defaults
	}
	locations := strings.Split(envLocations, ":")

	for d := range defaults {
		found := false
		for l := range locations {
			locations[l] = filepath.Clean(locations[l])
			if locations[l] == defaults[d] {
				found = true
				break
			}
		}
		if !found {
			locations = append(locations, defaults[d])
		}
	}

	return locations
}
