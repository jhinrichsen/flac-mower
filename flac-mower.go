package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var sourceDir = os.Getenv("HOME") + "/Usenext/wizard"

	fmt.Printf("Using source directory %s\n", sourceDir)
	var fds = flacDirs(sourceDir)
	for _, fd := range fds {
		fmt.Printf("Flac directory %s\n", fd)
	}
}

// List of directories that contain .flac files in first nested level
func flacDirs(basedir string) []string {
	// Filter directories that contain .flac files
	var pattern = fmt.Sprintf("%s/**/*.flac", basedir)
	matches, err := filepath.Glob(pattern)
	// Glob ignores IO errors, error means pattern is bad
	if err != nil {
		log.Fatalf("Bad wildcard %s\n", pattern)
	}
	var flacs = filter(matches, isFlacContent)
	var uniqueDirs = make(map[string]bool)
	for _, f := range flacs {
		var parent = filepath.Dir(f)
		uniqueDirs[parent] = true
	}
	return keys(uniqueDirs)
}

// Elaborate man's flac detection
func isFlacContent(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		log.Printf("Cannot read from %s, ignoring", filename)
		return false
	}
	buf := make([]byte, 4)
	n, err := f.Read(buf)
	if err != nil {
		log.Printf("Ignoring internal error %v\n", err)
		return false
	}
	if n != 4 {
		log.Printf("Cannot read from %s\n", filename)
		return false
	}
	var b = isFlacPrefix(buf)
	// log.Printf("File %s has flac content: %v\n", filename, b)
	return b
}

// Poor man's flac detection
func isFlacFilename(filename string) bool {
	return strings.HasSuffix(filename, ".flac")
}

func isFlacPrefix(buf []byte) bool {
	var expected = "fLaC"
	// log.Printf("Comparing expected %s against %s: ", expected, string(buf))
	var matches = 0 == bytes.Compare([]byte(expected), buf)
	// log.Printf("matches = %v\n", matches)
	return matches
}

func filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func keys(s map[string]bool) []string {
	keys := make([]string, len(s))

	i := 0
	for k := range s {
		keys[i] = k
		i++
	}
	return keys
}
