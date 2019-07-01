package h264

import (
	// "github.com/nareix/joy4/av"
	// 	"github.com/nareix/joy4/codec/h264parser"
	// "github.com/nareix/joy4/format/ts"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
)

// InitialNALU indicates the start of a h264 packet
// 0 - Forbidden 0 bit; always 0
// 1,2 - NRI
// 3,4,5,6,7 - Type
var (
	InitialNALU   = []byte{0, 0, 0, 1}
	Initial3BNALU = []byte{0, 0, 1}
	logger        *log.Logger
	streamOffset  = 0
)

func init() {
	logger = log.New(os.Stderr, "streamer ", log.Lshortfile|log.Lmicroseconds)
}

func ByteStreamReader(connection net.Conn) {
	logger.Printf("opened bytestream\n")
	defer connection.Close()
	handleConnection(connection)
}

func handleConnection(connection io.Reader) {
	logger.Printf("debug: handling connection\n")
	streamFilename := "/home/bruce/devel/go/src/github.com/mrmod/cvnightlife/output.mp4"
	_ = os.Remove(streamFilename)
	debugFile, err := os.Create(streamFilename)
	if err != nil {
		panic(err)
	}
	streamReader := &H264Reader{
		Stream:    connection,
		BitReader: &BitReader{bytes: []byte{}},
		DebugFile: debugFile,
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		logger.Printf("debug: waiting on signals\n")
		s := <-c
		logger.Printf("info: %v received, closing stream file\n", s)
		streamReader.DebugFile.Close()
		os.Exit(0)
	}()

	defer func() {
		if r := recover(); r != nil {
			logger.Printf("fatal: recovered: %v\n", r)
			logger.Printf("info: closing streamfile\n")
			streamReader.DebugFile.Close()
			os.Exit(1)
		}
	}()
	streamReader.Start()
}
