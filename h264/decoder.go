package h264

import "math"

// 8.2
// TODO: WARNING partial implementation
// returns topFieldOrderCnt, bottomFieldOrderCnt
func decodeFrame(s *SliceContext) (int, int) {
	if s.Slice.Data.MbFieldDecodingFlag {
		logger.Printf("debug: decoding field macroblock pair")
	} else {
		logger.Printf("debug: decoding frame macroblock pair")
	}
	logger.Printf("debug: decoding slice layer picture type %d", s.SPS.PicOrderCountType)
	switch s.SPS.PicOrderCountType {

	case 0:
		// 8.2.1.1
		if s.NalUnit.Type == IDR_PICTURE {
			return decodePictureType0(s)
		}
		logger.Printf("error: decoding of NAL %s not supported for picOrderCntType %d",
			NALUnitType[s.NalUnit.Type], s.SPS.PicOrderCountType)
	case 1:
		// 8.2.1.2
		logger.Printf("error: decoding of NAL %s not supported for picOrderCntType %d",
			NALUnitType[s.NalUnit.Type], s.SPS.PicOrderCountType)
		break

	case 2:
		// 8.2.1.3
		if s.NalUnit.Type == IDR_PICTURE {
			return decodePictureType2(s)
		}
		logger.Printf("error: decoding of NAL %s not supported for picOrderCntType %d",
			NALUnitType[s.NalUnit.Type], s.SPS.PicOrderCountType)
	}
	logger.Printf("error: frame has not been properly decoded")
	logger.Printf("error: decoding of NAL %s not supported for picOrderCntType %d",
		NALUnitType[s.NalUnit.Type], s.SPS.PicOrderCountType)
	return -1, -1
}

// 8.2.1.3
// returns either or both topFieldOrderCnt or BottomFieldOrderCnt
func decodePictureType2(s *SliceContext) (int, int) {
	var prevFrameNumOffset, frameNumOffset int
	var tempPicOrderCnt, maxFrameNum int
	var topFieldOrderCnt, bottomFieldOrderCnt int
	var idrPicFlag int
	// TODO: Should equal FrameNum of previous picture in decoding order
	prevFrameNum := s.Slice.Header.FrameNum
	maxFrameNum = MaxFrameNum(s)
	if s.NalUnit.Type == IDR_PICTURE {
		idrPicFlag = 1
	}
	// TODO: If not IDR
	// eq 8-11
	if idrPicFlag == 1 {
		frameNumOffset = 0
	} else if prevFrameNum > s.Slice.Header.FrameNum {
		frameNumOffset = prevFrameNumOffset + maxFrameNum
	} else {
		frameNumOffset = prevFrameNumOffset
	}
	// eq 8-12
	if idrPicFlag == 1 {
		tempPicOrderCnt = 0
	} else if s.NalUnit.RefIdc == 0 {
		tempPicOrderCnt = 2*(frameNumOffset+s.Slice.Header.FrameNum) - 1
	} else {
		tempPicOrderCnt = 2 * (frameNumOffset + s.Slice.Header.FrameNum)
	}
	// eq 8-13
	if !s.Slice.Header.FieldPic {
		topFieldOrderCnt = tempPicOrderCnt
		bottomFieldOrderCnt = tempPicOrderCnt
	} else if s.Slice.Header.BottomField {
		bottomFieldOrderCnt = tempPicOrderCnt
	} else {
		topFieldOrderCnt = tempPicOrderCnt
	}
	logger.Printf("debug: top/bottom field order cnt: %d/%d", topFieldOrderCnt, bottomFieldOrderCnt)

	return topFieldOrderCnt, bottomFieldOrderCnt
}

