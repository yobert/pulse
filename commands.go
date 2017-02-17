package pulse

import (
	"fmt"
	"io"
)

const (
	/* Generic commands */
	Command_ERROR uint32 = iota
	Command_TIMEOUT
	Command_REPLY // 2

	/* CLIENT->SERVER */
	Command_CREATE_PLAYBACK_STREAM
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
	Command_REMOVE_SAMPLE

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
	Command_SET_SOURCE_MUTE

	Command_CORK_PLAYBACK_STREAM
	Command_FLUSH_PLAYBACK_STREAM
	Command_TRIGGER_PLAYBACK_STREAM

	Command_SET_DEFAULT_SINK
	Command_SET_DEFAULT_SOURCE

	Command_SET_PLAYBACK_STREAM_NAME
	Command_SET_RECORD_STREAM_NAME

	Command_KILL_CLIENT
	Command_KILL_SINK_INPUT
	Command_KILL_SOURCE_OUTPUT

	Command_LOAD_MODULE
	Command_UNLOAD_MODULE

	Command_GET_RECORD_LATENCY
	Command_CORK_RECORD_STREAM
	Command_FLUSH_RECORD_STREAM
	Command_PREBUF_PLAYBACK_STREAM

	/* SERVER->CLIENT */
	Command_REQUEST
	Command_OVERFLOW
	Command_UNDERFLOW
	Command_PLAYBACK_STREAM_KILLED
	Command_RECORD_STREAM_KILLED
	Command_SUBSCRIBE_EVENT

	/* A few more client->server commands */

	Command_MOVE_SINK_INPUT
	Command_MOVE_SOURCE_OUTPUT
	Command_SET_SINK_INPUT_MUTE
	Command_SUSPEND_SINK
	Command_SUSPEND_SOURCE

	Command_SET_PLAYBACK_STREAM_BUFFER_ATTR
	Command_SET_RECORD_STREAM_BUFFER_ATTR

	Command_UPDATE_PLAYBACK_STREAM_SAMPLE_RATE
	Command_UPDATE_RECORD_STREAM_SAMPLE_RATE

	/* SERVER->CLIENT */
	Command_PLAYBACK_STREAM_SUSPENDED
	Command_RECORD_STREAM_SUSPENDED
	Command_PLAYBACK_STREAM_MOVED
	Command_RECORD_STREAM_MOVED

	Command_UPDATE_RECORD_STREAM_PROPLIST
	Command_UPDATE_PLAYBACK_STREAM_PROPLIST
	Command_UPDATE_CLIENT_PROPLIST
	Command_REMOVE_RECORD_STREAM_PROPLIST
	Command_REMOVE_PLAYBACK_STREAM_PROPLIST
	Command_REMOVE_CLIENT_PROPLIST

	/* SERVER->CLIENT */
	Command_STARTED

	Command_EXTENSION

	Command_GET_CARD_INFO
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

type Commander interface {
	String() string
	Cmd() uint32
	WriteTo(io.Writer) (int, error)
	ReadFrom(io.Reader) (int, error)
}

type CommandAuth struct {
	Version uint32
	Cookie  [256]byte
}

func (c *CommandAuth) String() string {
	return fmt.Sprintf("AUTH (version %d)", c.Version)
}
func (c *CommandAuth) Cmd() uint32 {
	return Command_AUTH
}
func (c *CommandAuth) WriteTo(w io.Writer) (int, error) {
	if err := bwrite(w, Uint32, c.Version, Arbitrary, uint32(len(c.Cookie))); err != nil {
		return 0, err
	}
	n, err := w.Write(c.Cookie[:])
	if err != nil {
		return 0, err
	}
	n += 10 // include the bytes written above
	return n, nil
}
func (c *CommandAuth) ReadFrom(r io.Reader) (int, error) {
	return 0, fmt.Errorf("Not implemented")
}
