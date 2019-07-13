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

func isEmulationPreventionThreeByte(b []byte) bool {
	if len(b) != 3 {
		return false
	}
	return b[0] == byte(0) && b[1] == byte(0) && b[2] == byte(3)
}

func NalUnitHeaderSvcExtension(nalUnit *NalUnit, br *bits.BitReader) error {
	// TODO: Annex G
	b, err := br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read IdrFlag")
	}
	nalUnit.IdrFlag = int(b)

	b, err = br.ReadBits(6)
	if err != nil {
		return errors.Wrap(err, "could not read PriorityId")
	}
	nalUnit.PriorityId = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read NoInterLayerPredFlag")
	}
	nalUnit.NoInterLayerPredFlag = int(b)

	b, err = br.ReadBits(3)
	if err != nil {
		return errors.Wrap(err, "could not read DependencyId")
	}
	nalUnit.DependencyId = int(b)

	b, err = br.ReadBits(4)
	if err != nil {
		return errors.Wrap(err, "could not read QualityId")
	}
	nalUnit.QualityId = int(b)

	b, err = br.ReadBits(3)
	if err != nil {
		return errors.Wrap(err, "could not read TemporalId")
	}
	nalUnit.TemporalId = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read UseRefBasePicFlag")
	}
	nalUnit.UseRefBasePicFlag = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read DiscardableFlag")
	}
	nalUnit.DiscardableFlag = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read OutputFlag")
	}
	nalUnit.OutputFlag = int(b)

	b, err = br.ReadBits(2)
	if err != nil {
		return errors.Wrap(err, "could not read ReservedThree2Bits")
	}
	nalUnit.ReservedThree2Bits = int(b)
	return nil
}

func NalUnitHeader3davcExtension(nalUnit *NalUnit, br *bits.BitReader) error {
	// TODO: Annex J
	b, err := br.ReadBits(8)
	if err != nil {
		return errors.Wrap(err, "could not read ViewIdx")
	}
	nalUnit.ViewIdx = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read DepthFlag")
	}
	nalUnit.DepthFlag = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read NonIdrFlag")
	}
	nalUnit.NonIdrFlag = int(b)

	b, err = br.ReadBits(3)
	if err != nil {
		return errors.Wrap(err, "could not read TemporalId")
	}
	nalUnit.TemporalId = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read AnchorPicFlag")
	}
	nalUnit.AnchorPicFlag = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read InterViewFlag")
	}
	nalUnit.InterViewFlag = int(b)
	return nil
}

func NalUnitHeaderMvcExtension(nalUnit *NalUnit, br *bits.BitReader) error {
	// TODO Annex H
	b, err := br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read NonIdrFlag")
	}
	nalUnit.NonIdrFlag = int(b)

	b, err = br.ReadBits(6)
	if err != nil {
		return errors.Wrap(err, "could not read PriorityId")
	}
	nalUnit.PriorityId = int(b)

	b, err = br.ReadBits(10)
	if err != nil {
		return errors.Wrap(err, "could not read ViewId")
	}
	nalUnit.ViewId = int(b)

	b, err = br.ReadBits(3)
	if err != nil {
		return errors.Wrap(err, "could not read TemporalId")
	}
	nalUnit.TemporalId = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read AnchorPicFlag")
	}
	nalUnit.AnchorPicFlag = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read InterViewFlag")
	}
	nalUnit.InterViewFlag = int(b)

	b, err = br.ReadBits(1)
	if err != nil {
		return errors.Wrap(err, "could not read ReservedOneBit")
	}
	nalUnit.ReservedOneBit = int(b)
	return nil
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

	b, err := br.ReadBits(1)
	if err != nil {
		return nil, errors.Wrap(err, "could not read ForbiddenZeroBit ")
	}
	nalUnit.ForbiddenZeroBit = int(b)

	b, err = br.ReadBits(2)
	if err != nil {
		return nil, errors.Wrap(err, "could not read RefIdc ")
	}
	nalUnit.RefIdc = int(b)

	b, err = br.ReadBits(5)
	if err != nil {
		return nil, errors.Wrap(err, "could not read Type")
	}
	nalUnit.Type = int(b)

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
		next3Bytes, err := br.PeekBytes(3)
		if err != nil {
			logger.Printf("error: while reading next 3 NAL bytes: %v\n", err)
			break
		}
		// Little odd, the err above and the i+2 check might be synonyms
		if i+2 < nalUnit.NumBytes && isEmulationPreventionThreeByte(next3Bytes) {
			_b, _ := br.ReadBytes(3)
			nalUnit.rbsp = append(nalUnit.rbsp, _b[:2]...)
			i += 2
			nalUnit.EmulationPreventionThreeByte = _b[2]
		} else {
			if _b, err := br.ReadByte(); err == nil {
				nalUnit.rbsp = append(nalUnit.rbsp, _b)
			} else {
				logger.Printf("error: while reading byte %d of %d nal bytes: %v\n", i, nalUnit.NumBytes, err)
				break
			}
		}
	}

	// nalUnit.rbsp = frame[nalUnit.HeaderBytes:]
	logger.Printf("info: decoded %s NAL with %d RBSP bytes\n", NALUnitType[nalUnit.Type], len(nalUnit.rbsp))
	return &nalUnit
}
