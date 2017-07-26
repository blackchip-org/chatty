package irc

import (
	"bufio"
	"errors"
	"log"
	"net"
	"time"
)

type TestServer struct {
	server  *Server
	clients []*TestClient
	err     error
}

type TestClient struct {
	conn  net.Conn
	recvq chan string
	w     *bufio.Writer
	err   error
}

func NewTestServer() (*TestServer, *TestClient) {
	ts := &TestServer{
		server: &Server{
			Name:  "example.com",
			Debug: true,
		},
		clients: make([]*TestClient, 0),
	}
	go func() {
		retries := 0
		for {
			if err := ts.server.ListenAndServe(); err != nil {
				if retries >= 10 {
					log.Printf("server error: %v\n", err)
					ts.err = err
					return
				}
				retries++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
	tc := ts.NewClient()
	return ts, tc
}

func (s *TestServer) NewClient() *TestClient {
	tc := &TestClient{
		recvq: make(chan string, 1024),
	}
	s.clients = append(s.clients, tc)
	if s.err != nil {
		tc.err = s.err
		return tc
	}
	err := tc.connect()
	if err != nil {
		tc.err = err
		return tc
	}
	go func() {
		if err := tc.reader(); err != nil {
			tc.err = err
		}
	}()
	return tc
}

func (s *TestServer) Quit() {
	for _, client := range s.clients {
		if client.conn != nil {
			client.conn.Close()
		}
	}
	s.server.Quit()
}

func (c *TestClient) connect() error {
	retries := 0
	for {
		conn, err := net.Dial("tcp", Addr)
		if err != nil {
			if retries >= 10 {
				return err
			}
		} else {
			c.conn = conn
			c.w = bufio.NewWriter(conn)
			return nil
		}
		retries++
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *TestClient) Send(line string) {
	if c.err != nil {
		return
	}
	_, err := c.w.WriteString(line + "\n")
	if err != nil {
		c.err = err
		return
	}
	if err := c.w.Flush(); err != nil {
		c.err = err
		return
	}
}

func (c *TestClient) Recv() string {
	if c.err != nil {
		return ""
	}
	retries := 0
	for {
		select {
		case line := <-c.recvq:
			return line
		default:
			retries++
			if retries > 10 {
				c.err = errors.New("recv timeout")
				return "recv timeout"
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *TestClient) reader() error {
	scanner := bufio.NewScanner(c.conn)
	for {
		if ok := scanner.Scan(); !ok {
			return scanner.Err()
		}
		line := scanner.Text()
		c.recvq <- line
	}
}
