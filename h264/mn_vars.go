package h264

type MN struct {
	M, N int
}

const NoCabacInitIdc = -1

// tables 9-12 to 9-13
var (
	// 0-39 : MB_Type
	// Maybe mapping all values in the range -128 to 128 to
	// a list of tuples for input vars would be less verbose
	// map[ctxIdx]MN
	MNVars = map[int]map[int]MN{
		0:  {NoCabacInitIdc: {20, -15}},
		1:  {NoCabacInitIdc: {2, 54}},
		2:  {NoCabacInitIdc: {3, 74}},
		3:  {NoCabacInitIdc: {20, -15}},
		4:  {NoCabacInitIdc: {2, 54}},
		5:  {NoCabacInitIdc: {3, 74}},
		6:  {NoCabacInitIdc: {-28, 127}},
		7:  {NoCabacInitIdc: {-23, 104}},
		8:  {NoCabacInitIdc: {-6, 53}},
		9:  {NoCabacInitIdc: {-1, 54}},
		10: {NoCabacInitIdc: {7, 51}},
		11: {
			0: {23, 33},
			1: {22, 25},
			2: {29, 16},
		},
		12: {
			0: {23, 2},
			1: {34, 0},
			2: {25, 0},
		},
		13: {
			0: {21, 0},
			1: {16, 0},
			2: {14, 0},
		},
		14: {
			0: {1, 9},
			1: {-2, 9},
			2: {-10, 51},
		},
		15: {
			0: {0, 49},
			1: {4, 41},
			2: {-3, 62},
		},
		16: {
			0: {-37, 118},
			1: {-29, 118},
			2: {-27, 99},
		},
		17: {
			0: {5, 57},
			1: {2, 65},
			2: {26, 16},
		},
		18: {
			0: {-13, 78},
			1: {-6, 71},
			2: {-4, 85},
		},
		19: {
			0: {-11, 65},
			1: {-13, 79},
			2: {-24, 102},
		},
		20: {
			0: {1, 62},
			1: {5, 52},
			2: {5, 57},
		},
		21: {
			0: {12, 49},
			1: {9, 50},
			2: {6, 57},
		},
		22: {
			0: {-4, 73},
			1: {-3, 70},
			2: {-17, 73},
		},
		23: {
			0: {17, 50},
			1: {10, 54},
			2: {14, 57},
		},
		// Table 9-14
		// Should use MNSecond to get the second M value if it exists
		// TODO: MNSecond determine when to provide second
		24: {
			0: {18, 64},
			1: {26, 34},
			2: {20, 40},
		},
		25: {
			0: {9, 43},
			1: {19, 22},
			2: {20, 10},
		},
		26: {
			0: {29, 0},
			1: {40, 0},
			2: {29, 0},
		},
		27: {
			0: {26, 67},
			1: {57, 2},
			2: {54, 0},
		},
		28: {
			0: {16, 90},
			1: {41, 36},
			2: {37, 42},
		},
		29: {
			0: {9, 104},
			1: {26, 59},
			2: {12, 97},
		},
		30: {
			0: {-4, 127}, // Second M: 6
			1: {-4, 127}, // Second M: 5
			2: {-3, 127}, // Second M: 2
		},
		31: {
			0: {-2, 104}, // Second M: 0
			1: {-1, 101}, // Second M: 5
			2: {-2, 117}, // Second M: 2
		},
		32: {
			0: {1, 67},
			1: {-4, 76},
			2: {-2, 74},
		},
		33: {
			0: {-1, 78}, // Second M: 3
			1: {-6, 71},
			2: {-4, 85},
		},
		34: {
			0: {-1, 65},  // Second M: 1
			1: {-1, 79},  // Second M: 3
			2: {-2, 102}, // Second M: 4
		},
		35: {
			0: {1, 62},
			1: {5, 52},
			2: {5, 57},
		},
		36: {
			0: {-6, 86},
			1: {6, 69},
			2: {-6, 93},
		},
		37: {
			0: {-1, 95}, // Second M: 7
			1: {-1, 90}, // Second M: 3
			2: {-1, 88}, // Second M: 4
		},
		38: {
			0: {-6, 61},
			1: {0, 52},
			2: {-6, 44},
		},
		39: {
			0: {9, 45},
			1: {8, 43},
			2: {4, 55},
		},
	}
)

// TODO: MNSecond determine when to provide second
func MNSecond(ctxIdx, cabacInitIdc int) {}

