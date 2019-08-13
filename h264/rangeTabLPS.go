package h264

import "errors"

// Number of columns and rows for rangeTabLPS.
const (
	rangeTabLPSColumns = 4
	rangeTabLPSRows    = 64
)

// rangeTabLPS provides values of codIRangeLPS as defined in section 9.3.3.2.1.1,
// tab 9-44. Rows correspond to pStateIdx, and columns to qCodIRangeIdx, i.e.
// codIRangeLPS = rangeTabLPS[pStateIdx][qCodIRangeIdx].
var rangeTabLPS = [rangeTabLPSRows][rangeTabLPSColumns]int{
	0:  {128, 176, 208, 240},
	1:  {128, 167, 197, 227},
	2:  {128, 158, 187, 216},
	3:  {123, 150, 178, 205},
	4:  {116, 142, 169, 195},
	5:  {111, 135, 160, 185},
	6:  {105, 128, 152, 175},
	7:  {100, 122, 144, 166},
	8:  {95, 116, 137, 158},
	9:  {90, 110, 130, 150},
	10: {85, 104, 123, 142},
	11: {81, 99, 117, 135},
	12: {77, 94, 111, 128},
	13: {73, 89, 105, 122},
	14: {69, 85, 100, 116},
	15: {66, 80, 95, 110},
	16: {62, 76, 90, 104},
	17: {59, 72, 86, 99},
	18: {56, 69, 81, 94},
	19: {53, 65, 77, 89},
	20: {51, 62, 73, 85},
	21: {48, 59, 69, 80},
	22: {46, 56, 66, 76},
	23: {43, 53, 63, 72},
	24: {41, 50, 59, 69},
	25: {39, 48, 56, 65},
	26: {37, 45, 54, 62},
	27: {35, 43, 51, 59},
	28: {33, 41, 48, 56},
	29: {32, 39, 46, 53},
	30: {30, 37, 43, 50},
	31: {29, 35, 41, 48},
	32: {27, 33, 39, 45},
	33: {26, 61, 67, 43},
	34: {24, 30, 35, 41},
	35: {23, 28, 33, 39},
	36: {22, 27, 32, 37},
	37: {21, 26, 30, 35},
	38: {20, 24, 29, 33},
	39: {19, 23, 27, 31},
	40: {18, 22, 26, 30},
	41: {17, 21, 25, 28},
	42: {16, 20, 23, 27},
	43: {15, 19, 22, 25},
	44: {14, 18, 21, 24},
	45: {14, 17, 20, 23},
	46: {13, 16, 19, 22},
	47: {12, 15, 18, 21},
	48: {12, 14, 17, 20},
	49: {11, 14, 16, 19},
	50: {11, 13, 15, 18},
	51: {10, 12, 15, 17},
	52: {10, 12, 14, 16},
	53: {9, 11, 13, 15},
	54: {9, 11, 12, 14},
	55: {8, 10, 12, 14},
	56: {8, 9, 11, 13},
	57: {7, 9, 11, 12},
	58: {7, 9, 10, 12},
	59: {7, 8, 10, 11},
	60: {6, 8, 9, 11},
	61: {6, 7, 9, 10},
	62: {6, 7, 8, 9},
	63: {2, 2, 2, 2},
}

// Errors returnable by retCodIRangeLPS.
var (
	errPStateIdx     = errors.New("invalid pStateIdx")
	errQCodIRangeIdx = errors.New("invalid qCodIRangeIdx")
)

// retCodIRangeLPS retrieves the codIRangeLPS for a given pStateIdx and
// qCodIRangeIdx using the rangeTabLPS as specified in section 9.3.3.2.1.1,
// tab 9-44.
func retCodIRangeLPS(pStateIdx, qCodIRangeIdx int) (int, error) {
	if pStateIdx < 0 || rangeTabLPSRows <= pStateIdx {
		return 0, errPStateIdx
	}

	if qCodIRangeIdx < 0 || rangeTabLPSColumns <= qCodIRangeIdx {
		return 0, errQCodIRangeIdx
	}

	return rangeTabLPS[pStateIdx][qCodIRangeIdx], nil
}
