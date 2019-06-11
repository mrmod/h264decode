package h264

const (
	NaCtxId            = 10000
	NA_SUFFIX          = -1
	MbAddrNotAvailable = 10000
)

// G.7.4.3.4 via G.7.3.3.4 via 7.3.2.13 for NalUnitType 20 or 21
// refLayerMbWidthC is equal to MbWidthC for the reference layer representation
func RefMbW(chromaFlag, refLayerMbWidthC int) int {
	if chromaFlag == 0 {
		return 16
	}
	return refLayerMbWidthC
}

// refLayerMbHeightC is equal to MbHeightC for the reference layer representation
func RefMbH(chromaFlag, refLayerMbHeightC int) int {
	if chromaFlag == 0 {
		return 16
	}
	return refLayerMbHeightC
}
func XOffset(xRefMin16, refMbW int) int {
	return (((xRefMin16 - 64) >> 8) << 4) - (refMbW >> 1)
}
func YOffset(yRefMin16, refMbH int) int {
	return (((yRefMin16 - 64) >> 8) << 4) - (refMbH >> 1)
}

func CondTermFlag(mbAddr, mbSkipFlag int) int {
	if mbAddr == MbAddrNotAvailable || mbSkipFlag == 1 {
		return 0
	}
	return 1
}

// s9.3.3 p 278: Returns the value of the syntax element
func (bin *Binarization) Decode(sliceContext *SliceContext, b *BitReader, rbsp []byte) {
	if bin.SyntaxElement == "MbType" {
		bin.binString = binIdxMbMap[sliceContext.Slice.Data.SliceTypeName][sliceContext.Slice.Data.MbType]
	} else {
		logger.Printf("TODO: no means to find binString for %s\n", bin.SyntaxElement)
	}
	logger.Printf("bin string of %s binarization: %#v\n", bin.SyntaxElement, bin.binString)
}

// 9.3.3.1.1 : returns ctxIdxInc
func Decoder9_3_3_1_1_1(condTermFlagA, condTermFlagB int) int {
	return condTermFlagA + condTermFlagB
}

// 9-5
// 7-30 p 112
func SliceQPy(pps *PPS, header *SliceHeader) int {
	return 26 + pps.PicInitQpMinus26 + header.SliceQpDelta
}

// 9-5
func PreCtxState(m, n, sliceQPy int) int {
	// slicQPy-subY
	return Clip3(1, 126, ((m*Clip3(0, 51, sliceQPy))>>4)+n)
}

func Clip1y(x, bitDepthY int) int {
	return Clip3(0, (1<<uint(bitDepthY))-1, x)
}
func Clipc(x, bitDepthC int) int {
	return Clip3(0, (1<<uint(bitDepthC))-1, x)
}

// 5-5
func Clip3(x, y, z int) int {
	if z < x {
		return x
	}
	if z > y {
		return y
	}
	return z
}

type CABAC struct {
	IsInitialized        bool
	CtxIdx               []MN
	ContextVariables     []ContextVariable
	DecodingEngine       *DecodingEngine
	ParsedSyntaxElements []string
	SECount              int
	Context              *SliceContext
	BitReader            *BitReader
}
type ContextVariable struct {
	PStateIdx, ValMPS, PreCtxState int
}
type DecodingEngine struct {
	CodIRange, CodIOffset int
}
type SEBinarization struct {
	SyntaxElement              string
	SliceTypeName              string
	MaxBinIdxCtx, CtxIdxOffset int
	MaxBinIdxCtxPrefix         int
	MaxBinIdxCtxSuffix         int
	CtxIdxOffsetPrefix         int
	CtxIdxOffsetSuffix         int
	BinString                  []int // aka: bins
	BinIdx                     int
	BypassFlag                 int
}

func NewCABAC(s *SliceContext, b *BitReader) *CABAC {
	return &CABAC{
		Context:   s,
		BitReader: b,
	}
}

