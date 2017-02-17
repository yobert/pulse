package pulse

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

func bwrite(w io.Writer, data ...interface{}) error {
	for _, v := range data {
		if err := binary.Write(w, binary.BigEndian, v); err != nil {
			return err
		}
	}
	return nil
}
func bread(r io.Reader, data ...interface{}) error {
	for _, v := range data {
		if err := binary.Read(r, binary.BigEndian, v); err != nil {
			return err
		}
	}
	return nil
}
func bread_uint32(r io.Reader, data ...interface{}) error {
	for _, v := range data {
		var t Type
		if err := binary.Read(r, binary.BigEndian, &t); err != nil {
			return err
		}
		if t != Uint32 {
			return fmt.Errorf("Protcol error: Got type %s but expected %s", t.String(), Uint32.String())
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

	resp, err := client.Request(&Request{
		Frame: &Frame{
			Command: auth,
		},
	})
	if err != nil {
		return err
	}

	fmt.Println(" Read frame", resp.Frame)

	return nil
}
