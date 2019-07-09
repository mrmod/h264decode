package h264

import (
	"fmt"
	"io"
	"math"
	"os"

	"github.com/mrmod/degolomb"
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

// 9.1 Table 9-2
func ue(bits []int) int {
	return bitVal(bits) - 1
}

// {codeNum: {codedBlockPattern: v}}
var meChroma1or2 = map[int]map[string]int{
	0:  map[string]int{"Intra_4x4": 47, "Intra_8x8": 47, "Inter": 0},
	1:  map[string]int{"Intra_4x4": 31, "Intra_8x8": 31, "Inter": 16},
	2:  map[string]int{"Intra_4x4": 15, "Intra_8x8": 15, "Inter": 1},
	3:  map[string]int{"Intra_4x4": 0, "Intra_8x8": 0, "Inter": 2},
	4:  map[string]int{"Intra_4x4": 23, "Intra_8x8": 23, "Inter": 4},
	5:  map[string]int{"Intra_4x4": 27, "Intra_8x8": 27, "Inter": 8},
	6:  map[string]int{"Intra_4x4": 29, "Intra_8x8": 29, "Inter": 32},
	7:  map[string]int{"Intra_4x4": 30, "Intra_8x8": 30, "Inter": 3},
	8:  map[string]int{"Intra_4x4": 7, "Intra_8x8": 7, "Inter": 5},
	9:  map[string]int{"Intra_4x4": 11, "Intra_8x8": 11, "Inter": 10},
	10: map[string]int{"Intra_4x4": 13, "Intra_8x8": 13, "Inter": 12},
	11: map[string]int{"Intra_4x4": 14, "Intra_8x8": 14, "Inter": 15},
	12: map[string]int{"Intra_4x4": 39, "Intra_8x8": 39, "Inter": 47},
	13: map[string]int{"Intra_4x4": 43, "Intra_8x8": 43, "Inter": 7},
	14: map[string]int{"Intra_4x4": 45, "Intra_8x8": 45, "Inter": 11},
	15: map[string]int{"Intra_4x4": 46, "Intra_8x8": 46, "Inter": 13},
	16: map[string]int{"Intra_4x4": 16, "Intra_8x8": 16, "Inter": 14},
	17: map[string]int{"Intra_4x4": 3, "Intra_8x8": 3, "Inter": 6},
	18: map[string]int{"Intra_4x4": 31, "Intra_8x8": 31, "Inter": 9},
	19: map[string]int{"Intra_4x4": 10, "Intra_8x8": 10, "Inter": 31},
	20: map[string]int{"Intra_4x4": 12, "Intra_8x8": 12, "Inter": 35},
	21: map[string]int{"Intra_4x4": 19, "Intra_8x8": 19, "Inter": 37},
	22: map[string]int{"Intra_4x4": 21, "Intra_8x8": 21, "Inter": 42},
	23: map[string]int{"Intra_4x4": 26, "Intra_8x8": 26, "Inter": 44},
	24: map[string]int{"Intra_4x4": 28, "Intra_8x8": 28, "Inter": 33},
	25: map[string]int{"Intra_4x4": 35, "Intra_8x8": 35, "Inter": 34},
	26: map[string]int{"Intra_4x4": 37, "Intra_8x8": 37, "Inter": 36},
	27: map[string]int{"Intra_4x4": 42, "Intra_8x8": 42, "Inter": 40},
	28: map[string]int{"Intra_4x4": 44, "Intra_8x8": 44, "Inter": 39},
	29: map[string]int{"Intra_4x4": 1, "Intra_8x8": 1, "Inter": 43},
	30: map[string]int{"Intra_4x4": 2, "Intra_8x8": 2, "Inter": 45},
	31: map[string]int{"Intra_4x4": 4, "Intra_8x8": 4, "Inter": 46},
	32: map[string]int{"Intra_4x4": 8, "Intra_8x8": 8, "Inter": 17},
	33: map[string]int{"Intra_4x4": 17, "Intra_8x8": 17, "Inter": 18},
	34: map[string]int{"Intra_4x4": 18, "Intra_8x8": 18, "Inter": 20},
	35: map[string]int{"Intra_4x4": 20, "Intra_8x8": 20, "Inter": 24},
	36: map[string]int{"Intra_4x4": 24, "Intra_8x8": 24, "Inter": 19},
	37: map[string]int{"Intra_4x4": 6, "Intra_8x8": 6, "Inter": 21},
	38: map[string]int{"Intra_4x4": 9, "Intra_8x8": 9, "Inter": 26},
	39: map[string]int{"Intra_4x4": 22, "Intra_8x8": 22, "Inter": 28},
	40: map[string]int{"Intra_4x4": 25, "Intra_8x8": 25, "Inter": 23},
	41: map[string]int{"Intra_4x4": 32, "Intra_8x8": 32, "Inter": 27},
	42: map[string]int{"Intra_4x4": 33, "Intra_8x8": 33, "Inter": 29},
	43: map[string]int{"Intra_4x4": 34, "Intra_8x8": 34, "Inter": 30},
	44: map[string]int{"Intra_4x4": 36, "Intra_8x8": 36, "Inter": 22},
	45: map[string]int{"Intra_4x4": 40, "Intra_8x8": 40, "Inter": 25},
	46: map[string]int{"Intra_4x4": 38, "Intra_8x8": 38, "Inter": 38},
	47: map[string]int{"Intra_4x4": 41, "Intra_8x8": 41, "Inter": 41},
}
var meChroma0or3 = map[int]map[string]int{
	0:  map[string]int{"Intra_4x4": 15, "Intra_8x8": 15, "Inter": 0},
	1:  map[string]int{"Intra_4x4": 0, "Intra_8x8": 0, "Inter": 1},
	2:  map[string]int{"Intra_4x4": 7, "Intra_8x8": 7, "Inter": 2},
	3:  map[string]int{"Intra_4x4": 11, "Intra_8x8": 11, "Inter": 4},
	4:  map[string]int{"Intra_4x4": 13, "Intra_8x8": 13, "Inter": 8},
	5:  map[string]int{"Intra_4x4": 14, "Intra_8x8": 14, "Inter": 3},
	6:  map[string]int{"Intra_4x4": 3, "Intra_8x8": 3, "Inter": 5},
	7:  map[string]int{"Intra_4x4": 5, "Intra_8x8": 5, "Inter": 10},
	8:  map[string]int{"Intra_4x4": 10, "Intra_8x8": 10, "Inter": 12},
	9:  map[string]int{"Intra_4x4": 12, "Intra_8x8": 12, "Inter": 15},
	10: map[string]int{"Intra_4x4": 1, "Intra_8x8": 1, "Inter": 7},
	11: map[string]int{"Intra_4x4": 2, "Intra_8x8": 2, "Inter": 11},
	12: map[string]int{"Intra_4x4": 4, "Intra_8x8": 4, "Inter": 13},
	13: map[string]int{"Intra_4x4": 8, "Intra_8x8": 8, "Inter": 14},
	14: map[string]int{"Intra_4x4": 6, "Intra_8x8": 6, "Inter": 6},
	15: map[string]int{"Intra_4x4": 9, "Intra_8x8": 9, "Inter": 9},
}

// 9.1.2 with Table 9-4
// macroBlockPredMode is equivalent to codedBlockPattern
func me(bits []int, chromaArrayType int, macroBlockPredMode string) int {
	codeNum := bitVal(bits) - 1
	if chromaArrayType == 1 || chromaArrayType == 2 {
		return meChroma1or2[codeNum][macroBlockPredMode]
	}
	return meChroma0or3[codeNum][macroBlockPredMode]
}

// truncated exp-golomb encoded
func te(bits []int, rangeMax int) int {
	if rangeMax > 1 {
		return ue(bits)
	}
	if bits[0] == 0 {
		return 1
	}
	return 0
}

// 9.1.1 Table 9-3
func se(bits []int) int {
	codeNum := bitVal(bits) - 1
	return int(math.Pow(float64(-1), float64(codeNum+1)) * math.Ceil(float64(codeNum/2)))
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

func (b *BitReader) golomb() []int {
	zeros := -1
	bit := 0
	bits := []int{}
	for bit != 1 {
		zeros += 1
		bit = degolomb.BitArray(b.bytes[b.byteOffset])[b.bitOffset]
		b.bitsRead += 1
		b.setOffset()
		bits = append(bits, bit)
	}
	if zeros == 0 {
		return bits
	}
	for i := 0; i < zeros; i++ {
		bit = degolomb.BitArray(b.bytes[b.byteOffset])[b.bitOffset]
		b.bitsRead += 1
		b.setOffset()
		bits = append(bits, bit)
	}

	return bits
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

	_, err := b.Read(buf)
	if err != nil {
		logger.Printf("error: unable to read more bits: %v", err)
		b.LogStreamPosition()
	}
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
	return byte(0), EOFBytes
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

var EOFBytes = fmt.Errorf("error: no more bytes")
var EOFBits = fmt.Errorf("error: no more bits")

// ReadBitsBE: Read N bits, return big endian base10 value
func (b *BitReader) ReadBitsBE(n int) (int, error) {
	var val int
	bits, err := b.ReadBitsSlice(n)
	if err != nil {
		return val, err
	}
	var bitPos int
	for i := len(bits) - 1; i >= 0; i-- {
		val += bits[i] << uint(7-bitPos)
		bitPos++
	}
	return val, nil
}

// ReadBits: Read N bits, return little endian base10 value
func (b *BitReader) ReadBits(n int) (int, error) {
	bits, err := b.ReadBitsSlice(n)
	if err != nil {
		return 0, err
	}
	return bitVal(bits), nil
}

// ReadBitsSlice : Read N bits, return binarized slice
func (b *BitReader) ReadBitsSlice(n int) ([]int, error) {
	var bits []int
	if b.byteOffset > len(b.bytes) {
		return bits, EOFBytes
	}
	bits = make([]int, n)
	_byte, err := b.ReadByte()
	if err != nil {
		return bits, err
	}
	for i := range bits {

		v := byte(1 << uint8(len(bits)-i-1))
		if _byte&v == v {
			bits[i] = 1
		}
		if i > 0 && i%8 == 0 {
			_byte, err = b.ReadByte()
			if err != nil {
				return bits, err
			}
		}
	}
	return bits, nil
}
func (b *BitReader) Read(buf []int) (int, error) {
	if b.byteOffset > len(b.bytes) {
		return 0, fmt.Errorf("EOF: %d > %d\n", b.byteOffset, len(b.bytes))
	}
	i := 0
	for {
		for _, bit := range degolomb.BitArray(b.bytes[b.byteOffset])[b.bitOffset:8] {
			buf[i] = bit
			i++
			b.bitsRead += 1
			b.setOffset()
			if i >= len(buf) {
				return len(buf), nil
			}
		}
		if b.byteOffset > len(b.bytes) {
			return len(buf), fmt.Errorf("EOF: %d > %d\n", b.byteOffset, len(b.bytes))
		}

	}
	return len(buf), nil

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