// 9.3.2
func (c *CABAC) GetBinarization(se string) *SEBinarization {
	binarization := &SEBinarization{
		SyntaxElement: se,
		SliceTypeName: sliceTypeMap[c.Context.Slice.Header.SliceType],
	}
	if !c.IsInitialized {
		c.InitContextVariables(c.Context, c.BitReader, se)
	}
	sliceTypeName := sliceTypeMap[c.Context.Slice.Header.SliceType]
	switch se {
	case "MBType":
		switch sliceTypeName {
		case "I":
			// 9.3.2.5
			binarization.MaxBinIdxCtx = 6
			binarization.CtxIdxOffset = 3
			binString := append([]int{1}, binIdxMbMap[sliceTypeName][c.Context.Slice.Data.MbType]...)
			binarization.BinString = binString
		case "SI":
			binarization.MaxBinIdxCtxPrefix = 0
			binarization.MaxBinIdxCtxSuffix = 6
			binarization.CtxIdxOffsetPrefix = 0
			binarization.CtxIdxOffsetSuffix = 3
			binString := append([]int{0}, binIdxMbMap["I"][c.Context.Slice.Data.MbType-1]...)
			binarization.BinString = binString
		}
	}
	binarization.BinIdx = -1
	return binarization
}

// 9.3.3
func (c *CABAC) Decode(binarization *SEBinarization) int {
	var parsedBits []int
	var maxBinIdx, ctxIdxOffset int
	sliceTypeName := sliceTypeMap[c.Context.Slice.Header.SliceType]
	// ref table 9-34
	switch binarization.SyntaxElement {
	case "MBType":
		if sliceTypeName == "I" {
			maxBinIdx = binarization.MaxBinIdxCtx
			ctxIdxOffset = binarization.CtxIdxOffset
		}
	}
	binarization.BinIdx += 1
	// Decoding process
	logger.Printf("debug: Decoding %s with maxBinIdx %d", binarization.SyntaxElement, maxBinIdx)
	for binarization.BinIdx < maxBinIdx {
		ctxIdx := c.GetCtxIdx(binarization, maxBinIdx, ctxIdxOffset)

		bit := c.BitReader.ReadOneBit()

		if bit == binarization.BinString[binarization.BinIdx] {
		} else {
			parsedBits = append(parsedBits, bit)
		}
		logger.Printf("debug: comparing\n\tbit string %#v\n\tbin string %#v", parsedBits, binarization.BinString)
		logger.Printf("debug: ctxIdx(%d) binIdx(%d): parsed bits %#v",
			ctxIdx, binarization.BinIdx, parsedBits)
		binarization.BinIdx += 1
	}
	return 0
}

// fig 9-1 CABAC parsing process
func (c *CABAC) GetCtxIdx(binarization *SEBinarization, maxBinIdx, ctxIdxOffset int) int {
	var ctxIdxInc int
	if binarization.SyntaxElement == "MBType" {
		mbTypeName := MbTypeName(sliceTypeMap[c.Context.Slice.Header.SliceType], c.Context.Slice.Data.MbType)
		ctxIdxInc = c.MBTypeCtxIdxInc(mbTypeName, ctxIdxOffset, binarization)
		return ctxIdxOffset + ctxIdxInc
	}
	return CtxIdx(binarization.BinIdx, maxBinIdx, ctxIdxOffset)
}

// 9.3.3.1.1.3
// returns ctxIdxInc
func (c *CABAC) MBTypeCtxIdxInc(mbType string, ctxIdxOffset int, binarization *SEBinarization) int {
	macroblocks := NeighborMacroblocks(c.Context, c.Context.Slice.Data.CurrMbAddr)
	condTermFlagN := func(mb Macroblock) int {
		if !mb.IsAvailable {
			return 0
		}
		// TODO: Should be MbType of macroblock[MbAddrN]
		if ctxIdxOffset == 0 && mbType == "SI" {
			return 0
		}
		if ctxIdxOffset == 3 && mbType == "I_NxN" {
			return 0
		}
		if ctxIdxOffset == 27 && mbType == "B_Skip" || mbType == "B_Direct_16x16" {
			return 0
		}
		return 1
	}
	// eq 9-9
	return condTermFlagN(macroblocks[MbAddrA]) + condTermFlagN(macroblocks[MbAddrB])
}

