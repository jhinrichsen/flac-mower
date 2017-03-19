package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	var sourceDir = os.Getenv("HOME") + "/Usenext/wizard"

	fmt.Printf("Using source directory %s\n", sourceDir)
	var flacs = findFlacFiles(sourceDir)
	sort.Strings(flacs)
	for _, d := range uniqueDirs(flacs) {
		// Only use flacs from current directory. While this is not strictly
		// necessary, it adds some security (same metadata for complete
		// folder e.a.)
		var fs = filter(flacs, func(filename string) bool {
			return strings.HasPrefix(filename, d)
		})
		fmt.Println(fs)
		break
	}
}

// Beginning from basedir, recursively find all .flac files
func findFlacFiles(basedir string) []string {
	var pattern = fmt.Sprintf("%s/**/*.flac", basedir)
	matches, err := filepath.Glob(pattern)
	// Glob ignores IO errors, error means pattern is bad
	if err != nil {
		log.Fatalf("Bad wildcard %s\n", pattern)
	}
	return matches
}

func filterFlacContent(filenames []string) []string {
	return filter(filenames, isFlacContent)
}

func uniqueDirs(dirs []string) []string {
	var u = make(map[string]bool)
	for _, f := range dirs {
		var parent = filepath.Dir(f)
		u[parent] = true
	}
	return keys(u)
}

// Poor man's flac detection
func isFlacFilename(filename string) bool {
	return strings.HasSuffix(filename, ".flac")
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
