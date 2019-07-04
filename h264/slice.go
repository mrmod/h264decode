package h264

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
)

// Chroma formats as defined in section 6.2, tab 6-1.
const (
	chromaMonochrome = iota
	chroma420
	chroma422
	chroma444
)

type VideoStream struct {
	SPS    *SPS
	PPS    *PPS
	Slices []*SliceContext
}
type SliceContext struct {
	*NalUnit
	*SPS
	*PPS
	*Slice
}
type Slice struct {
	Header *SliceHeader
	Data   *SliceData
}
type SliceHeader struct {
	FirstMbInSlice                   int
	SliceType                        int
	PPSID                            int
	ColorPlaneID                     int
	FrameNum                         int
	FieldPic                         bool
	BottomField                      bool
	IDRPicID                         int
	PicOrderCntLsb                   int
	DeltaPicOrderCntBottom           int
	DeltaPicOrderCnt                 []int
	RedundantPicCnt                  int
	DirectSpatialMvPred              bool
	NumRefIdxActiveOverride          bool
	NumRefIdxL0ActiveMinus1          int
	NumRefIdxL1ActiveMinus1          int
	CabacInit                        int
	SliceQpDelta                     int
	SpForSwitch                      bool
	SliceQsDelta                     int
	DisableDeblockingFilter          int
	SliceAlphaC0OffsetDiv2           int
	SliceBetaOffsetDiv2              int
	SliceGroupChangeCycle            int
	RefPicListModificationFlagL0     bool
	ModificationOfPicNums            int
	AbsDiffPicNumMinus1              int
	LongTermPicNum                   int
	RefPicListModificationFlagL1     bool
	LumaLog2WeightDenom              int
	ChromaLog2WeightDenom            int
	ChromaArrayType                  int
	LumaWeightL0Flag                 bool
	LumaWeightL0                     []int
	LumaOffsetL0                     []int
	ChromaWeightL0Flag               bool
	ChromaWeightL0                   [][]int
	ChromaOffsetL0                   [][]int
	LumaWeightL1Flag                 bool
	LumaWeightL1                     []int
	LumaOffsetL1                     []int
	ChromaWeightL1Flag               bool
	ChromaWeightL1                   [][]int
	ChromaOffsetL1                   [][]int
	NoOutputOfPriorPicsFlag          bool
	LongTermReferenceFlag            bool
	AdaptiveRefPicMarkingModeFlag    bool
	MemoryManagementControlOperation int
	DifferenceOfPicNumsMinus1        int
	LongTermFrameIdx                 int
	MaxLongTermFrameIdxPlus1         int
}

type SliceData struct {
	BitReader                *BitReader
	CabacAlignmentOneBit     int
	MbSkipRun                int
	MbSkipFlag               bool
	MbFieldDecodingFlag      bool
	EndOfSliceFlag           bool
	MbType                   int
	MbTypeName               string
	SliceTypeName            string
	PcmAlignmentZeroBit      int
	PcmSampleLuma            []int
	PcmSampleChroma          []int
	TransformSize8x8Flag     bool
	CodedBlockPattern        int
	MbQpDelta                int
	PrevIntra4x4PredModeFlag []int
	RemIntra4x4PredMode      []int
	PrevIntra8x8PredModeFlag []int
	RemIntra8x8PredMode      []int
	IntraChromaPredMode      int
	RefIdxL0                 []int
	RefIdxL1                 []int
	MvdL0                    [][][]int
	MvdL1                    [][][]int
}

// Table 7-6
var sliceTypeMap = map[int]string{
	0: "P",
	1: "B",
	2: "I",
	3: "SP",
	4: "SI",
	5: "P",
	6: "B",
	7: "I",
	8: "SP",
	9: "SI",
}

func flagVal(b bool) int {
	if b {
		return 1
	}
	return 0
}

// context-adaptive arithmetic entropy-coded element (CABAC)
// 9.3
// When parsing the slice date of a slice (7.3.4) the initialization is 9.3.1
func (d SliceData) ae(v int) int {
	// 9.3.1.1 : CABAC context initialization ctxIdx
	return 0
}

// 8.2.2
func MbToSliceGroupMap(sps *SPS, pps *PPS, header *SliceHeader) []int {
	mbaffFrameFlag := 0
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		mbaffFrameFlag = 1
	}
	mapUnitToSliceGroupMap := MapUnitToSliceGroupMap(sps, pps, header)
	mbToSliceGroupMap := []int{}
	for i := 0; i <= PicSizeInMbs(sps, header)-1; i++ {
		if sps.FrameMbsOnly || header.FieldPic {
			mbToSliceGroupMap = append(mbToSliceGroupMap, mapUnitToSliceGroupMap[i])
			continue
		}
		if mbaffFrameFlag == 1 {
			mbToSliceGroupMap = append(mbToSliceGroupMap, mapUnitToSliceGroupMap[i/2])
			continue
		}
		if !sps.FrameMbsOnly && !sps.MBAdaptiveFrameField && !header.FieldPic {
			mbToSliceGroupMap = append(
				mbToSliceGroupMap,
				mapUnitToSliceGroupMap[(i/(2*PicWidthInMbs(sps)))*PicWidthInMbs(sps)+(i%PicWidthInMbs(sps))])
		}
	}
	return mbToSliceGroupMap

}
func PicWidthInMbs(sps *SPS) int {
	return sps.PicWidthInMbsMinus1 + 1
}
func PicHeightInMapUnits(sps *SPS) int {
	return sps.PicHeightInMapUnitsMinus1 + 1
}
func PicSizeInMapUnits(sps *SPS) int {
	return PicWidthInMbs(sps) * PicHeightInMapUnits(sps)
}
func FrameHeightInMbs(sps *SPS) int {
	return (2 - flagVal(sps.FrameMbsOnly)) * PicHeightInMapUnits(sps)
}
func PicHeightInMbs(sps *SPS, header *SliceHeader) int {
	return FrameHeightInMbs(sps) / (1 + flagVal(header.FieldPic))
}
func PicSizeInMbs(sps *SPS, header *SliceHeader) int {
	return PicWidthInMbs(sps) * PicHeightInMbs(sps, header)
}

