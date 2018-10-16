package main

import "github.com/davecgh/go-spew/spew"

func play(conn *rtmpConn) {
	/*
		0000                                                     ....E..n,ï@.=.èò
		0010                                                     .@.j.è..&ÓÃa}8..
		0020                                                     ~......ÛRÙ......
		0030                           07 00 23 50 00 00 2e 09   Ð<É.}lV...#P....
		0040   01 00 00 00 17|00|00 00 00 01 64 00 20 ff e1 00   ..........d. ÿá.
		0050   19 67 64 00 20 ac d9 40 c0 29 b0 11 00 00 03 00   .gd. ¬Ù@À)°.....
		0060   01 00 00 03 00 32 0f 18 31 96 01 00 05 68 eb ec   .....2..1....hëì
		0070   b2 2c                                             ²,
	*/

	// rsp := []byte{
	// 	0x17, 0x00, 0x00, 0x00, 0x00, 0x01, 0x64,
	// 	0x00, 0x20, 0xff, 0xe1, 0x00, 0x19, 0x67,
	// 	0x64, 0x00, 0x20, 0xac, 0xd9, 0x40, 0xc0,
	// 	0x29, 0xb0, 0x11, 0x00, 0x00, 0x03, 0x00,
	// 	0x01, 0x00, 0x00, 0x03, 0x00, 0x32, 0x0f,
	// 	0x18, 0x31, 0x96, 0x01, 0x00, 0x05, 0x68,
	// 	0xeb, 0xec, 0xb2, 0x2c}

	// header

	hdr, err := gCache.Get(conn.app + ":" + conn.name + ":" + "hdr")

	spew.Dump(hdr, err)

	//BUG(lk): to get ts from publish
	// writeConn(conn.conn, mkChunkMsgHdr())
	// writeConn(conn.conn, conn.vHdr)

	// csid = 7
	// msgtypeidVideoMsg = 9

	for {
		// get data
	}
}
