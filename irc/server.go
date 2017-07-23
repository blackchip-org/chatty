package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
)

const defaultAddress = ":6667"
const maxQueueLen = 10

var quit = errors.New("QUIT")

type Server struct {
	Name    string
	Address string
	Debug   bool
}

func (s *Server) Run() error {
	if s.Address == "" {
		s.Address = defaultAddress
	}

	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}
	log.Printf("server started on %v", s.Address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("unable to accept connection: %v", err)
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
	user    User
	sendq   chan Message
	w       *bufio.Writer
	scanner *bufio.Scanner
}

func newHandler(server *Server, conn net.Conn) *handler {
	h := &handler{
		server:  server,
		sendq:   make(chan Message, maxQueueLen),
		w:       bufio.NewWriter(conn),
		scanner: bufio.NewScanner(conn),
	}
	return h
}

func (h *handler) service() {
	log.Printf("connection established")

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		log.Println("connection closed")
	}()

	go func() {
		defer cancel()
		h.processSendQueue(ctx)
	}()

	for h.scanner.Scan() {
		line := h.scanner.Text()
		if h.server.Debug {
			log.Printf(" -> %v", line)
		}
		m := DecodeMessage(line)
		if err := h.command(m); err != nil {
			if err != quit {
				log.Printf("error: %v", err)
			}
			return
		}
	}
	if err := h.scanner.Err(); err != nil {
		log.Printf("error: %v", err)
	}
}

func (h *handler) send(cmd string, args ...string) error {
	m := NewMessage(cmd, args...)
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
			msg.Prefix = h.server.Name
			msg.Target = h.user.Nick
			line := msg.Encode()
			if h.server.Debug {
				log.Printf("<-  %v", line)
			}
			if _, err := h.w.WriteString(line + "\n"); err != nil {
				log.Printf("error: %v", err)
				return
			}
			if err := h.w.Flush(); err != nil {
				log.Printf("error: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
