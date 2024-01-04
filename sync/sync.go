package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var sourceDir string
var targetDir string
var debug bool = true

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Parameters: <sourceDir> <targetDir>")
		return
	}

	sourceDir = strings.ReplaceAll(args[0], "\\", "/")
	targetDir = strings.ReplaceAll(args[1], "\\", "/")

	abs, err := filepath.Abs(sourceDir)
	check(err)

	log.Println("Walking directory " + abs + "...")
	// err = filepath.WalkDir(sourceDir, print)
	err = filepath.WalkDir(sourceDir, sync)
	check(err)
}

type fileDetails struct {
	Info  fs.FileInfo
	Attrs uint32
}

func getDetails(path string) *fileDetails {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	check(err)

	attrs := getFileAttributes(path)

	return &fileDetails{
		Info:  info,
		Attrs: attrs,
	}
}

func print(source string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(sourceDir, source)
	check(err)
	target := filepath.Join(targetDir, rel)

	sourceDetails := getDetails(source)
	targetDetails := getDetails(target)

	if targetDetails != nil {
		printFilesDiff("", rel, *sourceDetails, *targetDetails)
	}

	return nil
}

func sync(source string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	sourceDetails := getDetails(source)

	rel, err := filepath.Rel(sourceDir, source)
	check(err)

	target := filepath.Join(targetDir, rel)
	targetDetails := getDetails(target)
	newTargetDetails := targetDetails
	if d.Type().IsDir() {
		if debug {
			log.Printf("Walking directory %s...\n", rel)
		}
		if targetDetails == nil {
			log.Printf("Creating directory %s\n", rel)
			err := os.Mkdir(target, sourceDetails.Info.Mode())
			check(err)

			updateModTime(target, sourceDetails.Info.ModTime())
			newTargetDetails = printFilesDiffAfterwards(rel, *sourceDetails, target)
		} else {
			if !targetDetails.Info.IsDir() {
				log.Fatal(target + " should be a directory!")
			} else { // target directory is newer: sync the modification date
				deltaTime := sourceDetails.Info.ModTime().Sub(targetDetails.Info.ModTime())
				if deltaTime > 100 || deltaTime <= -1000000000 {
					printFilesDiff("BEFORE", rel, *sourceDetails, *targetDetails)
					log.Printf("Updating modification time of directory %s\n", rel)
					updateModTime(target, sourceDetails.Info.ModTime())
					newTargetDetails = printFilesDiffAfterwards(rel, *sourceDetails, target)
				}
			}
		}
	} else {
		if targetDetails == nil {
			log.Printf("Copying file %s\n", rel)
			syncFile(source, target, sourceDetails.Info)
			newTargetDetails = printFilesDiffAfterwards(rel, *sourceDetails, target)
		} else {
			check(err)
			if !targetDetails.Info.Mode().IsRegular() {
				log.Fatal(target + " should be a regular file!")
			}

			deltaTime := sourceDetails.Info.ModTime().Sub(targetDetails.Info.ModTime())
			deltaSize := sourceDetails.Info.Size() - targetDetails.Info.Size()
			if (deltaTime > 100 || deltaTime <= -1000000000) || deltaSize != 0 || sourceDetails.Attrs != targetDetails.Attrs {
				printFilesDiff("BEFORE", rel, *sourceDetails, *targetDetails)
				if deltaTime > 100 || deltaSize != 0 {
					log.Printf("Updating file %s\n", rel)
					if isWindowsHiddenFile(target) {
						log.Println("Hidden file cannot be updated in Windows: removing it...")
						os.Remove(target)
					}
					syncFile(source, target, sourceDetails.Info)
					newTargetDetails = printFilesDiffAfterwards(rel, *sourceDetails, target)
				} else if deltaTime <= -1000000000 { // same size, target file is newer: sync the modification date
					log.Printf("Updating modification time of file %s\n", rel)
					updateModTime(target, sourceDetails.Info.ModTime())
					newTargetDetails = printFilesDiffAfterwards(rel, *sourceDetails, target)
				}
			}
		}
	}
	if sourceDetails.Attrs != newTargetDetails.Attrs {
		printFilesDiff("BEFORE", rel, *sourceDetails, *newTargetDetails)
		log.Printf("Updating attributes of file %s\n", rel)
		setFileAttributes(target, sourceDetails.Attrs)
		_ = printFilesDiffAfterwards(rel, *sourceDetails, target)
	}
	return nil
}

func updateModTime(filePath string, modTime time.Time) {
	var noChange time.Time
	err := os.Chtimes(filePath, noChange, modTime)
	check(err)
}

func printFilesDiffAfterwards(fileName string, sourceDetails fileDetails, target string) *fileDetails {
	targetDetails := getDetails(target)
	printFilesDiff("AFTER", fileName, sourceDetails, *targetDetails)
	return targetDetails
}

func printFilesDiff(message, fileName string, sourceDetails, targetDetails fileDetails) {
	if debug {
		sourceMode := sourceDetails.Info.Mode()
		targetMode := targetDetails.Info.Mode()
		log.Printf("%s\nFile: %s\nSource: %-40v %12d bytes %12v %16b %s\nTarget: %-40v %12d bytes %12v %16b %s\nΔT: %d",
			message,
			fileName,
			sourceDetails.Info.ModTime(), sourceDetails.Info.Size(), sourceMode, uint32(sourceDetails.Attrs), detailed(sourceDetails.Attrs),
			targetDetails.Info.ModTime(), targetDetails.Info.Size(), targetMode, uint32(targetDetails.Attrs), detailed(targetDetails.Attrs),
			sourceDetails.Info.ModTime().Sub(targetDetails.Info.ModTime()))
	}
}

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

func syncFile(source string, target string, sourceInfo fs.FileInfo) {
	start := time.Now()

	bytes := copyFile(source, target, sourceInfo)
	updateModTime(target, sourceInfo.ModTime())

	elapsed := float64(time.Since(start)) / 1000000.0
	log.Printf("=> %d bytes copied in %.1f ms at %.1f KB/s\n", bytes, elapsed, float64(bytes)/elapsed)
}

func copyFile(source, target string, sourceInfo fs.FileInfo) int64 {
	sourceFile, err := os.Open(source)
	check(err)
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	check(err)
	defer targetFile.Close()

	bytes, err := io.Copy(targetFile, sourceFile)
	check(err)
	return bytes
}

func check(e error) {
	if e != nil {
		panic(e)
		// log.Fatal(e)
	}
}
