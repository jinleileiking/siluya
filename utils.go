package main

import (
	"encoding/binary"
	"net"
)

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

func readConn(conn net.Conn, len int) []byte {

	buf := make([]byte, len)
	retLen, err := conn.Read(buf)
	if err != nil {
		logger.Errorw("Error reading part:", "error", err.Error())
	}

	if retLen != len {
		logger.Errorw("Error reading part len not equal", "len", len, "retlen", retLen, "buf", buf)
	}

	return buf
}
func readConnPart(conn net.Conn, len int) []byte {

	buf := make([]byte, len)
	retLen, err := conn.Read(buf)
	if err != nil {
		logger.Errorw("Error reading part:", "error", err.Error())
	}

	if retLen != len {
		logger.Errorw("Error reading part len not equal", "len", len, "retlen", retLen, "buf", buf)
	}

	return buf
}

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
