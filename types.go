package pulse

type Type byte

const (
	InvalidType Type = 0
	String      Type = 't'
	StringNull  Type = 'N'
	Uint32      Type = 'L'
	Byte        Type = 'B'
	Uint64      Type = 'R'
	Int64       Type = 'r'
	SampleSpec  Type = 'a'
	Arbitrary   Type = 'x'
	True        Type = '1'
	False       Type = '0'
	Timeval     Type = 'T'
	Usec        Type = 'U'
	ChannelMap  Type = 'm'
	Cvolume     Type = 'v'
	PropList    Type = 'P'
	Volume      Type = 'V'
	FormatInfo  Type = 'f'
)

func (t Type) String() string {
	switch t {
	case InvalidType:
		return "InvalidType"
	case String:
		return "String"
	case StringNull:
		return "StringNull"
	case Uint32:
		return "Uint32"
	case Byte:
		return "Byte"
	case Uint64:
		return "Uint64"
	case Int64:
		return "Int64"
	case SampleSpec:
		return "SampleSpec"
	case Arbitrary:
		return "Arbitrary"
	case True:
		return "True"
	case False:
		return "False"
	case Timeval:
		return "Timeval"
	case Usec:
		return "Usec"
	case ChannelMap:
		return "ChannelMap"
	case Cvolume:
		return "Cvolume"
	case PropList:
		return "PropList"
	case Volume:
		return "Volume"
	case FormatInfo:
		return "FormatInfo"
	default:
		return "UnknownType"
	}
}
