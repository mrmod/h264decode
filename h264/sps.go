package h264

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Specification Page 43 7.3.2.1.1
// Range is always inclusive
// XRange is always exclusive
type SPS struct {
	// 8 bits
	Profile int
	// 6 bits
	Constraint0, Constraint1 int
	Constraint2, Constraint3 int
	Constraint4, Constraint5 int
	// 2 bit reserved 0 bits
	// 8 bits
	Level int
	// Range 0 - 31 ; 6 bits
	ID                         int
	ChromaFormat               int
	UseSeparateColorPlane      bool
	BitDepthLumaMinus8         int
	BitDepthChromaMinus8       int
	QPrimeYZeroTransformBypass bool
	SeqScalingMatrixPresent    bool
	// Delta is (0-12)-1 ; 4 bits
	SeqScalingList []bool // se
	// Range 0 - 12; 4 bits
	Log2MaxFrameNumMinus4 int
	// Range 0 - 2; 2 bits
	PicOrderCountType int
	// Range 0 - 12; 4 bits
	Log2MaxPicOrderCntLSBMin4 int
	DeltaPicOrderAlwaysZero   bool
	// Range (-2^31)+1 to (2^31)-1 ; 31 bits
	OffsetForNonRefPic int // Value - 1 (se)
	// Range (-2^31)+1 to (2^31)-1 ; 31 bits
	OffsetForTopToBottomField int // Value - 1 (se)
	// Range 0 - 255 ; 8 bits
	NumRefFramesInPicOrderCntCycle int
	// Range (-2^31)+1 to (2^31)-1 ; 31 bits
	OffsetForRefFrameList []int // Value - 1 ([]se)
	// Range 0 - MaxDpbFrames
	MaxNumRefFrames            int
	GapsInFrameNumValueAllowed bool
	// Page 77
	PicWidthInMbsMinus1 int
	// Page 77
	PicHeightInMapUnitsMinus1          int
	FrameMbsOnly                       bool
	MBAdaptiveFrameField               bool
	Direct8x8Inference                 bool
	FrameCropping                      bool
	FrameCropLeftOffset                int
	FrameCropRightOffset               int
	FrameCropTopOffset                 int
	FrameCropBottomOffset              int
	VuiParametersPresent               bool
	VuiParameters                      []int
	AspectRatioInfoPresent             bool
	AspectRatio                        int
	SarWidth                           int
	SarHeight                          int
	OverscanInfoPresent                bool
	OverscanAppropriate                bool
	VideoSignalTypePresent             bool
	VideoFormat                        int
	VideoFullRange                     bool
	ColorDescriptionPresent            bool
	ColorPrimaries                     int
	TransferCharacteristics            int
	MatrixCoefficients                 int
	ChromaLocInfoPresent               bool
	ChromaSampleLocTypeTopField        int
	ChromaSampleLocTypeBottomField     int
	CpbCntMinus1                       int
	BitRateScale                       int
	CpbSizeScale                       int
	BitRateValueMinus1                 []int
	Cbr                                []bool
	InitialCpbRemovalDelayLengthMinus1 int
	CpbRemovalDelayLengthMinus1        int
	CpbSizeValueMinus1                 []int
	DpbOutputDelayLengthMinus1         int
	TimeOffsetLength                   int
	TimingInfoPresent                  bool
	NumUnitsInTick                     int
	TimeScale                          int
	NalHrdParametersPresent            bool
	FixedFrameRate                     bool
	VclHrdParametersPresent            bool
	LowHrdDelay                        bool
	PicStructPresent                   bool
	BitstreamRestriction               bool
	MotionVectorsOverPicBoundaries     bool
	MaxBytesPerPicDenom                int
	MaxBitsPerMbDenom                  int
	Log2MaxMvLengthHorizontal          int
	Log2MaxMvLengthVertical            int
	MaxDecFrameBuffering               int
	MaxNumReorderFrames                int
}

