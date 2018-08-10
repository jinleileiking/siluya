package main

import (
	f "fmt"
	"net"

	"github.com/davecgh/go-spew/spew"
)

type rtmpConn struct {
	conn   net.Conn
	cksize int
}

func handshake(conn net.Conn) {

	// C0+C1
	data := readConn(conn, 1537)

	// spew.Dump(data)
	// C0
	pos := 0
	if int(data[0]) != 3 {
		logger.Panicw("version not  3")
	}
	pos++

	// C1
	// time
	t := get4bytes(data[pos:])
	pos += 4
	// zero
	pos += 4
	random := data[pos:]

	// spew.Dump(random)
	// spew.Dump(string(random))

	logger.Infow("handshake", "event", "start", "time", t, "random", random)

	// send s0 s1
	writeConn(conn, data)

	//TODO: must wait?
	// here, you must send s2, then receive c2

	// time.Sleep(2 * time.Second)
	writeConn(conn, data[1:])
	// time.Sleep(2 * time.Second)
	// spew.Dump(readConn(conn, 1536))

	readConn(conn, 1536)
	logger.Infow("handshake", "event", "done")

}

func waitMsg(conn rtmpConn) {

	// buf := make([]byte, 1)
	// _, err := conn.Read(buf)
	// if err != nil {
	// 	logger.Infow("Error reading:", err.Error())
	// }

	logger.Infow("chunk", "detail", "start receiving basic header")
	buf := readConnPart(conn.conn, 1)
	fmt := int(buf[0]) & 0xc0

	var csid int
	if fmt == 0 {
		csid = int(buf[0]) & 0x3f
	} else if fmt == 1 {
		logger.Infow("chunk", "detail", "fmt 1")
	} else if fmt == 2 {
		logger.Infow("chunk", "detail", "fmt 2")
	} else if fmt == 3 {
		logger.Infow("chunk", "detail", "fmt 3")
	} else {
		logger.Infow("chunk", "detail", "fmt not 1 2 3")
	}

	// spew.Dump(csid, fmt)

	logger.Infow("chunk", "csid", csid, "fmt", fmt)

	var msgLen, msgTypeId int

	if fmt == 0 {
		msgLen, msgTypeId = getChunkMsgHdr(conn.conn)
	}

	chunkData := getChunkData(conn.conn, msgLen, conn.cksize, csid)

	spew.Dump(chunkData)
	// os.Exit(0)

	var cmdMsg []amf0
	switch msgTypeId {
	case 20: //cmd
		cmdMsg, _ = parseAMF0(chunkData)
		// spew.Dump(cmdMsg)
		handleCmdMsg(&conn, cmdMsg)
	case 1: //set chunk size
		logger.Infow("message", "detail", "receive set chunk size")
	default:
		logger.Infow("message", "detail", f.Sprintf("not support msgtypeid: %d", msgTypeId))
	}

	//TODO: extended timestamp
}

func getChunkData(conn net.Conn, msgLen int, cksize int, csid int) []byte {

	var rawData []byte

	tail := msgLen
	for cnt := (msgLen / cksize) + 1; cnt > 0; cnt-- {
		var buf []byte
		if cnt == 1 {
			buf = readConn(conn, tail)
			rawData = append(rawData, buf...)
			break
		} else {
			buf = readConnPart(conn, cksize)
			// spew.Dump(buf)
			// read type3 header
			//TODO:
			readConnPart(conn, 1)
			// buf = readConnPart(conn, 1)
			// spew.Dump(buf)
			//TODO: check csid fmt
		}
		rawData = append(rawData, buf...)
		tail = tail - cksize
	}

	// spew.Dump(rawData)

	return rawData
}

func getChunkMsgHdr(conn net.Conn) (int, int) {

	buf := readConnPart(conn, 11)

	pos := 0

	// spew.Dump(buf)

	ts := get3bytes(buf)
	pos += 3

	// spew.Dump(ts)

	msgLen := get3bytes(buf[pos:])
	// spew.Dump(buf[pos : pos+3])
	pos += 3
	// spew.Dump(msgLen)

	msgTypeId := buf[pos]
	pos++

	msgStreamId := get4bytes(buf[pos:])
	pos++

	logger.Infow("chunk", "ts:", ts, "msgLen", msgLen, "msgTypeId:", msgTypeId, "msgStreamId:", msgStreamId)
	// spew.Dump("ts:", ts, "msgLen", msgLen, "msgTypeId:", msgTypeId, "msgStreamId:", msgStreamId)

	return msgLen, int(msgTypeId)
}

