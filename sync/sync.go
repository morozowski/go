package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

var sourceDir string
var targetDir string

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Parameters: <sourceDir> <targetDir>")
		return
	}

	sourceDir = args[0]
	targetDir = args[1]

	log.Println("Walking directory " + sourceDir + "...")
	err := filepath.WalkDir(sourceDir, visit)
	check(err)
}

func visit(source string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	sourceInfo, err := os.Stat(source)
	check(err)

	rel, err := filepath.Rel(sourceDir, source)
	check(err)

	target := filepath.Join(targetDir, rel)
	if d.Type().IsDir() {
		targetInfo, err := os.Stat(target)
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Creating directory %s\n", rel)
			err := os.Mkdir(target, sourceInfo.Mode())
			check(err)

			updateModTime(target, sourceInfo.ModTime())
			printFilesDiffAfterwards(rel, sourceInfo, target)
		} else {
			check(err)
			if !targetInfo.IsDir() {
				log.Fatal(target + " should be a directory!")
			}
		}
	} else {
		targetInfo, err := os.Stat(target)
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Copying file %s\n", rel)
			copyFile(source, target, sourceInfo)
			printFilesDiffAfterwards(rel, sourceInfo, target)
		} else {
			check(err)
			if !targetInfo.Mode().IsRegular() {
				log.Fatal(target + " should be a regular file!")
			}

			deltaTime := sourceInfo.ModTime().Sub(targetInfo.ModTime())
			deltaSize := sourceInfo.Size() - targetInfo.Size()
			if deltaTime != 0 || deltaSize != 0 {
				printFilesDiff("BEFORE", rel, sourceInfo, targetInfo)
				if deltaTime > 100 || deltaSize != 0 {
					log.Printf("Updating file %s\n", rel)
					copyFile(source, target, sourceInfo)
					printFilesDiffAfterwards(rel, sourceInfo, target)
				} else if deltaTime < 0 {
					log.Printf("WARNING: %s will not be updated because the target file is newer", rel)
				}
			}
		}
	}
	return nil
}

func updateModTime(filePath string, modTime time.Time) {
	var noChange time.Time
	err := os.Chtimes(filePath, noChange, modTime)
	check(err)
}

func printFilesDiffAfterwards(fileName string, sourceInfo fs.FileInfo, target string) fs.FileInfo {
	newTargetInfo, err := os.Stat(target)
	check(err)
	printFilesDiff("AFTER", fileName, sourceInfo, newTargetInfo)
	return newTargetInfo
}

func printFilesDiff(message, fileName string, sourceInfo, targetInfo fs.FileInfo) {
	log.Printf("%s\nFile: %s\nSource: %-40v %12d bytes %12v\nTarget: %-40v %12d bytes %12v\nΔT: %d",
		message,
		fileName,
		sourceInfo.ModTime(), sourceInfo.Size(), sourceInfo.Mode(),
		targetInfo.ModTime(), targetInfo.Size(), targetInfo.Mode(),
		sourceInfo.ModTime().Sub(targetInfo.ModTime()))
}

func copyFile(source string, target string, sourceInfo fs.FileInfo) {
	start := time.Now()

	sourceFile, err := os.Open(source)
	check(err)
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	check(err)
	defer targetFile.Close()

	bytes, err := io.Copy(targetFile, sourceFile)
	check(err)

	updateModTime(target, sourceInfo.ModTime())

	elapsed := float64(time.Since(start)) / 1000000.0
	log.Printf("=> %d bytes copied in %.1f ms at %.1f KB/s\n", bytes, elapsed, float64(bytes)/elapsed)
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
