/*
DESCRIPTION
  bitreader_test.go provides testing for functionality defined in bitreader.go.

AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>, The Australian Ocean Laboratory (AusOcean)

LICENSE
  Copyright (C) 2017-2018 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt.  If not, see http://www.gnu.org/licenses.
*/

package bits

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestReadBits(t *testing.T) {
	tests := []struct {
		in   []byte   // The bytes the source io.Reader will be initialised with.
		n    []int    // The values of n for the reads we wish to do.
		want []uint64 // The results we expect for each ReadBits call.
		err  []error  // The error expected from each ReadBits call.
	}{
		{
			in:   []byte{0xff},
			n:    []int{8},
			want: []uint64{0xff},
			err:  []error{nil},
		},
		{
			in:   []byte{0xff},
			n:    []int{4, 4},
			want: []uint64{0x0f, 0x0f},
			err:  []error{nil, nil},
		},
		{
			in:   []byte{0xff},
			n:    []int{1, 7},
			want: []uint64{0x01, 0x7f},
			err:  []error{nil, nil},
		},
		{
			in:   []byte{0xff, 0xff},
			n:    []int{8, 8},
			want: []uint64{0xff, 0xff},
			err:  []error{nil, nil},
		},
		{
			in:   []byte{0xff, 0xff},
			n:    []int{8, 10},
			want: []uint64{0xff, 0},
			err:  []error{nil, io.ErrUnexpectedEOF},
		},
		{
			in:   []byte{0xff, 0xff},
			n:    []int{4, 8, 4},
			want: []uint64{0x0f, 0xff, 0x0f},
			err:  []error{nil, nil, nil},
		},
		{
			in:   []byte{0xff, 0xff},
			n:    []int{16},
			want: []uint64{0xffff},
			err:  []error{nil},
		},
		{
			in:   []byte{0x8f, 0xe3},
			n:    []int{4, 2, 4, 6},
			want: []uint64{0x8, 0x3, 0xf, 0x23},
			err:  []error{nil, nil, nil, nil},
		},
	}

	for i, test := range tests {
		br := NewBitReader(bytes.NewReader(test.in))

		// For each value of n defined in test.reads, we call br.ReadBits, collect
		// the result and check the error.
		var got []uint64
		for j, n := range test.n {
			bits, err := br.ReadBits(n)
			if err != test.err[j] {
				t.Fatalf("did not expect error: %v for read: %d test: %d", err, j, i)
			}
			got = append(got, bits)
		}

		// Now we can check the read results.
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("did not get expected results from ReadBits for test: %d\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}

func TestPeekBits(t *testing.T) {
	tests := []struct {
		in   []byte
		n    []int
		want []uint64
		err  []error
	}{
		{
			in:   []byte{0xff},
			n:    []int{8},
			want: []uint64{0xff},
			err:  []error{nil},
		},
		{
			in:   []byte{0x8f, 0xe3},
			n:    []int{4, 8, 16},
			want: []uint64{0x8, 0x8f, 0x8fe3},
			err:  []error{nil, nil, nil},
		},
		{
			in:   []byte{0x8f, 0xe3, 0x8f, 0xe3},
			n:    []int{32},
			want: []uint64{0x8fe38fe3},
			err:  []error{nil},
		},
		{
			in:   []byte{0x8f, 0xe3},
			n:    []int{3, 5, 10},
			want: []uint64{0x4, 0x11, 0x23f},
			err:  []error{nil, nil, nil},
		},
		{
			in:   []byte{0x8f, 0xe3},
			n:    []int{3, 20, 10},
			want: []uint64{0x4, 0, 0x23f},
			err:  []error{nil, io.ErrUnexpectedEOF, nil},
		},
	}

	for i, test := range tests {
		br := NewBitReader(bytes.NewReader(test.in))

		// Call PeekBits for each value of n defined in test.
		var got []uint64
		for j, n := range test.n {
			bits, err := br.PeekBits(n)
			if err != test.err[j] {
				t.Fatalf("did not expect error: %v for peek: %d test: %d", err, j, i)
			}
			got = append(got, bits)
		}

		// Now we can check the peek results.
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("did not get expected results from PeekBits for test: %d\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}

func TestReadOrPeek(t *testing.T) {
	// The possible operations we might make.
	const (
		read = iota
		peek
	)

	tests := []struct {
		in   []byte   // The bytes the source io.Reader will be initialised with.
		op   []int    // The series of operations we want to perform (read or peek).
		n    []int    // The values of n for the reads/peeks we wish to do.
		want []uint64 // The results we expect for each ReadBits call.
	}{
		{
			in:   []byte{0x8f, 0xe3, 0x8f, 0xe3},
			op:   []int{read, peek, peek, read, peek},
			n:    []int{13, 3, 3, 7, 12},
			want: []uint64{0x11fc, 0x3, 0x3, 0x38, 0xfe3},
		},
	}

	for i, test := range tests {
		br := NewBitReader(bytes.NewReader(test.in))

		var (
			bits uint64
			got  []uint64
			err  error
		)

		// Go through the operations we wish to perform for this test and collect
		// results/errors.
		for j, op := range test.op {
			switch op {
			case read:
				bits, err = br.ReadBits(test.n[j])
			case peek:
				bits, err = br.PeekBits(test.n[j])
			default:
				panic(fmt.Sprintf("bad test: invalid operation: %d", op))
			}
			got = append(got, bits)
			if err != nil {
				t.Fatalf("did not expect error: %v for operation: %d test: %d", err, j, i)
			}
		}

		// Now we can check the results from the reads/peeks.
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("did not get expected results for test: %d\nGot: %v\nWant: %v\n", i, got, test.want)
		}
	}
}
