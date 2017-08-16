package irc

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const (
	Version = "chatty-0"
)

var (
	Addr       = ":6697"
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

type Server struct {
	Name       string
	Addr       string
	Debug      bool
	Insecure   bool
	CertFile   string
	KeyFile    string
	DataFile   string
	NoAutoInit bool

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

	boltOpts := bolt.Options{Timeout: 5 * time.Second}
	db, err := bolt.Open(s.DataFile, 0600, &boltOpts)
	if err != nil {
		return fmt.Errorf("unable to open database %v: %v", s.DataFile, err)
	}
	defer db.Close()

	s.service = newService(s.Name, db)
	s.quit = make(chan bool)

	var config tls.Config
	if !s.Insecure {
		cert, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
		if err != nil {
			return fmt.Errorf("certificate error: %v", err)
		}
		config = tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}

	if err := s.init(db); err != nil {
		return err
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
			if s.Insecure {
				defer conn.Close()
				s.handle(conn, s.Debug)
				return
			}
			tconn := tls.Server(conn, &config)
			defer tconn.Close()
			s.handle(tconn, s.Debug)
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

func (s *Server) init(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		config := tx.Bucket([]byte("config"))
		if config != nil {
			return nil
		}
		config, err := tx.CreateBucket([]byte("config"))
		if err != nil {
			return err
		}
		if s.NoAutoInit {
			return nil
		}
		password := "swordfish"
		config.Put([]byte("password"), []byte(password))
		fmt.Printf("server password is: %v\n", password)
		return nil
	})
	return err
}