func handleCmdMsg(conn *rtmpConn, cmdMsg []amf0) {

	/*
	                                  02 00 00 00 00 00 05 06   ..G.D.¾Y........
	   0040   00 00 00 00 00 4c 4b 40 02|02 00 00 00 00 00 04   .....LK@........
	   0050   01 00 00 00 00 00 00 0f a0|03 00 00 00 00 00 be   ........ ......¾
	   0060   14 00 00 00 00 02 00 07 5f 72 65 73 75 6c 74 00   ........_result.
	   0070   3f f0 00 00 00 00 00 00 03 00 06 66 6d 73 56 65   ?ð.........fmsVe
	   0080   72 02 00 0d 46 4d 53 2f 33 2c 30 2c 31 2c 31 32   r...FMS/3,0,1,12
	   0090   33 00 0c 63 61 70 61 62 69 6c 69 74 69 65 73 00   3..capabilities.
	   00a0   40 3f 00 00 00 00 00 00 00 00 09 03 00 05 6c 65   @?............le
	   00b0   76 65 6c 02 00 06 73 74 61 74 75 73 00 04 63 6f   vel...status..co
	   00c0   64 65 02 00 1d 4e 65 74 43 6f 6e 6e 65 63 74 69   de...NetConnecti
	   00d0   6f 6e 2e 43 6f 6e 6e 65 63 74 2e 53 75 63 63 65   on.Connect.Succe
	   00e0   73 73 00 0b 64 65 73 63 72 69 70 74 69 6f 6e 02   ss..description.
	   00f0   00 15 43 6f 6e 6e 65 63 74 69 6f 6e 20 73 75 63   ..Connection suc
	   0100   63 65 65 64 65 64 2e 00 0e 6f 62 6a 65 63 74 45   ceeded...objectE
	   0110   6e 63 6f 64 69 6e 67 00 00 00 00 00 00 00 00 00   ncoding.........
	   0120   00 00 09                                          ...
	*/

	msg, _ := getAmf0String(cmdMsg[0])
	if "connect" == msg {
		connectRsp := []byte{

			// 0x02,
			// 0x00, 0x00, 0x00,
			// 0x00, 0x00, 0x04,
			// 0x05,
			// 0x00, 0x00, 0x00, 0x00,
			// 0x00, 0x4c, 0x4b, 0x40, //  50 0000 ack size

			// 0x02,
			// 0x00, 0x00, 0x00,
			// 0x00, 0x00, 0x05,
			// 0x06,
			// 0x00, 0x00, 0x00, 0x00,
			// 0x00, 0x4c, 0x4b, 0x40, // ack size  50 0000
			// 0x02, //limit 2

			0x02,
			0x00, 0x00, 0x00,
			0x00, 0x00, 0x04,
			0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x0f, 0xa0, // set chunk size 4000

			0x03,
			0x00, 0x00, 0x00,
			0x00, 0x00, 0xbe,
			0x14, //20
			0x00, 0x00, 0x00, 0x00,
			0x02, 0x00, 0x07, 0x5f, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, //_result
			0x00, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // trans id
			0x03,
			0x00, 0x06, 0x66, 0x6d, 0x73, 0x56, 0x65, 0x72, // fmsver
			0x02, 0x00, 0x0d,
			0x46, 0x4d, 0x53, 0x2f, 0x33, 0x2c, 0x30, 0x2c,
			0x31, 0x2c, 0x31, 0x32, 0x33,
			0x00, 0x0c,
			0x63, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69,
			0x74, 0x69, 0x65, 0x73,
			0x00, 0x40,
			0x3f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x09, 0x03, 0x00, 0x05, 0x6c, 0x65, 0x76,
			0x65, 0x6c, 0x02, 0x00, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x00, 0x04, 0x63, 0x6f,
			0x64, 0x65, 0x02, 0x00, 0x1d, 0x4e, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69,
			0x6f, 0x6e, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x53, 0x75, 0x63, 0x63, 0x65,
			0x73, 0x73, 0x00, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x02,
			0x00, 0x15, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x20, 0x73, 0x75, 0x63,
			0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x2e, 0x00, 0x0e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x45,
			0x6e, 0x63, 0x6f, 0x64, 0x69, 0x6e, 0x67, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x09,
		}
		logger.Infow("connect", "detail", "send connect rsp done")
		conn.cksize = 4000
		writeConn(conn.conn, connectRsp)
	}
}
