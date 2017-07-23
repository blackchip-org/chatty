package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	Addr       = ":6667"
	ServerName = "localhost"
)

func init() {
	name, err := os.Hostname()
	if err == nil {
		ServerName = name
	}
}

const maxQueueLen = 10

var quit = errors.New("QUIT")

type Server struct {
	Name           string
	Addr           string
	Debug          bool
	NewHandlerFunc NewHandlerFunc
}

func (s *Server) ListenAndServe() error {
	if s.Addr == "" {
		s.Addr = Addr
	}
	if s.Name == "" {
		s.Name = ServerName
	}
	if s.NewHandlerFunc == nil {
		s.NewHandlerFunc = NewDefaultHandler
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}
	log.Printf("server started on %v", s.Addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connection: %v", err)
			continue
		}
		sconn := newServerConn(conn, s.Debug)
		go func() {
			defer conn.Close()
			s.service(sconn)
		}()
	}
}

func (s *Server) service(conn *serverConn) {
	log.Printf("connection established")

	sendq := make(chan Message, maxQueueLen)
	replySender := &ReplySender{
		ServerName: s.Name,
		sendq:      sendq,
	}
	handler := s.NewHandlerFunc(replySender)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		log.Println("connection closed")
	}()

	// Process the send queue
	go func() {
		defer cancel()
		for {
			select {
			case m := <-sendq:
				err := conn.Write(m)
				if err != nil {
					log.Printf("error: %v", err)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		m, err := conn.Read()
		if err != nil {
			log.Printf("error: %v", err)
			return
		}
		handler.HandleCommand(Command{Name: m.Cmd, Params: m.Params})
		err = replySender.err
		if err != nil {
			if err != quit {
				log.Printf("error: %v", err)
			}
			return
		}
	}
}

type Command struct {
	Name   string
	Params []string
}

type ReplySender struct {
	ServerName string
	Target     string
	err        error
	sendq      chan Message
}

func (r *ReplySender) Send(cmd string, args ...string) *ReplySender {
	if r.err != nil {
		return r
	}
	m := NewMessage(cmd, args...)
	m.Prefix = r.ServerName
	m.Target = r.Target
	select {
	case r.sendq <- m:
		return r
	default:
		r.err = errors.New("send queue full")
	}
	return r
}

func (r *ReplySender) Quit() {
	r.err = quit
}

type serverConn struct {
	debug   bool
	w       *bufio.Writer
	scanner *bufio.Scanner
}

func newServerConn(conn net.Conn, debug bool) *serverConn {
	s := &serverConn{
		debug:   debug,
		w:       bufio.NewWriter(conn),
		scanner: bufio.NewScanner(conn),
	}
	return s
}

func (c *serverConn) Read() (Message, error) {
	if ok := c.scanner.Scan(); !ok {
		return Message{}, c.scanner.Err()
	}
	line := c.scanner.Text()
	if c.debug {
		log.Printf(" -> %v", line)
	}
	return DecodeMessage(line), nil
}

func (c *serverConn) Write(m Message) error {
	line := m.Encode()
	if c.debug {
		log.Printf("<-  %v", line)
	}
	if _, err := c.w.WriteString(line + "\n"); err != nil {
		return err
	}
	if err := c.w.Flush(); err != nil {
		return err
	}
	return nil
}
