// +build darwin linux

package main

import (
	"fmt"
	"os"
	"syscall"
)

func getSys(info os.FileInfo, path string) string {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "+"
	}
	return fmt.Sprintf("%d+%d", stat.Dev, stat.Ino)
}
