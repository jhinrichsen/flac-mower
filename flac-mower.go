package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	var sourceDir = os.Getenv("HOME") + "/Usenext/wizard"

	fmt.Printf("Using source directory %s\n", sourceDir)
	flacDirs(sourceDir)
}

// List of directories that contain .flac files in first nested level
func flacDirs(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("Cannot list directory %s, check ${HOME} or permissions\n", dir)
	}
	// Filter directories that contain .flac files
	for _, fi := range files {
		if containsFlacFiles(fi.Name()) {
			fmt.Println(fi.Name())
		}
	}
}

func containsFlacFiles(dir string) bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Printf("Cannot list directory %s\n", dir)
	}
	var hasFlac = false
	for _, fi := range files {
		if strings.HasSuffix(fi.Name(), ".flac") {
			hasFlac = true
			break
		}
	}
	return hasFlac
}
