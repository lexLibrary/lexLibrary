// Copyright (c) 2017-2018 Townsourced Inc.
package app

import (
	"runtime"
	"strings"
	"time"

	"github.com/lexLibrary/lexLibrary/files"
	"github.com/pkg/errors"
)

var (
	version   = "unset"
	buildDate = time.Time{}
)

type RuntimeInfo struct {
	OS       string
	GoVer    string
	Arch     string
	Compiler string
	MaxProcs int
	NumCPU   int
}

var runtimeInfo = &RuntimeInfo{
	OS:       runtime.GOOS,
	GoVer:    runtime.Version(),
	Arch:     runtime.GOARCH,
	Compiler: runtime.Compiler,
	MaxProcs: runtime.GOMAXPROCS(-1),
	NumCPU:   runtime.NumCPU(),
}

func loadVersion() error {
	version = "unset"
	buildDate = time.Time{}

	b, err := files.Asset("version")
	if err != nil {
		return nil
	}

	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 {
		return nil
	}

	version = strings.TrimSpace(lines[0])
	buildDate, err = time.Parse(time.UnixDate, strings.TrimSpace(lines[1]))
	if err != nil {
		return errors.Wrap(err, "Parsing last modified date")
	}

	return nil
}

// Version returns the current app version
func Version() string {
	return version
}

func BuildDate() time.Time {
	return buildDate
}

// RuntimeInfo returns information about the currently running instance of Lex Library
func Runtime(who *User) *RuntimeInfo {
	if SettingMust("AllowRuntimeInfoInIssues").Bool() {
		return runtimeInfo
	}

	if who != nil && who.Admin {
		return runtimeInfo
	}
	return nil
}
