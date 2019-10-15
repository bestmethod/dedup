// +build windows

package main

import "os"

func getSys(info os.FileInfo) string {
	return "+"
}
