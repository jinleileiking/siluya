package main

import (
	"fmt"

	"github.com/nareix/joy4/format/flv/flvio"
)

func processVideo(conn *rtmpConn, chunkData []byte) {
	tag := flvio.Tag{Type: flvio.TAG_VIDEO}
	// var n int
	var err error
	// if n, err = (&tag).ParseHeader(chunkData); err != nil {
	if _, err = (&tag).ParseHeader(chunkData); err != nil {
		return
	}

	// sequence header
	if tag.FrameType == flvio.FRAME_KEY && tag.CodecID == flvio.VIDEO_H264 && tag.AACPacketType == flvio.AVC_SEQHDR {

		// conn.vHdr = nil
		// conn.vHdr = append(conn.vHdr, chunkData...)

		gCache.Set(conn.app+":"+conn.name+":"+"hdr", chunkData, 0)

		logger.Infow("publish", "detail", fmt.Sprintf("update header: %v", tag))

	}

	// rsp := []byte{
	// 	0x17, 0x00, 0x00, 0x00, 0x00, 0x01, 0x64,
	// 	0x00, 0x20, 0xff, 0xe1, 0x00, 0x19, 0x67,
	// 	0x64, 0x00, 0x20, 0xac, 0xd9, 0x40, 0xc0,
	// 	0x29, 0xb0, 0x11, 0x00, 0x00, 0x03, 0x00,
	// 	0x01, 0x00, 0x00, 0x03, 0x00, 0x32, 0x0f,
	// 	0x18, 0x31, 0x96, 0x01, 0x00, 0x05, 0x68,
	// 	0xeb, 0xec, 0xb2, 0x2c}

	// spew.Dump(n, err)
	// spew.Dump(tag)
}
