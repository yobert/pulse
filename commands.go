package pulse

import (
	"errors"
	"fmt"
	"io"
)

var ErrNotImplemented = errors.New("Not implemented")

const (
	/* Generic commands */
	Command_ERROR uint32 = iota
	Command_TIMEOUT
	Command_REPLY // 2

	/* CLIENT->SERVER */
	Command_CREATE_PLAYBACK_STREAM // 3
	Command_DELETE_PLAYBACK_STREAM
	Command_CREATE_RECORD_STREAM
	Command_DELETE_RECORD_STREAM
	Command_EXIT
	Command_AUTH // 8
	Command_SET_CLIENT_NAME
	Command_LOOKUP_SINK
	Command_LOOKUP_SOURCE
	Command_DRAIN_PLAYBACK_STREAM
	Command_STAT
	Command_GET_PLAYBACK_LATENCY
	Command_CREATE_UPLOAD_STREAM
	Command_DELETE_UPLOAD_STREAM
	Command_FINISH_UPLOAD_STREAM
	Command_PLAY_SAMPLE
	Command_REMOVE_SAMPLE // 19

	Command_GET_SERVER_INFO
	Command_GET_SINK_INFO
	Command_GET_SINK_INFO_LIST
	Command_GET_SOURCE_INFO
	Command_GET_SOURCE_INFO_LIST
	Command_GET_MODULE_INFO
	Command_GET_MODULE_INFO_LIST
	Command_GET_CLIENT_INFO
	Command_GET_CLIENT_INFO_LIST
	Command_GET_SINK_INPUT_INFO
	Command_GET_SINK_INPUT_INFO_LIST
	Command_GET_SOURCE_OUTPUT_INFO
	Command_GET_SOURCE_OUTPUT_INFO_LIST
	Command_GET_SAMPLE_INFO
	Command_GET_SAMPLE_INFO_LIST
	Command_SUBSCRIBE

	Command_SET_SINK_VOLUME
	Command_SET_SINK_INPUT_VOLUME
	Command_SET_SOURCE_VOLUME

	Command_SET_SINK_MUTE
	Command_SET_SOURCE_MUTE // 40

	Command_CORK_PLAYBACK_STREAM
	Command_FLUSH_PLAYBACK_STREAM
	Command_TRIGGER_PLAYBACK_STREAM // 43

	Command_SET_DEFAULT_SINK
	Command_SET_DEFAULT_SOURCE // 45

	Command_SET_PLAYBACK_STREAM_NAME
	Command_SET_RECORD_STREAM_NAME // 47

	Command_KILL_CLIENT
	Command_KILL_SINK_INPUT
	Command_KILL_SOURCE_OUTPUT // 50

	Command_LOAD_MODULE
	Command_UNLOAD_MODULE // 52

	Command_ADD_AUTOLOAD___OBSOLETE
	Command_REMOVE_AUTOLOAD___OBSOLETE
	Command_GET_AUTOLOAD_INFO___OBSOLETE
	Command_GET_AUTOLOAD_INFO_LIST___OBSOLETE //56

	Command_GET_RECORD_LATENCY
	Command_CORK_RECORD_STREAM
	Command_FLUSH_RECORD_STREAM
	Command_PREBUF_PLAYBACK_STREAM // 60

	/* SERVER->CLIENT */
	Command_REQUEST // 61
	Command_OVERFLOW
	Command_UNDERFLOW
	Command_PLAYBACK_STREAM_KILLED
	Command_RECORD_STREAM_KILLED
	Command_SUBSCRIBE_EVENT

	/* A few more client->server commands */

	Command_MOVE_SINK_INPUT // 67
	Command_MOVE_SOURCE_OUTPUT
	Command_SET_SINK_INPUT_MUTE
	Command_SUSPEND_SINK
	Command_SUSPEND_SOURCE

	Command_SET_PLAYBACK_STREAM_BUFFER_ATTR // 72
	Command_SET_RECORD_STREAM_BUFFER_ATTR

	Command_UPDATE_PLAYBACK_STREAM_SAMPLE_RATE // 74
	Command_UPDATE_RECORD_STREAM_SAMPLE_RATE

	/* SERVER->CLIENT */
	Command_PLAYBACK_STREAM_SUSPENDED // 76
	Command_RECORD_STREAM_SUSPENDED
	Command_PLAYBACK_STREAM_MOVED
	Command_RECORD_STREAM_MOVED

	Command_UPDATE_RECORD_STREAM_PROPLIST // 80
	Command_UPDATE_PLAYBACK_STREAM_PROPLIST
	Command_UPDATE_CLIENT_PROPLIST
	Command_REMOVE_RECORD_STREAM_PROPLIST
	Command_REMOVE_PLAYBACK_STREAM_PROPLIST
	Command_REMOVE_CLIENT_PROPLIST

	/* SERVER->CLIENT */
	Command_STARTED // 86

	Command_EXTENSION // 87

	Command_GET_CARD_INFO // 88
	Command_GET_CARD_INFO_LIST
	Command_SET_CARD_PROFILE

	Command_CLIENT_EVENT
	Command_PLAYBACK_STREAM_EVENT
	Command_RECORD_STREAM_EVENT

	/* SERVER->CLIENT */
	Command_PLAYBACK_BUFFER_ATTR_CHANGED
	Command_RECORD_BUFFER_ATTR_CHANGED

	Command_SET_SINK_PORT
	Command_SET_SOURCE_PORT

	Command_SET_SOURCE_OUTPUT_VOLUME
	Command_SET_SOURCE_OUTPUT_MUTE

	Command_SET_PORT_LATENCY_OFFSET

	/* BOTH DIRECTIONS */
	Command_ENABLE_SRBCHANNEL
	Command_DISABLE_SRBCHANNEL

	/* BOTH DIRECTIONS */
	Command_REGISTER_MEMFD_SHMID

	Command_MAX
)