// table 6-1
func SubWidthC(sps *SPS) int {
	n := 17
	if sps.UseSeparateColorPlane {
		if sps.ChromaFormat == chroma444 {
			return n
		}
	}

	switch sps.ChromaFormat {
	case chromaMonochrome:
		return n
	case chroma420:
		n = 2
	case chroma422:
		n = 2
	case chroma444:
		n = 1

	}
	return n
}
func SubHeightC(sps *SPS) int {
	n := 17
	if sps.UseSeparateColorPlane {
		if sps.ChromaFormat == chroma444 {
			return n
		}
	}
	switch sps.ChromaFormat {
	case chromaMonochrome:
		return n
	case chroma420:
		n = 2
	case chroma422:
		n = 1
	case chroma444:
		n = 1

	}
	return n
}

// 7-36
func CodedBlockPatternLuma(data *SliceData) int {
	return data.CodedBlockPattern % 16
}
func CodedBlockPatternChroma(data *SliceData) int {
	return data.CodedBlockPattern / 16
}

// dependencyId see Annex G.8.8.1
// Also G7.3.1.1 nal_unit_header_svc_extension
func DQId(nalUnit *NalUnit) int {
	return (nalUnit.DependencyId << 4) + nalUnit.QualityId
}

// Annex G p527
func NumMbPart(nalUnit *NalUnit, sps *SPS, header *SliceHeader, data *SliceData) int {
	sliceType := sliceTypeMap[header.SliceType]
	numMbPart := 0
	if MbTypeName(sliceType, CurrMbAddr(sps, header)) == "B_SKIP" || MbTypeName(sliceType, CurrMbAddr(sps, header)) == "B_Direct_16x16" {
		if DQId(nalUnit) == 0 && nalUnit.Type != 20 {
			numMbPart = 4
		} else if DQId(nalUnit) > 0 && nalUnit.Type == 20 {
			numMbPart = 1
		}
	} else if MbTypeName(sliceType, CurrMbAddr(sps, header)) != "B_SKIP" && MbTypeName(sliceType, CurrMbAddr(sps, header)) != "B_Direct_16x16" {
		numMbPart = CurrMbAddr(sps, header)

	}
	return numMbPart
}

