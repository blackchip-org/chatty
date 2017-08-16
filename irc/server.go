package irc

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const (
	Version = "chatty-0"
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

type Origin interface {
	Origin() string
}

const queueMaxLen = 10

//var quit = errors.New("QUIT")

type Server struct {
	Name                 string
	Addr                 string
	Debug                bool
	NewHandlerFunc       NewHandlerFunc
	RegistrationDeadline time.Duration

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
	if int(s.RegistrationDeadline) == 0 {
		s.RegistrationDeadline = 10 * time.Second
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
	cli := newClientUser(conn, s)
	handler := s.NewHandlerFunc(s.service, cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer cancel()
		if err := writer(ctx, conn, cli.User, cli.sendq, debug); err != nil {
			log.Printf("error: %v", err)
		}
	}()
	if err := reader(ctx, conn, cli.User, handler, debug); err != nil {
		log.Printf("error: %v", err)
	}
}

func reader(ctx context.Context, conn net.Conn, o Origin, handler Handler, debug bool) error {
	lreader := &io.LimitedReader{R: conn, N: MessageMaxLen}
	scanner := bufio.NewScanner(lreader)
	for {
		if ok := scanner.Scan(); !ok {
			return scanner.Err()
		}
		lreader.N = MessageMaxLen
		line := scanner.Text()
		if debug {
			log.Printf(" -> [%v] %v", o.Origin(), line)
		}
		m := DecodeMessage(line)
		if err := handler.Handle(Command{Name: m.Cmd, Params: m.Params}); err != nil {
			if err == Quit {
				return nil
			}
			log.Printf("error: %v", err)
			return err
		}
	}
}

func writer(ctx context.Context, conn net.Conn, o Origin, sendq <-chan Message, debug bool) error {
	w := bufio.NewWriter(conn)
	for {
		select {
		case m := <-sendq:
			line := m.Encode()
			if debug {
				log.Printf("<-  [%v] %v", o.Origin(), line)
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
