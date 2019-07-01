package h264

import "fmt"

const (
	PROFILE_IDC_BASELINE            = 66
	PROFILE_IDC_MAIN                = 77
	PROFILE_IDC_EXTENDED            = 88
	PROFILE_IDC_HIGH                = 100
	PROFILE_IDC_HIGH_10             = 110
	PROFILE_IDC_HIGH_422            = 122
	PROFILE_IDC_HIGH_444_PREDICTIVE = 244
)

var (
	ProfileIDC = map[int]string{
		PROFILE_IDC_BASELINE:            "Baseline",
		PROFILE_IDC_MAIN:                "Main",
		PROFILE_IDC_EXTENDED:            "Extended",
		PROFILE_IDC_HIGH:                "High",
		PROFILE_IDC_HIGH_10:             "High 10",
		PROFILE_IDC_HIGH_422:            "High 4:2:2",
		PROFILE_IDC_HIGH_444_PREDICTIVE: "High 4:4:4",
	}
)

// 7.3.2.11
func rbspTrailingBits(b *BitReader) {
	rbspStopOneBit := make([]int, 1)
	if _, err := b.Read(rbspStopOneBit); err != nil {
		fmt.Printf("error reading StopOneBit: %v\n", err)
	}
	// 7.2
	for !b.IsByteAligned() {
		// RBSPAlignmentZeroBit
		rbspAlignmentZeroBit := make([]int, 1)
		if _, err := b.Read(rbspAlignmentZeroBit); err != nil {
			fmt.Printf("error reading AligntmentZeroBit: %v\n", err)
			break
		}
	}
}
func NewRBSP(frame []byte) []byte {
	// TODO: NALUType 14,20,21 add padding to 3rd or 4th byte
	return frame[5:]
}

// TODO: Should be base-ten big endian bit arrays, not bytes
// ITU A.2.1.1 - Bit 9 is 1
func isConstrainedBaselineProfile(profile int, b []byte) bool {
	if profile != PROFILE_IDC_BASELINE {
		return false
	}
	if len(b) > 8 && b[8] == 1 {
		return true
	}
	return false
}

// ITU A2.4.2 - Bit 12 and 13 are 1
func isConstrainedHighProfile(profile int, b []byte) bool {
	if profile != PROFILE_IDC_HIGH {
		return false
	}
	if len(b) > 13 {
		if b[12] == 1 && b[13] == 1 {
			return true
		}
	}
	return false
}

// ITU A2.8 - Bit 11 is 1
func isHigh10IntraProfile(profile int, b []byte) bool {
	if profile != PROFILE_IDC_HIGH_10 {
		return false
	}
	if len(b) > 11 && b[11] == 1 {
		return true
	}
	return false
}