const (
	PA_PROTOCOL_FLAG_MASK    uint32 = 0xFFFF0000
	PA_PROTOCOL_VERSION_MASK uint32 = 0x0000FFFF

	PA_PROTOCOL_FLAG_SHM   uint32 = 0x80000000
	PA_PROTOCOL_FLAG_MEMFD uint32 = 0x40000000
)

type Commander interface {
	String() string
	Cmd() uint32
	WriteTo(io.Writer, uint32) (int, error)
	ReadFrom(io.Reader, uint32) error
}

type CommandAuth struct {
	Version uint32
	Shm     bool
	Memfd   bool
	Cookie  [256]byte
}

func (c *CommandAuth) String() string {
	return fmt.Sprintf("AUTH (v%d shm %v memfd %v)", c.Version, c.Shm, c.Memfd)
}
func (c *CommandAuth) Cmd() uint32 {
	return Command_AUTH
}
func (c *CommandAuth) WriteTo(w io.Writer, version uint32) (int, error) {

	v := c.Version

	if c.Version >= 13 && c.Shm {
		v &= PA_PROTOCOL_FLAG_SHM
	}
	if c.Version >= 31 && c.Memfd {
		v &= PA_PROTOCOL_FLAG_MEMFD
	}

	n, err := bwrite(w, Uint32Value, v, ArbitraryValue, uint32(len(c.Cookie)))
	if err != nil {
		return n, err
	}
	n2, err := w.Write(c.Cookie[:])
	n += n2
	if err != nil {
		return n, err
	}
	return n, nil
}
func (c *CommandAuth) ReadFrom(r io.Reader, version uint32) error {
	return ErrNotImplemented
}

type CommandAuthReply struct {
	Version uint32
	Shm     bool
	Memfd   bool
}

