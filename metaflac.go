// For the FLAC spec, see https://xiph.org/flac/format.html

package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"strings"
)

const (
	Streaminfo = iota
	Padding
	Application
	Seektable
	VorbisComment
	Cuesheet
	Picture
)

// MetadataBlock holds header and fixed streaminfo for now
type MetadataBlockHeader struct {
	lastMetadataBlockFlag    bool
	blockType                byte
	lengthOfBlockDataInBytes uint32
}

type MetadataBlock struct {
}

type Flac struct {
	streaminfoHeader MetadataBlockHeader
	streaminfo       MetadataBlockStreaminfo
	vorbisComments   VorbisComments
}

// ErrNotAFlacFile determined from header
var ErrNotAFlacFile = errors.New("Bad header")

var ErrIncompleteRead = errors.New("Incomplete read")

func parse(r io.Reader) (flac Flac, err error) {
	if !parseHeader(r) {
		return flac, ErrNotAFlacFile
	}
	flac.streaminfoHeader, err = parseMetadataBlockHeader(r)
	log.Printf("header 1: %v\n", flac.streaminfoHeader)
	// Spec requires exactly one STREAMINFO block
	flac.streaminfo, err = parseStreaminfo(r)

	// More headers?
	h := flac.streaminfoHeader
	for !h.lastMetadataBlockFlag {
		h, err = parseMetadataBlockHeader(r)
		log.Printf("Block header: %v\n", h)
		buf := make([]byte, h.lengthOfBlockDataInBytes)
		var _, err = r.Read(buf)
		if err != nil {
			if err == io.EOF {
				// Obviously 'last' flag is not set properly
				err = nil
				break
			}
			panic(err)
		}
		// For now, only care about comments
		if h.blockType == VorbisComment {
			log.Printf("Parsing vorbis comment")
			flac.vorbisComments = parseVorbisComments(buf)
		}
	}
	return flac, err
}

func parseHeader(r io.Reader) bool {
	var buf = make([]byte, 4)
	var _, err = r.Read(buf)
	if err != nil {
		panic(err)
	}
	return "fLaC" == string(buf)
}

func parseMetadataBlockHeader(r io.Reader) (mb MetadataBlockHeader, err error) {
	var buf = make([]byte, 4)
	var n, e = r.Read(buf)
	if n != 4 {
		return mb, ErrIncompleteRead
	}
	// log.Printf("Block header[0] = %v\n", buf[0])
	mb.lastMetadataBlockFlag = buf[0]&128 != 0
	mb.blockType = buf[0] & 255
	mb.lengthOfBlockDataInBytes = uint32(buf[1])*256*256 +
		uint32(buf[2])*256 + uint32(buf[3])

	return mb, e
}

// MetadataBlockStreaminfo is 272 bit = 34 byte
type MetadataBlockStreaminfo struct {
	// Byte 0:1
	minimumBlockSizeInFrames uint16
	// Byte 2:3
	maximumBlockSizeInFrames uint16
	// Byte 4:7
	minimumFrameSizeInBytes uint32
	// Byte 8:11
	maximumFrameSizeInBytes uint32

	// Byte 12:15 bit 0:19
	sampleRateInHz uint32

	// Byte 12:15 bit 20:22 bit == Byte 14:14 bit 4:6
	numberOfChannels byte

	bitsPersample        byte
	totalSamplesInStream uint64
}

func parseStreaminfo(r io.Reader) (mbs MetadataBlockStreaminfo, err error) {
	var buf = make([]byte, 34)
	var n, e = r.Read(buf)
	if n != 34 {
		return mbs, ErrIncompleteRead
	} else if e != nil {
		return mbs, e
	}
	var bo = binary.BigEndian
	mbs.minimumBlockSizeInFrames = bo.Uint16(buf)
	mbs.maximumBlockSizeInFrames = bo.Uint16(buf[2:])
	mbs.minimumFrameSizeInBytes = bo.Uint32(buf[4:])
	mbs.maximumFrameSizeInBytes = bo.Uint32(buf[8:])
	mbs.sampleRateInHz = bo.Uint32(buf[12:]) & (20 ^ 2)
	mbs.numberOfChannels = buf[14] & (8 + 4 + 2)
	// TODO add some more members
	return
}

type VorbisComments struct {
	vendor   string
	comments map[string]string
}

// Parse buffer according to spec https://xiph.org/vorbis/doc/v-comment.html
// Note: little endian instead of FLAC's big endian standard
func parseVorbisComments(buf []byte) (vc VorbisComments) {
	var bo = binary.LittleEndian
	// 1) [vendor_length] = read an unsigned integer of 32 bits
	var n = bo.Uint32(buf)

	// 2) [vendor_string] = read a UTF-8 vector as [vendor_length] octets
	var here = 4 + n
	vc.vendor = string(buf[4:here])
	log.Printf("Vendor string: %s\n", vc.vendor)

	// 3) [user_comment_list_length] = read an unsigned integer of 32 bits
	n = bo.Uint32(buf[here:])
	here += 4

	vc.comments = make(map[string]string, n)

	// 4) iterate [user_comment_list_length] times {
	for i := uint32(0); i < n; i++ {
		// 5) [length] = read an unsigned integer of 32 bits
		var length = bo.Uint32(buf[here:])
		here += 4

		// 6) this iteration's user comment = read a UTF-8 vector as [length]
		// octets
		var s = string(buf[here : here+length])
		log.Printf("comment[%v]: length %v, value %v\n", i, length, s)
		here += length

		var kv = strings.Split(s, "=")
		// Vorbis spec suggests uppercase keys
		vc.comments[strings.ToUpper(kv[0])] = kv[1]
	}
	return
}
