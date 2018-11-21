package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/gwuhaolin/livego/utils/pio"
)

type rtmpConn struct {
	conn    net.Conn
	cksize  int
	csid    int
	mLen    map[int]int
	mTypeId map[int]int
	rw      *bufio.ReadWriter
	app     string
	name    string
}

func handshake(conn rtmpConn) error {

	// C0+C1
	var err error
	var data []byte
	data, err = readConn(conn.rw, 1537)

	if err != nil {
		return errors.New("read error")
	}
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
	writeConn(conn.conn, data)

	writeConn(conn.conn, data[1:])
	// spew.Dump(readConn(conn, 1536))

	data, err = readConn(conn.rw, 1536)
	if err != nil {
		return errors.New("read error")
	}
	logger.Infow("handshake", "event", "done")

	return nil
}

func waitMsg(conn *rtmpConn) error {

	// logger.Infow("chunk", "detail", "start receiving basic header")
	buf, err := readConn(conn.rw, 1)

	if err != nil {
		return err
	}
	spew.Dump(buf)

	var csid, format int
	lower6bits := (int(buf[0]) & 0x3f)
	format = (int(buf[0]) & 0xc0) >> 6

	switch lower6bits {
	case 0:
		buf, _ := readConn(conn.rw, 1)
		csid = int(buf[0]) + 64
	case 1:
		buf, _ := readConn(conn.rw, 2)
		csid = get2bytes(buf) + 64
	default:
		csid = lower6bits
	}

	conn.csid = csid
	if format > 3 {
		logger.Errorw("chunk", "detail", fmt.Sprintf("format not 1 2 3, format: %d", format))
		return errors.New("format not 1 2 3")
	}

	logger.Infow("chunk", "csid", csid, "format", format)

	var msgLen, msgTypeId int

	// if format == 1 && csid != conn.csid {
	// 	logger.Error("chunk", "detail", fmt.Sprintf("format 3 csid not equal csid:%d, conn.csid:%d", csid, conn.csid))
	// 	return fmt.Errorf("format 3 csid not equal csid:%d, conn.csid:%d", csid, conn.csid)
	// }
	// if format == 1 {
	// 	logger.Error("chunk", "detail", fmt.Sprintf("format 3 csid not equal csid:%d, conn.csid:%d", csid, conn.csid))
	// 	return fmt.Errorf("format 3 csid not equal csid:%d, conn.csid:%d", csid, conn.csid)
	// }

	msgLen, msgTypeId = getChunkMsgHdr(conn, format)

	chunkData := getChunkData(*conn, msgLen, conn.cksize, csid)

	if len(chunkData) < 1000 {
		spew.Dump(chunkData)
	}

	var cmdMsg []amf0
	switch msgTypeId {
	case 20: //cmd
		cmdMsg, _ = parseAMF0(chunkData)
		// spew.Dump(cmdMsg)
		handleCmdMsg(conn, cmdMsg)
	case 1: //set chunk size
		logger.Infow("message", "type", "receive set chunk size")
	case 18: //onmetadata
		logger.Infow("message", "type", "receive data")
		// metadata, _ := parseAMF0(chunkData)
		// spew.Dump(metadata)
		// os.Exit(0)
	case 9:
		logger.Infow("message", "type", "receive video")
		processVideo(conn, chunkData)
	case 8:
		logger.Infow("message", "type", "receive audio")
	default:
		logger.Infow("message", "type", fmt.Sprintf("not support msgtypeid: %d", msgTypeId))
		os.Exit(0)
	}

	//TODO: extended timestamp
	return nil
}

func getChunkData(conn rtmpConn, msgLen int, cksize int, csid int) []byte {

	var rawData []byte

	tail := msgLen
	for cnt := (msgLen / cksize) + 1; cnt > 0; cnt-- {
		var buf []byte
		if cnt == 1 {
			buf, _ = readConn(conn.rw, tail)
			rawData = append(rawData, buf...)
			break
		} else {
			buf, _ = readConn(conn.rw, cksize)
			// spew.Dump(buf)
			// read type3 header
			//TODO:
			buf1, _ := readConn(conn.rw, 1)
			spew.Dump(buf1)
			// spew.Dump("------chunk continue--------", buf1)
			// buf = readConnPart(conn, 1)
			// spew.Dump(buf)
			//TODO: check csid format
		}
		rawData = append(rawData, buf...)
		tail = tail - cksize
	}

	// spew.Dump(rawData)

	return rawData
}

func getChunkMsgHdr(conn *rtmpConn, format int) (int, int) {

	var buf []byte

	if format == 3 {
		logger.Infow("chunk", "detail", "all use last chunk")
		return conn.mLen[conn.csid], conn.mTypeId[conn.csid]
	}

	buf, _ = readConn(conn.rw, 3)

	// ts delta
	ts := get3bytes(buf)

	if format == 2 {
		// logger.Infow("chunk", "ts:", ts, "msgLen", "use last chunk",
		// 	"msgTypeId:", "use last chunk", "msgStreamId:", "use last chunk")
		logger.Infow("chunk", "ts:", ts)
		return conn.mLen[conn.csid], conn.mTypeId[conn.csid]
	}

	// mlen
	buf, _ = readConn(conn.rw, 3)

	msgLen := get3bytes(buf)

	conn.mLen[conn.csid] = msgLen

	// mtype
	buf, _ = readConn(conn.rw, 1)

	msgTypeId := buf[0]
	conn.mTypeId[conn.csid] = int(msgTypeId)

	if format == 0 {
		buf, _ = readConn(conn.rw, 4)
		msgStreamId := get4bytes(buf)
		logger.Infow("chunk", "ts:", ts, "msgLen", msgLen, "msgTypeId:", msgTypeId, "msgStreamId:", msgStreamId)

	}
	if format == 1 {
		// logger.Infow("chunk", "ts:", ts, "msgLen", msgLen, "msgTypeId:", msgTypeId, "msgStreamId", "use last chunck")
		logger.Infow("chunk", "ts:", ts, "msgLen", msgLen, "msgTypeId:", msgTypeId)
	}

	// case 3:
	// 	if csid != conn.csid {
	// 		logger.Error("chunk", fmt.Sprintf("format 3 csid not equal csid:%d, conn.csid:%d", csid, conn.csid))
	// 	}

	// default:
	// 	logger.Error("chunk", "detail", "format not 0,1,2,3")
	// }

	info.Set(spew.Sdump(conn.mLen, conn.mTypeId))
	return msgLen, int(msgTypeId)
}

type chunkMsgHdr struct {
	format   int
	ts       int
	mLen     int
	mTypeId  int
	msIdtype int
}

func mkChunkMsgHdr(csid uint32, timestamp int32, msgtypeid uint8, msgsid uint32, msgdatalen int) []byte {
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |                   timestamp                   |message length |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |     message length (cont)     |message type id| msg stream id |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |           message stream id (cont)            |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	//
	//       Figure 9 Chunk Message Header – Type 0

	b := []byte{}
	n := 0

	b[n] = byte(csid) & 0x3f
	n++
	pio.PutU24BE(b[n:], uint32(timestamp))
	n += 3
	pio.PutU24BE(b[n:], uint32(msgdatalen))
	n += 3
	b[n] = msgtypeid
	n++
	pio.PutU32LE(b[n:], msgsid)
	n += 4

	return nil
}
