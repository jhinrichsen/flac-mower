package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestIsFlac(t *testing.T) {
	var buf = []byte("fLaCdata")
	var r = bytes.NewReader(buf)
	if !parseHeader(r) {
		t.Fail()
	}
}
func TestIsNotFlac(t *testing.T) {
	var buf = []byte("ffLaCdata")
	var r = bytes.NewReader(buf)
	if parseHeader(r) {
		t.Fail()
	}
}

func TestFlac(t *testing.T) {
	var buf, err = ioutil.ReadFile("test.flac")
	if err != nil {
		t.Errorf("Cannot locate file 'test.flac'")
	}
	var r = bytes.NewReader(buf)
	var flac, e = Parse(r)
	if e != nil {
		t.Fail()
	}
	fmt.Printf("flac: %v\n", flac)
}