func (c *CommandAuthReply) String() string {
	return fmt.Sprintf("AUTH REPLY (v%d shm %v memfd %v)", c.Version, c.Shm, c.Memfd)
}
func (c *CommandAuthReply) Cmd() uint32 {
	return Command_REPLY
}
func (c *CommandAuthReply) WriteTo(w io.Writer, version uint32) (int, error) {
	return 0, ErrNotImplemented
}
func (c *CommandAuthReply) ReadFrom(r io.Reader, version uint32) error {
	if err := bread_uint32(r, &c.Version); err != nil {
		return err
	}
	if (c.Version & PA_PROTOCOL_VERSION_MASK) >= 13 {
		if c.Version&PA_PROTOCOL_FLAG_SHM > 0 {
			c.Shm = true
		}
		if (c.Version & PA_PROTOCOL_VERSION_MASK) >= 31 {
			if c.Version&PA_PROTOCOL_FLAG_MEMFD > 0 {
				c.Memfd = true
			}
		}
		c.Version &= PA_PROTOCOL_VERSION_MASK
	}
	return nil
}

type CommandSetClientName struct {
	Props PropList
}

func (c *CommandSetClientName) String() string {
	return "SET CLIENT NAME (" + c.Props.String() + ")"
}
func (c *CommandSetClientName) Cmd() uint32 {
	return Command_SET_CLIENT_NAME
}
func (c *CommandSetClientName) WriteTo(w io.Writer, version uint32) (int, error) {
	return c.Props.WriteTo(w)
}
func (c *CommandSetClientName) ReadFrom(r io.Reader, version uint32) error {
	return ErrNotImplemented
}

type CommandSetClientNameReply struct {
	ClientIndex uint32
}

func (c *CommandSetClientNameReply) String() string {
	return fmt.Sprintf("SET CLIENT NAME REPLY (client index %d)", c.ClientIndex)
}
func (c *CommandSetClientNameReply) Cmd() uint32 {
	return Command_REPLY
}
func (c *CommandSetClientNameReply) WriteTo(w io.Writer, version uint32) (int, error) {
	return 0, ErrNotImplemented
}
func (c *CommandSetClientNameReply) ReadFrom(r io.Reader, version uint32) error {
	if err := bread_uint32(r, &c.ClientIndex); err != nil {
		return err
	}
	return nil
}

type CommandCreatePlaybackStream struct {
	Format        SampleType
	Channels      byte
	Rate          uint32
	ChannelMap    []byte
	ChannelVolume []uint32
	Props         PropList
}

