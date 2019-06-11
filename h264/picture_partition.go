package h264

// 6.4.1
// lumaXyForMb
// returns (x,y) of upper-left luma sample
func InverseMBScan(slice *SliceContext, mbAffFrameFlag int, mbAddr int) (int, int) {
	var x, y int
	if mbAffFrameFlag == 0 {
		// eq 6-3
		x = InverseRasterScan(
			mbAddr,
			16,
			16,
			PicWidthInSamplesSubL(PicWidthInMbs(slice.SPS)),
			0)
		// eq 6-4
		y = InverseRasterScan(
			mbAddr,
			16,
			16,
			PicWidthInSamplesSubL(PicWidthInMbs(slice.SPS)),
			1)
		return x, y
	} else {
		// eq 6-5
		x0 := InverseRasterScan(
			mbAddr/2,
			16,
			32,
			PicWidthInSamplesSubL(PicWidthInMbs(slice.SPS)),
			0)
		// eq 6-6
		y0 := InverseRasterScan(
			mbAddr/2,
			16,
			32,
			PicWidthInSamplesSubL(PicWidthInMbs(slice.SPS)),
			1)
		if slice.SPS.FrameMbsOnly || !slice.Slice.Data.MbFieldDecodingFlag {
			// eq 6-7
			x = x0
			// eq 6-8
			y = y0 + (mbAddr%2)*16
		} else {
			// eq 6-9
			x = x0
			// eq 6-10
			y = y0 + (mbAddr % 2)
		}
	}
	return x, y
}

// 6.4.2.1
// output is (x,y) of upper-left luma for mb partition mbPartIdx
//     relative to upper-left sample of MB (see fig 6-9)
func InverseMBPartitionScanning(s *SliceContext, mbPartIdx int) (int, int) {
	// eq 6-11
	x := InverseRasterScan(mbPartIdx, MbPartWidth(s), MbPartHeight(s), 16, 0)
	// eq 6-12
	y := InverseRasterScan(mbPartIdx, MbPartWidth(s), MbPartHeight(s), 16, 1)
	return x, y
}

// 6.4.2.2
// output is (x,y) of upper-left luma sample for sub-mb partition
//     subMbPartIdx relative to upper-left sample of sub-mb
// NOTE: May need subMbPartList[subMbPartIdx] value instead of the index
func InverseSubMBPartitionScanning(s *SliceContext, mbPartIdx, subMbPartIdx int) (int, int) {
	sliceType := sliceTypeMap[s.Slice.Header.SliceType]
	mbType := MbTypeName(sliceType, s.Slice.Data.MbType)
	var x, y int
	if mbType == "P_8x8" || mbType == "P_8x8ref0" || mbType == "B_8x8" {
		// eq 6-13
		x = InverseRasterScan(
			subMbPartIdx,
			SubMbPartWidth(s.Slice.Data.SubMbType[mbPartIdx]),
			SubMbPartHeight(s.Slice.Data.SubMbType[mbPartIdx]),
			8, 0)
		// eq 6-14
		y = InverseRasterScan(
			subMbPartIdx,
			SubMbPartWidth(s.Slice.Data.SubMbType[mbPartIdx]),
			SubMbPartHeight(s.Slice.Data.SubMbType[mbPartIdx]),
			8, 1)
	} else {
		// eq 6-15
		x = InverseRasterScan(subMbPartIdx, 4, 4, 8, 0)
		// eq 6-16
		y = InverseRasterScan(subMbPartIdx, 4, 4, 8, 1)
	}
	return x, y
}

// 6.4.3
func Inverse4x4LumaBlockScanning(luma4x4BlkIdx int) (int, int) {
	// eq 6-17
	xa := InverseRasterScan(luma4x4BlkIdx/4, 8, 8, 16, 0)
	xb := InverseRasterScan(luma4x4BlkIdx%4, 4, 4, 8, 0)
	x := xa + xb
	// eq 6-18
	ya := InverseRasterScan(luma4x4BlkIdx/4, 8, 8, 16, 1)
	yb := InverseRasterScan(luma4x4BlkIdx%4, 4, 4, 8, 1)
	y := ya + yb
	return x, y
}

// 6.4.5
func Inverse8x8LumaBlockScanning(luma8x8BlkIdx int) (int, int) {
	// eq 6-19
	x := InverseRasterScan(luma8x8BlkIdx, 8, 8, 16, 0)
	// eq 6-20
	y := InverseRasterScan(luma8x8BlkIdx, 8, 8, 16, 1)
	return x, y
}

// 6.4.7
func Inverse4x4ChromaBlockScanning(chroma4x4BlkIdx int) (int, int) {
	// eq 6-21
	x := InverseRasterScan(chroma4x4BlkIdx, 4, 4, 8, 0)
	// eq 6-22
	y := InverseRasterScan(chroma4x4BlkIdx, 4, 4, 8, 1)
	return x, y
}

