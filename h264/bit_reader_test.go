package h264

import (
	"fmt"
	"testing"
)

func TestReadBitsBE(t *testing.T) {
	testCases := []struct {
		in    []byte
		count int
		val   int
		err   error
	}{
		{
			[]byte{},
			1,
			0,
			EOFBytes,
		},
		{
			[]byte{0x01},
			1,
			128,
			nil,
		},
		{
			[]byte{0x02},
			2,
			64,
			nil,
		},
		{
			[]byte{0x0A},
			4,
			80,
			nil,
		},
		{
			[]byte{0x15},
			3,
			160,
			nil,
		},
	}

	for caseID, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", caseID), func(t *testing.T) {
			bitReader := &BitReader{bytes: testCase.in}
			v, err := bitReader.ReadBitsBE(testCase.count)
			if err != testCase.err {
				t.Errorf("expected error state %v, got %v", testCase.err, err)
			}
			if testCase.err != nil {
				t.SkipNow()
			}
			if v != testCase.val {
				t.Errorf("expected %d, got %d", testCase.val, v)
			}
		})
	}
}
func TestReadBits(t *testing.T) {
	testCases := []struct {
		in    []byte
		count int
		val   int
		err   error
	}{
		{
			[]byte{},
			1,
			0,
			EOFBytes,
		},
		{
			[]byte{0x01},
			1,
			1,
			nil,
		},
		{
			[]byte{0x0A},
			3,
			2,
			nil,
		},
		{
			[]byte{0x15},
			3,
			5,
			nil,
		},
	}

	for caseID, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", caseID), func(t *testing.T) {
			bitReader := &BitReader{bytes: testCase.in}
			v, err := bitReader.ReadBits(testCase.count)
			if err != testCase.err {
				t.Errorf("expected error state %v, got %v", testCase.err, err)
			}
			if testCase.err != nil {
				t.SkipNow()
			}
			if v != testCase.val {
				t.Errorf("expected %d, got %d", testCase.val, v)
			}
		})
	}
}
func TestReadBitsSlice(t *testing.T) {
	testCases := []struct {
		in  []byte
		n   int
		out []int
		err error
	}{
		{
			[]byte{},
			1,
			[]int{1},
			EOFBytes,
		},
		{
			[]byte{0x01},
			1,
			[]int{1},
			nil,
		},
		{
			[]byte{0x02},
			1,
			[]int{0},
			nil,
		},
		{
			[]byte{0x0A},
			8,
			[]int{0, 0, 0, 0, 1, 0, 1, 0},
			nil,
		},
		{
			[]byte{0x0A},
			9,
			[]int{0, 0, 0, 0, 1, 0, 1, 0},
			EOFBytes,
		},
	}

	for caseID, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", caseID), func(t *testing.T) {
			bitReader := &BitReader{bytes: testCase.in}
			v, err := bitReader.ReadBitsSlice(testCase.n)
			if err != testCase.err {
				t.Errorf("expected error state %v, got %v", testCase.err, err)
			}
			if testCase.err != nil {
				t.SkipNow()
			}
			if l := len(v); l != len(testCase.out) {
				t.Errorf("expected %d bits, got %d", len(testCase.out), l)
			}
			for i, bit := range testCase.out {
				if v[i] != bit {
					t.Errorf("expected bitIdx %d==%d, got %d", i, bit, v[i])
					t.Logf("failing: %#v\n", v)
				}
			}

		})
	}
}
