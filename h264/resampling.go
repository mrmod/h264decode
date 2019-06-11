package h264

import "math"

type ReferenceLayer struct {
	Slice                  *Slice
	ChromaPhaseXPlus1Flag  int
	ChromaPhaseYPlus1      int
	FrameMbsOnlyFlag       int
	FieldPicFlag           int
	PicWidthInSamplesSubZ  int
	PicHeightInSamplesSubZ int
}

// Definition: Coded (representation): 3.29: Data element in coded form
// Definition: Coded field 3.25: coded Field
// Definition: Field: 3.50: Alternating rows of frame; 2: top, bottom
// Definition: Field macroblock 3.51: All mb of coded field are, some of mbAff field

// g.6.1
// returns
//   :mb address mbAddrRefLayer
//   :luma location (xB, yB) rel to upper-left
func ReferenceLayerMacroblock(xP, yP, fieldMbFlag int, refLayerFieldMbFlag, refLayerSubMbType []int) {

}

// g.6.3
// returns xRef16, yRef16
func ReferenceLayerSampleLocations(slice *SliceContext, chromaFlag, xp, yp, fieldMbFlag int) (int, int) {
	// TODO: Derive refLayerFieldPicFlag
	refLayerFieldPicFlag := 1
	// TODO: Derive refLayerFrameMbsOnlyFlag
	refLayerFrameMbsOnlyFlag := 1
	botFieldFlag := 0
	frameMbsOnlyFlag := 0
	if slice.SPS.FrameMbsOnly {
		frameMbsOnlyFlag = 1
	}
	if slice.Header.BottomField {
		botFieldFlag = 1
	}
	currDQId := DQId(slice.NalUnit)
	_ = currDQId
	// TODO: Let levelIdc be SPS.LevelIdc referred to be code slice NALU with DQID ((currDQID>>4)<<4)
	levelIdc := slice.SPS.Level
	// if chromaFlag == 0: SubL, if chromaFlag == 1: SubC
	var refW, refH int
	if chromaFlag == 0 {
		refW = PicWidthInSamplesSubL(slice.SPS.PicWidthInMbsMinus1 + 1)
		refH = PicHeightInSamplesSubL(PicHeightInMbs(slice.SPS, slice.Header))
	} else {
		refW = PicWidthInSamplesSubC(PicWidthInMbs(slice.SPS), MbWidthC(slice.SPS))
		refH = PicHeightInSamplesSubC(PicHeightInMbs(slice.SPS, slice.Header), MbHeightC(slice.SPS))
	}
	refH = refH * (1 + refLayerFieldPicFlag)
	// TODO: Create the reference layer sample
	scaledW := refW // TODO: ScaledRefLayerPicWidthInSamplesSubZ()
	scaledH := refH // TODO: ScaledRefLayerPicHeightInSamplesSubZ()

	if !slice.SPS.FrameMbsOnly && refLayerFrameMbsOnlyFlag == 1 {
		scaledH = scaledH / 2
	}
	refPhaseX := 0
	refPhaseY := 0
	phaseX := 0
	phaseY := 0
	subW := 1
	subH := 1
	// TODO: derive refLayerChromaPhaseXPlus1Flag
	refLayerChromaPhaseXPlus1Flag := 1
	refLayerChromaPhaseYPlus1 := 1
	chromaPhaseXPlus1Flag := 1
	chromaPhaseYPlus1 := 1
	if chromaFlag == 1 {
		refPhaseX = refLayerChromaPhaseXPlus1Flag - 1
		refPhaseY = refLayerChromaPhaseYPlus1 - 1
		phaseX = chromaPhaseXPlus1Flag - 1
		phaseY = chromaPhaseYPlus1 - 1
		subW = SubWidthC(slice.SPS)
		subH = SubHeightC(slice.SPS)
	}

	if refLayerFrameMbsOnlyFlag == 0 || !slice.SPS.FrameMbsOnly {
		if refLayerFrameMbsOnlyFlag == 1 {
			phaseY = phaseY + 4*botFieldFlag + 3 - subH
			refPhaseY = 2*refPhaseY + 2
		} else {
			phaseY = phaseY + 4*botFieldFlag
			refPhaseY = refPhaseY + 4*botFieldFlag
		}
	}
	shiftX := 16
	shiftY := 16
	if levelIdc <= 30 {
		// WARNING: Will floor the float
		shiftX = int((31 - math.Ceil(math.Log2(float64(refW)))))
		shiftY = int((31 - math.Ceil(math.Log2(float64(refH)))))
	}
	scaleX := ((refW << uint(shiftX)) + (scaledW >> 1)) / scaledW
	scaleY := ((refH << uint(shiftY)) + (scaledH >> 1)) / scaledH
	// TODO: Derive MinNoInterLayerPredFlag properly
	// TODO: G.7.3.2.1.4 SPS SVC extension syntax
	scaledLeftOffset := 1
	scaledTopOffset := 1
	offsetX := ScaledRefLayerLeftOffset(scaledLeftOffset) / subW
	addX := (((refW * (2 + phaseX)) << uint(shiftX-2)) + (scaledW >> 1)) / scaledW
	deltaX := 4 * (2 + refPhaseX)
	var offsetY, addY, deltaY int
	if refLayerFrameMbsOnlyFlag == 1 && slice.SPS.FrameMbsOnly {
		offsetY = ScaledRefLayerTopOffset(scaledTopOffset, frameMbsOnlyFlag) / subH
		addY = (((refH * (2 + phaseY)) << uint(shiftY-2)) + (scaledH >> 1)) / scaledH
		deltaY = 4 * (2 + refPhaseY)
	}
	if refLayerFrameMbsOnlyFlag == 0 || !slice.SPS.FrameMbsOnly {
		offsetY = ScaledRefLayerTopOffset(scaledTopOffset, frameMbsOnlyFlag) / (2 * subH)
		addY = (((refH * (2 + phaseY)) << uint(shiftY-3)) + (scaledH >> 1)) / scaledH
		deltaY = 2 * (2 + refPhaseY)
	}
	// TODO: 6.4.1 for (xM, yM)
	xM := 1
	_ = xM
	yM := 1
	xC := Xc(Xp(), yM, subW)
	fieldPicFlag := 0
	if slice.Header.FieldPic {
		fieldPicFlag = 1
	}

	yC := Yc(Yp(), yM, subH, fieldMbFlag, fieldPicFlag)
	// g.6.3 2-3
	if refLayerFrameMbsOnlyFlag == 0 || !slice.SPS.FrameMbsOnly {
		yC = yC >> uint(1-fieldMbFlag)
	}

	xRef16 := (((xC-offsetX)*scaleX + addX) >> uint(shiftX-4)) - deltaX
	yRef16 := (((yC-offsetY)*scaleY + addY) >> uint(shiftY-4)) - deltaY
	return xRef16, yRef16
}

// g.6.3 - 2.2
func Xc(xp, xm, subW int) int {
	return xp + (xm >> uint(subW-1))
}

// g.6.3 - 2.2
func Yc(yp, ym, subH, fieldMbFlag, fieldPicFlag int) int {
	return yp + (ym >> uint(subH-1+fieldMbFlag-fieldPicFlag))
}

// ScaledRefLayerLeftOffset equation g-66
func ScaledRefLayerLeftOffset(scaledLeftOffset int) int {
	return 2 * scaledLeftOffset
}

// ScaledRefLayerTopOffset equation g-68
func ScaledRefLayerTopOffset(scaledTopOffset, frameMbsOnlyFlag int) int {
	return 2 * scaledTopOffset * (2 - frameMbsOnlyFlag)
}

// TODO: Proper implementation
func Xp() int { return 1 }
func Yp() int { return 1 }
