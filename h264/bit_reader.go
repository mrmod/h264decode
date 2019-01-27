package h264

import (
	"fmt"
	"github.com/mrmod/degolomb"
)

type BitReader struct {
	byteOffset int
	bitOffset  int
	bitsRead   int
}

func (b *BitReader) Fastforward(bits int) {
	b.byteOffset = bits / 8
	b.bitOffset = bits % 8
}
func (b *BitReader) setOffset() {
	b.byteOffset = b.bitsRead / 8
	b.bitOffset = b.bitsRead % 8
}

func (b *BitReader) golomb(ib []byte) []int {
	fmt.Printf("\t%d: bitReader golomb: %v\n", b.bitsRead, ib[b.byteOffset])

	zeros := -1
	bit := 0
	bits := []int{}
	for bit != 1 {
		zeros += 1
		bit = degolomb.BitArray(ib[b.byteOffset])[b.bitOffset]
		b.bitsRead += 1
		b.setOffset()
		bits = append(bits, bit)
	}
	if zeros == 0 {
		return bits
	}
	for i := 0; i < zeros; i++ {
		bit = degolomb.BitArray(ib[b.byteOffset])[b.bitOffset]
		b.bitsRead += 1
		b.setOffset()
		bits = append(bits, bit)
	}

	return bits
}

func (b *BitReader) Read(ib []byte, buf []int) (int, error) {
	fmt.Printf("\t%d: bitReader wants %d bits\n", b.bitsRead, len(buf))
	if b.byteOffset > len(ib) {
		return 0, fmt.Errorf("EOF: %d > %d\n", b.byteOffset, len(ib))
	}
	i := 0
	for {
		for _, bit := range degolomb.BitArray(ib[b.byteOffset])[b.bitOffset:8] {
			fmt.Printf("\t[%d:%d] -> buf[%d]\n", i, 8-b.bitOffset, bit)
			buf[i] = bit
			i++
			b.bitsRead += 1
			b.setOffset()
			if i >= len(buf) {
				goto BufferFilled
			}
		}
		fmt.Printf("\t -- %d\n", i)
		if b.byteOffset > len(ib) {
			return len(buf), fmt.Errorf("EOF: %d > %d\n", b.byteOffset, len(ib))
		}

	}
BufferFilled:
	return len(buf), nil

}
