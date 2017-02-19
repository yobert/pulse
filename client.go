package pulse

import (
	"fmt"
	"math"
	"net"
	"os"
	"sync"
)

//const use_debug_proxy = true
const use_debug_proxy = false

type Client struct {
	conn net.Conn

	requests_mu      sync.Mutex
	requests_next_id uint32
	requests         map[uint32]chan *Response

	version uint32
	shm     bool
	memfd   bool

	client_index uint32
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

func (c *Client) SetNegotiatedVersion(req *CommandAuth, resp *CommandAuthReply) {
	c.version = req.Version
	if resp.Version < req.Version {
		c.version = resp.Version
	}
	c.shm = req.Shm
	if !resp.Shm {
		c.shm = false
	}
	c.memfd = req.Memfd
	if !resp.Memfd {
		c.memfd = false
	}
	fmt.Printf("Negotiated native protocol version %d (shm %v memfd %v)\n", c.version, c.shm, c.memfd)
}

func (c *Client) GetNegotiatedVersion() uint32 {
	return c.version
}

func (c *Client) SetIndex(index uint32) {
	c.client_index = index
	fmt.Printf("Client index %d\n", c.client_index)
}

func (c *Client) reader() {
	defer c.conn.Close()

	for {
		frame, err := ReadFrame(c.conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		frame.Client = c

		c.requests_mu.Lock()
		rc := c.requests[frame.Tag]
		if rc != nil {
			delete(c.requests, frame.Tag)
		}
		c.requests_mu.Unlock()

		if rc == nil {
			if frame.Cmd == Command_REQUEST {
				// Ignore these for now.
			} else {
				fmt.Println("!Read frame", frame, "cmd")
			}
		} else {
			rc <- &Response{Frame: frame}
			close(rc)
		}
	}
}

func NewRequest(command Commander) *Request {
	return &Request{
		&Frame{
			Command: command,
		},
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

	req.Frame.Client = c
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

	if resp.Err != nil {
		return resp, resp.Err
	}

	resp.Frame.Origin = req.Frame.Command
	err = resp.Frame.ReadCommand()
	if err != nil {
		return resp, err
	}

	fmt.Println(req.Frame.Command, "->", resp.Frame.Command)

	return resp, nil
}
