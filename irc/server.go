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
	"sync"
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
	Name     string
	Addr     string
	Debug    bool
	Insecure bool
	DataFile string

	NewHandlerFunc       NewHandlerFunc
	RegistrationDeadline time.Duration

	service  *Service
	running  bool
	wg       sync.WaitGroup
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
	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range Buckets {
			tx.CreateBucketIfNotExists(bucket)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to initialize database %v: %v", s.DataFile, err)
	}

	s.service = newService(s.Name, db)
	s.quit = make(chan bool)

	var tlsConfig tls.Config
	if !s.Insecure {
		err := db.View(func(tx *bolt.Tx) error {
			config := tx.Bucket(BucketConfig)
			certPem := config.Get(ConfigCert)
			keyPem := config.Get(ConfigKey)
			cert, err := tls.X509KeyPair(certPem, keyPem)
			if err != nil {
				return fmt.Errorf("unable to load certificate: %v", err)
			}
			tlsConfig = tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("unable to load certificate: %v", err)
		}
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("unable to start server: %v", err)
	}

	secureMsg := "insecure"
	if !s.Insecure {
		secureMsg = "secure"
	}
	log.Printf("%v listening on %v (%v)", s.Name, s.Addr, secureMsg)

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
			log.Printf("[%v] connection established", conn.RemoteAddr())
			s.wg.Add(1)
			defer s.wg.Done()

			var err error
			if s.Insecure {
				defer conn.Close()
				err = s.handle(conn, s.Debug)
			} else {
				tconn := tls.Server(conn, &tlsConfig)
				defer tconn.Close()
				err = s.handle(tconn, s.Debug)
			}
			if err != nil {
				log.Printf("[%v] error: %v", conn.RemoteAddr(), err)
			} else {
				log.Printf("[%v] connection closed by remote host", conn.RemoteAddr())
			}
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
	s.wg.Wait()
}

func (s *Server) handle(conn net.Conn, debug bool) error {
	cli := newClientUser(conn, s)
	handler := s.NewHandlerFunc(s.service, cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer cancel()
		if err := writer(ctx, conn, cli.User, cli.sendq, debug); err != nil {
			log.Printf("[%v] %v", conn.RemoteAddr(), err)
		}
	}()
	if err := reader(ctx, conn, cli.User, handler, debug); err != nil {
		return err
	}
	return nil
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