// 9.3.1.1
func (c *CABAC) InitContextVariables(s *SliceContext, b *BitReader, se string) {
	var ctxIdx []MN
	var contextVariables []ContextVariable
	sliceTypeName := sliceTypeMap[s.Slice.Header.SliceType]
	switch se {
	case "MBType":
		switch sliceTypeName {
		case "SI":
			for i := 0; i <= 10; i++ {
				ctxIdx = append(ctxIdx, MNVars[i][NoCabacInitIdc])
			}
		case "I":
			for i := 3; i <= 10; i++ {
				ctxIdx = append(ctxIdx, MNVars[i][NoCabacInitIdc])
			}
		case "P", "SP":
			for i := 14; i <= 20; i++ {
				ctxIdx = append(ctxIdx, MNVars[i][s.Slice.Header.CabacInit])
			}
		}
	}
	for _, _ctxIdx := range ctxIdx {
		var pStateIdx, valMPS int
		// eq 9-5
		preCtxState := PreCtxState(_ctxIdx.M, _ctxIdx.N, SliceQPy(s.PPS, s.Header))
		if preCtxState <= 63 {
			pStateIdx = 63 - preCtxState
			valMPS = 0
		} else {
			pStateIdx = preCtxState - 64
			valMPS = 1
		}
		contextVariables = append(contextVariables, ContextVariable{
			pStateIdx,
			valMPS,
			preCtxState,
		})
	}
	c.Context = s
	c.CtxIdx = ctxIdx
	c.ContextVariables = contextVariables
	c.SECount += 1
	c.ParsedSyntaxElements = append(c.ParsedSyntaxElements, se)
	c.InitDecodingEngine(b)
}
func (c *CABAC) InitDecodingEngine(b *BitReader) {
	logger.Printf("debug: initializing arithmetic decoding engine\n")
	b.LogStreamPosition()
	codIRange := 510
	codIOffset := b.NextField("Initial CodIOffset", 9)
	logger.Printf("debug: codIRange: %d :: codIOffsset: %d\n", codIRange, codIOffset)
	c.DecodingEngine = &DecodingEngine{codIRange, codIOffset}
	c.IsInitialized = true
}

// table 9-1
// 9.3.1.1: initialization of context variables
func initContextVariables(se string, binarization *Binarization, context *SliceContext) *CABAC {
	return initCabac(se, binarization, context)
}

// TODO: refactor to initContextVariables globally
func initCabac(se string, binarization *Binarization, context *SliceContext) *CABAC {
	logger.Printf("init sliceType %s context variables for SE: %s\n", sliceTypeMap[context.Slice.Header.SliceType], se)
	var valMPS, pStateIdx int
	// TODO: When to use prefix, when to use suffix? See 9.3.2 on BinZ process
	ctxIdx := CtxIdx(
		binarization.binIdx,
		binarization.MaxBinIdxCtx.Prefix,
		binarization.CtxIdxOffset.Prefix)
	mn := MNVars[ctxIdx]
	logger.Printf("debug: initialized ctxIdx to %v\n", ctxIdx)
	// eq 9-5
	preCtxState := PreCtxState(mn[0].M, mn[0].N, SliceQPy(context.PPS, context.Header))
	if preCtxState <= 63 {
		pStateIdx = 63 - preCtxState
		valMPS = 0
	} else {
		pStateIdx = preCtxState - 64
		valMPS = 1
	}
	_ = pStateIdx
	_ = valMPS
	// Initialization of context variables
	// Initialization of decoding engine
	return &CABAC{
		// PStateIdx: pStateIdx,
		// ValMPS:    valMPS,
		// Context:   context,
	}
}

