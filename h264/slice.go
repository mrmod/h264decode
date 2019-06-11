package h264

import (
	"fmt"
	"math"
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
	SubMbType                []int
	CurrMbAddr               int
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

// Annex G p527, G.8.4.1 undeclared subsection for derivation of numMbPart
func NumMbPart(nalUnit *NalUnit, sps *SPS, header *SliceHeader, data *SliceData) int {
	sliceType := sliceTypeMap[header.SliceType]
	if sliceType == "B" {
		logger.Printf("error: Dropping B slice on the floor")
		return -1
	}
	numMbPart := 0
	if MbTypeName(sliceType, data.CurrMbAddr) == "B_SKIP" || MbTypeName(sliceType, data.CurrMbAddr) == "B_Direct_16x16" {
		if DQId(nalUnit) == 0 && nalUnit.Type != 20 {
			numMbPart = 4
		} else if DQId(nalUnit) > 0 && nalUnit.Type == 20 {
			numMbPart = 1
		}
	} else if MbTypeName(sliceType, data.CurrMbAddr) != "B_SKIP" && MbTypeName(sliceType, data.CurrMbAddr) != "B_Direct_16x16" {
		numMbPart = data.CurrMbAddr

	}
	return numMbPart
}
func SubMbPredMode(s *SliceContext, subMbType int) string {
	switch subMbType {
	case 1, 2, 3, 4:
		return "Pred_L0"
	}
	return "na"
}

// 7.3.5.2 subMbPred
// returns []int : []subMbType
func SubMbPred(s *SliceContext, b *BitReader, mbTypeName string) []int {
	subMbTypeList := []int{}
	// 1
	for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
		var v int
		if s.PPS.EntropyCodingMode == 1 {
			// TODO: ae(v)
			logger.Printf("debug: TODO ae(v) implementation")
			v = 0
		} else {
			v = ue(b.golomb())
		}
		subMbTypeList = append(subMbTypeList, v)
	}
	// 2
	for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
		rangeMax := 8
		subMbTypeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], subMbTypeList[mbPartIdx])
		subMbPredModeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], SubMbPred(s, b, subMbTypeName)[mbPartIdx])
		if (s.Slice.Header.NumRefIdxL0ActiveMinus1 > 0 || s.Slice.Data.MbFieldDecodingFlag != s.Slice.Header.FieldPic) && mbTypeName != "P_8x8ref0" && subMbTypeName != "B_Direct_8x8" && subMbPredModeName != "Pred_L1" {
			var v int
			if s.PPS.EntropyCodingMode == 1 {
				// TODO: ae(v)
				logger.Printf("debug: TODO ae(v) refIdxL0")
				v = 0
			} else {
				v = te(b.golomb(), rangeMax) // TODO: What is rangeMax?

			}
			if mbPartIdx < len(s.Slice.Data.RefIdxL0) {
				s.Slice.Data.RefIdxL0[mbPartIdx] = v
			} else {
				s.Slice.Data.RefIdxL0 = append(s.Slice.Data.RefIdxL0, v)
			}
		}
	}
	// 3
	for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
		rangeMax := 8

		subMbTypeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], subMbTypeList[mbPartIdx])
		subMbPredModeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], SubMbPred(s, b, subMbTypeName)[mbPartIdx])
		if (s.Slice.Header.NumRefIdxL1ActiveMinus1 > 0 || s.Slice.Data.MbFieldDecodingFlag != s.Slice.Header.FieldPic) && subMbTypeName != "B_Direct_8x8" && subMbPredModeName != "Pred_L0" {
			var v int
			if s.PPS.EntropyCodingMode == 1 {
				// TODO: ae(v)
				logger.Printf("debug: TODO ae(v) refIdxL1")
				v = 0
			} else {
				v = te(b.golomb(), rangeMax) // TODO: What is rangeMax?

			}
			if mbPartIdx < len(s.Slice.Data.RefIdxL0) {
				s.Slice.Data.RefIdxL1[mbPartIdx] = v
			} else {
				s.Slice.Data.RefIdxL1 = append(s.Slice.Data.RefIdxL1, v)
			}
		}
	}

	// 4
	for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
		subMbTypeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], subMbTypeList[mbPartIdx])
		subMbPredModeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], SubMbPred(s, b, subMbTypeName)[mbPartIdx])

		if subMbTypeName != "B_Direct_8x8" && subMbPredModeName != "Pred_L1" {
			for subMbPartIdx := 0; subMbPartIdx < NumSubMbPart(subMbTypeList[mbPartIdx]); subMbPartIdx++ {
				for compIdx := 0; compIdx < 2; compIdx++ {
					var v int
					if s.PPS.EntropyCodingMode == 1 {
						logger.Printf("debug: TODO ae(v) mvdL0 sub MB prediction")
						v = 0
					} else {
						v = se(b.golomb())
					}
					if len(s.Slice.Data.MvdL0) > mbPartIdx {
						if len(s.Slice.Data.MvdL0[mbPartIdx]) > subMbPartIdx {
							if len(s.Slice.Data.MvdL0[mbPartIdx][subMbPartIdx]) > compIdx {
								s.Slice.Data.MvdL0[mbPartIdx][subMbPartIdx][compIdx] = v
							} else {
								s.Slice.Data.MvdL0[mbPartIdx][subMbPartIdx] = append(s.Slice.Data.MvdL0[mbPartIdx][subMbPartIdx], v)
							}
						} else {
							s.Slice.Data.MvdL0[mbPartIdx] = [][]int{
								[]int{compIdx}}
						}
					} else {
						s.Slice.Data.MvdL0 = [][][]int{
							[][]int{
								[]int{compIdx}}}
					}
				}
			}
		}
	}
	// 5
	for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
		subMbTypeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], subMbTypeList[mbPartIdx])
		subMbPredModeName := MbTypeName(sliceTypeMap[s.Slice.Header.SliceType], SubMbPred(s, b, subMbTypeName)[mbPartIdx])
		if subMbTypeName != "B_Direct_8x8" && subMbPredModeName != "Pred_L0" {
			for subMbPartIdx := 0; subMbPartIdx < NumSubMbPart(subMbTypeList[mbPartIdx]); subMbPartIdx++ {
				for compIdx := 0; compIdx < 2; compIdx++ {
					var v int
					if s.PPS.EntropyCodingMode == 1 {
						logger.Printf("debug: TODO ae(v) mvdL0 sub MB prediction")
						v = 0
					} else {
						v = se(b.golomb())
					}
					if len(s.Slice.Data.MvdL1) > mbPartIdx {
						if len(s.Slice.Data.MvdL1[mbPartIdx]) > subMbPartIdx {
							if len(s.Slice.Data.MvdL1[mbPartIdx][subMbPartIdx]) > compIdx {
								s.Slice.Data.MvdL1[mbPartIdx][subMbPartIdx][compIdx] = v
							} else {
								s.Slice.Data.MvdL1[mbPartIdx][subMbPartIdx] = append(s.Slice.Data.MvdL1[mbPartIdx][subMbPartIdx], v)
							}
						} else {
							s.Slice.Data.MvdL1[mbPartIdx] = append(s.Slice.Data.MvdL1[mbPartIdx], [][]int{
								[]int{v}}...)
						}
					} else {
						s.Slice.Data.MvdL1 = append(s.Slice.Data.MvdL1, [][][]int{
							[][]int{
								[]int{compIdx}}}...)
					}
				}
			}
		}
	}

	return subMbTypeList
}