// 6.4.8
func IsMacroblockAvailable(s *SliceContext, mbAddr int) bool {
	if mbAddr < 0 {
		return false
	}
	if mbAddr > s.Slice.Data.CurrMbAddr {
		return false
	}
	// TODO: False if MB belongs to different slice than CurrMBAddr. How to tell?
	return true
}

type Macroblock struct {
	Address     int
	IsAvailable bool
}

const (
	MbAddrA = iota
	MbAddrB
	MbAddrC
	MbAddrD
)

// 6.4.9, 6,4,10
// returns [A,B,C,D] macroblock addresses
func NeighborMacroblocks(s *SliceContext, mbAddr int) []Macroblock {
	mbAddrs := make([]Macroblock, 4)
	currMbAddr := s.Slice.Data.CurrMbAddr
	picWidthInMbs := PicWidthInMbs(s.SPS)
	mbAffFrameFlag := MbaffFrameFlag(s.SPS, s.Slice.Header)
	if mbAffFrameFlag == 0 {
		// 6.4.9
		mbAddrs[MbAddrA] = Macroblock{
			Address:     currMbAddr - 1,
			IsAvailable: IsMacroblockAvailable(s, currMbAddr-1),
		}
		if currMbAddr%picWidthInMbs == 0 {
			mbAddrs[MbAddrA].IsAvailable = false
		}

		mbAddrs[MbAddrB] = Macroblock{
			Address:     currMbAddr - picWidthInMbs,
			IsAvailable: IsMacroblockAvailable(s, currMbAddr-picWidthInMbs),
		}

		mbAddrs[MbAddrC] = Macroblock{
			Address:     currMbAddr - picWidthInMbs + 1,
			IsAvailable: IsMacroblockAvailable(s, currMbAddr-picWidthInMbs+1),
		}
		if (currMbAddr+1)%picWidthInMbs == 0 {
			mbAddrs[MbAddrC].IsAvailable = false
		}

		mbAddrs[MbAddrD] = Macroblock{
			Address:     currMbAddr - picWidthInMbs - 1,
			IsAvailable: IsMacroblockAvailable(s, currMbAddr-picWidthInMbs-1),
		}
		if currMbAddr%picWidthInMbs == 0 {
			mbAddrs[MbAddrD].IsAvailable = false
		}
	} else {
		// 6.4.10
		mbAddrs[MbAddrA] = Macroblock{
			Address:     2 * (currMbAddr/2 - 1),
			IsAvailable: IsMacroblockAvailable(s, 2*(currMbAddr/2-1)),
		}
		if (currMbAddr/2)%picWidthInMbs == 0 {
			mbAddrs[MbAddrA].IsAvailable = false
		}

		mbAddrs[MbAddrB] = Macroblock{
			Address:     2 * (currMbAddr/2 - picWidthInMbs),
			IsAvailable: IsMacroblockAvailable(s, 2*(currMbAddr/2-picWidthInMbs)),
		}

		mbAddrs[MbAddrC] = Macroblock{
			Address:     2 * (currMbAddr/2 - picWidthInMbs + 1),
			IsAvailable: IsMacroblockAvailable(s, 2*(currMbAddr/2-picWidthInMbs+1)),
		}
		if (currMbAddr/2+1)%picWidthInMbs == 0 {
			mbAddrs[MbAddrC].IsAvailable = false
		}
		mbAddrs[MbAddrD] = Macroblock{
			Address:     2 * (currMbAddr/2 - picWidthInMbs - 1),
			IsAvailable: IsMacroblockAvailable(s, 2*(currMbAddr/2-picWidthInMbs-1)),
		}
		if (currMbAddr/2)%picWidthInMbs == 0 {
			mbAddrs[MbAddrD].IsAvailable = false
		}
	}
	return mbAddrs
}

// 6.4.11.4
func Neighbor4x4LumaBlock(blkIdx int) {
	logger.Printf("debug: finding neighbor luma4x4 blocks of %d", blkIdx)
}

// table 7-13, table 7-11
func MbPartWidth(s *SliceContext) int {
	// table 7-13
	switch s.Slice.Data.MbType {
	case 0:
		fallthrough
	case 1:
		return 16
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		return 8

	}

	return 16
}

// table 7-13, table 7-11
func MbPartHeight(s *SliceContext) int {
	// table 7-13
	switch s.Slice.Data.MbType {
	case 0:
		return 16
	case 1:
		return 8
	case 2:
		return 16
	case 3:
		fallthrough
	case 4:
		return 8
	default:
		return 16
	}
}

// 5-7
func InverseRasterScan(a, b, c, d, e int) int {
	if e == 0 {
		return (a % (d / b)) * b
	}
	return (a / (d / b)) * c
}