// Table 9-36, 9-37
// func BinIdx(mbType int, sliceTypeName string) []int {
// Map of SliceTypeName[MbType][]int{binString}
// {"SliceTypeName": {MbTypeCode: []BinString}}
var (
	binIdxMbMap = map[string]map[int][]int{
		"I": map[int][]int{
			0:  []int{0},
			1:  []int{1, 0, 0, 0, 0, 0},
			2:  []int{1, 0, 0, 0, 0, 1},
			3:  []int{1, 0, 0, 0, 1, 0},
			4:  []int{1, 0, 0, 0, 1, 1},
			5:  []int{1, 0, 0, 1, 0, 0, 0},
			6:  []int{1, 0, 0, 1, 0, 0, 1},
			7:  []int{1, 0, 0, 1, 0, 1, 0},
			8:  []int{1, 0, 0, 1, 0, 1, 1},
			9:  []int{1, 0, 0, 1, 1, 0, 0},
			10: []int{1, 0, 0, 1, 1, 0, 1},
			11: []int{1, 0, 0, 1, 1, 1, 0},
			12: []int{1, 0, 0, 1, 1, 1, 1},
			13: []int{1, 0, 1, 0, 0, 0},
			14: []int{1, 0, 1, 0, 0, 1},
			15: []int{1, 0, 1, 0, 1, 0},
			16: []int{1, 0, 1, 0, 1, 1},
			17: []int{1, 0, 1, 1, 0, 0, 0},
			18: []int{1, 0, 1, 1, 0, 0, 1},
			19: []int{1, 0, 1, 1, 0, 1, 0},
			20: []int{1, 0, 1, 1, 0, 1, 1},
			21: []int{1, 0, 1, 1, 1, 0, 0},
			22: []int{1, 0, 1, 1, 1, 0, 1},
			23: []int{1, 0, 1, 1, 1, 1, 0},
			24: []int{1, 0, 1, 1, 1, 1, 1},
			25: []int{1, 1},
		},
		// Table 9-37
		"P": map[int][]int{
			0:  []int{0, 0, 0},
			1:  []int{0, 1, 1},
			2:  []int{0, 1, 0},
			3:  []int{0, 0, 1},
			4:  []int{},
			5:  []int{1},
			6:  []int{1},
			7:  []int{1},
			8:  []int{1},
			9:  []int{1},
			10: []int{1},
			11: []int{1},
			12: []int{1},
			13: []int{1},
			14: []int{1},
			15: []int{1},
			16: []int{1},
			17: []int{1},
			18: []int{1},
			19: []int{1},
			20: []int{1},
			21: []int{1},
			22: []int{1},
			23: []int{1},
			24: []int{1},
			25: []int{1},
			26: []int{1},
			27: []int{1},
			28: []int{1},
			29: []int{1},
			30: []int{1},
		},
		// Table 9-37
		"SP": map[int][]int{
			0:  []int{0, 0, 0},
			1:  []int{0, 1, 1},
			2:  []int{0, 1, 0},
			3:  []int{0, 0, 1},
			4:  []int{},
			5:  []int{1},
			6:  []int{1},
			7:  []int{1},
			8:  []int{1},
			9:  []int{1},
			10: []int{1},
			11: []int{1},
			12: []int{1},
			13: []int{1},
			14: []int{1},
			15: []int{1},
			16: []int{1},
			17: []int{1},
			18: []int{1},
			19: []int{1},
			20: []int{1},
			21: []int{1},
			22: []int{1},
			23: []int{1},
			24: []int{1},
			25: []int{1},
			26: []int{1},
			27: []int{1},
			28: []int{1},
			29: []int{1},
			30: []int{1},
		},
		// TODO: B Slice table 9-37
	}

	// Map of SliceTypeName[SubMbType][]int{binString}
	binIdxSubMbMap = map[string]map[int][]int{
		"P": map[int][]int{
			0: []int{1},
			1: []int{0, 0},
			2: []int{0, 1, 1},
			3: []int{0, 1, 0},
		},
		"SP": map[int][]int{
			0: []int{1},
			1: []int{0, 0},
			2: []int{0, 1, 1},
			3: []int{0, 1, 0},
		},
		// TODO: B slice table 9-38
	}

	// Table 9-36, 9-37
	MbBinIdx = []int{1, 2, 3, 4, 5, 6}

	// Table 9-38
	SubMbBinIdx = []int{0, 1, 2, 3, 4, 5}
)

