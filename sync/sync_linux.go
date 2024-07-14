//go:build !windows
// +build !windows

package main

func detailed(attrs uint32) string {
	return ""
}

func isWindowsHiddenFile(path string) bool {
	return false
}

func getFileAttributes(path string) uint32 {
	return 0
}

func setFileAttributes(path string, attributes uint32) {
}
