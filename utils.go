package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

func get2bytes(buf []byte) int {
	return int(binary.BigEndian.Uint16(buf))
}

func get3bytes(buf []byte) int {

	bufFix := []byte{0x00}
	bufFix = append(bufFix, buf...)
	// spew.Dump(bufFix)
	return int(binary.BigEndian.Uint32(bufFix) & 0xffffff)
}

func get4bytes(buf []byte) int {
	return int(binary.BigEndian.Uint32(buf))
}

// func readConn(conn net.Conn, len int) []byte {

// 	c := bufio.NewReader(conn)
// 	buf := make([]byte, len)
// 	retLen, err := io.ReadAtLeast(c, buf, len)
// 	if err != nil {
// 		logger.Errorw("Error reading:", "error", err.Error())
// 	}

// 	if retLen != len {
// 		logger.Errorw("Error reading len not equal", "len", len, "retlen", retLen)
// 	}

// 	return buf
// }

func readConn(rw *bufio.ReadWriter, len int) ([]byte, error) {

	var e error

	var tail int
	var ret []byte

	tail = len
	for tail != 0 {
		buf := make([]byte, tail)
		retLen, err := rw.Read(buf)
		if err != nil {
			logger.Errorw("conn", "error", err.Error())
			e = errors.New("rw Read error")
			return ret, e
		}

		if retLen != tail {
			// logger.Infow("conn", "detail",
			// 	fmt.Sprintf("Error reading part len not equal, len:%d,  ret:%d, buf:%v", len, retLen, buf))
			// logger.Infow("conn", "detail",
			// 	fmt.Sprintf("Error reading part len not equal, len:%d,  ret:%d, continue...", len, retLen))
			// e = errors.New("rw Read len error")
		}

		ret = append(ret, buf...)
		tail = tail - retLen
	}
	return ret, e
}

func readConnPart(rw bufio.ReadWriter, len int) []byte {

	buf := make([]byte, len)
	retLen, err := rw.Read(buf)
	if err != nil {
		logger.Errorw("conn", "error", err.Error())
	}

	if retLen != len {
		logger.Errorw("conn", "detail",
			fmt.Sprintf("Error reading part len not equal, len:%d,  ret:%d, buf:%v", len, retLen, buf))
	}

	return buf
}

// func readConnPart(conn net.Conn, len int) []byte {

// 	buf := make([]byte, len)
// 	retLen, err := conn.Read(buf)
// 	if err != nil {
// 		logger.Errorw("Error reading part:", "error", err.Error())
// 	}

// 	if retLen != len {
// 		logger.Errorw("Error reading part len not equal", "len", len, "retlen", retLen, "buf", buf)
// 	}

// 	return buf
// }

func writeConn(conn net.Conn, buf []byte) {
	retLen, err := conn.Write(buf)
	if err != nil {
		logger.Infow("Error reading:", err.Error())
	}

	if retLen != len(buf) {
		logger.Error("Error writing len != retLen")
	}

	logger.Infow("session", "[S]write len", len(buf))
	return
}