// Table 9-34
type MaxBinIdxCtx struct {
	// When false, Prefix is the MaxBinIdxCtx
	IsPrefixSuffix bool
	Prefix, Suffix int
}
type CtxIdxOffset struct {
	// When false, Prefix is the MaxBinIdxCtx
	IsPrefixSuffix bool
	Prefix, Suffix int
}

// Table 9-34
type Binarization struct {
	SyntaxElement string
	BinarizationType
	MaxBinIdxCtx
	CtxIdxOffset
	UseDecodeBypass int
	// TODO: Why are these private but others aren't?
	binIdx    int
	binString []int
}
type BinarizationType struct {
	PrefixSuffix   bool
	FixedLength    bool
	Unary          bool
	TruncatedUnary bool
	CMax           bool
	// 9.3.2.3
	UEGk      bool
	CMaxValue int
}

// 9.3.2.5
func GetBinarization(syntaxElement string, data *SliceData) *Binarization {
	return NewBinarization(syntaxElement, data)
}

// TODO: Rename to GetBinarization globally
func NewBinarization(syntaxElement string, data *SliceData) *Binarization {
	sliceTypeName := data.SliceTypeName
	logger.Printf("debug: binarization of %s in sliceType %s\n", syntaxElement, sliceTypeName)
	binarization := &Binarization{SyntaxElement: syntaxElement}
	switch syntaxElement {
	case "CodedBlockPattern":
		binarization.BinarizationType = BinarizationType{PrefixSuffix: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 3, Suffix: 1}
		binarization.CtxIdxOffset = CtxIdxOffset{IsPrefixSuffix: true, Prefix: 73, Suffix: 77}
	case "IntraChromaPredMode":
		binarization.BinarizationType = BinarizationType{
			TruncatedUnary: true, CMax: true, CMaxValue: 3}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 1}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 64}
	case "MbQpDelta":
		binarization.BinarizationType = BinarizationType{}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 2}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 60}
	case "MvdLnEnd0":
		binarization.UseDecodeBypass = 1
		binarization.BinarizationType = BinarizationType{UEGk: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 4, Suffix: NA_SUFFIX}
		binarization.CtxIdxOffset = CtxIdxOffset{
			IsPrefixSuffix: true,
			Prefix:         40,
			Suffix:         NA_SUFFIX,
		}
	case "MvdLnEnd1":
		binarization.UseDecodeBypass = 1
		binarization.BinarizationType = BinarizationType{UEGk: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{
			IsPrefixSuffix: true,
			Prefix:         4,
			Suffix:         NA_SUFFIX,
		}
		binarization.CtxIdxOffset = CtxIdxOffset{
			IsPrefixSuffix: true,
			Prefix:         47,
			Suffix:         NA_SUFFIX,
		}
		// 9.3.2.5
	case "MbType":
		logger.Printf("debug: \tMbType %d is %s\n", data.MbType, data.MbTypeName)
		switch sliceTypeName {
		case "SI":
			binarization.BinarizationType = BinarizationType{PrefixSuffix: true}
			binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 0, Suffix: 6}
			binarization.CtxIdxOffset = CtxIdxOffset{IsPrefixSuffix: true, Prefix: 0, Suffix: 3}
		case "I":
			binarization.BinarizationType = BinarizationType{}
			binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 6}
			binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 3}
		case "SP":
			fallthrough
		case "P":
			binarization.BinarizationType = BinarizationType{PrefixSuffix: true}
			binarization.MaxBinIdxCtx = MaxBinIdxCtx{IsPrefixSuffix: true, Prefix: 2, Suffix: 5}
			binarization.CtxIdxOffset = CtxIdxOffset{IsPrefixSuffix: true, Prefix: 14, Suffix: 17}
		}
	case "MbFieldDecodingFlag":
		binarization.BinarizationType = BinarizationType{
			FixedLength: true, CMax: true, CMaxValue: 1}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 70}
	case "PrevIntra4x4PredModeFlag":
		fallthrough
	case "PrevIntra8x8PredModeFlag":
		binarization.BinarizationType = BinarizationType{FixedLength: true, CMax: true, CMaxValue: 1}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 68}
	case "RefIdxL0":
		fallthrough
	case "RefIdxL1":
		binarization.BinarizationType = BinarizationType{Unary: true}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 2}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 54}
	case "RemIntra4x4PredMode":
		fallthrough
	case "RemIntra8x8PredMode":
		binarization.BinarizationType = BinarizationType{FixedLength: true, CMax: true, CMaxValue: 7}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 69}
	case "TransformSize8x8Flag":
		binarization.BinarizationType = BinarizationType{FixedLength: true, CMax: true, CMaxValue: 1}
		binarization.MaxBinIdxCtx = MaxBinIdxCtx{Prefix: 0}
		binarization.CtxIdxOffset = CtxIdxOffset{Prefix: 399}
	}
	logger.Printf("debug: new %s binarization\n\t%+v", syntaxElement, binarization)
	return binarization
}
func (b *Binarization) IsBinStringMatch(bits []int) bool {
	for i, _b := range bits {
		if b.binString[i] != _b {
			return false
		}
	}
	return len(b.binString) == len(bits)
}