func MbPred(sliceContext *SliceContext, b *BitReader, rbsp []byte) error {
	var cabac *CABAC
	sliceType := sliceTypeMap[sliceContext.Slice.Header.SliceType]
	mbPartPredMode := MbPartPredMode(sliceContext.Slice.Data, sliceType, sliceContext.Slice.Data.MbType, 0)
	if mbPartPredMode == "Intra_4x4" || mbPartPredMode == "Intra_8x8" || mbPartPredMode == "Intra_16x16" {
		if mbPartPredMode == "Intra_4x4" {
			for luma4x4BlkIdx := 0; luma4x4BlkIdx < 16; luma4x4BlkIdx++ {
				var v int
				if sliceContext.PPS.EntropyCodingMode == 1 {
					// TODO: 1 bit or ae(v)
					binarization := NewBinarization(
						"PrevIntra4x4PredModeFlag",
						sliceContext.Slice.Data)
					binarization.Decode(sliceContext, b, rbsp)

					cabac = initCabac(binarization, sliceContext)
					_ = cabac
					logger.Printf("TODO: ae for PevIntra4x4PredModeFlag[%d]\n", luma4x4BlkIdx)
				} else {
					v = b.NextField(fmt.Sprintf("PrevIntra4x4PredModeFlag[%d]", luma4x4BlkIdx), 1)
				}
				sliceContext.Slice.Data.PrevIntra4x4PredModeFlag = append(
					sliceContext.Slice.Data.PrevIntra4x4PredModeFlag,
					v)
				if sliceContext.Slice.Data.PrevIntra4x4PredModeFlag[luma4x4BlkIdx] == 0 {
					if sliceContext.PPS.EntropyCodingMode == 1 {
						// TODO: 3 bits or ae(v)
						binarization := NewBinarization(
							"RemIntra4x4PredMode",
							sliceContext.Slice.Data)
						binarization.Decode(sliceContext, b, rbsp)

						logger.Printf("TODO: ae for RemIntra4x4PredMode[%d]\n", luma4x4BlkIdx)
					} else {
						v = b.NextField(fmt.Sprintf("RemIntra4x4PredMode[%d]", luma4x4BlkIdx), 3)
					}
					if len(sliceContext.Slice.Data.RemIntra4x4PredMode) < luma4x4BlkIdx {
						sliceContext.Slice.Data.RemIntra4x4PredMode = append(
							sliceContext.Slice.Data.RemIntra4x4PredMode,
							make([]int, luma4x4BlkIdx-len(sliceContext.Slice.Data.RemIntra4x4PredMode)+1)...)
					}
					sliceContext.Slice.Data.RemIntra4x4PredMode[luma4x4BlkIdx] = v
				}
			}
		}
		if mbPartPredMode == "Intra_8x8" {
			for luma8x8BlkIdx := 0; luma8x8BlkIdx < 4; luma8x8BlkIdx++ {
				sliceContext.Update(sliceContext.Slice.Header, sliceContext.Slice.Data)
				var v int
				if sliceContext.PPS.EntropyCodingMode == 1 {
					// TODO: 1 bit or ae(v)
					binarization := NewBinarization("PrevIntra8x8PredModeFlag", sliceContext.Slice.Data)
					binarization.Decode(sliceContext, b, rbsp)

					logger.Printf("TODO: ae for PrevIntra8x8PredModeFlag[%d]\n", luma8x8BlkIdx)
				} else {
					v = b.NextField(fmt.Sprintf("PrevIntra8x8PredModeFlag[%d]", luma8x8BlkIdx), 1)
				}
				sliceContext.Slice.Data.PrevIntra8x8PredModeFlag = append(
					sliceContext.Slice.Data.PrevIntra8x8PredModeFlag, v)
				if sliceContext.Slice.Data.PrevIntra8x8PredModeFlag[luma8x8BlkIdx] == 0 {
					if sliceContext.PPS.EntropyCodingMode == 1 {
						// TODO: 3 bits or ae(v)
						binarization := NewBinarization(
							"RemIntra8x8PredMode",
							sliceContext.Slice.Data)
						binarization.Decode(sliceContext, b, rbsp)

						logger.Printf("TODO: ae for RemIntra8x8PredMode[%d]\n", luma8x8BlkIdx)
					} else {
						v = b.NextField(fmt.Sprintf("RemIntra8x8PredMode[%d]", luma8x8BlkIdx), 3)
					}
					if len(sliceContext.Slice.Data.RemIntra8x8PredMode) < luma8x8BlkIdx {
						sliceContext.Slice.Data.RemIntra8x8PredMode = append(
							sliceContext.Slice.Data.RemIntra8x8PredMode,
							make([]int, luma8x8BlkIdx-len(sliceContext.Slice.Data.RemIntra8x8PredMode)+1)...)
					}
					sliceContext.Slice.Data.RemIntra8x8PredMode[luma8x8BlkIdx] = v
				}
			}

		}
		if sliceContext.Slice.Header.ChromaArrayType == 1 || sliceContext.Slice.Header.ChromaArrayType == 2 {
			if sliceContext.PPS.EntropyCodingMode == 1 {
				// TODO: ue(v) or ae(v)
				binarization := NewBinarization(
					"IntraChromaPredMode",
					sliceContext.Slice.Data)
				binarization.Decode(sliceContext, b, rbsp)

				logger.Printf("TODO: ae for IntraChromaPredMode\n")
			} else {
				var err error
				sliceContext.Slice.Data.IntraChromaPredMode, err = readUe(nil)
				if err != nil {
					return errors.Wrap(err, "could not parse IntraChromaPredMode")
				}
			}
		}

	} else if mbPartPredMode != "Direct" {
		for mbPartIdx := 0; mbPartIdx < NumMbPart(sliceContext.NalUnit, sliceContext.SPS, sliceContext.Slice.Header, sliceContext.Slice.Data); mbPartIdx++ {
			sliceContext.Update(sliceContext.Slice.Header, sliceContext.Slice.Data)
			if (sliceContext.Slice.Header.NumRefIdxL0ActiveMinus1 > 0 || sliceContext.Slice.Data.MbFieldDecodingFlag != sliceContext.Slice.Header.FieldPic) && MbPartPredMode(sliceContext.Slice.Data, sliceType, sliceContext.Slice.Data.MbType, mbPartIdx) != "Pred_L1" {
				logger.Printf("\tTODO: refIdxL0[%d] te or ae(v)\n", mbPartIdx)
				if len(sliceContext.Slice.Data.RefIdxL0) < mbPartIdx {
					sliceContext.Slice.Data.RefIdxL0 = append(
						sliceContext.Slice.Data.RefIdxL0, make([]int, mbPartIdx-len(sliceContext.Slice.Data.RefIdxL0)+1)...)
				}
				if sliceContext.PPS.EntropyCodingMode == 1 {
					// TODO: te(v) or ae(v)
					binarization := NewBinarization(
						"RefIdxL0",
						sliceContext.Slice.Data)
					binarization.Decode(sliceContext, b, rbsp)

					logger.Printf("TODO: ae for RefIdxL0[%d]\n", mbPartIdx)
				} else {
					// TODO: Only one reference picture is used for inter-prediction,
					// then the value should be 0
					if MbaffFrameFlag(sliceContext.SPS, sliceContext.Slice.Header) == 0 || !sliceContext.Slice.Data.MbFieldDecodingFlag {
						sliceContext.Slice.Data.RefIdxL0[mbPartIdx], _ = readTe(
							nil,
							uint(sliceContext.Slice.Header.NumRefIdxL0ActiveMinus1))
					} else {
						rangeMax := 2*sliceContext.Slice.Header.NumRefIdxL0ActiveMinus1 + 1
						sliceContext.Slice.Data.RefIdxL0[mbPartIdx], _ = readTe(
							nil,
							uint(rangeMax))
					}
				}
			}
		}
		for mbPartIdx := 0; mbPartIdx < NumMbPart(sliceContext.NalUnit, sliceContext.SPS, sliceContext.Slice.Header, sliceContext.Slice.Data); mbPartIdx++ {
			if MbPartPredMode(sliceContext.Slice.Data, sliceType, sliceContext.Slice.Data.MbType, mbPartIdx) != "Pred_L1" {
				for compIdx := 0; compIdx < 2; compIdx++ {
					if len(sliceContext.Slice.Data.MvdL0) < mbPartIdx {
						sliceContext.Slice.Data.MvdL0 = append(
							sliceContext.Slice.Data.MvdL0,
							make([][][]int, mbPartIdx-len(sliceContext.Slice.Data.MvdL0)+1)...)
					}
					if len(sliceContext.Slice.Data.MvdL0[mbPartIdx][0]) < compIdx {
						sliceContext.Slice.Data.MvdL0[mbPartIdx][0] = append(
							sliceContext.Slice.Data.MvdL0[mbPartIdx][0],
							make([]int, compIdx-len(sliceContext.Slice.Data.MvdL0[mbPartIdx][0])+1)...)
					}
					if sliceContext.PPS.EntropyCodingMode == 1 {
						// TODO: se(v) or ae(v)
						if compIdx == 0 {
							binarization := NewBinarization(
								"MvdLnEnd0",
								sliceContext.Slice.Data)
							binarization.Decode(sliceContext, b, rbsp)

						} else if compIdx == 1 {
							binarization := NewBinarization(
								"MvdLnEnd1",
								sliceContext.Slice.Data)
							binarization.Decode(sliceContext, b, rbsp)

						}
						logger.Printf("TODO: ae for MvdL0[%d][0][%d]\n", mbPartIdx, compIdx)
					} else {
						sliceContext.Slice.Data.MvdL0[mbPartIdx][0][compIdx], _ = readSe(nil)
					}
				}
			}
		}
		for mbPartIdx := 0; mbPartIdx < NumMbPart(sliceContext.NalUnit, sliceContext.SPS, sliceContext.Slice.Header, sliceContext.Slice.Data); mbPartIdx++ {
			sliceContext.Update(sliceContext.Slice.Header, sliceContext.Slice.Data)
			if MbPartPredMode(sliceContext.Slice.Data, sliceType, sliceContext.Slice.Data.MbType, mbPartIdx) != "Pred_L0" {
				for compIdx := 0; compIdx < 2; compIdx++ {
					if len(sliceContext.Slice.Data.MvdL1) < mbPartIdx {
						sliceContext.Slice.Data.MvdL1 = append(
							sliceContext.Slice.Data.MvdL1,
							make([][][]int, mbPartIdx-len(sliceContext.Slice.Data.MvdL1)+1)...)
					}
					if len(sliceContext.Slice.Data.MvdL1[mbPartIdx][0]) < compIdx {
						sliceContext.Slice.Data.MvdL1[mbPartIdx][0] = append(
							sliceContext.Slice.Data.MvdL0[mbPartIdx][0],
							make([]int, compIdx-len(sliceContext.Slice.Data.MvdL1[mbPartIdx][0])+1)...)
					}
					if sliceContext.PPS.EntropyCodingMode == 1 {
						if compIdx == 0 {
							binarization := NewBinarization(
								"MvdLnEnd0",
								sliceContext.Slice.Data)
							binarization.Decode(sliceContext, b, rbsp)

						} else if compIdx == 1 {
							binarization := NewBinarization(
								"MvdLnEnd1",
								sliceContext.Slice.Data)
							binarization.Decode(sliceContext, b, rbsp)

						}
						// TODO: se(v) or ae(v)
						logger.Printf("TODO: ae for MvdL1[%d][0][%d]\n", mbPartIdx, compIdx)
					} else {
						sliceContext.Slice.Data.MvdL1[mbPartIdx][0][compIdx], _ = readSe(nil)
					}
				}
			}
		}
	}
	return nil
}

