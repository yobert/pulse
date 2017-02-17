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
	Length uint32
	Buf    *bytes.Buffer

	Channel    uint32
	OffsetHigh uint32
	OffsetLow  uint32
	Flags      uint32

	Cmd uint32
	Tag uint32

	Command Commander
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

	switch f.Cmd {
	case Command_AUTH:
	case Command_REPLY:
	}

	return f, nil
}

func (f *Frame) WriteTo(w io.Writer) error {
	// build out the frame's buffer

	f.Buf = &bytes.Buffer{}

	if err := bwrite(f.Buf,
		f.Length, // dummy value-- we'll overwrite at the end when we know our final length
		f.Channel,
		f.OffsetHigh,
		f.OffsetLow,
		f.Flags,
	); err != nil {
		return err
	}

	if err := bwrite(f.Buf,
		Uint32,
		f.Cmd,
		Uint32,
		f.Tag,
	); err != nil {
		return err
	}

	n, err := f.Command.WriteTo(f.Buf)
	if err != nil {
		return err
	}

	n += 10 // For the command and tag entries above
	if n > FRAME_SIZE_MAX_ALLOW {
		return fmt.Errorf("Frame size %d is too long (only %d allowed)", n, FRAME_SIZE_MAX_ALLOW)
	}
	f.Length = uint32(n)

	// Rewrite size entry at start of buffer
	start := &bytes.Buffer{}
	if err = bwrite(start, f.Length); err != nil {
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