// 9.3.1.2: output is codIRange and codIOffset
func initDecodingEngine(bitReader *BitReader) (int, int) {
	logger.Printf("debug: initializing arithmetic decoding engine\n")
	bitReader.LogStreamPosition()
	codIRange := 510
	codIOffset := bitReader.NextField("Initial CodIOffset", 9)
	logger.Printf("debug: codIRange: %d :: codIOffsset: %d\n", codIRange, codIOffset)
	return codIRange, codIOffset
}

// 9.3.3.2.3
// Invoked when bypassFlag is equal to 1
func (a ArithmeticDecoding) DecodeBypass(sliceData *SliceData, codIRange, codIOffset int) (int, int) {
	// Decoded value binVal
	codIOffset = codIOffset << uint(1)
	// TODO: Concurrency check
	// TODO: Possibly should be codIOffset | ReadOneBit
	codIOffset = codIOffset << uint(sliceData.BitReader.ReadOneBit())
	if codIOffset >= codIRange {
		a.BinVal = 1
		codIOffset -= codIRange
	} else {
		a.BinVal = 0
	}
	return codIOffset, a.BinVal
}

// 9.3.3.2.4
// Decodes endOfSliceFlag and I_PCM
// Returns codIRange, codIOffSet, decoded value of binVal
func (a ArithmeticDecoding) DecodeTerminate(sliceData *SliceData, codIRange, codIOffset int) (int, int, int) {
	codIRange -= 2
	if codIOffset >= codIRange {
		a.BinVal = 1
		// Terminate CABAC decoding, last bit inserted into codIOffset is = 1
		// this is now also the rbspStopOneBit
		// TODO: How is this denoting termination?
		return codIRange, codIOffset, a.BinVal
	}
	a.BinVal = 0
	codIRange, codIOffset = a.RenormD(sliceData, codIRange, codIOffset)

	return codIRange, codIOffset, a.BinVal
}

// 9.3.3.2.2 Renormalization process of ADecEngine
// Returns codIRange, codIOffset
func (a ArithmeticDecoding) RenormD(sliceData *SliceData, codIRange, codIOffset int) (int, int) {
	if codIRange >= 256 {
		return codIRange, codIOffset
	}
	codIRange = codIRange << uint(1)
	codIOffset = codIOffset << uint(1)
	codIOffset = codIOffset | sliceData.BitReader.ReadOneBit()
	return a.RenormD(sliceData, codIRange, codIOffset)
}

type ArithmeticDecoding struct {
	Context      *SliceContext
	Binarization *Binarization
	BinVal       int
}

func (c *CABAC) ae(se string) int {
	c.SECount += 1
	return 0
}

