// +build windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func getSys(info os.FileInfo, path string) string {
	nInfo := new(windows.ByHandleFileInformation)
	handle, err := windows.Open(path, windows.O_RDONLY, 0)
	if err != nil {
		return "+"
	}
	defer windows.Close(handle)
	err = windows.GetFileInformationByHandle(handle, nInfo)
	if err != nil {
		return "+"
	}
	return fmt.Sprintf("%d+%d-%d", nInfo.VolumeSerialNumber, nInfo.FileIndexHigh, nInfo.FileIndexLow)
}
