package pulse

import (
	"bytes"
	"fmt"
	"io"
)

const (
	FRAME_SIZE_MAX_ALLOW = 1024 * 1024 * 16

	PA_FLAG_SHMDATA             uint32 = 0x80000000
	PA_FLAG_SHMDATA_MEMFD_BLOCK uint32 = 0x20000000
	PA_FLAG_SHMRELEASE          uint32 = 0x40000000
	PA_FLAG_SHMREVOKE           uint32 = 0xC0000000
	PA_FLAG_SHMMASK             uint32 = 0xFF000000
	PA_FLAG_SEEKMASK            uint32 = 0x000000FF
	PA_FLAG_SHMWRITABLE         uint32 = 0x00800000
)

type Frame struct {
	Client *Client
	Length uint32
	Buf    *bytes.Buffer

	Channel    uint32
	OffsetHigh uint32
	OffsetLow  uint32
	Flags      uint32

	Cmd uint32
	Tag uint32

	Command Commander
	Origin  Commander
}

func (f *Frame) String() string {
	r := fmt.Sprintf("channel %08x flags %08x offset %08x / %08x cmd %d tag %d (%d bytes)",
		f.Channel, f.Flags, f.OffsetHigh, f.OffsetLow, f.Cmd, f.Tag, f.Length)
	return r
}

func ReadFrame(r io.Reader) (*Frame, error) {
	f := &Frame{}
	f.Buf = &bytes.Buffer{}

	if _, err := io.CopyN(f.Buf, r, 4); err != nil {
		return nil, err
	}

	err := bread(f.Buf, &f.Length)
	if err != nil {
		return nil, err
	}

	if f.Length > FRAME_SIZE_MAX_ALLOW {
		return nil, fmt.Errorf("Frame size %d is too long (only %d allowed)", f.Length, FRAME_SIZE_MAX_ALLOW)
	}

	f.Buf.Grow(int(f.Length) + 16)

	_, err = io.CopyN(f.Buf, r, int64(f.Length)+16)
	if err != nil {
		return nil, err
	}

	if err = bread(f.Buf, &f.Channel, &f.OffsetHigh, &f.OffsetLow, &f.Flags); err != nil {
		return nil, err
	}

	if err = bread_uint32(f.Buf, &f.Cmd, &f.Tag); err != nil {
		return nil, err
	}

	// Don't decode the command yet. We need to associate a reply with
	// it's original request before we can do it easily.
	// See Decode()

	return f, nil
}

func (f *Frame) ReadCommand() error {
	var c Commander

	switch f.Cmd {
	case Command_REPLY:
		if f.Origin != nil {
			switch f.Origin.Cmd() {
			case Command_AUTH:
				c = &CommandAuthReply{}
			case Command_SET_CLIENT_NAME:
				c = &CommandSetClientNameReply{}
			case Command_CREATE_PLAYBACK_STREAM:
				c = &CommandCreatePlaybackStreamReply{}
			}
		}
	}

	if c == nil {
		if f.Origin == nil {
			return fmt.Errorf("Not sure how to read command code %d", f.Cmd)
		} else {
			return fmt.Errorf("Not sure how to read command code %d (from origin %s)", f.Origin.String())
		}
	}

	err := c.ReadFrom(f.Buf, f.Client.GetNegotiatedVersion())
	if err != nil {
		return err
	}

	// success!
	f.Command = c
	return nil
}

func (f *Frame) WriteTo(w io.Writer) error {
	// build out the frame's buffer

	f.Buf = &bytes.Buffer{}

	n, err := bwrite(f.Buf,
		f.Length, // dummy value-- we'll overwrite at the end when we know our final length
		f.Channel,
		f.OffsetHigh,
		f.OffsetLow,
		f.Flags,
	)
	if err != nil {
		return err
	}

	// apparently we don't want to actually count the first 20 bytes.
	n = 0

	n2, err := bwrite(f.Buf,
		Uint32Value,
		f.Cmd,
		Uint32Value,
		f.Tag,
	)
	n += n2
	if err != nil {
		return err
	}

	n2, err = f.Command.WriteTo(f.Buf, f.Client.GetNegotiatedVersion())
	n += n2
	if err != nil {
		return err
	}

	if n > FRAME_SIZE_MAX_ALLOW {
		return fmt.Errorf("Frame size %d is too long (only %d allowed)", n, FRAME_SIZE_MAX_ALLOW)
	}
	f.Length = uint32(n)

	// Rewrite size entry at start of buffer
	start := &bytes.Buffer{}
	if _, err = bwrite(start, f.Length); err != nil {
		return err
	}
	copy(f.Buf.Bytes(), start.Bytes())

	// Done! Do the actual write.
	wn, err := f.Buf.WriteTo(w)
	if err != nil {
		return err
	}

	fmt.Println("Wrote frame", f, "in", wn, "bytes")
	return nil
}
