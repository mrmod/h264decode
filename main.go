package main

import "github.com/saxon-milton/h264decode/h264"
import "net"
import "fmt"

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
