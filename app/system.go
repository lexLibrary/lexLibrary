// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"os"
	"path/filepath"
	"syscall"
	"time"

	humanize "github.com/dustin/go-humanize"
)

type sysInfo struct {
	Uptime     string
	Totalram   string
	Freeram    string
	Sharedram  string
	Totalswap  string
	Freeswap   string
	TotalSpace string
	FreeSpace  string
}

func systemInfo() sysInfo {
	result := sysInfo{}

	info := &syscall.Sysinfo_t{}

	err := syscall.Sysinfo(info)
	if err != nil {
		return result
	}
	result.Uptime = humanize.RelTime(time.Now().Add(time.Duration(info.Uptime)*-1*time.Second), time.Now(), "", "")
	result.Totalram = humanize.Bytes(info.Totalram)
	result.Freeram = humanize.Bytes(info.Freeram)
	result.Sharedram = humanize.Bytes(info.Sharedram)
	result.Totalswap = humanize.Bytes(info.Totalswap)
	result.Freeswap = humanize.Bytes(info.Freeswap)

	cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return result
	}

	total, free, err := diskSpace(cwd)
	if err != nil {
		return result
	}
	result.TotalSpace = humanize.Bytes(uint64(total))
	result.FreeSpace = humanize.Bytes(uint64(free))
	return result
}
