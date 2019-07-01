package h264

import (
	"fmt"
	"io"
	"os"
)

type BitReader struct {
	bytes      []byte
	byteOffset int
	bitOffset  int
	bitsRead   int
	Debug      bool
}
type H264Reader struct {
	IsStarted    bool
	Stream       io.Reader
	NalUnits     []*BitReader
	VideoStreams []*VideoStream
	DebugFile    *os.File
	*BitReader
}

func (h *H264Reader) BufferToReader(cntBytes int) error {
	buf := make([]byte, cntBytes)
	if _, err := h.Stream.Read(buf); err != nil {
		logger.Printf("error: while reading %d bytes: %v\n", cntBytes, err)
		return err
	}
	h.bytes = append(h.bytes, buf...)
	if h.DebugFile != nil {
		h.DebugFile.Write(buf)
	}
	h.byteOffset += cntBytes
	return nil
}

func (h *H264Reader) Discard(cntBytes int) error {
	buf := make([]byte, cntBytes)
	if _, err := h.Stream.Read(buf); err != nil {
		logger.Printf("error: while discarding %d bytes: %v\n", cntBytes, err)
		return err
	}
	h.byteOffset += cntBytes
	return nil
}

// TODO: what does this do ?
func bitVal(bits []int) int {
	t := 0
	for i, b := range bits {
		if b == 1 {
			t += 1 << uint((len(bits)-1)-i)
		}
	}
	// fmt.Printf("\t bitVal: %d\n", t)
	return t
}

func (h *H264Reader) Start() {
	for {
		nalUnit, _ := h.readNalUnit()
		switch nalUnit.Type {
		case NALU_TYPE_SPS:
			// TODO: handle this error
			sps, _ := NewSPS(nalUnit.rbsp, false)
			h.VideoStreams = append(
				h.VideoStreams,
				&VideoStream{SPS: sps},
			)
		case NALU_TYPE_PPS:
			videoStream := h.VideoStreams[len(h.VideoStreams)-1]
			// TODO: handle this error
			videoStream.PPS, _ = NewPPS(videoStream.SPS, nalUnit.RBSP(), false)
		case NALU_TYPE_SLICE_IDR_PICTURE:
			fallthrough
		case NALU_TYPE_SLICE_NON_IDR_PICTURE:
			videoStream := h.VideoStreams[len(h.VideoStreams)-1]
			logger.Printf("info: frame number %d\n", len(videoStream.Slices))
			// TODO: handle this error
			sliceContext, _ := NewSliceContext(videoStream, nalUnit, nalUnit.RBSP(), true)
			videoStream.Slices = append(videoStream.Slices, sliceContext)
		}
	}
}

func (r *H264Reader) readNalUnit() (*NalUnit, *BitReader) {
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

func (b *BitReader) Bytes() []byte {
	return b.bytes
}
func (b *BitReader) Fastforward(bits int) {
	b.bitsRead += bits
	b.setOffset()
}
func (b *BitReader) setOffset() {
	b.byteOffset = b.bitsRead / 8
	b.bitOffset = b.bitsRead % 8
}

// TODO: MoreRBSPData Section 7.2 p 62
func (b *BitReader) MoreRBSPData() bool {
	logger.Printf("moreRBSPData: %d [byteO: %d, bitO: %d]\n", len(b.bytes), b.byteOffset, b.bitOffset)
	if len(b.bytes)-b.byteOffset == 0 {
		return false
	}
	// Read until the least significant bit of any remaining bytes
	// If the least significant bit is 1, that marks the first bit
	// of the rbspTrailingBits() struct. If the bits read is more
	// than 0, then there is more RBSP data
	buf := make([]int, 1)
	cnt := 0
	for buf[0] != 1 {
		if _, err := b.Read(buf); err != nil {
			logger.Printf("moreRBSPData error: %v\n", err)
			return false
		}
		cnt++
	}
	logger.Printf("moreRBSPData: read %d additional bits\n", cnt)
	return cnt > 0
}
func (b *BitReader) HasMoreData() bool {
	if b.Debug {
		logger.Printf("\tHasMoreData: %+v\n", b)
		logger.Printf("\tHas %d more bytes\n", len(b.bytes)-b.byteOffset)
	}
	return len(b.bytes)-b.byteOffset > 0
}

func (b *BitReader) IsByteAligned() bool {
	return b.bitOffset == 0
}

func (b *BitReader) ReadOneBit() int {
	buf := make([]int, 1)
	_, _ = b.Read(buf)
	return buf[0]
}
func (b *BitReader) RewindBits(n int) error {
	if n > 8 {
		nBytes := n / 8
		if err := b.RewindBytes(nBytes); err != nil {
			return err
		}
		b.bitsRead -= n
		b.setOffset()
		return nil
	}
	b.bitsRead -= n
	b.setOffset()
	return nil
}

func (b *BitReader) RewindBytes(n int) error {
	if b.byteOffset-n < 0 {
		return fmt.Errorf("attempted to seek below 0")
	}
	b.byteOffset -= n
	b.bitsRead -= n * 8
	b.setOffset()
	return nil
}

// Get bytes without advancing
func (b *BitReader) PeekBytes(n int) ([]byte, error) {
	if len(b.bytes) >= b.byteOffset+n {
		return b.bytes[b.byteOffset : b.byteOffset+n], nil
	}
	return []byte{}, fmt.Errorf("EOF: not enough bytes to give %d (%d @ offset %d", n, len(b.bytes), b.byteOffset)

}

// io.ByteReader interface
func (b *BitReader) ReadByte() (byte, error) {
	if len(b.bytes) > b.byteOffset {
		bt := b.bytes[b.byteOffset]
		b.byteOffset += 1
		return bt, nil
	}
	return byte(0), fmt.Errorf("EOF:  no more bytes")
}
func (b *BitReader) ReadBytes(n int) ([]byte, error) {
	buf := []byte{}
	for i := 0; i < n; i++ {
		if _b, err := b.ReadByte(); err == nil {
			buf = append(buf, _b)
		} else {
			return buf, err
		}
	}
	return buf, nil
}

func (b *BitReader) Read(buf []int) (int, error) {
	return 0, nil

}
func (b *BitReader) NextField(name string, bits int) int {
	buf := make([]int, bits)
	if _, err := b.Read(buf); err != nil {
		fmt.Printf("error reading %d bits for %s: %v\n", bits, name, err)
		return -1
	}
	if b.Debug {
		logger.Printf("\t[%s] %d bits = value[%d]\n", name, bits, bitVal(buf))
	}
	return bitVal(buf)
}
func (b *BitReader) StreamPosition() (int, int, int) {
	return len(b.bytes), b.byteOffset, b.bitOffset
}

func (b *BitReader) LogStreamPosition() {
	logger.Printf("debug: %d byte stream @ byte %d bit %d\n", len(b.bytes), b.byteOffset, b.bitOffset)
}
