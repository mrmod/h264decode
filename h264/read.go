package h264

import (
	"fmt"
	"io"
	"os"

	"github.com/ausocean/h264decode/h264/bits"
	"github.com/pkg/errors"
)

type H264Reader struct {
	IsStarted    bool
	Stream       io.Reader
	NalUnits     []*bits.BitReader
	VideoStreams []*VideoStream
	DebugFile    *os.File
	bytes        []byte
	byteOffset   int
	*bits.BitReader
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
		// TODO: need to handle error from this.
		nalUnit, _, _ := h.readNalUnit()
		switch nalUnit.Type {
		case naluTypeSPS:
			// TODO: handle this error
			sps, _ := NewSPS(nalUnit.rbsp, false)
			h.VideoStreams = append(
				h.VideoStreams,
				&VideoStream{SPS: sps},
			)
		case naluTypePPS:
			videoStream := h.VideoStreams[len(h.VideoStreams)-1]
			// TODO: handle this error
			videoStream.PPS, _ = NewPPS(videoStream.SPS, nalUnit.RBSP(), false)
		case naluTypeSliceIDRPicture:
			fallthrough
		case naluTypeSliceNonIDRPicture:
			videoStream := h.VideoStreams[len(h.VideoStreams)-1]
			logger.Printf("info: frame number %d\n", len(videoStream.Slices))
			// TODO: handle this error
			sliceContext, _ := NewSliceContext(videoStream, nalUnit, nalUnit.RBSP(), true)
			videoStream.Slices = append(videoStream.Slices, sliceContext)
		}
	}
}

func (r *H264Reader) readNalUnit() (*NalUnit, *bits.BitReader, error) {
	// Read to start of NAL
	logger.Printf("debug: Seeking NAL %d start\n", len(r.NalUnits))

	// TODO: Fix this.
	for !isStartSequence(nil) {
		if err := r.BufferToReader(1); err != nil {
			// TODO: should this return an error here.
			return nil, nil, nil
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
	startOffset := r.BytesRead()
	logger.Printf("debug: Seeking next NAL start\n")
	// Read to start of next NAL
	so := r.BytesRead()
	for so == startOffset || !isStartSequence(nil) {
		so = r.BytesRead()
		if err := r.BufferToReader(1); err != nil {
			// TODO: should this return an error here?
			return nil, nil, nil
		}
	}
	// logger.Printf("debug: PreRewind %#v\n", r.Bytes())
	// Rewind back the length of the start sequence
	// r.RewindBytes(4)
	// logger.Printf("debug: PostRewind %#v\n", r.Bytes())
	endOffset := r.BytesRead()
	logger.Printf("debug: found NAL unit with %d bytes from %d to %d\n", endOffset-startOffset, startOffset, endOffset)
	nalUnitReader := bits.NewBitReader(nil)
	r.NalUnits = append(r.NalUnits, nalUnitReader)
	// TODO: this should really take an io.Reader rather than []byte. Need to fix nil
	// once this is fixed.
	nalUnit, err := NewNalUnit(nil, 0)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot create new nal unit")
	}
	return nalUnit, nalUnitReader, nil
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

// TODO: complete this.
func moreRBSPData(br *bits.BitReader) bool {
	// Read until the least significant bit of any remaining bytes
	// If the least significant bit is 1, that marks the first bit
	// of the rbspTrailingBits() struct. If the bits read is more
	// than 0, then there is more RBSP data
	var bits uint64
	cnt := 0
	for bits != 1 {
		if _, err := br.ReadBits(8); err != nil {
			logger.Printf("moreRBSPData error: %v\n", err)
			return false
		}
		cnt++
	}
	logger.Printf("moreRBSPData: read %d additional bits\n", cnt)
	return cnt > 0
}

type field struct {
	loc  *int
	name string
	n    int
}

func readFields(br *bits.BitReader, fields []field) error {
	for _, f := range fields {
		b, err := br.ReadBits(f.n)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not read %s", f.name))
		}
		*f.loc = int(b)
	}
	return nil
}

type flag struct {
	loc  *bool
	name string
}

func readFlags(br *bits.BitReader, flags []flag) error {
	for _, f := range flags {
		b, err := br.ReadBits(1)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not read %s", f.name))
		}
		*f.loc = b == 1
	}
	return nil
}