// 9.3.3.1
// Returns ctxIdx
func CtxIdx(binIdx, maxBinIdxCtx, ctxIdxOffset int) int {
	logger.Printf("deriving ctxIdx with binIdx: %d, maxBinIdxCtx: %d, ctxIdxOffset: %d",
		binIdx, maxBinIdxCtx, ctxIdxOffset)
	ctxIdx := NaCtxId
	// table 9-39
	switch ctxIdxOffset {
	case 0:
		if binIdx != 0 {
			return NaCtxId
		}
		// see 9.3.3.1.1.3
	case 3:
		switch binIdx {
		case 0:
			// see 9.3.3.1.1.3
		case 1:
			ctxIdx = 276
		case 2:
			ctxIdx = 3
		case 3:
			ctxIdx = 4
		case 4:
			// 9.3.3.1.2
		case 5:
			// 9.3.3.1.2
		default:
			ctxIdx = 7
		}
	case 11:
		if binIdx != 0 {
			return NaCtxId
		}

		// see 9.3.3.1.1.3
	case 14:
		if binIdx == 0 {
			ctxIdx = 0
		}
		if binIdx == 1 {
			ctxIdx = 1
		}
		if binIdx == 2 {
			// 9.3.3.1.2
		}
		if binIdx > 2 {
			return NaCtxId
		}
	case 17:
		switch binIdx {
		case 0:
			ctxIdx = 0
		case 1:
			ctxIdx = 276
		case 2:
			ctxIdx = 1
		case 3:
			ctxIdx = 2
		case 4:
			// 9.3.3.1.2
		default:
			ctxIdx = 3
		}

	case 21:
		if binIdx < 3 {
			ctxIdx = binIdx
		} else {
			return NaCtxId
		}
	case 24:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.1
	case 27:
		switch binIdx {
		case 0:
			// see 9.3.3.1.1.3
		case 1:
			ctxIdx = 3
		case 2:
			// 9.3.3.1.2
		default:
			ctxIdx = 5
		}
	case 32:
		switch binIdx {
		case 0:
			ctxIdx = 0
		case 1:
			ctxIdx = 276
		case 2:
			ctxIdx = 1
		case 3:
			ctxIdx = 2
		case 4:
			// 9.3.3.1.2
		default:
			ctxIdx = 3
		}
	case 36:
		if binIdx == 0 || binIdx == 1 {
			ctxIdx = binIdx
		}
		if binIdx == 2 {
			// 9.3.3.1.2
		}
		if binIdx > 2 && binIdx < 6 {
			ctxIdx = 3
		}

	case 40:
		fallthrough
	case 47:
		switch binIdx {
		case 0:
			// 9.3.3.1.1.7
		case 1:
			ctxIdx = 3
		case 2:
			ctxIdx = 4
		case 3:
			ctxIdx = 5
		default:
			ctxIdx = 6
		}
	case 54:
		if binIdx == 0 {
			// 9.3.3.1.1.6
		}
		if binIdx == 1 {
			ctxIdx = 4
		}
		if binIdx > 1 {
			ctxIdx = 5
		}
	case 60:
		if binIdx == 0 {
			// 9.3.3.1.1.5
		}
		if binIdx == 1 {
			ctxIdx = 2
		}
		if binIdx > 1 {
			ctxIdx = 3
		}
	case 64:
		if binIdx == 0 {
			// 9.3.3.1.1.8
		} else if binIdx == 1 || binIdx == 2 {
			ctxIdx = 3
		} else {
			return NaCtxId
		}
	case 68:
		if binIdx != 0 {
			return NaCtxId
		}
		ctxIdx = 0
	case 69:
		if binIdx >= 0 && binIdx < 3 {
			ctxIdx = 0
		}
		return NaCtxId
	case 70:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.2
	case 73:
		switch binIdx {
		case 0:
			fallthrough
		case 1:
			fallthrough
		case 2:
			fallthrough
		case 3:
			// 9.3.3.1.1.4
		default:
			return NaCtxId
		}
	case 77:
		if binIdx == 0 {
			// 9.3.3.1.1.4
		} else if binIdx == 1 {
			// 9.3.3.1.1.4
		} else {
			return NaCtxId
		}
	case 276:
		if binIdx != 0 {
			return NaCtxId
		}
		ctxIdx = 0
	case 399:
		if binIdx != 0 {
			return NaCtxId
		}
		// 9.3.3.1.1.10
	}

	return ctxIdx
}
