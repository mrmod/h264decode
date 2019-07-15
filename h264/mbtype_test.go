package h264

import (
	"testing"
)

func TestMbPartPredMode(t *testing.T) {
	tests := []struct {
		sliceType string
		mbType    int
		data      *SliceData
		want      mbPartPredMode
		err       error
	}{
		// Table 7-11 (I-slices).
		0:  {"I", 0, &SliceData{TransformSize8x8Flag: false}, intra4x4, nil},
		1:  {"I", 0, &SliceData{TransformSize8x8Flag: true}, intra8x8, nil},
		2:  {"I", 1, nil, intra16x16, nil},
		3:  {"I", 2, nil, intra16x16, nil},
		4:  {"I", 3, nil, intra16x16, nil},
		5:  {"I", 4, nil, intra16x16, nil},
		6:  {"I", 5, nil, intra16x16, nil},
		7:  {"I", 6, nil, intra16x16, nil},
		8:  {"I", 7, nil, intra16x16, nil},
		9:  {"I", 8, nil, intra16x16, nil},
		10: {"I", 9, nil, intra16x16, nil},
		11: {"I", 10, nil, intra16x16, nil},
		12: {"I", 11, nil, intra16x16, nil},
		13: {"I", 12, nil, intra16x16, nil},
		14: {"I", 13, nil, intra16x16, nil},
		15: {"I", 14, nil, intra16x16, nil},
		16: {"I", 15, nil, intra16x16, nil},
		17: {"I", 16, nil, intra16x16, nil},
		18: {"I", 17, nil, intra16x16, nil},
		19: {"I", 18, nil, intra16x16, nil},
		20: {"I", 19, nil, intra16x16, nil},
		21: {"I", 20, nil, intra16x16, nil},
		22: {"I", 21, nil, intra16x16, nil},
		23: {"I", 22, nil, intra16x16, nil},
		24: {"I", 23, nil, intra16x16, nil},
		25: {"I", 24, nil, intra16x16, nil},
		26: {"I", 25, nil, naMbPartPredMode, errNaMode},

		// Table 7-12 (SI-slices).
		27: {"SI", 0, nil, intra4x4, nil},

		// Table 7-13 (SP-slices).
		28: {"SP", 0, nil, predL0, nil},
		29: {"SP", 1, nil, predL0, nil},
		30: {"SP", 2, nil, predL0, nil},
		31: {"SP", 3, nil, naMbPartPredMode, errNaMode},
		32: {"SP", 4, nil, naMbPartPredMode, errNaMode},
		// TODO: work out what to do with inferred case.

		// Table 7-14 (B-slices).
		33: {"B", 0, nil, direct, nil},
		34: {"B", 1, nil, predL0, nil},
		35: {"B", 2, nil, predL1, nil},
		36: {"B", 3, nil, biPred, nil},
		37: {"B", 4, nil, predL0, nil},
		38: {"B", 5, nil, predL0, nil},
		39: {"B", 6, nil, predL1, nil},
		40: {"B", 7, nil, predL1, nil},
		41: {"B", 8, nil, predL0, nil},
		42: {"B", 9, nil, predL0, nil},
		43: {"B", 10, nil, predL1, nil},
		44: {"B", 11, nil, predL1, nil},
		45: {"B", 12, nil, predL0, nil},
		46: {"B", 13, nil, predL0, nil},
		47: {"B", 14, nil, predL1, nil},
		48: {"B", 15, nil, predL1, nil},
		49: {"B", 16, nil, biPred, nil},
		50: {"B", 17, nil, biPred, nil},
		51: {"B", 18, nil, biPred, nil},
		52: {"B", 19, nil, biPred, nil},
		53: {"B", 20, nil, biPred, nil},
		54: {"B", 21, nil, biPred, nil},
		55: {"B", 22, nil, naMbPartPredMode, errNaMode},

		// Test some cases where we expect error.
		// TODO: write some error test cases.
	}

	for i, test := range tests {
		m, err := MbPartPredMode(test.data, test.sliceType, test.mbType, 0)
		if err != test.err {
			t.Errorf("unexpected error %v from test %d", err, i)
		}

		if m != test.want {
			t.Errorf("did not get expected result for test %d.\nGot: %v\nWant: %v\n", i, m, test.want)
		}
	}
}
