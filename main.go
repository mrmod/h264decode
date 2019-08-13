package main

import (
	"fmt"
	"net"

	"github.com/ausocean/h264decode/h264"
)

func main() {
	port := "8000"
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		panic(fmt.Sprintf("failed to listen %s\n", err))
	}
	fmt.Printf("listening for h264 bytestreams on %s\n", port)
	defer server.Close()
	for {
		connection, err := server.Accept()
		if err != nil {
			panic(fmt.Sprintf("connection failed %s\n", err))
		}
		go h264.ByteStreamReader(connection)
		// hand connection to ReadMuxer
	}
}
