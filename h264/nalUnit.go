package h264

import (
	"github.com/ausocean/h264decode/h264/bits"
	"github.com/pkg/errors"
)

type NalUnit struct {
	NumBytes                     int
	ForbiddenZeroBit             int
	RefIdc                       int
	Type                         int
	SvcExtensionFlag             int
	Avc3dExtensionFlag           int
	IdrFlag                      int
	PriorityId                   int
	NoInterLayerPredFlag         int
	DependencyId                 int
	QualityId                    int
	TemporalId                   int
	UseRefBasePicFlag            int
	DiscardableFlag              int
	OutputFlag                   int
	ReservedThree2Bits           int
	HeaderBytes                  int
	NonIdrFlag                   int
	ViewId                       int
	AnchorPicFlag                int
	InterViewFlag                int
	ReservedOneBit               int
	ViewIdx                      int
	DepthFlag                    int
	EmulationPreventionThreeByte byte
	rbsp                         []byte
}

func NalUnitHeaderSvcExtension(nalUnit *NalUnit, br *bits.BitReader) error {
	return readFields(br, []field{
		{&nalUnit.IdrFlag, "IdrFlag", 1},
		{&nalUnit.PriorityId, "PriorityId", 6},
		{&nalUnit.NoInterLayerPredFlag, "NoInterLayerPredFlag", 1},
		{&nalUnit.DependencyId, "DependencyId", 3},
		{&nalUnit.QualityId, "QualityId", 4},
		{&nalUnit.TemporalId, "TemporalId", 3},
		{&nalUnit.UseRefBasePicFlag, "UseRefBasePicFlag", 1},
		{&nalUnit.DiscardableFlag, "DiscardableFlag", 1},
		{&nalUnit.OutputFlag, "OutputFlag", 1},
		{&nalUnit.ReservedThree2Bits, "ReservedThree2Bits", 2},
	})
}

func NalUnitHeader3davcExtension(nalUnit *NalUnit, br *bits.BitReader) error {
	return readFields(br, []field{
		{&nalUnit.ViewIdx, "ViewIdx", 8},
		{&nalUnit.DepthFlag, "DepthFlag", 1},
		{&nalUnit.NonIdrFlag, "NonIdrFlag", 1},
		{&nalUnit.TemporalId, "TemporalId", 3},
		{&nalUnit.AnchorPicFlag, "AnchorPicFlag", 1},
		{&nalUnit.InterViewFlag, "InterViewFlag", 1},
	})
}

func NalUnitHeaderMvcExtension(nalUnit *NalUnit, br *bits.BitReader) error {
	return readFields(br, []field{
		{&nalUnit.NonIdrFlag, "NonIdrFlag", 1},
		{&nalUnit.PriorityId, "PriorityId", 6},
		{&nalUnit.ViewId, "ViewId", 10},
		{&nalUnit.TemporalId, "TemporalId", 3},
		{&nalUnit.AnchorPicFlag, "AnchorPicFlag", 1},
		{&nalUnit.InterViewFlag, "InterViewFlag", 1},
		{&nalUnit.ReservedOneBit, "ReservedOneBit", 1},
	})
}

func (n *NalUnit) RBSP() []byte {
	return n.rbsp
}

func NewNalUnit(frame []byte, numBytesInNal int) (*NalUnit, error) {
	logger.Printf("debug: reading %d byte NAL\n", numBytesInNal)
	nalUnit := NalUnit{
		NumBytes:    numBytesInNal,
		HeaderBytes: 1,
	}
	// TODO: pass in actual io.Reader to NewBitReader
	br := bits.NewBitReader(nil)

	err := readFields(br, []field{
		{&nalUnit.ForbiddenZeroBit, "ForbiddenZeroBit", 1},
		{&nalUnit.RefIdc, "NalRefIdc", 2},
		{&nalUnit.Type, "NalUnitType", 5},
	})
	if err != nil {
		return nil, err
	}

	if nalUnit.Type == 14 || nalUnit.Type == 20 || nalUnit.Type == 21 {
		if nalUnit.Type != 21 {
			b, err := br.ReadBits(1)
			if err != nil {
				return nil, errors.Wrap(err, "could not read SvcExtensionFlag")
			}
			nalUnit.SvcExtensionFlag = int(b)
		} else {
			b, err := br.ReadBits(1)
			if err != nil {
				return nil, errors.Wrap(err, "could not read Avc3dExtensionFlag")
			}
			nalUnit.Avc3dExtensionFlag = int(b)
		}
		if nalUnit.SvcExtensionFlag == 1 {
			NalUnitHeaderSvcExtension(&nalUnit, br)
			nalUnit.HeaderBytes += 3
		} else if nalUnit.Avc3dExtensionFlag == 1 {
			NalUnitHeader3davcExtension(&nalUnit, br)
			nalUnit.HeaderBytes += 2
		} else {
			NalUnitHeaderMvcExtension(&nalUnit, br)
			nalUnit.HeaderBytes += 3

		}
	}

	logger.Printf("debug: found %d byte header. Reading body\n", nalUnit.HeaderBytes)
	for i := nalUnit.HeaderBytes; i < nalUnit.NumBytes; i++ {
		next3Bytes, err := br.PeekBits(24)
		if err != nil {
			logger.Printf("error: while reading next 3 NAL bytes: %v\n", err)
			break
		}
		// Little odd, the err above and the i+2 check might be synonyms
		if i+2 < nalUnit.NumBytes && next3Bytes == 0x000003 {
			for j := 0; j < 2; j++ {
				rbspByte, err := br.ReadBits(8)
				if err != nil {
					return nil, errors.Wrap(err, "could not read rbspByte")
				}
				nalUnit.rbsp = append(nalUnit.rbsp, byte(rbspByte))
			}
			i += 2

			// Read Emulation prevention three byte.
			eptByte, err := br.ReadBits(8)
			if err != nil {
				return nil, errors.Wrap(err, "could not read eptByte")
			}
			nalUnit.EmulationPreventionThreeByte = byte(eptByte)
		} else {
			if b, err := br.ReadBits(8); err == nil {
				nalUnit.rbsp = append(nalUnit.rbsp, byte(b))
			} else {
				logger.Printf("error: while reading byte %d of %d nal bytes: %v\n", i, nalUnit.NumBytes, err)
				break
			}
		}
	}

	// nalUnit.rbsp = frame[nalUnit.HeaderBytes:]
	logger.Printf("info: decoded %s NAL with %d RBSP bytes\n", NALUnitType[nalUnit.Type], len(nalUnit.rbsp))
	return &nalUnit, nil
}
