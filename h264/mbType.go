package h264

import (
	"errors"
)

const MB_TYPE_INFERRED = 1000

var (
	ISliceMbType = map[int]string{
		0:  "I_NxN",
		1:  "I_16x16_0_0_0",
		2:  "I_16x16_1_0_0",
		3:  "I_16x16_2_0_0",
		4:  "I_16x16_3_0_0",
		5:  "I_16x16_0_1_0",
		6:  "I_16x16_1_1_0",
		7:  "I_16x16_2_1_0",
		8:  "I_16x16_3_1_0",
		9:  "I_16x16_0_2_0",
		10: "I_16x16_1_2_0",
		11: "I_16x16_2_2_0",
		12: "I_16x16_3_2_0",
		13: "I_16x16_0_0_1",
		14: "I_16x16_1_0_1",
		15: "I_16x16_2_0_1",
		16: "I_16x16_3_0_1",
		17: "I_16x16_0_1_1",
		18: "I_16x16_1_1_1",
		19: "I_16x16_2_1_1",
		20: "I_16x16_3_1_1",
		21: "I_16x16_0_2_1",
		22: "I_16x16_1_2_1",
		23: "I_16x16_2_2_1",
		24: "I_16x16_3_2_1",
		25: "I_PCM",
	}
	SISliceMbType = map[int]string{
		0: "SI",
	}
	PSliceMbType = map[int]string{
		0:                "P_L0_16x16",
		1:                "P_L0_16x8",
		2:                "P_L0_L0_8x16",
		3:                "P_8x8",
		4:                "P_8x8ref0",
		MB_TYPE_INFERRED: "P_Skip",
	}
	BSliceMbType = map[int]string{
		0:                "B_Direct_16x16",
		1:                "B_L0_16x16",
		2:                "B_L1_16x16",
		3:                "B_Bi_16x16",
		4:                "B_L0_L0_16x8",
		5:                "B_L0_L0_8x16",
		6:                "B_L1_L1_16x8",
		7:                "B_L1_L1_8x16",
		8:                "B_L0_L1_16x8",
		9:                "B_L0_L1_8x16",
		10:               "B_L1_L0_16x8",
		11:               "B_L1_L0_8x16",
		12:               "B_L0_Bi_16x8",
		13:               "B_L0_Bi_8x16",
		14:               "B_L1_Bi_16x8",
		15:               "B_L1_Bi_8x16",
		16:               "B_Bi_L0_16x8",
		17:               "B_Bi_L0_8x16",
		18:               "B_Bi_l1_16x8",
		19:               "B_Bi_L1_8x16",
		20:               "B_Bi_Bi_16x8",
		21:               "B_Bi_Bi_8x16",
		22:               "B_8x8",
		MB_TYPE_INFERRED: "B_Skip",
	}
)

func MbTypeName(sliceType string, mbType int) string {
	sliceTypeName := "NaSliceType"
	switch sliceType {
	case "I":
		sliceTypeName = ISliceMbType[mbType]
	case "SI":
		sliceTypeName = SISliceMbType[mbType]
	case "P":
		sliceTypeName = PSliceMbType[mbType]
	case "B":
		sliceTypeName = BSliceMbType[mbType]
	}
	return sliceTypeName
}

// Errors used by MbPartPredMode.
var (
	errNaMode    = errors.New("no mode for given slice and mb type")
	errPartition = errors.New("partition must be 0")
)

func MbPartPredMode(data *SliceData, sliceType string, mbType, partition int) (mbPartPredMode, error) {
	if partition == 0 {
		switch sliceType {
		case "I":
			if mbType == 0 {
				if data.TransformSize8x8Flag {
					return intra8x8, nil
				}
				return intra4x4, nil
			}
			if mbType > 0 && mbType < 25 {
				return intra16x16, nil
			}
			return -1, errNaMode
		case "SI":
			return intra4x4, nil
		case "P":
			fallthrough
		case "SP":
			if mbType >= 0 && mbType < 3 {
				return predL0, nil
			} else if mbType == 3 || mbType == 4 {
				return -1, errNaMode
			} else {
				return predL0, nil
			}
		case "B":
			switch mbType {
			case 0:
				return direct, nil
			case 1:
				fallthrough
			case 4:
				fallthrough
			case 5:
				fallthrough
			case 8:
				fallthrough
			case 9:
				fallthrough
			case 12:
				fallthrough
			case 13:
				return predL0, nil
			case 2:
				fallthrough
			case 6:
				fallthrough
			case 7:
				fallthrough
			case 10:
				fallthrough
			case 11:
				fallthrough
			case 14:
				fallthrough
			case 15:
				return predL1, nil
			case 22:
				return -1, errNaMode
			default:
				if mbType > 15 && mbType < 22 {
					return biPred, nil
				}
				return direct, nil
			}
		}
	}
	return -1, errPartition
}