// table 7-17
func SubMbPartHeight(subMbType int) int {
	switch subMbType {
	case 0:
		fallthrough
	case 2:
		return 8
	case 1:
		fallthrough
	case 3:
		return 4
	}
	// na
	return -1
}

// table 7-17
func SubMbPartWidth(subMbType int) int {
	switch subMbType {
	case 0:
		fallthrough
	case 2:
		return 8
	case 1:
		fallthrough
	case 3:
		return 4
	}
	// na
	return -1
}

// table 7-17
func NumSubMbPart(subMbType int) int {
	switch subMbType {
	case 0:
		return 1
	case 1:
		fallthrough
	case 2:
		return 2
	case 3:
		return 4
	}
	// na
	return -1
}

// 7.3.5.1
// Macroblock prediction
func MbPred(sliceContext *SliceContext, currMbAddr int, b *BitReader, rbsp []byte) {
	logger.Printf("debug: entering macroblock prediction for currentMbAddr %d", currMbAddr)
	var cabac *CABAC
	sliceType := sliceTypeMap[sliceContext.Slice.Header.SliceType]
	mbPartPredMode := MbPartPredMode(sliceContext.Slice.Data, sliceType, sliceContext.Slice.Data.MbType, 0)
	x, y := InverseMBScan(
		sliceContext,
		MbaffFrameFlag(sliceContext.SPS, sliceContext.Header),
		currMbAddr)
	logger.Printf("debug: derived (%d,%d) top-left block location", x, y)
	/*
		// TODO: Properly derive chroma flag
		chromaFlag := 0
		// TODO: This doesn't account for ref_layer_dq_id
		refLayerMbWidthC := MbWidthC(sliceContext.SPS)
		refLayerMbHeightC := MbHeightC(sliceContext.SPS)
		_ = refLayerMbHeightC
		refMbW := RefMbW(chromaFlag, refLayerMbWidthC)
		// g.6.3 - xRefMin16
		xRefMin16 := 1
		xOffset := XOffset(xRefMin16, refMbW)
		xr := Xr(x, xOffset, RefMbW(chromaFlag, MbWidthC(sliceContext.SPS)))
		xd := Xd(xr, refMbW)
		_ = xd
	*/

	logger.Printf("debug: macroblock prediction mode %s for %s slice type", mbPartPredMode, sliceType)
	if mbPartPredMode == "Intra_4x4" || mbPartPredMode == "Intra_8x8" || mbPartPredMode == "Intra_16x16" {
		if mbPartPredMode == "Intra_4x4" {
			for luma4x4BlkIdx := 0; luma4x4BlkIdx < 16; luma4x4BlkIdx++ {
				// See 6.4.11.4
				logger.Printf("debug: processing luma 4x4 block %d", luma4x4BlkIdx)
				Neighbor4x4LumaBlock(luma4x4BlkIdx)
				var v int
				if sliceContext.PPS.EntropyCodingMode == 1 {
					// TODO: 1 bit or ae(v)
					binarization := NewBinarization(
						"PrevIntra4x4PredModeFlag",
						sliceContext.Slice.Data)
					binarization.Decode(sliceContext, b, rbsp)

					cabac = initContextVariables("PrevIntra4x4PredModeFlag", binarization, sliceContext)
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
				// See 6.4.11.2
				logger.Printf("debug: processing luma 8x8 block %d", luma8x8BlkIdx)
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
				sliceContext.Slice.Data.IntraChromaPredMode = ue(b.golomb())
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
						sliceContext.Slice.Data.RefIdxL0[mbPartIdx] = te(
							b.golomb(),
							sliceContext.Slice.Header.NumRefIdxL0ActiveMinus1)
					} else {
						rangeMax := 2*sliceContext.Slice.Header.NumRefIdxL0ActiveMinus1 + 1
						sliceContext.Slice.Data.RefIdxL0[mbPartIdx] = te(
							b.golomb(),
							rangeMax)
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
						sliceContext.Slice.Data.MvdL0[mbPartIdx][0][compIdx] = se(b.golomb())
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
						sliceContext.Slice.Data.MvdL1[mbPartIdx][0][compIdx] = se(b.golomb())
					}
				}
			}
		}

	}
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
	logger.Printf("debug: Getting nextMbAddress for N %d", n)
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

