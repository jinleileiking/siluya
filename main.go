package main

import (
	"net"
	"os"

	"go.uber.org/zap"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "8765"
	CONN_TYPE = "tcp"
)

var logger *zap.SugaredLogger

func main() {

	logg, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	logger = logg.Sugar()

	logger.Info("cls started")

	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		logger.Errorw("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	logger.Infow("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Errorw("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.

		// bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {

	defer conn.Close()

	handshake(conn)

	rtmpConnIns := rtmpConn{
		conn:   conn,
		cksize: 128,
	}
	for {
		waitMsg(rtmpConnIns)
	}
}
