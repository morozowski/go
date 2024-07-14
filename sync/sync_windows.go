//go:build windows
// +build windows

package main

import (
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func detailed(attrs uint32) string {
	var buf strings.Builder
	if attrs&syscall.FILE_ATTRIBUTE_READONLY != 0 {
		buf.WriteString(" RO")
	}
	if attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0 {
		buf.WriteString(" H")
	}
	if attrs&syscall.FILE_ATTRIBUTE_SYSTEM != 0 {
		buf.WriteString(" S")
	}
	if attrs&syscall.FILE_ATTRIBUTE_DIRECTORY != 0 {
		buf.WriteString(" D")
	}
	if attrs&syscall.FILE_ATTRIBUTE_ARCHIVE != 0 {
		buf.WriteString(" A")
	}
	if attrs&syscall.FILE_ATTRIBUTE_NORMAL != 0 {
		buf.WriteString(" N")
	}
	if attrs&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		buf.WriteString(" RP")
	}
	return buf.String()
}

// https://freshman.tech/snippets/go/detect-hidden-file/
// isHidden checks if a file is hidden on Windows.
func isWindowsHiddenFile(path string) bool {
	if runtime.GOOS != "windows" {
		return false
	}

	return getFileAttributes(path)&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}

func getFileAttributes(path string) uint32 {
	absPath, err := filepath.Abs(path)
	check(err)

	// Appending `\\?\` to the absolute path helps with
	// preventing 'Path Not Specified Error' when accessing
	// long paths and filenames
	// https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation?tabs=cmd
	pointer, err := syscall.UTF16PtrFromString(`\\?\` + absPath)
	check(err)

	attributes, err := syscall.GetFileAttributes(pointer)
	check(err)

	return attributes & 0xFF
}

func setFileAttributes(path string, attributes uint32) {
	absPath, err := filepath.Abs(path)
	check(err)

	// Appending `\\?\` to the absolute path helps with
	// preventing 'Path Not Specified Error' when accessing
	// long paths and filenames
	// https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation?tabs=cmd
	pointer, err := syscall.UTF16PtrFromString(`\\?\` + absPath)
	check(err)

	err = syscall.SetFileAttributes(pointer, attributes)
	check(err)
}