// 8.2.1.2
// returns either or both topFieldOrderCnt or BottomFieldOrderCnt
func decodePictureType1(s *SliceContext) (int, int) {
	var prevFrameNumOffset, frameNumOffset int
	var topFieldOrderCnt, bottomFieldOrderCnt int
	var absFrameNum, maxFrameNum int
	var picOrderCntCycleCnt, frameNumInPicOrderCntCycle int
	var expectedPicOrderCnt, expectedDeltaPerPicOrderCntCycle int
	var idrPicFlag int
	// TODO: Should equal FrameNum of previous picture in decoding order
	// TODO: Derive for non-IDR picture
	prevFrameNum := s.Slice.Header.FrameNum
	maxFrameNum = MaxFrameNum(s)
	expectedDeltaPerPicOrderCntCycle = ExpectedDeltaPerPicOrderCntCycle(s)
	if s.NalUnit.Type == IDR_PICTURE {
		idrPicFlag = 1
	}
	// eq 8-6
	if idrPicFlag == 1 {
		frameNumOffset = 0
	} else if prevFrameNum > s.Slice.Header.FrameNum {
		frameNumOffset = prevFrameNumOffset + maxFrameNum
	}
	// eq 8-7
	if s.SPS.NumRefFramesInPicOrderCntCycle != 0 {
		absFrameNum = frameNumOffset + s.Slice.Header.FrameNum
	} else {
		absFrameNum = 0
	}
	if s.NalUnit.RefIdc == 0 && absFrameNum > 0 {
		absFrameNum = absFrameNum - 1
	}
	// eq 8-8
	if absFrameNum > 0 {
		picOrderCntCycleCnt = (absFrameNum - 1) / s.SPS.NumRefFramesInPicOrderCntCycle
		frameNumInPicOrderCntCycle = (absFrameNum - 1) % s.SPS.NumRefFramesInPicOrderCntCycle
	}
	// eq 8-9
	if absFrameNum > 0 {
		expectedPicOrderCnt = picOrderCntCycleCnt * expectedDeltaPerPicOrderCntCycle
		for i := 0; i < frameNumInPicOrderCntCycle; i++ {
			expectedPicOrderCnt = expectedPicOrderCnt + s.SPS.OffsetForRefFrameList[i]
		}
	} else {
		expectedPicOrderCnt = 0
	}
	if s.NalUnit.RefIdc == 0 {
		expectedPicOrderCnt = expectedPicOrderCnt + s.SPS.OffsetForNonRefPic
	}
	// eq 8-10
	if !s.Slice.Header.FieldPic {
		topFieldOrderCnt = expectedPicOrderCnt + s.Slice.Header.DeltaPicOrderCnt[0]
		bottomFieldOrderCnt = topFieldOrderCnt + s.SPS.OffsetForTopToBottomField + s.Slice.Header.DeltaPicOrderCnt[1]
	} else if !s.Slice.Header.BottomField {
		topFieldOrderCnt = expectedPicOrderCnt + s.Slice.Header.DeltaPicOrderCnt[0]
	} else {
		bottomFieldOrderCnt = expectedPicOrderCnt + s.SPS.OffsetForTopToBottomField + s.Slice.Header.DeltaPicOrderCnt[0]
	}
	logger.Printf("debug: top/bottom field order cnt: %d/%d", topFieldOrderCnt, bottomFieldOrderCnt)

	return topFieldOrderCnt, bottomFieldOrderCnt
}

// 8.2.1.1
// returns either or both topFieldOrderCnt or BottomFieldOrderCnt
func decodePictureType0(s *SliceContext) (int, int) {
	var picOrderCntMsb, prevPicOrderCntMsb, prevPicOrderCntLsb int
	var topFieldOrderCnt, bottomFieldOrderCnt int
	maxPicOrderCntLsb := MaxPicOrderCntLsb(s)
	picOrderCntLsb := s.Header.PicOrderCntLsb
	if (picOrderCntLsb < prevPicOrderCntLsb) && ((prevPicOrderCntLsb - picOrderCntLsb) >= (maxPicOrderCntLsb / 2)) {
		picOrderCntMsb = prevPicOrderCntMsb + maxPicOrderCntLsb
	} else if (picOrderCntLsb > prevPicOrderCntLsb) && ((picOrderCntLsb - prevPicOrderCntLsb) > (maxPicOrderCntLsb / 2)) {
		picOrderCntMsb = prevPicOrderCntMsb - maxPicOrderCntLsb
	} else {
		picOrderCntMsb = prevPicOrderCntMsb
	}
	// Pic is not a bottom field
	if !s.Slice.Header.BottomField {
		topFieldOrderCnt = picOrderCntMsb + picOrderCntLsb
	}
	if !s.Slice.Header.FieldPic {
		bottomFieldOrderCnt = topFieldOrderCnt + s.Slice.Header.DeltaPicOrderCntBottom
	} else {
		bottomFieldOrderCnt = picOrderCntMsb + picOrderCntLsb
	}
	logger.Printf("debug: top/bottom field order cnt: %d/%d", topFieldOrderCnt, bottomFieldOrderCnt)
	return topFieldOrderCnt, bottomFieldOrderCnt

	// TODO: Derive non-IDR picture
	/*
		prevFrameNum := s.Header.FrameNum
		prevFrameNumOffset := 0

			var maxFrameNum int
			if s.NalUnit.Type != IDR_PICTURE {
				// TODO: Derive PrevFrameNum via 8.2.1.3
			}
			frameNumOffset := FrameNumOffset(s, prevFrameNum, prevFrameNumOffset, maxFrameNum)
			_ = frameNumOffset
	*/
}

