package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
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

type Source interface {
	Prefix() string
}

const maxQueueLen = 10

var quit = errors.New("QUIT")

type Server struct {
	Name           string
	Addr           string
	Debug          bool
	NewHandlerFunc NewHandlerFunc

	service  *Service
	running  bool
	quit     chan bool
	quitting bool
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
	s.service = newService(s.Name)
	s.quit = make(chan bool)

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}

	go func() {
		s.running = true
		<-s.quit
		s.quitting = true
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if s.quitting {
				return nil
			}
			return err
		}
		go func() {
			defer conn.Close()
			s.handle(conn, s.Debug)
		}()
	}
}

func (s *Server) Prefix() string {
	return s.Name
}

func (s *Server) Quit() {
	if s.running {
		s.quit <- true
	}
}

func (s *Server) handle(conn net.Conn, debug bool) {
	sendq := make(chan Message, maxQueueLen)
	user := &User{
		ServerName: s.Name,
		Host:       hostnameFromAddr(conn.RemoteAddr().String()),
		sendq:      sendq,
	}
	handler := s.NewHandlerFunc(s.service, user)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer cancel()
		if err := writer(ctx, conn, user, sendq, debug); err != nil {
			log.Printf("error: %v", err)
		}
	}()
	if err := reader(ctx, conn, user, handler, debug); err != nil {
		log.Printf("error: %v", err)
	}
}

func reader(ctx context.Context, conn net.Conn, source Source, handler Handler, debug bool) error {
	lreader := &io.LimitedReader{R: conn, N: MessageMaxLen}
	scanner := bufio.NewScanner(lreader)
	for {
		if ok := scanner.Scan(); !ok {
			return scanner.Err()
		}
		lreader.N = MessageMaxLen
		line := scanner.Text()
		if debug {
			log.Printf(" -> [%v] %v", source.Prefix(), line)
		}
		m := DecodeMessage(line)
		if _, err := handler.Handle(Command{Name: m.Cmd, Params: m.Params}); err != nil {
			if err == quit {
				return nil
			}
			log.Printf("error: %v", err)
			return err
		}
	}
}

func writer(ctx context.Context, conn net.Conn, source Source, sendq <-chan Message, debug bool) error {
	w := bufio.NewWriter(conn)
	for {
		select {
		case m := <-sendq:
			line := m.Encode()
			if debug {
				log.Printf("<-  [%v] %v", source.Prefix(), line)
			}
			if _, err := w.WriteString(line + "\n"); err != nil {
				return err
			}
			if err := w.Flush(); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func hostnameFromAddr(addr string) string {
	i := strings.LastIndex(addr, ":")
	ipAddr := addr[:i]
	name, err := net.LookupAddr(ipAddr)
	if err != nil {
		return ipAddr
	}
	return name[0]
}
