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
func isStartSequence(packet []byte) bool {
	if len(packet) < len(InitialNALU) {
		return false
	}
	naluSegment := packet[len(packet)-4:]
	for i := range InitialNALU {
		if naluSegment[i] != InitialNALU[i] {
			return false
		}
	}
	return true
}
func isEmpty3Byte(buf []byte) bool {
	if len(buf) < 3 {
		return false
	}
	for _, i := range buf[len(buf)-3:] {
		if i != 0 {
			return false
		}
	}
	return true
}
func isStartCodeOnePrefix(buf []byte) bool {
	for i, b := range buf {
		if i < 2 && b != byte(0) {
			return false
		}
		// byte 3 may be 0 or 1
		if i == 3 && b != byte(0) || b != byte(1) {
			return false
		}
	}
	logger.Printf("debug: found start code one prefix byte\n")
	return true
}
func readNalUnit(r *H264Reader) (*NalUnit, *BitReader) {
	// Read to start of NAL
	logger.Printf("debug: Seeking NAL %d start\n", len(r.NalUnits))
	r.LogStreamPosition()
	for !isStartSequence(r.Bytes()) {
		if err := r.BufferToReader(1); err != nil {
			return nil, nil
		}
	}
	/*
		if !r.IsStarted {
			logger.Printf("debug: skipping initial NAL zero byte spaces\n")
			r.LogStreamPosition()
			// Annex B.2 Step 1
			if err := r.Discard(1); err != nil {
				logger.Printf("error: while discarding empty byte (Annex B.2:1): %v\n", err)
				return nil
			}
			if err := r.Discard(2); err != nil {
				logger.Printf("error: while discarding start code prefix one 3bytes (Annex B.2:2): %v\n", err)
				return nil
			}
		}
	*/
	_, startOffset, _ := r.StreamPosition()
	logger.Printf("debug: Seeking next NAL start\n")
	r.LogStreamPosition()
	// Read to start of next NAL
	_, so, _ := r.StreamPosition()
	for so == startOffset || !isStartSequence(r.Bytes()) {
		_, so, _ = r.StreamPosition()
		if err := r.BufferToReader(1); err != nil {
			return nil, nil
		}
	}
	// logger.Printf("debug: PreRewind %#v\n", r.Bytes())
	// Rewind back the length of the start sequence
	// r.RewindBytes(4)
	// logger.Printf("debug: PostRewind %#v\n", r.Bytes())
	_, endOffset, _ := r.StreamPosition()
	logger.Printf("debug: found NAL unit with %d bytes from %d to %d\n", endOffset-startOffset, startOffset, endOffset)
	nalUnitReader := &BitReader{bytes: r.Bytes()[startOffset:]}
	r.NalUnits = append(r.NalUnits, nalUnitReader)
	r.LogStreamPosition()
	logger.Printf("debug: NAL Header: %#v\n", nalUnitReader.Bytes()[0:8])
	nalUnit := NewNalUnit(nalUnitReader.Bytes(), len(nalUnitReader.Bytes()))
	return nalUnit, nalUnitReader
}

func handleConnection(connection io.Reader) {
	logger.Printf("debug: handling connection\n")
	cwd, _ := os.Getwd()
	streamFilename := cwd + "/output.mp4"
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
	for {
		nalUnit, nalReader := readNalUnit(streamReader)
		_ = nalReader
		switch nalUnit.Type {
		case NALU_TYPE_SPS:
			sps := NewSPS(nalUnit.rbsp, false)
			streamReader.VideoStreams = append(
				streamReader.VideoStreams,
				&VideoStream{SPS: sps},
			)
		case NALU_TYPE_PPS:
			videoStream := streamReader.VideoStreams[len(streamReader.VideoStreams)-1]
			videoStream.PPS = NewPPS(videoStream.SPS, nalUnit.RBSP(), false)
		case NALU_TYPE_SLICE_IDR_PICTURE:
			fallthrough
		case NALU_TYPE_SLICE_NON_IDR_PICTURE:
			videoStream := streamReader.VideoStreams[len(streamReader.VideoStreams)-1]
			logger.Printf("info: frame number %d\n", len(videoStream.Slices))
			sliceContext := NewSliceContext(videoStream, nalUnit, nalUnit.RBSP(), true)
			videoStream.Slices = append(videoStream.Slices, sliceContext)
		}
	}
}

func ByteStreamReader(connection net.Conn) {
	logger.Printf("opened bytestream\n")
	defer connection.Close()
	handleConnection(connection)
}