// eq 7-10
func MaxFrameNum(s *SliceContext) int {
	f := math.Pow(float64(2), float64(s.SPS.Log2MaxFrameNumMinus4+4))
	// Safe
	return int(f)
}

// eq 7-11
func MaxPicOrderCntLsb(s *SliceContext) int {
	f := math.Pow(float64(2), float64(s.SPS.Log2MaxPicOrderCntLSBMin4+4))
	// Safe
	return int(f)
}

// eq 7-12
func ExpectedDeltaPerPicOrderCntCycle(s *SliceContext) int {
	v := 0
	for i := 0; i < s.SPS.NumRefFramesInPicOrderCntCycle; i++ {
		v += s.SPS.OffsetForRefFrameList[i]
	}
	return v
}

// eq 8-13
func TopBottomFieldCnt(s *SliceContext, tempPicOrderCnt int) (int, int) {
	var topFieldOrderCnt, bottomFieldOrderCnt int
	if !s.Header.FieldPic {
		topFieldOrderCnt = tempPicOrderCnt
		bottomFieldOrderCnt = tempPicOrderCnt

	} else if s.Header.BottomField {
		bottomFieldOrderCnt = tempPicOrderCnt
	} else {
		topFieldOrderCnt = tempPicOrderCnt
	}
	// TODO: At times, only one value may be real. Pointers to ints would help
	//       as nil could represent emptiness
	return topFieldOrderCnt, bottomFieldOrderCnt
}

// eq 8-11
func FrameNumOffset(s *SliceContext, prevFrameNum, prevFrameNumOffset, maxFrameNum int) int {
	if s.NalUnit.Type == IDR_PICTURE {
		return 0
	}
	if prevFrameNum > s.Header.FrameNum {
		return prevFrameNumOffset + maxFrameNum
	}
	return prevFrameNumOffset
}

// eq 8-3
func PicOrderCntMsb(s *SliceContext, prevPicOrderCntMsb, prevPicOrderCntLsb, maxPicOrderCntLsb int) int {
	picOrderCntLsb := s.Header.PicOrderCntLsb
	// TODO: Calculate real values
	if picOrderCntLsb < prevPicOrderCntLsb && ((prevPicOrderCntLsb - picOrderCntLsb) >= (maxPicOrderCntLsb / 2)) {
		return prevPicOrderCntMsb + maxPicOrderCntLsb
	} else if (picOrderCntLsb > prevPicOrderCntLsb) && ((picOrderCntLsb - prevPicOrderCntLsb) > (maxPicOrderCntLsb / 2)) {
		return prevPicOrderCntMsb - maxPicOrderCntLsb
	}
	return prevPicOrderCntMsb
}

// eq 8-4
func TopFieldOrderCnt(s *SliceContext, picOrderCntMsb int) int {
	return picOrderCntMsb + s.Header.PicOrderCntLsb
}

// eq 8-5
func BottomFieldOrderCnt(s *SliceContext, picOrderCntMsb, topFieldOrderCnt int) int {
	if !s.Header.FieldPic {
		return topFieldOrderCnt + s.Header.DeltaPicOrderCntBottom
	}
	return picOrderCntMsb + s.Header.PicOrderCntLsb
}
