// Copyright (c) 2017-2018 Townsourced Inc.

package app

import (
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type sysInfo struct {
	Uptime     time.Duration
	Totalram   uint64
	Freeram    uint64
	Sharedram  uint64
	Totalswap  uint64
	Freeswap   uint64
	TotalSpace uint64
	FreeSpace  uint64
}

func systemInfo() sysInfo {
	result := sysInfo{}

	info := &syscall.Sysinfo_t{}

	err := syscall.Sysinfo(info)
	if err != nil {
		return result
	}
	result.Uptime = time.Duration(info.Uptime) * time.Second
	result.Totalram = info.Totalram
	result.Freeram = info.Freeram
	result.Sharedram = info.Sharedram
	result.Totalswap = info.Totalswap
	result.Freeswap = info.Freeswap

	cwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return result
	}

	total, free, err := diskSpace(cwd)
	if err != nil {
		return result
	}
	result.TotalSpace = uint64(total)
	result.FreeSpace = uint64(free)
	return result
}
