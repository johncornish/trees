package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"trees/internal/store"
)

// TCPConnection wraps a net.Conn and implements the Connection interface.
// This is the "humble object" - thin wrapper around TCP with no logic.
type TCPConnection struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

// NewTCPConnection creates a new TCP connection wrapper.
func NewTCPConnection(conn net.Conn) *TCPConnection {
	return &TCPConnection{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

// ReadLine reads a newline-delimited line from the connection.
func (c *TCPConnection) ReadLine() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Remove trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	return line, nil
}

// WriteLine writes a newline-delimited line to the connection.
func (c *TCPConnection) WriteLine(line string) error {
	if _, err := c.writer.WriteString(line + "\n"); err != nil {
		return err
	}
	return c.writer.Flush()
}

// Close closes the underlying connection.
func (c *TCPConnection) Close() error {
	return c.conn.Close()
}

// TCPServer manages TCP listening and client connections.
type TCPServer struct {
	server *Server
	addr   string
}

// NewTCPServer creates a new TCP server.
func NewTCPServer(addr string, store *store.Store) *TCPServer {
	return &TCPServer{
		server: NewServer(store),
		addr:   addr,
	}
}

// Listen starts listening for TCP connections.
func (ts *TCPServer) Listen() error {
	listener, err := net.Listen("tcp", ts.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", ts.addr, err)
	}
	defer listener.Close()

	log.Printf("Trees server listening on %s", ts.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go ts.handleConnection(conn)
	}
}

func (ts *TCPServer) handleConnection(conn net.Conn) {
	tcpConn := NewTCPConnection(conn)
	defer tcpConn.Close()
	defer ts.server.Unsubscribe(tcpConn)

	log.Printf("Client connected: %s", conn.RemoteAddr())

	for {
		line, err := tcpConn.ReadLine()
		if err != nil {
			log.Printf("Connection closed: %s", conn.RemoteAddr())
			return
		}

		if err := ts.server.HandleMessage(tcpConn, line); err != nil {
			log.Printf("Error handling message: %v", err)
			return
		}
	}
}