// 8.2.2.1
func MapUnitToSliceGroupMap(sps *SPS, pps *PPS, header *SliceHeader) []int {
	mapUnitToSliceGroupMap := []int{}
	picSizeInMapUnits := PicSizeInMapUnits(sps)
	if pps.NumSliceGroupsMinus1 == 0 {
		// 0 to PicSizeInMapUnits -1 inclusive
		for i := 0; i <= picSizeInMapUnits-1; i++ {
			mapUnitToSliceGroupMap = append(mapUnitToSliceGroupMap, 0)
		}
	} else {
		switch pps.SliceGroupMapType {
		case 0:
			// 8.2.2.1
			i := 0
			for i < picSizeInMapUnits {
				// iGroup should be incremented in the pps.RunLengthMinus1 index operation. There may be a bug here
				for iGroup := 0; iGroup <= pps.NumSliceGroupsMinus1 && i < picSizeInMapUnits; i += pps.RunLengthMinus1[iGroup+1] + 1 {
					for j := 0; j < pps.RunLengthMinus1[iGroup] && i+j < picSizeInMapUnits; j++ {
						if len(mapUnitToSliceGroupMap) < i+j {
							mapUnitToSliceGroupMap = append(
								mapUnitToSliceGroupMap,
								make([]int, (i+j)-len(mapUnitToSliceGroupMap)+1)...)
						}
						mapUnitToSliceGroupMap[i+j] = iGroup
					}
				}
			}
		case 1:
			// 8.2.2.2
			for i := 0; i < picSizeInMapUnits; i++ {
				v := ((i % PicWidthInMbs(sps)) + (((i / PicWidthInMbs(sps)) * (pps.NumSliceGroupsMinus1 + 1)) / 2)) % (pps.NumSliceGroupsMinus1 + 1)
				mapUnitToSliceGroupMap = append(mapUnitToSliceGroupMap, v)
			}
		case 2:
			// 8.2.2.3
			for i := 0; i < picSizeInMapUnits; i++ {
				mapUnitToSliceGroupMap = append(mapUnitToSliceGroupMap, pps.NumSliceGroupsMinus1)
			}
			for iGroup := pps.NumSliceGroupsMinus1 - 1; iGroup >= 0; iGroup-- {
				yTopLeft := pps.TopLeft[iGroup] / PicWidthInMbs(sps)
				xTopLeft := pps.TopLeft[iGroup] % PicWidthInMbs(sps)
				yBottomRight := pps.BottomRight[iGroup] / PicWidthInMbs(sps)
				xBottomRight := pps.BottomRight[iGroup] % PicWidthInMbs(sps)
				for y := yTopLeft; y <= yBottomRight; y++ {
					for x := xTopLeft; x <= xBottomRight; x++ {
						idx := y*PicWidthInMbs(sps) + x
						if len(mapUnitToSliceGroupMap) < idx {
							mapUnitToSliceGroupMap = append(
								mapUnitToSliceGroupMap,
								make([]int, idx-len(mapUnitToSliceGroupMap)+1)...)
							mapUnitToSliceGroupMap[idx] = iGroup
						}
					}
				}
			}

		case 3:
			// 8.2.2.4
			// TODO
		case 4:
			// 8.2.2.5
			// TODO
		case 5:
			// 8.2.2.6
			// TODO
		case 6:
			// 8.2.2.7
			// TODO
		}
	}
	// 8.2.2.8
	// Convert mapUnitToSliceGroupMap to MbToSliceGroupMap
	return mapUnitToSliceGroupMap
}
func nextMbAddress(n int, sps *SPS, pps *PPS, header *SliceHeader) int {
	i := n + 1
	// picSizeInMbs is the number of macroblocks in picture 0
	// 7-13
	// PicWidthInMbs = sps.PicWidthInMbsMinus1 + 1
	// PicHeightInMapUnits = sps.PicHeightInMapUnitsMinus1 + 1
	// 7-29
	// picSizeInMbs = PicWidthInMbs * PicHeightInMbs
	// 7-26
	// PicHeightInMbs = FrameHeightInMbs / (1 + header.fieldPicFlag)
	// 7-18
	// FrameHeightInMbs = (2 - ps.FrameMbsOnly) * PicHeightInMapUnits
	picWidthInMbs := sps.PicWidthInMbsMinus1 + 1
	picHeightInMapUnits := sps.PicHeightInMapUnitsMinus1 + 1
	frameHeightInMbs := (2 - flagVal(sps.FrameMbsOnly)) * picHeightInMapUnits
	picHeightInMbs := frameHeightInMbs / (1 + flagVal(header.FieldPic))
	picSizeInMbs := picWidthInMbs * picHeightInMbs
	mbToSliceGroupMap := MbToSliceGroupMap(sps, pps, header)
	for i < picSizeInMbs && mbToSliceGroupMap[i] != mbToSliceGroupMap[i] {
		i++
	}
	return i
}

func CurrMbAddr(sps *SPS, header *SliceHeader) int {
	mbaffFrameFlag := 0
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		mbaffFrameFlag = 1
	}

	return header.FirstMbInSlice * (1 * mbaffFrameFlag)
}