func MbaffFrameFlag(sps *SPS, header *SliceHeader) int {
	if sps.MBAdaptiveFrameField && !header.FieldPic {
		return 1
	}
	return 0
}

// 7.3.4 - SliceData
// Decode with 8.2
func NewSliceData(sliceContext *SliceContext, b *BitReader) *SliceData {
	cabac := NewCABAC(sliceContext, b)
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
	// initContextVariables(sliceContext)
	if sliceContext.PPS.EntropyCodingMode == 1 {
		for !b.IsByteAligned() {
			sliceContext.Slice.Data.CabacAlignmentOneBit = b.NextField("CabacAlignmentOneBit", 1)
		}
	}
	// 6.3
	// Macroblock-adaptive frame field decoding flag; true indicating slices of pairs
	mbaffFrameFlag := MbaffFrameFlag(sliceContext.SPS, sliceContext.Header)
	sliceContext.Slice.Data.CurrMbAddr = sliceContext.Slice.Header.FirstMbInSlice * (1 * mbaffFrameFlag)
	logger.Printf("debug: macroblock-adaptive frame field is %d", mbaffFrameFlag)
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
				sliceContext.Slice.Data.MbSkipRun = ue(b.golomb())
				if sliceContext.Slice.Data.MbSkipRun > 0 {
					prevMbSkipped = 1
				}
				for i := 0; i < sliceContext.Slice.Data.MbSkipRun; i++ {
					// nextMbAddress(currMbAdd
					sliceContext.Slice.Data.CurrMbAddr = nextMbAddress(sliceContext.Slice.Data.CurrMbAddr, sliceContext.SPS, sliceContext.PPS, sliceContext.Slice.Header)
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
		decodeFrame(sliceContext)
		if moreDataFlag {
			if mbaffFrameFlag == 1 && (sliceContext.Slice.Data.CurrMbAddr%2 == 0 || (sliceContext.Slice.Data.CurrMbAddr%2 == 1 && prevMbSkipped == 1)) {
				if sliceContext.PPS.EntropyCodingMode == 1 {
					_ = cabac.GetBinarization("MbFieldDecodingFlag")
					// TODO: ae implementation
					binarization := NewBinarization("MbFieldDecodingFlag", sliceContext.Slice.Data)
					binarization.Decode(sliceContext, b, b.Bytes())

					logger.Printf("TODO: ae for MbFieldDecodingFlag\n")
				} else {
					sliceContext.Slice.Data.MbFieldDecodingFlag = flagField()
				}
			}

			// BEGIN: macroblockLayer() 7.3.5
			if sliceContext.PPS.EntropyCodingMode == 1 {
				logger.Printf("debug: parsing macroblock layer syntax elements")
				binarization := cabac.GetBinarization("MBType")
				// TODO: ae implementation
				v := cabac.Decode(binarization)
				sliceContext.Slice.Data.MbType = v
				logger.Printf("debug: macroblock type is %d: %s",
					v,
					MbTypeName(sliceTypeMap[sliceContext.Slice.Header.SliceType], v))
				/*
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
				*/
			} else {
				sliceContext.Slice.Data.MbType = ue(b.golomb())
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
				// cabac = initContextVariables(binarization, sliceContext)
				// 6-1 p 47
				mbWidthC := 16 / SubWidthC(sliceContext.SPS)
				mbHeightC := 16 / SubHeightC(sliceContext.SPS)
				// if monochrome
				if sliceContext.SPS.ChromaFormat == 0 || sliceContext.SPS.UseSeparateColorPlane {
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
				// cabac = initContextVariables(binarization, sliceContext)

			} else {
				noSubMbPartSizeLessThan8x8Flag := 1
				if sliceContext.Slice.Data.MbTypeName == "I_NxN" && MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0) != "Intra_16x16" && NumMbPart(sliceContext.NalUnit, sliceContext.SPS, sliceContext.Slice.Header, sliceContext.Slice.Data) == 4 {
					logger.Printf("\tTODO: subMbPred\n")
					sliceContext.Slice.Data.SubMbType = SubMbPred(sliceContext, b, sliceContext.Data.MbTypeName)

					for mbPartIdx := 0; mbPartIdx < 4; mbPartIdx++ {
						subMbTypeName := MbTypeName(sliceTypeMap[sliceContext.Slice.Header.SliceType], sliceContext.Slice.Data.SubMbType[mbPartIdx])
						if subMbTypeName != "B_Direct_8x8" {
							if NumSubMbPart(sliceContext.Slice.Data.SubMbType[mbPartIdx]) > 1 {
								noSubMbPartSizeLessThan8x8Flag = 0
							}
						} else if !sliceContext.SPS.Direct8x8Inference {
							noSubMbPartSizeLessThan8x8Flag = 0
						}
					}

				} else {
					if sliceContext.PPS.Transform8x8Mode == 1 && sliceContext.Slice.Data.MbTypeName == "I_NxN" {
						// TODO
						// 1 bit or ae(v)
						// If sliceContext.PPS.EntropyCodingMode == 1, use ae(v)
						if sliceContext.PPS.EntropyCodingMode == 1 {
							binarization := NewBinarization("TransformSize8x8Flag", sliceContext.Slice.Data)
							cabac = initContextVariables("TransformSize8x8Flag", binarization, sliceContext)
							binarization.Decode(sliceContext, b, b.Bytes())

							logger.Println("TODO: ae(v) for TransformSize8x8Flag")
						} else {
							sliceContext.Slice.Data.TransformSize8x8Flag = flagField()
						}
					}
					MbPred(sliceContext, sliceContext.Slice.Data.CurrMbAddr, b, b.Bytes())
				}
				if MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0) != "Intra_16x16" {
					// TODO: me, ae
					logger.Printf("TODO: CodedBlockPattern pending me/ae implementation\n")
					if sliceContext.PPS.EntropyCodingMode == 1 {
						binarization := NewBinarization("CodedBlockPattern", sliceContext.Slice.Data)
						cabac = initContextVariables("CodedBlockPattern", binarization, sliceContext)
						binarization.Decode(sliceContext, b, b.Bytes())

						logger.Printf("TODO: ae for CodedBlockPattern\n")
					} else {
						sliceContext.Slice.Data.CodedBlockPattern = me(
							b.golomb(),
							sliceContext.Slice.Header.ChromaArrayType,
							MbPartPredMode(sliceContext.Slice.Data, sliceContext.Slice.Data.SliceTypeName, sliceContext.Slice.Data.MbType, 0))
					}

					// sliceContext.Slice.Data.CodedBlockPattern = me(v) | ae(v)
					if CodedBlockPatternLuma(sliceContext.Slice.Data) > 0 && sliceContext.PPS.Transform8x8Mode == 1 && sliceContext.Slice.Data.MbTypeName != "I_NxN" && noSubMbPartSizeLessThan8x8Flag == 1 && (sliceContext.Slice.Data.MbTypeName != "B_Direct_16x16" || sliceContext.SPS.Direct8x8Inference) {
						// TODO: 1 bit or ae(v)
						if sliceContext.PPS.EntropyCodingMode == 1 {
							binarization := NewBinarization("Transform8x8Flag", sliceContext.Slice.Data)
							cabac = initContextVariables("Transform8x8Flag", binarization, sliceContext)
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
						cabac = initContextVariables("MbQpDelta", binarization, sliceContext)
						binarization.Decode(sliceContext, b, b.Bytes())

						logger.Printf("TODO: ae for MbQpDelta\n")
					} else {
						sliceContext.Slice.Data.MbQpDelta = se(b.golomb())
					}

				}
			}
			logger.Printf("debug: finished parsing macroblock layer syntax elements")
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
			if mbaffFrameFlag == 1 && sliceContext.Slice.Data.CurrMbAddr%2 == 0 {
				logger.Printf("debug: \tNon-I/SI: More sliceContext.Slice.Data at currMbAddr[%v] sliceContext.Slice.Data %d:%d:%d\n", sliceContext.Slice.Data.CurrMbAddr, b.byteOffset, b.bitOffset, len(b.Bytes()))
				moreDataFlag = true
			} else {
				// TODO: ae implementation
				sliceContext.Slice.Data.EndOfSliceFlag = flagField() // ae(b.golomb())
				logger.Printf("debug: \tNon-I/SI: End of slice[%v] %d:%d:%d\n", sliceContext.Slice.Data.EndOfSliceFlag, b.byteOffset, b.bitOffset, len(b.Bytes()))
				moreDataFlag = !sliceContext.Slice.Data.EndOfSliceFlag
			}
		}
		sliceContext.Slice.Data.CurrMbAddr = nextMbAddress(sliceContext.Slice.Data.CurrMbAddr, sliceContext.SPS, sliceContext.PPS, sliceContext.Slice.Header)
	} // END while moreDataFlag
	return sliceContext.Slice.Data
}

