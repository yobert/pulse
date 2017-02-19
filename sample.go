package pulse

type SampleType byte

const (
	SampleU8 SampleType = iota
	SampleAlaw
	SampleUlaw
	SampleS16LE
	SampleS16BE
	SampleFloat32LE
	SampleFloat32BE
	SampleS32LE
	SampleS32BE
	SampleS24LE
	SampleS24BE
	SampleS24_32LE
	SampleS24_32BE
)

type EncodingType byte

const (
	EncodingAny EncodingType = iota
	EncodingPCM
	EncodingAC3Padded
	EncodingEAC3Padded
	EncodingMpegPadded // mpeg1 or mpeg2 (part 3, not aac)
	EncodingDTSPadded
	EncodingMpeg2AACPadded
)
