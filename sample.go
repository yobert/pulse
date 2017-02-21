package pulse

import (
	"fmt"
)

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

func (s SampleType) String() string {
	switch s {
	case SampleU8:
		return "SampleU8"
	case SampleAlaw:
		return "SampleAlaw"
	case SampleUlaw:
		return "SampleUlaw"
	case SampleS16LE:
		return "SampleS16LE"
	case SampleS16BE:
		return "SampleS16BE"
	case SampleFloat32LE:
		return "SampleFloat32LE"
	case SampleFloat32BE:
		return "SampleFloat32BE"
	case SampleS32LE:
		return "SampleS32LE"
	case SampleS32BE:
		return "SampleS32BE"
	case SampleS24LE:
		return "SampleS24LE"
	case SampleS24BE:
		return "SampleS24BE"
	case SampleS24_32LE:
		return "SampleS24_32LE"
	case SampleS24_32BE:
		return "SampleS24_32BE"
	default:
		return fmt.Sprintf("UnknownSampleType(%d)", s)
	}
}

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

func (e EncodingType) String() string {
	switch e {
	case EncodingAny:
		return "EncodingAny"
	case EncodingPCM:
		return "EncodingPCM"
	case EncodingAC3Padded:
		return "EncodingAC3Padded"
	case EncodingEAC3Padded:
		return "EncodingEAC3Padded"
	case EncodingMpegPadded:
		return "EncodingMpegPadded"
	case EncodingDTSPadded:
		return "EncodingDTSPadded"
	case EncodingMpeg2AACPadded:
		return "EncodingMpeg2AACPadded"
	default:
		return "UnknownEncodingType"
	}
}
