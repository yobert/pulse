package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func main() {
	server, err := net.Listen("unix", "/tmp/pulsedebug")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go serve(conn)
	}
}

func dbg(buf []byte) {
	for i, v := range buf {
		fmt.Printf("\t%d\t%08b %02x %03d", i, v, v, v)
		if i > 512 {
			fmt.Println("...")
			return
		}
		if v > 31 && v < 127 {
			fmt.Printf(" '%c'", v)
		}

		if v == 'L' && i < len(buf)+4 {
			r := bytes.NewReader(buf[i+1:])
			var li uint32
			err := binary.Read(r, binary.BigEndian, &li)
			if err == nil {
				fmt.Printf(" %d", li)
			} else {
				fmt.Printf(" %v", err)
			}
		}
		if v == 't' {
			str := ""
			for ii := i + 1; ii < len(buf); ii++ {
				if buf[ii] == 0 {
					break
				}
				str += string(buf[ii])
			}
			fmt.Printf(" %#v", str)
		}

		fmt.Println()
	}
}

func serve(conn net.Conn) {
	defer conn.Close()

	proxy, err := net.Dial("unix", fmt.Sprintf("/run/user/%d/pulse/native", os.Getuid()))
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		defer proxy.Close()

		buf := make([]byte, 1024*1024)
		for {
			n, err := proxy.Read(buf)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("server %d\n", n)
			dbg(buf[:n])

			_, err = conn.Write(buf[:n])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()

	buf := make([]byte, 1024*1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("client %d\n", n)
		dbg(buf[:n])

		_, err = proxy.Write(buf[:n])
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