// picture width for luma
func PicWidthInSamplesSubL(picWidthInMbs int) int {
	return picWidthInMbs * 16
}

// picture width for chroma
func PicWidthInSamplesSubC(picWidthInMbs, mbWidthC int) int {
	return picWidthInMbs * mbWidthC
}

func PicHeightInSamplesSubL(picHeightInMbs int) int {
	return picHeightInMbs * 16
}

func PicHeightInSamplesSubC(picHeightInMbs, mbHeightC int) int {
	return picHeightInMbs * mbHeightC
}

func MbWidthC(sps *SPS) int {
	mbWidthC := 16 / SubWidthC(sps)
	if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
		mbWidthC = 0
	}
	return mbWidthC
}
func MbHeightC(sps *SPS) int {
	mbHeightC := 16 / SubHeightC(sps)
	if sps.ChromaFormat == 0 || sps.UseSeparateColorPlane {
		mbHeightC = 0
	}
	return mbHeightC
}

// table 6-1
func SubWidthC(sps *SPS) int {
	n := 17
	if sps.UseSeparateColorPlane {
		if sps.ChromaFormat == 3 {
			return n
		}
	}

	switch sps.ChromaFormat {
	case 0:
		return n
	case 1:
		n = 2
	case 2:
		n = 2
	case 3:
		n = 1

	}
	return n
}
func SubHeightC(sps *SPS) int {
	n := 17
	if sps.UseSeparateColorPlane {
		if sps.ChromaFormat == 3 {
			return n
		}
	}
	switch sps.ChromaFormat {
	case 0:
		return n
	case 1:
		n = 2
	case 2:
		n = 1
	case 3:
		n = 1

	}
	return n
}

// G.8.6.2.2.2
func Xr(x, xOffset, refMbW int) int {
	return (x + xOffset) % refMbW
}
func Yr(y, yOffset, refMbH int) int {
	return (y + yOffset) % refMbH
}

// G.8.6.2.2.2
func Xd(xr, refMbW int) int {
	if xr >= refMbW/2 {
		return xr - refMbW
	}
	return xr + 1
}
func Yd(yr, refMbH int) int {
	if yr >= refMbH/2 {
		return yr - refMbH
	}
	return yr + 1
}
func Ya(yd, refMbH, signYd int) int {
	return yd - (refMbH/2+1)*signYd
}

// table 6-2
// returnx (xD, yD)
func XdYd(n string, predPartWidth int) (int, int) {
	switch n {
	case "A":
		return -1, 0
	case "B":
		return 0, -1
	case "C":
		return predPartWidth, -1
	}
	// D
	return -1, -1
}

