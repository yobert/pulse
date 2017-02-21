package pulse

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/user"
	"path"
	"time"
)

func bwrite(w io.Writer, data ...interface{}) (int, error) {
	n := 0
	for _, v := range data {
		if err := binary.Write(w, binary.BigEndian, v); err != nil {
			return n, err
		}
		n += binary.Size(v)
	}
	return n, nil
}
func bread(r io.Reader, data ...interface{}) error {
	for _, v := range data {
		t, ok := v.(ValueType)
		if ok {
			var tt ValueType
			if err := binary.Read(r, binary.BigEndian, &tt); err != nil {
				return err
			}
			if tt != t {
				return fmt.Errorf("Protcol error: Got type %s but expected %s", tt, t)
			}
			continue
		}

		sptr, ok := v.(*string)
		if ok {
			buf := make([]byte, 1024) // max string length i guess.
			i := 0
			for {
				_, err := r.Read(buf[i : i+1])
				if err != nil {
					return err
				}
				if buf[i] == 0 {
					*sptr = string(buf[:i])
					break
				} else {
					if i > len(buf) {
						return fmt.Errorf("String is too long (max %d bytes)", len(buf))
					}
					i++
				}
			}
			continue
		}

		bptr, ok := v.(*bool)
		if ok {
			var tt ValueType
			if err := binary.Read(r, binary.BigEndian, &tt); err != nil {
				return err
			}
			if tt == TrueValue {
				*bptr = true
			} else if tt == FalseValue {
				*bptr = false
			} else {
				return fmt.Errorf("Protcol error: Got type %s but expected boolean true or false", tt)
			}
			continue
		}

		if err := binary.Read(r, binary.BigEndian, v); err != nil {
			return err
		}
	}
	return nil
}
func bread_uint32(r io.Reader, data ...interface{}) error {
	for _, v := range data {
		var t ValueType
		if err := binary.Read(r, binary.BigEndian, &t); err != nil {
			return err
		}
		if t != Uint32Value {
			return fmt.Errorf("Protcol error: Got type %s but expected %s", t.String(), Uint32Value)
		}
		if err := binary.Read(r, binary.BigEndian, v); err != nil {
			return err
		}
	}
	return nil
}

func Ding() error {

	client, err := NewClient()
	if err != nil {
		return err
	}
	defer client.Close()

	auth := &CommandAuth{
		Version: 32,
	}

	cookie_path := os.Getenv("HOME") + "/.config/pulse/cookie"
	cookie, err := ioutil.ReadFile(cookie_path)
	if err != nil {
		return err
	}
	if len(cookie) != len(auth.Cookie) {
		return fmt.Errorf("Pulse audio client cookie has incorrect length %d: Expected %d (path %#v)",
			len(cookie), len(auth.Cookie), cookie_path)
	}
	copy(auth.Cookie[:], cookie)

	resp, err := client.Request(NewRequest(auth))
	if err != nil {
		return err
	}

	fmt.Println(" Read frame", resp.Frame)

	ar, ok := resp.Frame.Command.(*CommandAuthReply)
	if !ok {
		return fmt.Errorf("Unexpected command %v: Wanted CommandAuthReply", resp.Frame.Command)
	}

	client.SetNegotiatedVersion(auth, ar)

	current, err := user.Current()
	if err != nil {
		return err
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	set_client_name := &CommandSetClientName{
		Props: PropList{
			V: map[string]string{
				"media.format":               "WAV (Microsoft)",
				"media.name":                 "Ding!",
				"application.name":           path.Base(os.Args[0]),
				"application.process.id":     fmt.Sprintf("%d", os.Getpid()),
				"application.process.user":   current.Username,
				"application.process.host":   hostname,
				"application.process.binary": os.Args[0],
				"application.language":       "en_US.UTF-8",
				"window.x11.display":         os.Getenv("DISPLAY"),
			},
		},
	}

	resp, err = client.Request(NewRequest(set_client_name))
	if err != nil {
		return err
	}

	fmt.Println(" Read frame", resp.Frame)

	sr, ok := resp.Frame.Command.(*CommandSetClientNameReply)
	if !ok {
		return fmt.Errorf("Unexpected command %v: Wanted CommandSetClientNameReply", resp.Frame.Command)
	}

	client.SetIndex(sr.ClientIndex)

	rate := uint32(44100)

	create_playback_stream := &CommandCreatePlaybackStream{
		Format:        SampleFloat32LE,
		Channels:      1,
		Rate:          rate,
		ChannelMap:    []byte{0},
		ChannelVolume: []uint32{256},
		Props: PropList{
			V: map[string]string{
				"media.format":     set_client_name.Props.V["media.format"],
				"application.name": set_client_name.Props.V["application.name"],
				"media.name":       set_client_name.Props.V["media.name"],
			},
		},
	}

	resp, err = client.Request(NewRequest(create_playback_stream))
	if err != nil {
		return err
	}

	fmt.Println(" Read frame", resp.Frame)

	_, ok = resp.Frame.Command.(*CommandCreatePlaybackStreamReply)
	if !ok {
		return fmt.Errorf("Unexpected command %v: Wanted CommandCreatePlaybackStreamReply", resp.Frame.Command)
	}

	bytes_per_sample := 4
	max_frame_bytes := 64000
	samples_per_frame := int(max_frame_bytes / bytes_per_sample)
	seconds_per_frame := float64(samples_per_frame) / float64(rate)

	fmt.Printf("samples per frame: %d (%.4f seconds)\n", samples_per_frame, seconds_per_frame)

	t := 0.0

	for {
		f := &Frame{}
		f.Length = uint32(samples_per_frame * bytes_per_sample)
		f.Buf = &bytes.Buffer{}
		_, err := bwrite(f.Buf,
			f.Length,
			f.Channel,
			f.OffsetHigh,
			f.OffsetLow,
			f.Flags,
		)
		if err != nil {
			return err
		}

		for i := 0; i < samples_per_frame; i++ {

			v := math.Sin(t * 2 * math.Pi * 440)
			v *= 0.1

			if err := binary.Write(f.Buf, binary.LittleEndian, float32(v)); err != nil {
				return err
			}

			t += 1.0 / float64(rate)
		}

		n, err := f.Buf.WriteTo(client.conn)
		if err != nil {
			return err
		}
		fmt.Printf("wrote %d length audio frame (%d bytes)\n", f.Length, n)

		time.Sleep(time.Duration(float64(time.Second) * seconds_per_frame * 0.999))
	}

	fmt.Println("blocking")
	select {}

	return nil
}