// Table 9-18
// Coded block pattern (luma y chroma)
// map[ctxIdx][cabacInitIdc]MN
func CodedblockPatternMN(ctxIdx, cabacInitIdc int, sliceType string) MN {
	var mn MN
	if sliceType != "I" && sliceType != "SI" {
		logger.Printf("warning: trying to initialize %s slice type\n", sliceType)
	}
	switch ctxIdx {
	case 70:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{0, 45}, {13, 15}, {7, 34},
			}[cabacInitIdc]
		}
		return MN{0, 11}
	case 71:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-4, 78}, {7, 51}, {-9, 88},
			}[cabacInitIdc]
		}
		return MN{1, 55}
	case 72:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-3, 96}, {2, 80}, {-20, 127},
			}[cabacInitIdc]
		}
		return MN{0, 69}
	case 73:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-27, 126}, {-39, 127}, {-36, 127},
			}[cabacInitIdc]
		}
		return MN{-17, 127}
	case 74:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-28, 98}, {-18, 91}, {-17, 91},
			}[cabacInitIdc]
		}
		return MN{-13, 102}
	case 75:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-25, 101}, {-17, 96}, {-14, 95},
			}[cabacInitIdc]
		}
		return MN{0, 82}
	case 76:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-23, 67}, {-26, 81}, {-25, 84},
			}[cabacInitIdc]
		}
		return MN{-7, 24}
	case 77:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-28, 82}, {-35, 98}, {-25, 86},
			}[cabacInitIdc]
		}
		return MN{-21, 107}
	case 78:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-20, 94}, {-24, 102}, {-12, 89},
			}[cabacInitIdc]
		}
		return MN{-27, 127}
	case 79:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-16, 83}, {-23, 97}, {-17, 91},
			}[cabacInitIdc]
		}
		return MN{-31, 127}
	case 80:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-22, 110}, {-27, 119}, {-31, 127},
			}[cabacInitIdc]
		}
		return MN{-24, 127}
	case 81:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-21, 91}, {-24, 99}, {-14, 76},
			}[cabacInitIdc]
		}
		return MN{-18, 95}
	case 82:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-18, 102}, {-21, 110}, {-18, 103},
			}[cabacInitIdc]
		}
		return MN{-27, 127}
	case 83:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-13, 93}, {-18, 102}, {-13, 90},
			}[cabacInitIdc]
		}
		return MN{-21, 114}
	case 84:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-29, 127}, {-36, 127}, {-37, 127},
			}[cabacInitIdc]
		}
		return MN{-30, 127}
	case 85:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-7, 92}, {0, 80}, {11, 80},
			}[cabacInitIdc]
		}
		return MN{-17, 123}
	case 86:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-5, 89}, {-5, 89}, {5, 76},
			}[cabacInitIdc]
		}
		return MN{-12, 115}
	case 87:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-7, 96}, {-7, 94}, {2, 84},
			}[cabacInitIdc]
		}
		return MN{-16, 122}
		// TODO: 88 to 104
	case 88:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-13, 108}, {-4, 92}, {5, 78},
			}[cabacInitIdc]
		}
		return MN{-11, 115}
	case 89:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-3, 46}, {0, 39}, {-6, 55},
			}[cabacInitIdc]
		}
		return MN{-12, 63}
	case 90:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-1, 65}, {0, 65}, {4, 61},
			}[cabacInitIdc]
		}
		return MN{-2, 68}
	case 91:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-1, 57}, {-15, 84}, {-14, 83},
			}[cabacInitIdc]
		}
		return MN{-15, 85}
	case 92:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-9, 93}, {-36, 127}, {-37, 127},
			}[cabacInitIdc]
		}
		return MN{-13, 104}
	case 93:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-3, 74}, {-2, 73}, {-5, 79},
			}[cabacInitIdc]
		}
		return MN{-3, 70}
	case 94:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-9, 92}, {-12, 104}, {-11, 104},
			}[cabacInitIdc]
		}
		return MN{-8, 93}
	case 95:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-8, 87}, {-9, 91}, {-11, 91},
			}[cabacInitIdc]
		}
		return MN{-10, 90}
	case 96:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-23, 126}, {-31, 127}, {-30, 127},
			}[cabacInitIdc]
		}
		return MN{-30, 127}
	case 97:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{5, 54}, {3, 55}, {0, 65},
			}[cabacInitIdc]
		}
		return MN{-1, 74}
	case 98:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{6, 60}, {7, 56}, {-2, 79},
			}[cabacInitIdc]
		}
		return MN{-6, 97}
	case 99:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{6, 59}, {7, 55}, {0, 72},
			}[cabacInitIdc]
		}
		return MN{-7, 91}
	case 100:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{6, 69}, {8, 61}, {-4, 92},
			}[cabacInitIdc]
		}
		return MN{-20, 127}
	case 101:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-1, 48}, {-3, 53}, {-6, 56},
			}[cabacInitIdc]
		}
		return MN{-4, 56}
	case 102:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{0, 68}, {0, 68}, {3, 68},
			}[cabacInitIdc]
		}
		return MN{-5, 82}
	case 103:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-4, 69}, {-7, 74}, {-8, 71},
			}[cabacInitIdc]
		}
		return MN{-7, 76}
	case 104:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				{-8, 88}, {-9, 88}, {-13, 98},
			}[cabacInitIdc]
		}
		return MN{-22, 125}

	}

	return mn
}
