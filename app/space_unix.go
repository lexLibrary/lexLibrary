// +build darwin dragonfly freebsd linux netbsd openbsd

// Copyright (c) 2017-2018 Townsourced Inc.
package app

import (
	"syscall"
)

func diskSpace(path string) (total, free int, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		return
	}
	total = int(s.Bsize) * int(s.Blocks)
	free = int(s.Bsize) * int(s.Bfree)
	return
}
