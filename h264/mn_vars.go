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
	// map[ctxIdx][cabacInitIdc]MN
	MNVars = map[int]map[int]MN{
		0:  map[int]MN{NoCabacInitIdc: MN{20, -15}},
		1:  map[int]MN{NoCabacInitIdc: MN{2, 54}},
		2:  map[int]MN{NoCabacInitIdc: MN{3, 74}},
		3:  map[int]MN{NoCabacInitIdc: MN{20, -15}},
		4:  map[int]MN{NoCabacInitIdc: MN{2, 54}},
		5:  map[int]MN{NoCabacInitIdc: MN{3, 74}},
		6:  map[int]MN{NoCabacInitIdc: MN{-28, 127}},
		7:  map[int]MN{NoCabacInitIdc: MN{-23, 104}},
		8:  map[int]MN{NoCabacInitIdc: MN{-6, 53}},
		9:  map[int]MN{NoCabacInitIdc: MN{-1, 54}},
		10: map[int]MN{NoCabacInitIdc: MN{7, 51}},
		11: map[int]MN{
			0: MN{23, 33},
			1: MN{22, 25},
			2: MN{29, 16},
		},
		12: map[int]MN{
			0: MN{23, 2},
			1: MN{34, 0},
			2: MN{25, 0},
		},
		13: map[int]MN{
			0: MN{21, 0},
			1: MN{16, 0},
			2: MN{14, 0},
		},
		14: map[int]MN{
			0: MN{1, 9},
			1: MN{-2, 9},
			2: MN{-10, 51},
		},
		15: map[int]MN{
			0: MN{0, 49},
			1: MN{4, 41},
			2: MN{-3, 62},
		},
		16: map[int]MN{
			0: MN{-37, 118},
			1: MN{-29, 118},
			2: MN{-27, 99},
		},
		17: map[int]MN{
			0: MN{5, 57},
			1: MN{2, 65},
			2: MN{26, 16},
		},
		18: map[int]MN{
			0: MN{-13, 78},
			1: MN{-6, 71},
			2: MN{-4, 85},
		},
		19: map[int]MN{
			0: MN{-11, 65},
			1: MN{-13, 79},
			2: MN{-24, 102},
		},
		20: map[int]MN{
			0: MN{1, 62},
			1: MN{5, 52},
			2: MN{5, 57},
		},
		21: map[int]MN{
			0: MN{12, 49},
			1: MN{9, 50},
			2: MN{6, 57},
		},
		22: map[int]MN{
			0: MN{-4, 73},
			1: MN{-3, 70},
			2: MN{-17, 73},
		},
		23: map[int]MN{
			0: MN{17, 50},
			1: MN{10, 54},
			2: MN{14, 57},
		},
		// Table 9-14
		// Should use MNSecond to get the second M value if it exists
		// TODO: MNSecond determine when to provide second
		24: map[int]MN{
			0: MN{18, 64},
			1: MN{26, 34},
			2: MN{20, 40},
		},
		25: map[int]MN{
			0: MN{9, 43},
			1: MN{19, 22},
			2: MN{20, 10},
		},
		26: map[int]MN{
			0: MN{29, 0},
			1: MN{40, 0},
			2: MN{29, 0},
		},
		27: map[int]MN{
			0: MN{26, 67},
			1: MN{57, 2},
			2: MN{54, 0},
		},
		28: map[int]MN{
			0: MN{16, 90},
			1: MN{41, 36},
			2: MN{37, 42},
		},
		29: map[int]MN{
			0: MN{9, 104},
			1: MN{26, 59},
			2: MN{12, 97},
		},
		30: map[int]MN{
			0: MN{-4, 127}, // Second M: 6
			1: MN{-4, 127}, // Second M: 5
			2: MN{-3, 127}, // Second M: 2
		},
		31: map[int]MN{
			0: MN{-2, 104}, // Second M: 0
			1: MN{-1, 101}, // Second M: 5
			2: MN{-2, 117}, // Second M: 2
		},
		32: map[int]MN{
			0: MN{1, 67},
			1: MN{-4, 76},
			2: MN{-2, 74},
		},
		33: map[int]MN{
			0: MN{-1, 78}, // Second M: 3
			1: MN{-6, 71},
			2: MN{-4, 85},
		},
		34: map[int]MN{
			0: MN{-1, 65},  // Second M: 1
			1: MN{-1, 79},  // Second M: 3
			2: MN{-2, 102}, // Second M: 4
		},
		35: map[int]MN{
			0: MN{1, 62},
			1: MN{5, 52},
			2: MN{5, 57},
		},
		36: map[int]MN{
			0: MN{-6, 86},
			1: MN{6, 69},
			2: MN{-6, 93},
		},
		37: map[int]MN{
			0: MN{-1, 95}, // Second M: 7
			1: MN{-1, 90}, // Second M: 3
			2: MN{-1, 88}, // Second M: 4
		},
		38: map[int]MN{
			0: MN{-6, 61},
			1: MN{0, 52},
			2: MN{-6, 44},
		},
		39: map[int]MN{
			0: MN{9, 45},
			1: MN{8, 43},
			2: MN{4, 55},
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
				MN{0, 45}, MN{13, 15}, MN{7, 34},
			}[cabacInitIdc]
		}
		return MN{0, 11}
	case 71:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-4, 78}, MN{7, 51}, MN{-9, 88},
			}[cabacInitIdc]
		}
		return MN{1, 55}
	case 72:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-3, 96}, MN{2, 80}, MN{-20, 127},
			}[cabacInitIdc]
		}
		return MN{0, 69}
	case 73:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-27, 126}, MN{-39, 127}, MN{-36, 127},
			}[cabacInitIdc]
		}
		return MN{-17, 127}
	case 74:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-28, 98}, MN{-18, 91}, MN{-17, 91},
			}[cabacInitIdc]
		}
		return MN{-13, 102}
	case 75:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-25, 101}, MN{-17, 96}, MN{-14, 95},
			}[cabacInitIdc]
		}
		return MN{0, 82}
	case 76:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-23, 67}, MN{-26, 81}, MN{-25, 84},
			}[cabacInitIdc]
		}
		return MN{-7, 24}
	case 77:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-28, 82}, MN{-35, 98}, MN{-25, 86},
			}[cabacInitIdc]
		}
		return MN{-21, 107}
	case 78:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-20, 94}, MN{-24, 102}, MN{-12, 89},
			}[cabacInitIdc]
		}
		return MN{-27, 127}
	case 79:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-16, 83}, MN{-23, 97}, MN{-17, 91},
			}[cabacInitIdc]
		}
		return MN{-31, 127}
	case 80:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-22, 110}, MN{-27, 119}, MN{-31, 127},
			}[cabacInitIdc]
		}
		return MN{-24, 127}
	case 81:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-21, 91}, MN{-24, 99}, MN{-14, 76},
			}[cabacInitIdc]
		}
		return MN{-18, 95}
	case 82:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-18, 102}, MN{-21, 110}, MN{-18, 103},
			}[cabacInitIdc]
		}
		return MN{-27, 127}
	case 83:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-13, 93}, MN{-18, 102}, MN{-13, 90},
			}[cabacInitIdc]
		}
		return MN{-21, 114}
	case 84:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-29, 127}, MN{-36, 127}, MN{-37, 127},
			}[cabacInitIdc]
		}
		return MN{-30, 127}
	case 85:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-7, 92}, MN{0, 80}, MN{11, 80},
			}[cabacInitIdc]
		}
		return MN{-17, 123}
	case 86:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-5, 89}, MN{-5, 89}, MN{5, 76},
			}[cabacInitIdc]
		}
		return MN{-12, 115}
	case 87:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-7, 96}, MN{-7, 94}, MN{2, 84},
			}[cabacInitIdc]
		}
		return MN{-16, 122}
		// TODO: 88 to 104
	case 88:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-13, 108}, MN{-4, 92}, MN{5, 78},
			}[cabacInitIdc]
		}
		return MN{-11, 115}
	case 89:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-3, 46}, MN{0, 39}, MN{-6, 55},
			}[cabacInitIdc]
		}
		return MN{-12, 63}
	case 90:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-1, 65}, MN{0, 65}, MN{4, 61},
			}[cabacInitIdc]
		}
		return MN{-2, 68}
	case 91:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-1, 57}, MN{-15, 84}, MN{-14, 83},
			}[cabacInitIdc]
		}
		return MN{-15, 85}
	case 92:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-9, 93}, MN{-36, 127}, MN{-37, 127},
			}[cabacInitIdc]
		}
		return MN{-13, 104}
	case 93:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-3, 74}, MN{-2, 73}, MN{-5, 79},
			}[cabacInitIdc]
		}
		return MN{-3, 70}
	case 94:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-9, 92}, MN{-12, 104}, MN{-11, 104},
			}[cabacInitIdc]
		}
		return MN{-8, 93}
	case 95:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-8, 87}, MN{-9, 91}, MN{-11, 91},
			}[cabacInitIdc]
		}
		return MN{-10, 90}
	case 96:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-23, 126}, MN{-31, 127}, MN{-30, 127},
			}[cabacInitIdc]
		}
		return MN{-30, 127}
	case 97:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{5, 54}, MN{3, 55}, MN{0, 65},
			}[cabacInitIdc]
		}
		return MN{-1, 74}
	case 98:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{6, 60}, MN{7, 56}, MN{-2, 79},
			}[cabacInitIdc]
		}
		return MN{-6, 97}
	case 99:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{6, 59}, MN{7, 55}, MN{0, 72},
			}[cabacInitIdc]
		}
		return MN{-7, 91}
	case 100:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{6, 69}, MN{8, 61}, MN{-4, 92},
			}[cabacInitIdc]
		}
		return MN{-20, 127}
	case 101:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-1, 48}, MN{-3, 53}, MN{-6, 56},
			}[cabacInitIdc]
		}
		return MN{-4, 56}
	case 102:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{0, 68}, MN{0, 68}, MN{3, 68},
			}[cabacInitIdc]
		}
		return MN{-5, 82}
	case 103:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-4, 69}, MN{-7, 74}, MN{-8, 71},
			}[cabacInitIdc]
		}
		return MN{-7, 76}
	case 104:
		if cabacInitIdc >= 0 && cabacInitIdc <= 2 {
			return []MN{
				MN{-8, 88}, MN{-9, 88}, MN{-13, 98},
			}[cabacInitIdc]
		}
		return MN{-22, 125}

	}

	return mn
}
