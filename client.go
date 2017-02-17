package pulse

import (
	"fmt"
	"math"
	"net"
	"os"
	"sync"
)

const use_debug_proxy = true

type Client struct {
	conn net.Conn

	requests_mu      sync.Mutex
	requests_next_id uint32
	requests         map[uint32]chan *Response
}

type Request struct {
	Frame *Frame
}

type Response struct {
	Frame *Frame
	Err   error
}

func NewClient() (*Client, error) {
	addr := fmt.Sprintf("/run/user/%d/pulse/native", os.Getuid())
	if use_debug_proxy {
		addr = "/tmp/pulsedebug"
	}

	conn, err := net.Dial("unix", addr)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:             conn,
		requests_next_id: 1,
		requests:         make(map[uint32]chan *Response),
	}

	go c.reader()

	return c, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) reader() {
	defer c.conn.Close()

	for {
		frame, err := ReadFrame(c.conn)
		if err != nil {
			fmt.Println(err)
			return
		}

		c.requests_mu.Lock()
		rc := c.requests[frame.Tag]
		if rc != nil {
			delete(c.requests, frame.Tag)
		}
		c.requests_mu.Unlock()

		if rc == nil {
			fmt.Println("!Read frame", frame)
		} else {
			rc <- &Response{Frame: frame}
			close(rc)
		}
	}
}

func (c *Client) Request(req *Request) (*Response, error) {
	rc := make(chan *Response)

	c.requests_mu.Lock()
	for c.requests[c.requests_next_id] != nil {
		c.requests_next_id++
		if c.requests_next_id == math.MaxUint32 {
			c.requests_next_id = 0
		}
	}
	id := c.requests_next_id
	c.requests[id] = rc
	c.requests_mu.Unlock()

	req.Frame.Length = 0
	req.Frame.Channel = 0xffffffff
	req.Frame.OffsetHigh = 0
	req.Frame.OffsetLow = 0
	req.Frame.Flags = 0
	req.Frame.Cmd = req.Frame.Command.Cmd()
	req.Frame.Tag = id

	err := req.Frame.WriteTo(c.conn)
	if err != nil {
		// error during write: assume we aren't going
		// to get a response back.
		c.requests_mu.Lock()
		if c.requests[id] == nil {
			// whoa, apparently we did get an insanely fast response back.
			// the message should come on the channel below.
			fmt.Println("Error writing frame:", err, "but we got a response back anyways?!?", req.Frame)
			err = nil // don't early out
		} else {
			delete(c.requests, id)
		}
		c.requests_mu.Unlock()

		if err != nil {
			return nil, err
		}
	}

	resp := <-rc
	return resp, resp.Err
}
