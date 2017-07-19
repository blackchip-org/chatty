package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

const defaultAddress = ":6667"
const maxQueueLen = 10

var quit = errors.New("QUIT")

type Server struct {
	Name    string
	Address string
	Debug   bool
	log     *log.Logger
	dlog    *log.Logger
}

func (s *Server) Run() error {
	if s.Address == "" {
		s.Address = defaultAddress
	}
	s.log = log.New(os.Stdout, s.Name, log.LstdFlags)
	s.dlog = log.New(ioutil.Discard, s.Name, log.LstdFlags)
	if s.Debug {
		s.dlog.SetOutput(os.Stdout)
	}

	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}
	s.log.Printf("server started on %v", s.Address)
	s.dlog.Printf("debug logging enabled")
	for {
		conn, err := listener.Accept()
		if err != nil {
			s.log.Printf("unable to accept -- %v", err)
			continue
		}
		handler := newHandler(s, conn)
		go func() {
			defer conn.Close()
			handler.service()
		}()
	}
}

type handler struct {
	server  *Server
	sendq   chan Message
	w       *bufio.Writer
	scanner *bufio.Scanner
	log     *log.Logger
	dlog    *log.Logger
}

func newHandler(server *Server, conn net.Conn) *handler {
	prefix := fmt.Sprintf("%v ", conn.RemoteAddr())
	h := &handler{
		server:  server,
		sendq:   make(chan Message, maxQueueLen),
		w:       bufio.NewWriter(conn),
		scanner: bufio.NewScanner(conn),
		log:     log.New(os.Stdout, prefix, log.LstdFlags),
		dlog:    log.New(ioutil.Discard, prefix, log.LstdFlags),
	}
	if server.Debug {
		h.dlog.SetOutput(os.Stdout)
	}
	return h
}

func (h *handler) service() {
	h.log.Println("connection established")

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		h.log.Println("connection closed")
	}()

	go func() {
		defer cancel()
		h.processSendQueue(ctx)
	}()

	for h.scanner.Scan() {
		line := h.scanner.Text()
		h.dlog.Printf(" -> %v", line)
		m, err := DecodeMessage(line)
		if err != nil {
			h.log.Printf("error: %v", err)
			return
		}
		if err := h.command(m); err != nil {
			if err != quit {
				h.log.Printf("error: %v", err)
			}
			return
		}
	}
	if err := h.scanner.Err(); err != nil {
		h.log.Printf("error: %v", err)
	}
}

func (h *handler) send(cmd string, args ...string) error {
	m := Message{Cmd: cmd, Args: args}
	select {
	case h.sendq <- m:
		return nil
	default:
		return errors.New("send queue full")
	}
}

func (h *handler) processSendQueue(ctx context.Context) {
	for {
		select {
		case msg := <-h.sendq:
			line, err := msg.Encode()
			if err != nil {
				h.log.Printf("error: cannot encode message: %v", err)
				continue
			}
			h.dlog.Printf("<-  %v", line)
			if _, err := h.w.WriteString(line + "\n"); err != nil {
				h.log.Printf("error: %v", err)
				return
			}
			if err := h.w.Flush(); err != nil {
				h.log.Printf("error: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
