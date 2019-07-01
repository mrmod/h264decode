/*
NAME
  parse_test.go

DESCRIPTION
  parse_test.go provides testing for parsing utilities provided in parse.go

AUTHOR
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)
*/

package h264

import (
	"bytes"
	"testing"

	"github.com/icza/bitio"
)

// TestReadUe checks that readUe correctly parses an Exp-Golomb-coded element
// to a code number.
func TestReadUe(t *testing.T) {
	// tests has been derived from Table 9-2 in ITU-T H.H264, showing bit strings
	// and corresponding codeNums.
	tests := []struct {
		in   []byte // The bitstring we wish to read.
		want uint   // The expected codeNum.
	}{
		{[]byte{0x80}, 0},  // Bit string: 1, codeNum: 0
		{[]byte{0x40}, 1},  // Bit string: 010, codeNum: 1
		{[]byte{0x60}, 2},  // Bit string: 011, codeNum: 2
		{[]byte{0x20}, 3},  // Bit string: 00100, codeNum: 3
		{[]byte{0x28}, 4},  // Bit string: 00101, codeNum: 4
		{[]byte{0x30}, 5},  // Bit string: 00110, codeNum: 5
		{[]byte{0x38}, 6},  // Bit string: 00111, codeNum: 6
		{[]byte{0x10}, 7},  // Bit string: 0001000, codeNum: 7
		{[]byte{0x12}, 8},  // Bit string: 0001001, codeNum: 8
		{[]byte{0x14}, 9},  // Bit string: 0001010, codeNum: 9
		{[]byte{0x16}, 10}, // Bit string: 0001011, codeNum: 10
	}

	for i, test := range tests {
		got, err := readUe(bitio.NewReader(bytes.NewReader(test.in)))
		if err != nil {
			t.Fatalf("did not expect error: %v from readUe", err)
		}

		if test.want != uint(got) {
			t.Errorf("did not get expected result for test: %v\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}

// TestReadTe checks that readTe correctly parses a truncated Exp-Golomb-coded
// syntax element. Expected results are outlined in section 9.1 pg209 Rec ITU-T
// H.264 (04/2017)
func TestReadTe(t *testing.T) {
	tests := []struct {
		in   []byte // The bitstring we will read.
		x    uint   // The upper bound of the range.
		want uint   // Expected result from readTe.
		err  error  // Expected error from readTe.
	}{
		{[]byte{0x30}, 1, 1, nil},
		{[]byte{0x80}, 1, 0, nil},
		{[]byte{0x30}, 5, 5, nil},
		{[]byte{0x30}, 0, 0, errReadTeBadX},
	}

	for i, test := range tests {
		got, err := readTe(bitio.NewReader(bytes.NewReader(test.in)), test.x)
		if err != test.err {
			t.Fatalf("did not get expected error for test: %v\nGot: %v\nWant: %v\n", i, err, test.err)
		}

		if test.want != uint(got) {
			t.Errorf("did not get expected result for test: %v\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}

// TestReadSe checks that readSe correctly parses an se(v) signed integer
// Exp-Golomb-coded syntax element. Expected behaviour is found in section 9.1
// and 9.1.1 of the Rec. ITU-T H.264(04/2017).
func TestReadSe(t *testing.T) {
	// tests has been derived from table 9-3 of the specifications.
	tests := []struct {
		in   []byte // Bitstring to read.
		want int    // Expected value from se(v) parsing process.
	}{
		{[]byte{0x80}, 0},
		{[]byte{0x40}, 1},
		{[]byte{0x60}, -1},
		{[]byte{0x20}, 2},
		{[]byte{0x28}, -2},
		{[]byte{0x30}, 3},
		{[]byte{0x38}, -3},
	}

	for i, test := range tests {
		got, err := readSe(bitio.NewReader(bytes.NewReader(test.in)))
		if err != nil {
			t.Fatalf("did not expect error: %v from readSe", err)
		}

		if test.want != got {
			t.Errorf("did not get expected result for test: %v\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}

// TestReadMe checks that readMe correctly parses a me(v) mapped
// Exp-Golomb-coded element. Expected behaviour is described in  in sections 9.1
// and 9.1.2 in Rec. ITU-T H.264 (04/2017).
func TestReadMe(t *testing.T) {
	in := []byte{0x38}          // Bit string: 00111, codeNum: 6.
	inErr := []byte{0x07, 0xe0} // Bit string: 0000 0111 111, codeNum: 62 (will give invalid codeNum err)

	tests := []struct {
		in   []byte // Input data.
		cat  uint   // Chroma array..
		mpm  macroblockPredictionMode
		want uint  // Expected result from readMe.
		err  error // Expected value of err from readMe.
	}{
		{in, 1, intra4x4, 29, nil},
		{in, 1, intra8x8, 29, nil},
		{in, 1, inter, 32, nil},
		{in, 2, intra4x4, 29, nil},
		{in, 2, intra8x8, 29, nil},
		{in, 2, inter, 32, nil},
		{in, 0, intra4x4, 3, nil},
		{in, 0, intra8x8, 3, nil},
		{in, 0, inter, 5, nil},
		{in, 3, intra4x4, 3, nil},
		{in, 3, intra8x8, 3, nil},
		{in, 3, inter, 5, nil},
		{inErr, 1, intra4x4, 0, errInvalidCodeNum},
		{in, 4, intra4x4, 0, errInvalidCAT},
		{in, 0, 4, 0, errInvalidMPM},
	}

	for i, test := range tests {
		got, err := readMe(bitio.NewReader(bytes.NewReader(test.in)), test.cat, test.mpm)
		if err != test.err {
			t.Fatalf("did not expect to get error: %v for test: %v", err, i)
		}

		if test.want != got {
			t.Errorf("did not get expected result for test: %v\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}
