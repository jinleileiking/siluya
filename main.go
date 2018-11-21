package main

import (
	cache "Gout/go-cache"
	"bufio"
	"expvar"
	"net"
	"net/http"
	"os"

	"go.uber.org/zap"
)

const (
	CONN_HOST = "0.0.0.0"
	// CONN_HOST = "120.92.8.170"
	CONN_PORT = "8765"
	CONN_TYPE = "tcp"
)

var logger *zap.SugaredLogger
var info = expvar.NewString("info")
var evCache = expvar.NewString("cache")
var gCache = cache.New(0, 0)

func main() {

	go http.ListenAndServe(":9876", nil)

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

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {

	defer conn.Close()

	rtmpConnIns := rtmpConn{
		conn:    conn,
		cksize:  128,
		rw:      bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		mLen:    make(map[int]int, 0),
		mTypeId: make(map[int]int, 0),
	}

	if nil != handshake(rtmpConnIns) {
		return
	}

	for {
		err := waitMsg(&rtmpConnIns)
		if err != nil {
			logger.Infow("session", "detail", "EOF")
			return
		}
	}
}
