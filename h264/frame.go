package h264

// NALU types, as defined in table 7-1 in specifications.
const (
	naluTypeUnspecified = iota
	naluTypeSliceNonIDRPicture
	naluTypeSlicePartA
	naluTypeSlicePartB
	naluTypeSlicePartC
	naluTypeSliceIDRPicture
	naluTypeSEI
	naluTypeSPS
	naluTypePPS
	naluTypeAccessUnitDelimiter
	naluTypeEndOfSequence
	naluTypeEndOfStream
	naluTypeFillerData
	naluTypeSPSExtension
	naluTypePrefixNALU
	naluTypeSubsetSPS
	naluTypeDepthParamSet
)

var (
	// Refer to ITU-T H.264 4/10/2017
	// Specifieds the RBSP structure in the NAL unit
	NALUnitType = map[int]string{
		0: "unspecified",
		// slice_layer_without_partitioning_rbsp
		1: "coded slice of non-IDR picture",
		// slice_data_partition_a_layer_rbsp
		2: "coded slice data partition a",
		// slice_data_partition_b_layer_rbsp
		3: "coded slice data partition b",
		// slice_data_partition_c_layer_rbsp
		4: "coded slice data partition c",
		// slice_layer_without_partitioning_rbsp
		5: "coded IDR slice of picture",
		// sei_rbsp
		6: "sei suppl. enhancem. info",
		// seq_parameter_set_rbsp
		7: "sequence parameter set",
		// pic_parameter_set_rbsp
		8: "picture parameter set",
		// access_unit_delimiter_rbsp
		9: "access unit delimiter",
		// end_of_seq_rbsp
		10: "end of sequence",
		// end_of_stream_rbsp
		11: "end of stream",
		// filler_data_rbsp
		12: "filler data",
		// seq_parameter_set_extension_rbsp
		13: "sequence parameter set extensions",
		// prefix_nal_unit_rbsp
		14: "prefix NAL unit",
		// subset sequence parameter set
		15: "subset SPS",
		// depth_parameter_set_rbsp
		16: "depth parameter set",
		// 17, 18 are reserved
		17: "reserved",
		18: "reserved",
		// slice_layer_without_partitioning_rbsp
		19: "coded slice of aux coded pic w/o partit.",
		// slice_layer_extension_rbsp
		20: "coded slice extension",
		// slice_layer_extension_rbsp
		21: "slice ext. for depth of view or 3Davc view comp.",
		22: "reserved",
		23: "reserved",
		// 24 - 31 undefined
	}
	// ITU-T H.265 Section 7.4.1 nal_ref_idc
	NALRefIDC = map[int]string{
		0: "only nal_unit_type 6, 9, 10, 11, or 12",
		1: "anything",
		2: "anything",
		3: "anything",
		4: "anything",
	}
)

func rbspBytes(frame []byte) []byte {
	if len(frame) > 8 {
		return frame[8:]
	}
	return frame
}