// 6.4.12
// returns mbAddrN, (xW, yW)
func NeighborLocations(s *SliceContext, mbAddr, lumaFlag, xN, yN int) (int, int, int) {
	var mbAddrN int
	var xW, yW int
	// eq 6-31
	maxW := 16
	maxH := 16

	if lumaFlag == 0 {
		// eq 6-32
		maxW = MbWidthC(s.SPS)
		// eq 6-33
		maxH = MbHeightC(s.SPS)
	}

	macroblocks := NeighborMacroblocks(s, mbAddr)

	if MbaffFrameFlag(s.SPS, s.Slice.Header) == 0 {
		// 6.4.12.1
		// table 6-3
		if xN < 0 && yN < 0 {
			mbAddrN = macroblocks[MbAddrD].Address
		} else if xN < 0 && yN >= 0 && yN < maxH {
			mbAddrN = macroblocks[MbAddrA].Address
		} else if xN >= 0 && xN < maxW && yN < 0 {
			mbAddrN = macroblocks[MbAddrB].Address
		} else if xN >= 0 && xN < maxW && yN >= 0 && yN < maxH {
			mbAddrN = s.Slice.Data.CurrMbAddr
		} else if xN > maxW-1 && yN < 0 {
			mbAddrN = macroblocks[MbAddrC].Address
		} else if xN > maxW-1 && yN >= 0 && yN < maxH {
			// na
		} else if yN > maxH-1 {
			// na
		}
		// eq 6-34
		xW = (xN + maxW) % maxW
		// eq 6-35
		yW = (yN + maxH) % maxH
	} else {
		// 6.4.12.2
		var currMbFrameFlag, mbIsTopMbFlag, mbAddrXFrameFlag int
		var yM int
		var mbAddrX int
		if !s.Slice.Data.MbFieldDecodingFlag || s.SPS.FrameMbsOnly {
			currMbFrameFlag = 1
		}
		if c := s.Slice.Data.CurrMbAddr; c%2 == 0 {
			mbIsTopMbFlag = 1
		}
		// TODO: Check mbAddrX is frame MB
		if s.SPS.FrameMbsOnly {
			mbAddrXFrameFlag = 1
		}
		// table 6-4
		if xN < 0 && yN < 0 {
			// Row 1
			if currMbFrameFlag == 1 {
				if mbIsTopMbFlag == 1 {
					mbAddrX = macroblocks[MbAddrD].Address
					mbAddrN = mbAddrX + 1
					yM = yN
				} else {
					mbAddrX = macroblocks[MbAddrA].Address
					if mbAddrXFrameFlag == 1 {
						mbAddrN = mbAddrX
						yM = yN
					} else {
						mbAddrN = mbAddrX + 1
						yM = (yN + maxH) >> 1
					}
				}
			} else {
				mbAddrX = macroblocks[MbAddrD].Address
				if mbIsTopMbFlag == 1 {
					if mbAddrXFrameFlag == 1 {
						mbAddrN = mbAddrX + 1
						yM = 2 * yN
					} else {
						mbAddr = mbAddrX
						yM = yN
					}
				} else {
					mbAddrN = mbAddrX + 1
					yM = yN
				}
			}
		} else if xN < 0 && yN >= 0 && yN < maxH {
			// Row 2
			mbAddrX = macroblocks[MbAddrA].Address
			if currMbFrameFlag == 1 {
				if mbIsTopMbFlag == 1 {
					if mbAddrXFrameFlag == 1 {
						mbAddrN = mbAddrX
						yM = yN
					} else {
						if yN%2 == 0 {
							mbAddrN = mbAddrX
							yM = yN >> 1
						} else {
							mbAddrN = mbAddrX + 1
							yM = yN >> 1
						}
					}
				} else {
					if mbAddrXFrameFlag == 1 {
						mbAddrN = mbAddrX + 1
						yM = yN
					} else {
						if yN%2 == 0 {
							mbAddrN = mbAddrX
							yM = (yN + maxH) >> 1
						} else {
							mbAddrN = mbAddrX + 1
							yM = (yN + maxH) >> 1
						}
					}
				}
			} else {
				if mbIsTopMbFlag == 1 {
					if mbAddrXFrameFlag == 1 {
						if yN < (maxH / 2) {
							mbAddrN = mbAddrX
							yM = yN << 1
						} else {
							mbAddrN = mbAddrX + 1
							yM = (yN << 1) - maxH
						}
					} else {
						mbAddrN = mbAddrX
						yM = yN
					}
				} else {
					if mbAddrXFrameFlag == 1 {
						if yN < (maxH / 2) {
							mbAddrN = mbAddrX
							yM = (yN << 1) + 1
						} else {
							mbAddrN = mbAddrX + 1
							yM = (yN << 1) + 1 - maxH
						}
					} else {
						mbAddrN = mbAddrX + 1
						yM = yN
					}
				}
			}
		} else if xN >= 0 && xN < maxW && yN < 0 {
			// Row 3
			if currMbFrameFlag == 1 {
				if mbIsTopMbFlag == 1 {
					mbAddrX = macroblocks[MbAddrB].Address
					mbAddrN = mbAddrX + 1
					yM = yN
				} else {
					mbAddrX = s.Slice.Data.CurrMbAddr
					mbAddrN = mbAddrX - 1
					yM = yN
				}
			} else {
				mbAddrX = macroblocks[MbAddrB].Address
				if mbIsTopMbFlag == 1 {
					if mbAddrXFrameFlag == 1 {
						mbAddrN = mbAddrX + 1
						yM = 2 * yN
					} else {
						mbAddrN = mbAddrX
						yM = yN
					}
				} else {
					mbAddrN = mbAddrX + 1
					yM = yN
				}
			}
		} else if xN >= 0 && xN < maxW && yN >= 0 && yN < maxH {
			// Row 4
			mbAddrX = s.Slice.Data.CurrMbAddr
			mbAddrN = mbAddrX
			yM = yN
		} else if xN > maxW-1 && yN < 0 {
			if currMbFrameFlag == 1 && mbIsTopMbFlag == 1 {
				mbAddrX = macroblocks[MbAddrC].Address
				mbAddrN = mbAddrX + 1
				yM = yN
			} // NA on other cases
			if currMbFrameFlag == 0 {
				mbAddrX = macroblocks[MbAddrC].Address
				if mbIsTopMbFlag == 1 {
					if mbAddrXFrameFlag == 1 {
						mbAddrN = mbAddrX + 1
						yM = 2 * yN
					} else {
						mbAddrN = mbAddrX
						yM = yN
					}
				} else {
					mbAddrN = mbAddrX + 1
					yM = yN
				}
			}
		} else if xN > maxW-1 && yN >= 0 && yN < maxH {
			// Row 5
			// NA
		} else if yN > maxH-1 {
			// Row 6
			// NA
		}
		// eq 6-36
		xW = (xN + maxW) % maxW
		// eq 6-37
		yW = (yM + maxH) % maxH
	}
	return mbAddrN, xW, yW
}
