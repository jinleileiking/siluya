package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
)

// stupid realization

type amf0 struct {
	marker  int
	len     int
	data    []byte
	number  int
	str     string
	boolean bool
	obj     []objInner
}

type objInner struct {
	name  string
	value amf0 // just reuse it
}

func parseAMF0(buf []byte) (ret []amf0, parsedLen int) {
	for parsedLen < len(buf) {
		one, parsed := parseAMF0one(buf[parsedLen:])
		// spew.Dump(parsed)
		// spew.Dump(one, parsed)
		parsedLen += parsed
		ret = append(ret, one)
	}
	return
}

func parseAMF0one(buf []byte) (amf0, int) {
	var parsedLen int

	ret := amf0{}

	ret.marker = int(buf[0])

	if ret.marker == 2 {
		// string
		ret.len = int(binary.BigEndian.Uint16(buf[1:]))
		// spew.Dump(buf[3 : 3+ret.len])
		ret.str = string(buf[3 : 3+ret.len])
		parsedLen = 1 + 2 + ret.len
	} else if ret.marker == 0 {
		// number
		ret.number = int(math.Float64frombits(binary.BigEndian.Uint64(buf[1:])))
		parsedLen = 1 + 8
	} else if ret.marker == 1 {
		// bool
		if buf[1] == 0 {
			ret.boolean = false
		} else {
			ret.boolean = true
		}
		parsedLen = 2
	} else if ret.marker == 3 {
		// obj
		ret.obj, parsedLen = parseObj(buf[1:])
		// spew.Dump(ret.obj, parsedLen)
	} else if ret.marker == 5 {
		// null
		parsedLen = 1
		// spew.Dump(ret.obj, parsedLen)
	} else {
		logger.Errorw("marker not support")
	}
	// spew.Dump(ret, parsedLen)
	return ret, parsedLen
}

func parseObj(buf []byte) (ret []objInner, parsedLen int) {

	for parsedLen < len(buf) {
		// spew.Dump(buf[parsedLen : parsedLen+3])
		if bytes.Equal(buf[parsedLen:parsedLen+3], []byte{0x00, 0x00, 0x09}) {
			parsedLen += 4 //header 0x03 + 0 0 9
			break
		}
		inner, parsedLenInner := parseObjInner(buf[parsedLen:])
		parsedLen += parsedLenInner
		ret = append(ret, inner)
		// spew.Dump(inner)
	}
	return
}

func parseObjInner(buf []byte) (ret objInner, parsedLen int) {
	//name len
	pos := 0
	keyLen := int(binary.BigEndian.Uint16(buf))
	//name
	pos += 2
	ret.name = string(buf[pos : pos+keyLen])
	// value
	pos += keyLen
	var amf0Len int

	ret.value, amf0Len = parseAMF0one(buf[pos:])

	parsedLen = 2 + keyLen + amf0Len
	return
}

func getAmf0String(in amf0) (string, error) {
	if in.marker == 2 {
		return in.str, nil
	}
	return "", errors.New("not amf0 string")
}