var (
	DefaultScalingMatrix4x4 = [][]int{
		{6, 13, 20, 28, 13, 20, 28, 32, 20, 28, 32, 37, 28, 32, 37, 42},
		{10, 14, 20, 24, 14, 20, 24, 27, 20, 24, 27, 30, 24, 27, 30, 34},
	}

	DefaultScalingMatrix8x8 = [][]int{
		{6, 10, 13, 16, 18, 23, 25, 27,
			10, 11, 16, 18, 23, 25, 27, 29,
			13, 16, 18, 23, 25, 27, 29, 31,
			16, 18, 23, 25, 27, 29, 31, 33,
			18, 23, 25, 27, 29, 31, 33, 36,
			23, 25, 27, 29, 31, 33, 36, 38,
			25, 27, 29, 31, 33, 36, 38, 40,
			27, 29, 31, 33, 36, 38, 40, 42},
		{9, 13, 15, 17, 19, 21, 22, 24,
			13, 13, 17, 19, 21, 22, 24, 25,
			15, 17, 19, 21, 22, 24, 25, 27,
			17, 19, 21, 22, 24, 25, 27, 28,
			19, 21, 22, 24, 25, 27, 28, 30,
			21, 22, 24, 25, 27, 28, 30, 32,
			22, 24, 25, 27, 28, 30, 32, 33,
			24, 25, 27, 28, 30, 32, 33, 35},
	}
	Default4x4IntraList = []int{6, 13, 13, 20, 20, 20, 38, 38, 38, 38, 32, 32, 32, 37, 37, 42}
	Default4x4InterList = []int{10, 14, 14, 20, 20, 20, 24, 24, 24, 24, 27, 27, 27, 30, 30, 34}
	Default8x8IntraList = []int{
		6, 10, 10, 13, 11, 13, 16, 16, 16, 16, 18, 18, 18, 18, 18, 23,
		23, 23, 23, 23, 23, 25, 25, 25, 25, 25, 25, 25, 27, 27, 27, 27,
		27, 27, 27, 27, 29, 29, 29, 29, 29, 29, 29, 31, 31, 31, 31, 31,
		31, 33, 33, 33, 33, 33, 36, 36, 36, 36, 38, 38, 38, 40, 40, 42}
	Default8x8InterList = []int{
		9, 13, 13, 15, 13, 15, 17, 17, 17, 17, 19, 19, 19, 19, 19, 21,
		21, 21, 21, 21, 21, 22, 22, 22, 22, 22, 22, 22, 24, 24, 24, 24,
		24, 24, 24, 24, 25, 25, 25, 25, 25, 25, 25, 27, 27, 27, 27, 27,
		27, 28, 28, 28, 28, 28, 30, 30, 30, 30, 32, 32, 32, 33, 33, 35}
	ScalingList4x4 = map[int][]int{
		0:  Default4x4IntraList,
		1:  Default4x4IntraList,
		2:  Default4x4IntraList,
		3:  Default4x4InterList,
		4:  Default4x4InterList,
		5:  Default4x4InterList,
		6:  Default8x8IntraList,
		7:  Default8x8InterList,
		8:  Default8x8IntraList,
		9:  Default8x8InterList,
		10: Default8x8IntraList,
		11: Default8x8InterList,
	}
	ScalingList8x8 = ScalingList4x4
)