func (c *SliceContext) Update(header *SliceHeader, data *SliceData) {
	c.Slice = &Slice{Header: header, Data: data}
}
func NewSliceContext(videoStream *VideoStream, nalUnit *NalUnit, rbsp []byte, showPacket bool) *SliceContext {
	sps := videoStream.SPS
	pps := videoStream.PPS
	logger.Printf("debug: (%d)%s RBSP %d bytes %d bits == \n", nalUnit.Type, NALUnitType[nalUnit.Type], len(rbsp), len(rbsp)*8)
	logger.Printf("debug: \t%#v\n", rbsp[0:8])
	if nalUnit.NoInterLayerPredFlag == 0 {
		// g.6.2
		logger.Printf("debug: no inter layer prediction flag is off\n")
	}
	idrPic := nalUnit.Type == 5
	header := SliceHeader{}
	if sps.UseSeparateColorPlane {
		// 6.3
		logger.Printf("debug: using separate color planes; three slices per picture")
		header.ChromaArrayType = 0
	} else {
		// 6.3
		logger.Printf("debug: not using separate color plane; one macroblock per slice")
		header.ChromaArrayType = sps.ChromaFormat
	}
	logger.Printf("debug: chroma array type is %d\n", header.ChromaArrayType)
	b := &BitReader{bytes: rbsp}
	flagField := func() bool {
		if v := b.NextField("", 1); v == 1 {
			return true
		}
		return false
	}
	header.FirstMbInSlice = ue(b.golomb())
	header.SliceType = ue(b.golomb())
	sliceType := sliceTypeMap[header.SliceType]
	if sliceType == "B" {
		logger.Printf("error: B slices are unsupported. Dropping frame")
		goto DropFrame
	}
	logger.Printf("debug: %s (%s) slice of %d bytes\n", NALUnitType[nalUnit.Type], sliceType, len(rbsp))
	header.PPSID = ue(b.golomb())
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
		header.IDRPicID = ue(b.golomb())
		logger.Printf("debug: decoding IDR pic %d with pic order count type %d",
			header.IDRPicID, sps.PicOrderCountType)
	}

	if sps.PicOrderCountType == 0 {
		header.PicOrderCntLsb = b.NextField("PicOrderCntLsb", sps.Log2MaxPicOrderCntLSBMin4+4)
		if pps.BottomFieldPicOrderInFramePresent && !header.FieldPic {
			header.DeltaPicOrderCntBottom = se(b.golomb())
		}
	}
	if sps.PicOrderCountType == 1 && !sps.DeltaPicOrderAlwaysZero {
		header.DeltaPicOrderCnt[0] = se(b.golomb())
		if pps.BottomFieldPicOrderInFramePresent && !header.FieldPic {
			header.DeltaPicOrderCnt[1] = se(b.golomb())
		}
	}
	if pps.RedundantPicCntPresent {
		header.RedundantPicCnt = ue(b.golomb())
	}
	if sliceType == "B" {
		header.DirectSpatialMvPred = flagField()
	}
	if sliceType == "B" || sliceType == "SP" || sliceType == "B" {
		header.NumRefIdxActiveOverride = flagField()
		if header.NumRefIdxActiveOverride {
			header.NumRefIdxL0ActiveMinus1 = ue(b.golomb())
			if sliceType == "B" {
				header.NumRefIdxL1ActiveMinus1 = ue(b.golomb())
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
					header.ModificationOfPicNums = ue(b.golomb())
					if header.ModificationOfPicNums == 0 || header.ModificationOfPicNums == 1 {
						header.AbsDiffPicNumMinus1 = ue(b.golomb())
					} else if header.ModificationOfPicNums == 2 {
						header.LongTermPicNum = ue(b.golomb())
					}
				}
			}

		}
		if header.SliceType%5 == 1 {
			header.RefPicListModificationFlagL1 = flagField()
			if header.RefPicListModificationFlagL1 {
				for header.ModificationOfPicNums != 3 {
					header.ModificationOfPicNums = ue(b.golomb())
					if header.ModificationOfPicNums == 0 || header.ModificationOfPicNums == 1 {
						header.AbsDiffPicNumMinus1 = ue(b.golomb())
					} else if header.ModificationOfPicNums == 2 {
						header.LongTermPicNum = ue(b.golomb())
					}
				}
			}
		}
		// refPicListModification()
	}

	if (pps.WeightedPred && (sliceType == "P" || sliceType == "SP")) || (pps.WeightedBipred == 1 && sliceType == "B") {
		// predWeightTable()
		header.LumaLog2WeightDenom = ue(b.golomb())
		if header.ChromaArrayType != 0 {
			header.ChromaLog2WeightDenom = ue(b.golomb())
		}
		for i := 0; i <= header.NumRefIdxL0ActiveMinus1; i++ {
			header.LumaWeightL0Flag = flagField()
			if header.LumaWeightL0Flag {
				header.LumaWeightL0 = append(header.LumaWeightL0, se(b.golomb()))
				header.LumaOffsetL0 = append(header.LumaOffsetL0, se(b.golomb()))
			}
			if header.ChromaArrayType != 0 {
				header.ChromaWeightL0Flag = flagField()
				if header.ChromaWeightL0Flag {
					header.ChromaWeightL0 = append(header.ChromaWeightL0, []int{})
					header.ChromaOffsetL0 = append(header.ChromaOffsetL0, []int{})
					for j := 0; j < 2; j++ {
						header.ChromaWeightL0[i] = append(header.ChromaWeightL0[i], se(b.golomb()))
						header.ChromaOffsetL0[i] = append(header.ChromaOffsetL0[i], se(b.golomb()))
					}
				}
			}
		}
		if header.SliceType%5 == 1 {
			for i := 0; i <= header.NumRefIdxL1ActiveMinus1; i++ {
				header.LumaWeightL1Flag = flagField()
				if header.LumaWeightL1Flag {
					header.LumaWeightL1 = append(header.LumaWeightL1, se(b.golomb()))
					header.LumaOffsetL1 = append(header.LumaOffsetL1, se(b.golomb()))
				}
				if header.ChromaArrayType != 0 {
					header.ChromaWeightL1Flag = flagField()
					if header.ChromaWeightL1Flag {
						header.ChromaWeightL1 = append(header.ChromaWeightL1, []int{})
						header.ChromaOffsetL1 = append(header.ChromaOffsetL1, []int{})
						for j := 0; j < 2; j++ {
							header.ChromaWeightL1[i] = append(header.ChromaWeightL1[i], se(b.golomb()))
							header.ChromaOffsetL1[i] = append(header.ChromaOffsetL1[i], se(b.golomb()))
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
				header.MemoryManagementControlOperation = ue(b.golomb())
				for header.MemoryManagementControlOperation != 0 {
					if header.MemoryManagementControlOperation == 1 || header.MemoryManagementControlOperation == 3 {
						header.DifferenceOfPicNumsMinus1 = ue(b.golomb())
					}
					if header.MemoryManagementControlOperation == 2 {
						header.LongTermPicNum = ue(b.golomb())
					}
					if header.MemoryManagementControlOperation == 3 || header.MemoryManagementControlOperation == 6 {
						header.LongTermFrameIdx = ue(b.golomb())
					}
					if header.MemoryManagementControlOperation == 4 {
						header.MaxLongTermFrameIdxPlus1 = ue(b.golomb())
					}
				}
			}
		} // end decRefPicMarking
	}
	if pps.EntropyCodingMode == 1 && sliceType != "I" && sliceType != "SI" {
		header.CabacInit = ue(b.golomb())
	}
	header.SliceQpDelta = se(b.golomb())
	if sliceType == "SP" || sliceType == "SI" {
		if sliceType == "SP" {
			header.SpForSwitch = flagField()
		}
		header.SliceQsDelta = se(b.golomb())
	}
	if pps.DeblockingFilterControlPresent {
		header.DisableDeblockingFilter = ue(b.golomb())
		if header.DisableDeblockingFilter != 1 {
			header.SliceAlphaC0OffsetDiv2 = se(b.golomb())
			header.SliceBetaOffsetDiv2 = se(b.golomb())
		}
	}
	if pps.NumSliceGroupsMinus1 > 0 && pps.SliceGroupMapType >= 3 && pps.SliceGroupMapType <= 5 {
		header.SliceGroupChangeCycle = b.NextField(
			"SliceGroupChangeCycle",
			int(math.Ceil(math.Log2(float64(pps.PicSizeInMapUnitsMinus1/pps.SliceGroupChangeRateMinus1+1)))))
	}
DropFrame:
	sliceContext := &SliceContext{
		NalUnit: nalUnit,
		SPS:     sps,
		PPS:     pps,
		Slice: &Slice{
			Header: &header,
		},
	}
	sliceContext.Slice.Data = NewSliceData(sliceContext, b)
	// TODO: Decode slice data using 8.2
	if showPacket {
		debugPacket("debug: Header", sliceContext.Slice.Header)
		debugPacket("debug: Data", sliceContext.Slice.Data)
	}
	return sliceContext
}
