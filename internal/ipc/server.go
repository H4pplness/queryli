package ipc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

// Handler processes incoming requests and returns responses.
type Handler interface {
	HandleRequest(req *Request) *Response
}

// Server listens on a Unix socket and dispatches requests to a handler.
type Server struct {
	socketPath string
	listener   net.Listener
	handler    Handler
	wg         sync.WaitGroup
	quit       chan struct{}
}

// NewServer creates a new IPC server.
func NewServer(socketPath string, handler Handler) *Server {
	return &Server{
		socketPath: socketPath,
		handler:    handler,
		quit:       make(chan struct{}),
	}
}

// Listen starts the server and accepts connections.
func (s *Server) Listen() error {
	// Remove stale socket file
	os.Remove(s.socketPath)

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.socketPath, err)
	}
	s.listener = listener

	// Restrict socket permissions
	os.Chmod(s.socketPath, 0600)

	// Use fmt instead of log since log may be redirected
	fmt.Fprintf(os.Stderr, "Server listening on %s\n", s.socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				log.Printf("Server: quit signal received, stopping accept loop")
				return nil
			default:
				log.Printf("Server: accept error: %v", err)
				return fmt.Errorf("accept: %w", err)
			}
		}

		s.wg.Add(1)
		go s.handleConn(conn)
	}
}

// Close stops the server and cleans up.
func (s *Server) Close() error {
	close(s.quit)
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	os.Remove(s.socketPath)
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	var req Request
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Server: decode error: %v", err)
		return
	}

	resp := s.handler.HandleRequest(&req)

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("Server: encode error: %v", err)
	}
}