func isInList(l []int, term int) bool {
	for _, m := range l {
		if m == term {
			return true
		}
	}
	return false
}
func debugPacket(name string, packet interface{}) {
	logger.Printf("debug: %s packet\n", name)
	for _, line := range strings.Split(fmt.Sprintf("%+v", packet), " ") {
		logger.Printf("debug: \t%#v\n", line)
	}
}
func scalingList(b *BitReader, scalingList []int, sizeOfScalingList int, defaultScalingMatrix []int) error {
	lastScale := 8
	nextScale := 8
	for i := 0; i < sizeOfScalingList; i++ {
		if nextScale != 0 {
			deltaScale, err := readSe(nil)
			if err != nil {
				return errors.Wrap(err, "could not parse deltaScale")
			}
			nextScale = (lastScale + deltaScale + 256) % 256
			if i == 0 && nextScale == 0 {
				// Scaling list should use the default list for this point in the matrix
				_ = defaultScalingMatrix
			}
		}
		if nextScale == 0 {
			scalingList[i] = lastScale
		} else {
			scalingList[i] = nextScale
		}
		lastScale = scalingList[i]
	}
	return nil
}
func NewSPS(rbsp []byte, showPacket bool) (*SPS, error) {
	logger.Printf("debug: SPS RBSP %d bytes %d bits\n", len(rbsp), len(rbsp)*8)
	logger.Printf("debug: \t%#v\n", rbsp[0:8])
	sps := SPS{}
	b := &BitReader{bytes: rbsp}
	var err error
	hrdParameters := func() error {
		sps.CpbCntMinus1, err = readUe(nil)
		if err != nil {
			return errors.Wrap(err, "could not parse CpbCntMinus1")
		}

		sps.BitRateScale = b.NextField("BitRateScale", 4)
		sps.CpbSizeScale = b.NextField("CPBSizeScale", 4)
		// SchedSelIdx E1.2
		for sseli := 0; sseli <= sps.CpbCntMinus1; sseli++ {
			ue, err := readUe(nil)
			if err != nil {
				return errors.Wrap(err, "could not parse BitRateValueMinus1")
			}
			sps.BitRateValueMinus1 = append(sps.BitRateValueMinus1, ue)

			ue, err = readUe(nil)
			if err != nil {
				return errors.Wrap(err, "could not parse CpbSizeValueMinus1")
			}
			sps.CpbSizeValueMinus1 = append(sps.CpbSizeValueMinus1, ue)

			if v := b.NextField(fmt.Sprintf("CBR[%d]", sseli), 1); v == 1 {
				sps.Cbr = append(sps.Cbr, true)
			} else {
				sps.Cbr = append(sps.Cbr, false)
			}

			sps.InitialCpbRemovalDelayLengthMinus1 = b.NextField("InitialCpbRemovalDelayLengthMinus1", 5)
			sps.CpbRemovalDelayLengthMinus1 = b.NextField("CpbRemovalDelayLengthMinus1", 5)
			sps.DpbOutputDelayLengthMinus1 = b.NextField("DpbOutputDelayLengthMinus1", 5)
			sps.TimeOffsetLength = b.NextField("TimeOffsetLength", 5)
		}
		return nil
	}

	sps.Profile = b.NextField("ProfileIDC", 8)
	sps.Constraint0 = b.NextField("Constraint0", 1)
	sps.Constraint1 = b.NextField("Constraint1", 1)
	sps.Constraint2 = b.NextField("Constraint2", 1)
	sps.Constraint3 = b.NextField("Constraint3", 1)
	sps.Constraint4 = b.NextField("Constraint4", 1)
	sps.Constraint5 = b.NextField("Constraint5", 1)
	_ = b.NextField("ReservedZeroBits", 2)
	sps.Level = b.NextField("LevelIDC", 8)
	// sps.ID = b.NextField("SPSID", 6) // proper
	sps.ID, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse ID")
	}

	sps.ChromaFormat, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse ChromaFormat")
	}

	// This should be done only for certain ProfileIDC:
	isProfileIDC := []int{100, 110, 122, 244, 44, 83, 86, 118, 128, 138, 139, 134, 135}
	// SpecialProfileCase1
	if isInList(isProfileIDC, sps.Profile) {
		if sps.ChromaFormat == chroma444 {
			if v := b.NextField("SeperateColorPlaneFlag", 1); v == 1 {
				sps.UseSeparateColorPlane = true
			} else {
				sps.UseSeparateColorPlane = false
			}
		}

		sps.BitDepthLumaMinus8, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse BitDepthLumaMinus8")
		}

		sps.BitDepthChromaMinus8, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse BitDepthChromaMinus8")
		}

		if v := b.NextField("QPrimeYZeroTransformBypassFlag", 1); v == 1 {
			sps.QPrimeYZeroTransformBypass = true
		} else {
			sps.QPrimeYZeroTransformBypass = false
		}
		if v := b.NextField("SequenceScalingMatrixPresentFlag", 1); v == 1 {
			sps.SeqScalingMatrixPresent = true
		} else {
			sps.SeqScalingMatrixPresent = false
		}
		if sps.SeqScalingMatrixPresent {
			max := 12
			if sps.ChromaFormat != chroma444 {
				max = 8
			}
			logger.Printf("debug: \tbuilding Scaling matrix for %d elements\n", max)
			for i := 0; i < max; i++ {
				if v := b.NextField(fmt.Sprintf("SeqScalingListPresentFlag[%d]", i), 1); v == 1 {
					sps.SeqScalingList = append(sps.SeqScalingList, true)
				} else {
					sps.SeqScalingList = append(sps.SeqScalingList, false)
				}
				if sps.SeqScalingList[i] {
					if i < 6 {
						scalingList(
							b,
							ScalingList4x4[i],
							16,
							DefaultScalingMatrix4x4[i])
						// 4x4: Page 75 bottom
					} else {
						// 8x8 Page 76 top
						scalingList(
							b,
							ScalingList8x8[i],
							64,
							DefaultScalingMatrix8x8[i-6])
					}
				}
			}
		}
	} // End SpecialProfileCase1

	// showSPS()
	// return sps
	// Possibly wrong due to no scaling list being built
	sps.Log2MaxFrameNumMinus4, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse Log2MaxFrameNumMinus4")
	}

	sps.PicOrderCountType, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse PicOrderCountType")
	}

	if sps.PicOrderCountType == 0 {
		sps.Log2MaxPicOrderCntLSBMin4, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse Log2MaxPicOrderCntLSBMin4")
		}
	} else if sps.PicOrderCountType == 1 {
		if v := b.NextField("DeltaPicOrderAlwaysZeroFlag", 1); v == 1 {
			sps.DeltaPicOrderAlwaysZero = true
		} else {
			sps.DeltaPicOrderAlwaysZero = false
		}

		sps.OffsetForNonRefPic, err = readSe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse OffsetForNonRefPic")
		}

		sps.OffsetForTopToBottomField, err = readSe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse OffsetForTopToBottomField")
		}

		sps.NumRefFramesInPicOrderCntCycle, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse NumRefFramesInPicOrderCntCycle")
		}

		for i := 0; i < sps.NumRefFramesInPicOrderCntCycle; i++ {
			se, err := readSe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse OffsetForRefFrameList")
			}
			sps.OffsetForRefFrameList = append(
				sps.OffsetForRefFrameList,
				se)
		}

	}

	sps.MaxNumRefFrames, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse MaxNumRefFrames")
	}

	if v := b.NextField("GapsInFrameNumValueAllowedFlag", 1); v == 1 {
		sps.GapsInFrameNumValueAllowed = true
	}

	sps.PicWidthInMbsMinus1, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse PicWidthInMbsMinus1")
	}

	sps.PicHeightInMapUnitsMinus1, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse PicHeightInMapUnitsMinus1")
	}

	if v := b.NextField("FrameMbsOnlyFlag", 1); v == 1 {
		sps.FrameMbsOnly = true
	}
	if !sps.FrameMbsOnly {
		if v := b.NextField("MBAdaptiveFrameFieldFlag", 1); v == 1 {
			sps.MBAdaptiveFrameField = true
		}
	}
	if v := b.NextField("Direct8x8InferenceFlag", 1); v == 1 {
		sps.Direct8x8Inference = true
	}
	if v := b.NextField("FrameCroppingFlag", 1); v == 1 {
		sps.FrameCropping = true
	}
	if sps.FrameCropping {
		sps.FrameCropLeftOffset, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse FrameCropLeftOffset")
		}

		sps.FrameCropRightOffset, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse FrameCropRightOffset")
		}

		sps.FrameCropTopOffset, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse FrameCropTopOffset")
		}

		sps.FrameCropBottomOffset, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse FrameCropBottomOffset")
		}
	}
	if v := b.NextField("VUIParametersPresentFlag", 1); v == 1 {
		sps.VuiParametersPresent = true
	}
	if sps.VuiParametersPresent {
		// vui_parameters
		if v := b.NextField("AspectRatioInfoPresentFlag", 1); v == 1 {
			sps.AspectRatioInfoPresent = true
		}
		if sps.AspectRatioInfoPresent {
			sps.AspectRatio = b.NextField("AspectRatioIDC", 8)
			EXTENDED_SAR := 999
			if sps.AspectRatio == EXTENDED_SAR {
				sps.SarWidth = b.NextField("SARWidth", 16)
				sps.SarHeight = b.NextField("SARHeight", 16)
			}
		}
		if v := b.NextField("OverscanInfoPresentFlag", 1); v == 1 {
			sps.OverscanInfoPresent = true
		}
		if sps.OverscanInfoPresent {
			if v := b.NextField("OverscanAppropriateFlag", 1); v == 1 {
				sps.OverscanAppropriate = true
			}
		}
		if v := b.NextField("VideoSignalPresentFlag", 1); v == 1 {
			sps.VideoSignalTypePresent = true
		}
		if sps.VideoSignalTypePresent {
			sps.VideoFormat = b.NextField("VideoFormat", 3)
		}
		if sps.VideoSignalTypePresent {
			if v := b.NextField("VideoFullRangeFlag", 1); v == 1 {
				sps.VideoFullRange = true
			}
			if v := b.NextField("ColorDescriptionPresentFlag", 1); v == 1 {
				sps.ColorDescriptionPresent = true
			}
			if sps.ColorDescriptionPresent {
				sps.ColorPrimaries = b.NextField("ColorPrimaries", 8)
				sps.TransferCharacteristics = b.NextField("TransferCharacteristics", 8)
				sps.MatrixCoefficients = b.NextField("MatrixCoefficients", 8)
			}
		}
		if v := b.NextField("ChromaLocInfoPresentFlag", 1); v == 1 {
			sps.ChromaLocInfoPresent = true
		}
		if sps.ChromaLocInfoPresent {
			sps.ChromaSampleLocTypeTopField, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse ChromaSampleLocTypeTopField")
			}

			sps.ChromaSampleLocTypeBottomField, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse ChromaSampleLocTypeBottomField")
			}
		}

		if v := b.NextField("TimingInfoPresentFlag", 1); v == 1 {
			sps.TimingInfoPresent = true
		}
		if sps.TimingInfoPresent {
			sps.NumUnitsInTick = b.NextField("NumUnitsInTick", 32)
			sps.TimeScale = b.NextField("TimeScale", 32)
			if v := b.NextField("FixedFramerateFlag", 1); v == 1 {
				sps.FixedFrameRate = true
			}
		}
		if v := b.NextField("NALHRDParametersPresent", 1); v == 1 {
			sps.NalHrdParametersPresent = true
		}
		if sps.NalHrdParametersPresent {
			err = hrdParameters()
			if err != nil {
				return nil, errors.Wrap(err, "could not get hrdParameters")
			}
		}
		if v := b.NextField("VCLHRDParametersPresent", 1); v == 1 {
			sps.VclHrdParametersPresent = true
		}
		if sps.VclHrdParametersPresent {
			err = hrdParameters()
			if err != nil {
				return nil, errors.Wrap(err, "could not get hrdParameters")
			}
		}
		if sps.NalHrdParametersPresent || sps.VclHrdParametersPresent {
			if v := b.NextField("LowHRDDelayFlag", 1); v == 1 {
				sps.LowHrdDelay = true
			}
		}
		if v := b.NextField("PicStructPresentFlag", 1); v == 1 {
			sps.PicStructPresent = true
		}
		if v := b.NextField("BitstreamRestrictionFlag", 1); v == 1 {
			sps.BitstreamRestriction = true
		}
		if sps.BitstreamRestriction {
			if v := b.NextField("MotionVectorsOverPicBoundaries", 1); v == 1 {
				sps.MotionVectorsOverPicBoundaries = true
			}
			sps.MaxBytesPerPicDenom, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse MaxBytesPerPicDenom")
			}

			sps.MaxBitsPerMbDenom, err = readUe(nil)
			if err != nil {
			}
			return nil, errors.Wrap(err, "could not parse MaxBitsPerMbDenom")

			sps.Log2MaxMvLengthHorizontal, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse Log2MaxMvLengthHorizontal")
			}

			sps.Log2MaxMvLengthVertical, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse Log2MaxMvLengthVertical")
			}

			sps.MaxNumReorderFrames, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse MaxNumReorderFrames")
			}

			sps.MaxDecFrameBuffering, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse MaxDecFrameBuffering")
			}
		}

	} // End VuiParameters Annex E.1.1
	if showPacket {
		debugPacket("SPS", sps)
	}
	return &sps, nil
}