func MbaffFrameFlag(sps *SPS, header *SliceHeader) int {
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		return 1
	}
	return 0
}

func NewSliceData(sliceContext *SliceContext, b *BitReader) (*SliceData, error) {
	var cabac *CABAC
	var err error
	logger.Printf("debug: SliceData starts at ByteOffset: %d BitOffset %d\n", b.byteOffset, b.bitOffset)
	logger.Printf("debug: \t== %d bytes remain ==\n", len(b.bytes)-b.byteOffset)
	sliceContext.Slice.Data = &SliceData{BitReader: b}
	flagField := func() bool {
		if v := b.NextField("", 1); v == 1 {
			return true
		}
		return false
	}
	// TODO: Why is this being initialized here?
	// initCabac(sliceContext)
	if sliceContext.PPS.EntropyCodingMode == 1 {
		for !b.IsByteAligned() {
			sliceContext.Slice.Data.CabacAlignmentOneBit = b.NextField("CabacAlignmentOneBit", 1)
		}
	}
	mbaffFrameFlag := 0
	if sliceContext.SPS.MBAdaptiveFrameField && !sliceContext.Slice.Header.FieldPic {
		mbaffFrameFlag = 1
	}
	currMbAddr := sliceContext.Slice.Header.FirstMbInSlice * (1 * mbaffFrameFlag)

	moreDataFlag := true
	prevMbSkipped := 0
	sliceContext.Slice.Data.SliceTypeName = sliceTypeMap[sliceContext.Slice.Header.SliceType]
	sliceContext.Slice.Data.MbTypeName = MbTypeName(sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType)
	logger.Printf("debug: \tSliceData: Processing moreData: %v\n", moreDataFlag)
	for moreDataFlag {
		logger.Printf("debug: \tLooking for more sliceContext.Slice.Data in slice type %s\n", sliceContext.Slice.Data.SliceTypeName)
		if sliceContext.Slice.Data.SliceTypeName != "I" && sliceContext.Slice.Data.SliceTypeName != "SI" {
			logger.Printf("debug: \tNonI/SI slice, processing moreData\n")
			if sliceContext.PPS.EntropyCodingMode == 0 {
				sliceContext.Slice.Data.MbSkipRun, err = readUe(nil)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse MbSkipRun")
				}

				if sliceContext.Slice.Data.MbSkipRun > 0 {
					prevMbSkipped = 1
				}
				for i := 0; i < sliceContext.Slice.Data.MbSkipRun; i++ {
					// nextMbAddress(currMbAdd
					currMbAddr = nextMbAddress(currMbAddr, sliceContext.SPS, sliceContext.PPS, sliceContext.Slice.Header)
				}
				if sliceContext.Slice.Data.MbSkipRun > 0 {
					logger.Printf("debug: \tNon-I/SI: Checking for more sliceContext.Slice.Data %d:%d:%d\n", b.byteOffset, b.bitOffset, len(b.Bytes()))
					moreDataFlag = b.MoreRBSPData()
				}
			} else {
				sliceContext.Slice.Data.MbSkipFlag = flagField()

				logger.Printf("debug: \tNon-I/SI: Eval MbSkipFlag[%v] %d:%d:%d\n", sliceContext.Slice.Data.MbSkipFlag, b.byteOffset, b.bitOffset, len(b.Bytes()))
				moreDataFlag = !sliceContext.Slice.Data.MbSkipFlag
			}
		}
		if moreDataFlag {
			if mbaffFrameFlag == 1 && (currMbAddr%2 == 0 || (currMbAddr%2 == 1 && prevMbSkipped == 1)) {
				if sliceContext.PPS.EntropyCodingMode == 1 {
					// TODO: ae implementation
					binarization := NewBinarization("MbFieldDecodingFlag", sliceContext.Slice.Data)
					binarization.Decode(sliceContext, b, b.Bytes())

					logger.Printf("TODO: ae for MbFieldDecodingFlag\n")
				} else {
					sliceContext.Slice.Data.MbFieldDecodingFlag = flagField()
				}
			}

			// BEGIN: macroblockLayer()
			if sliceContext.PPS.EntropyCodingMode == 1 {
				// TODO: ae implementation
				binarization := NewBinarization("MbType", sliceContext.Slice.Data)
				cabac = initCabac(binarization, sliceContext)
				_ = cabac
				binarization.Decode(sliceContext, b, b.Bytes())
				if binarization.PrefixSuffix {
					logger.Printf("debug: MBType binarization has prefix and suffix\n")
				}
				bits := []int{}
				for binIdx := 0; binarization.IsBinStringMatch(bits); binIdx++ {
					newBit := b.ReadOneBit()
					if binarization.UseDecodeBypass == 1 {
						// DecodeBypass
						logger.Printf("TODO: decodeBypass is set: 9.3.3.2.3")
						codIRange, codIOffset := initDecodingEngine(sliceContext.Slice.Data.BitReader)
						// Initialize the decoder
						// TODO: When should the suffix of MaxBinIdxCtx be used and when just the prefix?
						// TODO: When should the suffix of CtxIdxOffset be used?
						arithmeticDecoder := NewArithmeticDecoding(
							sliceContext,
							binarization,
							CtxIdx(
								binarization.binIdx,
								binarization.MaxBinIdxCtx.Prefix,
								binarization.CtxIdxOffset.Prefix,
							),
							codIRange,
							codIOffset,
						)
						// Bypass decoding
						codIOffset, _ = arithmeticDecoder.DecodeBypass(
							sliceContext.Slice.Data,
							codIRange,
							codIOffset,
						)
						// End DecodeBypass

					} else {
						// DO 9.3.3.1
						ctxIdx := CtxIdx(
							binIdx,
							binarization.MaxBinIdxCtx.Prefix,
							binarization.CtxIdxOffset.Prefix)
						if binarization.MaxBinIdxCtx.IsPrefixSuffix {
							logger.Printf("TODO: Handle PrefixSuffix binarization\n")
						}
						logger.Printf("debug: MBType ctxIdx for %d is %d\n", binIdx, ctxIdx)
						// Then 9.3.3.2
						codIRange, codIOffset := initDecodingEngine(b)
						logger.Printf("debug: coding engine initialized: %d/%d\n", codIRange, codIOffset)
					}
					bits = append(bits, newBit)
				}

				logger.Printf("TODO: ae for MBType\n")
			} else {
				sliceContext.Slice.Data.MbType, err = readUe(nil)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse MbType")
				}
			}
			if sliceContext.Slice.Data.MbTypeName == "I_PCM" {
				for !b.IsByteAligned() {
					_ = b.NextField("PCMAlignmentZeroBit", 1)
				}
				// 7-3 p95
				bitDepthY := 8 + sliceContext.SPS.BitDepthLumaMinus8
				for i := 0; i < 256; i++ {
					sliceContext.Slice.Data.PcmSampleLuma = append(
						sliceContext.Slice.Data.PcmSampleLuma,
						b.NextField(fmt.Sprintf("PcmSampleLuma[%d]", i), bitDepthY))
				}
				// 9.3.1 p 246
				// cabac = initCabac(binarization, sliceContext)
				// 6-1 p 47
				mbWidthC := 16 / SubWidthC(sliceContext.SPS)
				mbHeightC := 16 / SubHeightC(sliceContext.SPS)
				// if monochrome
				if sliceContext.SPS.ChromaFormat == chromaMonochrome || sliceContext.SPS.UseSeparateColorPlane {
					mbWidthC = 0
					mbHeightC = 0
				}

				bitDepthC := 8 + sliceContext.SPS.BitDepthChromaMinus8
				for i := 0; i < 2*mbWidthC*mbHeightC; i++ {
					sliceContext.Slice.Data.PcmSampleChroma = append(
						sliceContext.Slice.Data.PcmSampleChroma,
						b.NextField(fmt.Sprintf("PcmSampleChroma[%d]", i), bitDepthC))
				}
				// 9.3.1 p 246
				// cabac = initCabac(binarization, sliceContext)

			} else {
				noSubMbPartSizeLessThan8x8Flag := 1
				if sliceContext.Slice.Data.MbTypeName == "I_NxN" && MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0) != "Intra_16x16" && NumMbPart(sliceContext.NalUnit, sliceContext.SPS, sliceContext.Slice.Header, sliceContext.Slice.Data) == 4 {
					logger.Printf("\tTODO: subMbPred\n")
					/*
						subMbType := SubMbPred(sliceContext.Slice.Data.MbType)
						for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
							if subMbType[mbPartIdx] != "B_Direct_8x8" {
								if NumbSubMbPart(subMbType[mbPartIdx]) > 1 {
									noSubMbPartSizeLessThan8x8Flag = 0
								}
							} else if !sliceContext.SPS.Direct8x8Inference {
								noSubMbPartSizeLessThan8x8Flag = 0
							}
						}
					*/
				} else {
					if sliceContext.PPS.Transform8x8Mode == 1 && sliceContext.Slice.Data.MbTypeName == "I_NxN" {
						// TODO
						// 1 bit or ae(v)
						// If sliceContext.PPS.EntropyCodingMode == 1, use ae(v)
						if sliceContext.PPS.EntropyCodingMode == 1 {
							binarization := NewBinarization("TransformSize8x8Flag", sliceContext.Slice.Data)
							cabac = initCabac(binarization, sliceContext)
							binarization.Decode(sliceContext, b, b.Bytes())

							logger.Println("TODO: ae(v) for TransformSize8x8Flag")
						} else {
							sliceContext.Slice.Data.TransformSize8x8Flag = flagField()
						}
					}
					MbPred(sliceContext, b, b.Bytes())
				}
				if MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0) != "Intra_16x16" {
					// TODO: me, ae
					logger.Printf("TODO: CodedBlockPattern pending me/ae implementation\n")
					if sliceContext.PPS.EntropyCodingMode == 1 {
						binarization := NewBinarization("CodedBlockPattern", sliceContext.Slice.Data)
						cabac = initCabac(binarization, sliceContext)
						binarization.Decode(sliceContext, b, b.Bytes())

						logger.Printf("TODO: ae for CodedBlockPattern\n")
					} else {
						me, _ := readMe(
							nil,
							uint(sliceContext.Slice.Header.ChromaArrayType),
							// TODO: fix this
							//MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0)))
							0)
						sliceContext.Slice.Data.CodedBlockPattern = int(me)
					}

					// sliceContext.Slice.Data.CodedBlockPattern = me(v) | ae(v)
					if CodedBlockPatternLuma(sliceContext.Slice.Data) > 0 && sliceContext.PPS.Transform8x8Mode == 1 && sliceContext.Slice.Data.MbTypeName != "I_NxN" && noSubMbPartSizeLessThan8x8Flag == 1 && (sliceContext.Slice.Data.MbTypeName != "B_Direct_16x16" || sliceContext.SPS.Direct8x8Inference) {
						// TODO: 1 bit or ae(v)
						if sliceContext.PPS.EntropyCodingMode == 1 {
							binarization := NewBinarization("Transform8x8Flag", sliceContext.Slice.Data)
							cabac = initCabac(binarization, sliceContext)
							binarization.Decode(sliceContext, b, b.Bytes())

							logger.Printf("TODO: ae for TranformSize8x8Flag\n")
						} else {
							sliceContext.Slice.Data.TransformSize8x8Flag = flagField()
						}
					}
				}
				if CodedBlockPatternLuma(sliceContext.Slice.Data) > 0 || CodedBlockPatternChroma(sliceContext.Slice.Data) > 0 || MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0) == "Intra_16x16" {
					// TODO: se or ae(v)
					if sliceContext.PPS.EntropyCodingMode == 1 {
						binarization := NewBinarization("MbQpDelta", sliceContext.Slice.Data)
						cabac = initCabac(binarization, sliceContext)
						binarization.Decode(sliceContext, b, b.Bytes())

						logger.Printf("TODO: ae for MbQpDelta\n")
					} else {
						sliceContext.Slice.Data.MbQpDelta, _ = readSe(nil)
					}

				}
			}

		} // END MacroblockLayer
		if sliceContext.PPS.EntropyCodingMode == 0 {
			logger.Printf("debug: \tNon-I/SI: Again Checking for more sliceContext.Slice.Data %d:%d:%d\n", b.byteOffset, b.bitOffset, len(b.Bytes()))
			moreDataFlag = b.MoreRBSPData()
		} else {
			if sliceContext.Slice.Data.SliceTypeName != "I" && sliceContext.Slice.Data.SliceTypeName != "SI" {
				if sliceContext.Slice.Data.MbSkipFlag {
					prevMbSkipped = 1
				} else {
					prevMbSkipped = 0
				}
			}
			if mbaffFrameFlag == 1 && currMbAddr%2 == 0 {
				logger.Printf("debug: \tNon-I/SI: More sliceContext.Slice.Data at currMbAddr[%v] sliceContext.Slice.Data %d:%d:%d\n", currMbAddr, b.byteOffset, b.bitOffset, len(b.Bytes()))
				moreDataFlag = true
			} else {
				// TODO: ae implementation
				sliceContext.Slice.Data.EndOfSliceFlag = flagField() // ae(nil)
				logger.Printf("debug: \tNon-I/SI: End of slice[%v] %d:%d:%d\n", sliceContext.Slice.Data.EndOfSliceFlag, b.byteOffset, b.bitOffset, len(b.Bytes()))
				moreDataFlag = !sliceContext.Slice.Data.EndOfSliceFlag
			}
		}
		currMbAddr = nextMbAddress(currMbAddr, sliceContext.SPS, sliceContext.PPS, sliceContext.Slice.Header)
	} // END while moreDataFlag
	return sliceContext.Slice.Data, nil
}

