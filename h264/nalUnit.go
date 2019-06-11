package h264

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

const (
	CODED_IDR_SLICE = 5
	IDR_PICTURE     = 5 // Matching ITU rec convention
)

func isEmulationPreventionThreeByte(b []byte) bool {
	if len(b) != 3 {
		return false
	}
	return b[0] == byte(0) && b[1] == byte(0) && b[2] == byte(3)
}

func NalUnitHeaderSvcExtension(nalUnit *NalUnit, b *BitReader) {
	// TODO: Annex G
	nalUnit.IdrFlag = b.NextField("IdrFlag", 1)
	nalUnit.PriorityId = b.NextField("PriorityId", 6)
	nalUnit.NoInterLayerPredFlag = b.NextField("NoInterLayerPredFlag", 1)
	nalUnit.DependencyId = b.NextField("DependencyId", 3)
	nalUnit.QualityId = b.NextField("QualityId", 4)
	nalUnit.TemporalId = b.NextField("TemporalId", 3)
	nalUnit.UseRefBasePicFlag = b.NextField("UseRefBasePicFlag", 1)
	nalUnit.DiscardableFlag = b.NextField("DiscardableFlag", 1)
	nalUnit.OutputFlag = b.NextField("OutputFlag", 1)
	nalUnit.ReservedThree2Bits = b.NextField("ReservedThree2Bits", 2)
}

func NalUnitHeader3davcExtension(nalUnit *NalUnit, b *BitReader) {
	// TODO: Annex J
	nalUnit.ViewIdx = b.NextField("ViewIdx", 8)
	nalUnit.DepthFlag = b.NextField("DepthFlag", 1)
	nalUnit.NonIdrFlag = b.NextField("NonIdrFlag", 1)
	nalUnit.TemporalId = b.NextField("TemporalId", 3)
	nalUnit.AnchorPicFlag = b.NextField("AnchorPicFlag", 1)
	nalUnit.InterViewFlag = b.NextField("InterViewFlag", 1)
}
func NalUnitHeaderMvcExtension(nalUnit *NalUnit, b *BitReader) {
	// TODO: Annex H
	nalUnit.NonIdrFlag = b.NextField("NonIdrFlag", 1)
	nalUnit.PriorityId = b.NextField("PriorityId", 6)
	nalUnit.ViewId = b.NextField("ViewId", 10)
	nalUnit.TemporalId = b.NextField("TemporalId", 3)
	nalUnit.AnchorPicFlag = b.NextField("AnchorPicFlag", 1)
	nalUnit.InterViewFlag = b.NextField("InterViewFlag", 1)
	nalUnit.ReservedOneBit = b.NextField("ReservedOneBit", 1)
}
func (n *NalUnit) RBSP() []byte {
	return n.rbsp
}
func NewNalUnit(frame []byte, numBytesInNal int) *NalUnit {
	logger.Printf("debug: reading %d byte NAL\n", numBytesInNal)
	nalUnit := NalUnit{
		NumBytes:    numBytesInNal,
		HeaderBytes: 1,
	}
	b := &BitReader{bytes: frame}
	nalUnit.ForbiddenZeroBit = b.NextField("ForbiddenZeroBit", 1)
	nalUnit.RefIdc = b.NextField("NalRefIdc", 2)
	nalUnit.Type = b.NextField("NalUnitType", 5)

	if nalUnit.Type == 14 || nalUnit.Type == 20 || nalUnit.Type == 21 {
		if nalUnit.Type != 21 {
			nalUnit.SvcExtensionFlag = b.NextField("SvcExtensionFlag", 1)
		} else {
			nalUnit.Avc3dExtensionFlag = b.NextField("Avc3dExtensionFlag", 1)
		}
		if nalUnit.SvcExtensionFlag == 1 {
			NalUnitHeaderSvcExtension(&nalUnit, b)
			nalUnit.HeaderBytes += 3
		} else if nalUnit.Avc3dExtensionFlag == 1 {
			NalUnitHeader3davcExtension(&nalUnit, b)
			nalUnit.HeaderBytes += 2
		} else {
			NalUnitHeaderMvcExtension(&nalUnit, b)
			nalUnit.HeaderBytes += 3

		}
	}
	b.LogStreamPosition()
	logger.Printf("debug: found %d byte header. Reading body\n", nalUnit.HeaderBytes)
	for i := nalUnit.HeaderBytes; i < nalUnit.NumBytes; i++ {
		next3Bytes, err := b.PeekBytes(3)
		if err != nil {
			logger.Printf("error: while reading next 3 NAL bytes: %v\n", err)
			break
		}
		// Little odd, the err above and the i+2 check might be synonyms
		if i+2 < nalUnit.NumBytes && isEmulationPreventionThreeByte(next3Bytes) {
			_b, _ := b.ReadBytes(3)
			nalUnit.rbsp = append(nalUnit.rbsp, _b[:2]...)
			i += 2
			nalUnit.EmulationPreventionThreeByte = _b[2]
		} else {
			if _b, err := b.ReadByte(); err == nil {
				nalUnit.rbsp = append(nalUnit.rbsp, _b)
			} else {
				logger.Printf("error: while reading byte %d of %d nal bytes: %v\n", i, nalUnit.NumBytes, err)
				break
			}
		}
	}
	if nalUnit.RefIdc != 0 {
		logger.Printf("debug: Decoding reference picture")
	}

	// nalUnit.rbsp = frame[nalUnit.HeaderBytes:]
	logger.Printf("info: decoded [%d] %s NAL with %d RBSP bytes\n",
		nalUnit.Type,
		NALUnitType[nalUnit.Type],
		len(nalUnit.rbsp))
	return &nalUnit
}
