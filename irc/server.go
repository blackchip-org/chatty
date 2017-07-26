package irc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
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

	mutex    sync.RWMutex
	channels map[string]*Channel
	nicks    map[string]*User
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
	s.quit = make(chan bool)
	s.channels = make(map[string]*Channel)
	s.nicks = make(map[string]*User)

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}
	log.Printf("server started on %v", s.Addr)

	go func() {
		<-s.quit
		s.quitting = true
		log.Printf("server shutting down")
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
			s.service(conn, s.Debug)
		}()
	}
}

func (s *Server) Prefix() string {
	return s.Name
}

func (s *Server) Quit() {
	s.quit <- true
}

func (s *Server) service(conn net.Conn, debug bool) {
	log.Printf("connection established")

	sendq := make(chan Message, maxQueueLen)
	user := &User{
		ServerName: s.Name,
		Host:       conn.RemoteAddr().String(),
		sendq:      sendq,
	}
	handler := s.NewHandlerFunc(s, user)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		log.Println("connection closed")
	}()

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
	scanner := bufio.NewScanner(conn)
	for {
		if ok := scanner.Scan(); !ok {
			return scanner.Err()
		}
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

/*
func (s *Server) JoinChannel(u *User, name string) (*Channel, *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.channels[name]
	if !exists {
		ch = NewChannel(name)
		s.channels[name] = ch
	}
	ch.Join(u)
	return ch, nil
}
*/

func (s *Server) Nick(u *User, nick string) *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exists := s.nicks[nick]; exists {
		return NewError(ErrNickNameInUse, nick)
	}
	delete(s.nicks, u.Nick)
	s.nicks[nick] = u
	u.Nick = nick
	return nil
}