func (c *SliceContext) Update(header *SliceHeader, data *SliceData) {
	c.Slice = &Slice{Header: header, Data: data}
}
func NewSliceContext(videoStream *VideoStream, nalUnit *NalUnit, rbsp []byte, showPacket bool) (*SliceContext, error) {
	var err error
	sps := videoStream.SPS
	pps := videoStream.PPS
	logger.Printf("debug: %s RBSP %d bytes %d bits == \n", NALUnitType[nalUnit.Type], len(rbsp), len(rbsp)*8)
	logger.Printf("debug: \t%#v\n", rbsp[0:8])
	var idrPic bool
	if nalUnit.Type == 5 {
		idrPic = true
	}
	header := SliceHeader{}
	if sps.UseSeparateColorPlane {
		header.ChromaArrayType = 0
	} else {
		header.ChromaArrayType = sps.ChromaFormat
	}
	b := &BitReader{bytes: rbsp}
	flagField := func() bool {
		if v := b.NextField("", 1); v == 1 {
			return true
		}
		return false
	}
	header.FirstMbInSlice, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse FirstMbInSlice")
	}

	header.SliceType, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse SliceType")
	}

	sliceType := sliceTypeMap[header.SliceType]
	logger.Printf("debug: %s (%s) slice of %d bytes\n", NALUnitType[nalUnit.Type], sliceType, len(rbsp))
	header.PPSID, err = readUe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse PPSID")
	}

	if sps.UseSeparateColorPlane {
		header.ColorPlaneID = b.NextField("ColorPlaneID", 2)
	}
	// TODO: See 7.4.3
	// header.FrameNum = b.NextField("FrameNum", 0)
	if !sps.FrameMbsOnly {
		header.FieldPic = flagField()
		if header.FieldPic {
			header.BottomField = flagField()
		}
	}
	if idrPic {
		header.IDRPicID, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse IDRPicID")
		}
	}
	if sps.PicOrderCountType == 0 {
		header.PicOrderCntLsb = b.NextField("PicOrderCntLsb", sps.Log2MaxPicOrderCntLSBMin4+4)
		if pps.BottomFieldPicOrderInFramePresent && !header.FieldPic {
			header.DeltaPicOrderCntBottom, err = readSe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse DeltaPicOrderCntBottom")
			}
		}
	}
	if sps.PicOrderCountType == 1 && !sps.DeltaPicOrderAlwaysZero {
		header.DeltaPicOrderCnt[0], err = readSe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse DeltaPicOrderCnt")
		}

		if pps.BottomFieldPicOrderInFramePresent && !header.FieldPic {
			header.DeltaPicOrderCnt[1], err = readSe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse DeltaPicOrderCnt")
			}
		}
	}
	if pps.RedundantPicCntPresent {
		header.RedundantPicCnt, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse RedundantPicCnt")
		}
	}
	if sliceType == "B" {
		header.DirectSpatialMvPred = flagField()
	}
	if sliceType == "B" || sliceType == "SP" {
		header.NumRefIdxActiveOverride = flagField()
		if header.NumRefIdxActiveOverride {
			header.NumRefIdxL0ActiveMinus1, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse NumRefIdxL0ActiveMinus1")
			}
			if sliceType == "B" {
				header.NumRefIdxL1ActiveMinus1, err = readUe(nil)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse NumRefIdxL1ActiveMinus1")
				}
			}
		}
	}

	if nalUnit.Type == 20 || nalUnit.Type == 21 {
		// Annex H
		// H.7.3.3.1.1
		// refPicListMvcModifications()
	} else {
		// 7.3.3.1
		if header.SliceType%5 != 2 && header.SliceType%5 != 4 {
			header.RefPicListModificationFlagL0 = flagField()
			if header.RefPicListModificationFlagL0 {
				for header.ModificationOfPicNums != 3 {
					header.ModificationOfPicNums, err = readUe(nil)
					if err != nil {
						return nil, errors.Wrap(err, "could not parse ModificationOfPicNums")
					}

					if header.ModificationOfPicNums == 0 || header.ModificationOfPicNums == 1 {
						header.AbsDiffPicNumMinus1, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse AbsDiffPicNumMinus1")
						}
					} else if header.ModificationOfPicNums == 2 {
						header.LongTermPicNum, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse LongTermPicNum")
						}
					}
				}
			}

		}
		if header.SliceType%5 == 1 {
			header.RefPicListModificationFlagL1 = flagField()
			if header.RefPicListModificationFlagL1 {
				for header.ModificationOfPicNums != 3 {
					header.ModificationOfPicNums, err = readUe(nil)
					if err != nil {
						return nil, errors.Wrap(err, "could not parse ModificationOfPicNums")
					}

					if header.ModificationOfPicNums == 0 || header.ModificationOfPicNums == 1 {
						header.AbsDiffPicNumMinus1, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse AbsDiffPicNumMinus1")
						}
					} else if header.ModificationOfPicNums == 2 {
						header.LongTermPicNum, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse LongTermPicNum")
						}
					}
				}
			}
		}
		// refPicListModification()
	}

	if (pps.WeightedPred && (sliceType == "P" || sliceType == "SP")) || (pps.WeightedBipred == 1 && sliceType == "B") {
		// predWeightTable()
		header.LumaLog2WeightDenom, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse LumaLog2WeightDenom")
		}

		if header.ChromaArrayType != 0 {
			header.ChromaLog2WeightDenom, err = readUe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse ChromaLog2WeightDenom")
			}
		}
		for i := 0; i <= header.NumRefIdxL0ActiveMinus1; i++ {
			header.LumaWeightL0Flag = flagField()
			if header.LumaWeightL0Flag {
				se, err := readSe(nil)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse LumaWeightL0")
				}
				header.LumaWeightL0 = append(header.LumaWeightL0, se)

				se, err = readSe(nil)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse LumaOffsetL0")
				}
				header.LumaOffsetL0 = append(header.LumaOffsetL0, se)
			}
			if header.ChromaArrayType != 0 {
				header.ChromaWeightL0Flag = flagField()
				if header.ChromaWeightL0Flag {
					header.ChromaWeightL0 = append(header.ChromaWeightL0, []int{})
					header.ChromaOffsetL0 = append(header.ChromaOffsetL0, []int{})
					for j := 0; j < 2; j++ {
						se, err := readSe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse ChromaWeightL0")
						}
						header.ChromaWeightL0[i] = append(header.ChromaWeightL0[i], se)

						se, err = readSe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse ChromaOffsetL0")
						}
						header.ChromaOffsetL0[i] = append(header.ChromaOffsetL0[i], se)
					}
				}
			}
		}
		if header.SliceType%5 == 1 {
			for i := 0; i <= header.NumRefIdxL1ActiveMinus1; i++ {
				header.LumaWeightL1Flag = flagField()
				if header.LumaWeightL1Flag {
					se, err := readSe(nil)
					if err != nil {
						return nil, errors.Wrap(err, "could not parse LumaWeightL1")
					}
					header.LumaWeightL1 = append(header.LumaWeightL1, se)

					se, err = readSe(nil)
					if err != nil {
						return nil, errors.Wrap(err, "could not parse LumaOffsetL1")
					}
					header.LumaOffsetL1 = append(header.LumaOffsetL1, se)
				}
				if header.ChromaArrayType != 0 {
					header.ChromaWeightL1Flag = flagField()
					if header.ChromaWeightL1Flag {
						header.ChromaWeightL1 = append(header.ChromaWeightL1, []int{})
						header.ChromaOffsetL1 = append(header.ChromaOffsetL1, []int{})
						for j := 0; j < 2; j++ {
							se, err := readSe(nil)
							if err != nil {
								return nil, errors.Wrap(err, "could not parse ChromaWeightL1")
							}
							header.ChromaWeightL1[i] = append(header.ChromaWeightL1[i], se)

							se, err = readSe(nil)
							if err != nil {
								return nil, errors.Wrap(err, "could not parse ChromaOffsetL1")
							}
							header.ChromaOffsetL1[i] = append(header.ChromaOffsetL1[i], se)
						}
					}
				}
			}
		}
	} // end predWeightTable
	if nalUnit.RefIdc != 0 {
		// devRefPicMarking()
		if idrPic {
			header.NoOutputOfPriorPicsFlag = flagField()
			header.LongTermReferenceFlag = flagField()
		} else {
			header.AdaptiveRefPicMarkingModeFlag = flagField()
			if header.AdaptiveRefPicMarkingModeFlag {
				header.MemoryManagementControlOperation, err = readUe(nil)
				if err != nil {
					return nil, errors.Wrap(err, "could not parse MemoryManagementControlOperation")
				}
				for header.MemoryManagementControlOperation != 0 {
					if header.MemoryManagementControlOperation == 1 || header.MemoryManagementControlOperation == 3 {
						header.DifferenceOfPicNumsMinus1, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse MemoryManagementControlOperation")
						}
					}
					if header.MemoryManagementControlOperation == 2 {
						header.LongTermPicNum, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse LongTermPicNum")
						}
					}
					if header.MemoryManagementControlOperation == 3 || header.MemoryManagementControlOperation == 6 {
						header.LongTermFrameIdx, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse LongTermFrameIdx")
						}
					}
					if header.MemoryManagementControlOperation == 4 {
						header.MaxLongTermFrameIdxPlus1, err = readUe(nil)
						if err != nil {
							return nil, errors.Wrap(err, "could not parse MaxLongTermFrameIdxPlus1")
						}
					}
				}
			}
		} // end decRefPicMarking
	}
	if pps.EntropyCodingMode == 1 && sliceType != "I" && sliceType != "SI" {
		header.CabacInit, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse CabacInit")
		}
	}
	header.SliceQpDelta, err = readSe(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse SliceQpDelta")
	}

	if sliceType == "SP" || sliceType == "SI" {
		if sliceType == "SP" {
			header.SpForSwitch = flagField()
		}
		header.SliceQsDelta, err = readSe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse SliceQsDelta")
		}
	}
	if pps.DeblockingFilterControlPresent {
		header.DisableDeblockingFilter, err = readUe(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse DisableDeblockingFilter")
		}

		if header.DisableDeblockingFilter != 1 {
			header.SliceAlphaC0OffsetDiv2, err = readSe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse SliceAlphaC0OffsetDiv2")
			}

			header.SliceBetaOffsetDiv2, err = readSe(nil)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse SliceBetaOffsetDiv2")
			}
		}
	}
	if pps.NumSliceGroupsMinus1 > 0 && pps.SliceGroupMapType >= 3 && pps.SliceGroupMapType <= 5 {
		header.SliceGroupChangeCycle = b.NextField(
			"SliceGroupChangeCycle",
			int(math.Ceil(math.Log2(float64(pps.PicSizeInMapUnitsMinus1/pps.SliceGroupChangeRateMinus1+1)))))
	}

	sliceContext := &SliceContext{
		NalUnit: nalUnit,
		SPS:     sps,
		PPS:     pps,
		Slice: &Slice{
			Header: &header,
		},
	}
	sliceContext.Slice.Data, err = NewSliceData(sliceContext, b)
	if err != nil {
		return nil, errors.Wrap(err, "could not create slice data")
	}
	if showPacket {
		debugPacket("debug: Header", sliceContext.Slice.Header)
		debugPacket("debug: Data", sliceContext.Slice.Data)
	}
	return sliceContext, nil
}
