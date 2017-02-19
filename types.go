package pulse

import (
	"fmt"
	"io"
	"sort"
)

type ValueType byte

const (
	InvalidValue    ValueType = 0
	StringValue     ValueType = 't'
	StringNullValue ValueType = 'N'
	Uint32Value     ValueType = 'L'
	ByteValue       ValueType = 'B'
	Uint64Value     ValueType = 'R'
	Int64Value      ValueType = 'r'
	SampleSpecValue ValueType = 'a'
	ArbitraryValue  ValueType = 'x'
	TrueValue       ValueType = '1'
	FalseValue      ValueType = '0'
	TimeValue       ValueType = 'T'
	UsecValue       ValueType = 'U'
	ChannelMapValue ValueType = 'm'
	CvolumeValue    ValueType = 'v'
	PropListValue   ValueType = 'P'
	VolumeValue     ValueType = 'V'
	FormatInfoValue ValueType = 'f'
)

func (t ValueType) String() string {
	switch t {
	case InvalidValue:
		return "InvalidValue"
	case StringValue:
		return "StringValue"
	case StringNullValue:
		return "StringNullValue"
	case Uint32Value:
		return "Uint32Value"
	case ByteValue:
		return "ByteValue"
	case Uint64Value:
		return "Uint64Value"
	case Int64Value:
		return "Int64Value"
	case SampleSpecValue:
		return "SampleSpecValue"
	case ArbitraryValue:
		return "ArbitraryValue"
	case TrueValue:
		return "TrueValue"
	case FalseValue:
		return "FalseValue"
	case TimeValue:
		return "TimeValue"
	case UsecValue:
		return "UsecValue"
	case ChannelMapValue:
		return "ChannelMapValue"
	case CvolumeValue:
		return "CvolumeValue"
	case PropListValue:
		return "PropListValue"
	case VolumeValue:
		return "VolumeValue"
	case FormatInfoValue:
		return "FormatInfoValue"
	default:
		return fmt.Sprintf("UnknownValue(%d)", t)
	}
}

type PropList struct {
	V map[string]string
}

func (p *PropList) String() string {
	d := ""
	var keys []string
	for k := range p.V {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i != 0 {
			d += ", "
		}
		d += fmt.Sprintf("%s %#v", k, p.V[k])
	}
	if len(d) > 512 {
		d = d[:512] + " ..."
	}
	return d
}
func (p *PropList) WriteTo(w io.Writer) (int, error) {
	n, err := bwrite(w, PropListValue)
	if err != nil {
		return n, err
	}

	for k, v := range p.V {
		if v == "" {
			continue
		}

		l := uint32(len(v) + 1) // +1 for null at the end of string
		n2, err := bwrite(w,
			StringValue, []byte(k), byte(0),
			Uint32Value, l,
			ArbitraryValue, l,
			[]byte(v), byte(0),
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	n2, err := bwrite(w, StringNullValue)
	n += n2
	if err != nil {
		return n, err
	}

	return n, nil
}
func (p *PropList) ReadFrom(r io.Reader) error {
	if p.V == nil {
		p.V = make(map[string]string)
	}
	err := bread(r, PropListValue)
	if err != nil {
		return err
	}
	for {
		var t ValueType
		if err = bread(r, &t); err != nil {
			return err
		}

		if t == StringNullValue {
			// end of the proplist.
			break
		}
		if t != StringValue {
			return fmt.Errorf("Protcol error: Got type %s but expected %s", t, StringValue)
		}

		var k, v string
		var l1, l2 uint32
		if err = bread(r,
			&k,
			Uint32Value, &l1,
			ArbitraryValue, &l2,
			&v,
		); err != nil {
			return err
		}
		if len(v) != int(l1-1) || len(v) != int(l2-1) {
			return fmt.Errorf("Protocol error: Proplist value length mismatch (len %d, arb len %d, value len %d)",
				l1, l2, len(v))
		}
		p.V[k] = v
	}
	return nil
}
