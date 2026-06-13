package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Client communicates with the queryli daemon over a Unix socket.
type Client struct {
	socketPath string
	timeout    time.Duration
}

// NewClient creates a new IPC client.
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    30 * time.Second,
	}
}

// Send sends a request to the daemon and returns the response.
func (c *Client) Send(req *Request) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return nil, fmt.Errorf("connect to daemon: %w (is the daemon running? run 'queryli connect')", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(c.timeout))

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	var resp Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &resp, nil
}

// Ping sends a ping request and returns true if daemon is reachable.
func (c *Client) Ping() (*Response, time.Duration, error) {
	start := time.Now()
	resp, err := c.Send(&Request{Type: "ping"})
	elapsed := time.Since(start)
	return resp, elapsed, err
}
