//go:build linux

package main

import (
	"fmt"
	"os"
	"syscall"
)

func getOwnerInfo(info os.FileInfo, path string) string {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return fmt.Sprintf("%d:%d", stat.Uid, stat.Gid)
	}
	return "0:0"
}