func (c *CommandCreatePlaybackStream) String() string {
	return fmt.Sprintf("CREATE PLAYBACK STREAM")
}
func (c *CommandCreatePlaybackStream) Cmd() uint32 {
	return Command_CREATE_PLAYBACK_STREAM
}
func (c *CommandCreatePlaybackStream) WriteTo(w io.Writer, version uint32) (int, error) {
	n, err := bwrite(w,
		SampleSpecValue,
		c.Format,
		c.Channels,
		c.Rate,

		ChannelMapValue,
		byte(len(c.ChannelMap)),
		c.ChannelMap,

		Uint32Value,
		uint32(0xffffffff), // sink index
		StringNullValue,    // sink name

		Uint32Value,
		uint32(0xffffffff), // buffer max length
		FalseValue,         // corked
		Uint32Value,
		uint32(0xffffffff), // buffer target length
		Uint32Value,
		uint32(0xffffffff), // buffer pre-buffer length
		Uint32Value,
		uint32(0xffffffff), // buffer minimum request
		Uint32Value,
		uint32(0), // sync id-- no idea what that does

		CvolumeValue,
		byte(len(c.ChannelVolume)),
		c.ChannelVolume,
	)
	if err != nil {
		return n, err
	}

	if version >= 12 {
		n2, err := bwrite(w,
			FalseValue, // no remap
			FalseValue, // no remix
			FalseValue, // fix format
			FalseValue, // fix rate
			FalseValue, // fix channels
			FalseValue, // no move
			FalseValue, // variable rate
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	if version >= 13 {
		n2, err := bwrite(w,
			FalseValue, // muted
			FalseValue, // adjust latency
		)
		n += n2
		if err != nil {
			return n, err
		}
		n2, err = c.Props.WriteTo(w)
		n += n2
		if err != nil {
			return n, err
		}
	}

	if version >= 14 {
		n2, err := bwrite(w,
			FalseValue, // volume set
			FalseValue, // early requests
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	if version >= 15 {
		n2, err := bwrite(w,
			FalseValue, // muted set
			FalseValue, // dont inhibit auto suspend
			FalseValue, // fail on suspend
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	if version >= 17 {
		n2, err := bwrite(w,
			FalseValue, // relative volume
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	if version >= 18 {
		n2, err := bwrite(w,
			FalseValue, // passthrough
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	if version >= 21 {
		n2, err := bwrite(w,
			ByteValue,
			byte(0), // n_formats
		)
		n += n2
		if err != nil {
			return n, err
		}
	}

	return n, nil
}
func (c *CommandCreatePlaybackStream) ReadFrom(r io.Reader, version uint32) error {
	return ErrNotImplemented
}

type CommandCreatePlaybackStreamReply struct {
	StreamIndex    uint32
	SinkInputIndex uint32
	Missing        uint32

	BufferMaxLength       uint32
	BufferTargetLength    uint32
	BufferPrebufferLength uint32
	BufferMinimumRequest  uint32

	Format     SampleType
	Channels   byte
	Rate       uint32
	ChannelMap []byte

	SinkInputSinkIndex     uint32
	SinkInputSinkName      string
	SinkInputSinkSuspended bool

	SinkLatency uint64

	Encoding EncodingType
	Props    PropList
}

func (c *CommandCreatePlaybackStreamReply) String() string {
	return fmt.Sprintf("CREATE PLAYBACK STREAM REPLY (index %d/%d/%d missing %d %#v, format %s %d %dhz) %s", c.StreamIndex, c.SinkInputIndex, c.SinkInputSinkIndex, c.Missing, c.SinkInputSinkName, c.Format, c.Channels, c.Rate, c.Props.String())
}
func (c *CommandCreatePlaybackStreamReply) Cmd() uint32 {
	return Command_REPLY
}
func (c *CommandCreatePlaybackStreamReply) WriteTo(w io.Writer, version uint32) (int, error) {
	return 0, ErrNotImplemented
}
func (c *CommandCreatePlaybackStreamReply) ReadFrom(r io.Reader, version uint32) error {
	if err := bread_uint32(r,
		&c.StreamIndex,
		&c.SinkInputIndex,
		&c.Missing,
	); err != nil {
		return err
	}

	if version >= 9 {
		if err := bread_uint32(r,
			&c.BufferMaxLength,
			&c.BufferTargetLength,
			&c.BufferPrebufferLength,
			&c.BufferMinimumRequest,
		); err != nil {
			return err
		}
	}

	if version >= 12 {
		if err := bread(r,
			SampleSpecValue,
			&c.Format,
			&c.Channels,
			&c.Rate,
		); err != nil {
			return err
		}

		var l byte
		if err := bread(r, ChannelMapValue, &l); err != nil {
			return err
		}
		c.ChannelMap = make([]byte, l)
		if err := bread(r, &c.ChannelMap); err != nil {
			return err
		}

		if err := bread(r,
			Uint32Value,
			&c.SinkInputSinkIndex,
			StringValue,
			&c.SinkInputSinkName,
			&c.SinkInputSinkSuspended,
		); err != nil {
			return err
		}
	}

	if version >= 13 {
		if err := bread(r, UsecValue, &c.SinkLatency); err != nil {
			return err
		}
	}

	if version >= 21 {
		if err := bread(r, FormatInfoValue, ByteValue, &c.Encoding); err != nil {
			return err
		}
		if err := c.Props.ReadFrom(r); err != nil {
			return err
		}
	}

	return nil
}
