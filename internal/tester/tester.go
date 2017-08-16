package tester

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/blackchip-org/chatty/irc"
)

var (
	RealServer bool
	nextPort   = 6667
)

func init() {
	flag.BoolVar(&RealServer, "real-server", false, "run tests using a real server")
}

var errRecvTimeout = errors.New("recv timeout")

type Server struct {
	Actual    *irc.Server
	server    *irc.Server
	clients   []*Client
	err       error
	t         *testing.T
	nextID    int
	timeStart time.Time
}

type Client struct {
	id     int
	conn   net.Conn
	recvq  chan string
	w      *bufio.Writer
	debug  bool
	err    error
	t      *testing.T
	server *Server
	nsent  int
}

func NewServer(t *testing.T) (*Server, *Client) {
	addr := ":" + strconv.Itoa(nextPort)
	if !RealServer {
		nextPort++
		if nextPort > 6668 {
			nextPort = 6667
		}
	}
	ts := &Server{
		server: &irc.Server{
			Name:       "irc.localhost",
			Addr:       addr,
			Insecure:   true,
			DataFile:   "fntests.db",
			NoAutoInit: true,
		},
		clients:   make([]*Client, 0),
		t:         t,
		nextID:    1,
		timeStart: time.Now(),
	}
	ts.Actual = ts.server
	if !RealServer {
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
				} else {
					return
				}
			}
		}()
	} else {
		out, err := exec.Command("docker", "start", "irc-test").CombinedOutput()
		if err != nil {
			t.Fatalf("unable to start docker container:\n%v", string(out))
		}
	}
	tc := ts.NewClient()
	return ts, tc
}

func (s *Server) NewClient() *Client {
	tc := &Client{
		recvq:  make(chan string, 1024),
		t:      s.t,
		server: s,
		id:     s.nextID,
	}
	s.nextID++
	s.clients = append(s.clients, tc)
	if s.err != nil {
		tc.err = s.err
		return tc
	}
	err := tc.connect(s.server.Addr)
	if err != nil {
		tc.err = err
		return tc
	}
	go func() {
		if err := tc.reader(); err != nil {
			tc.err = err
		}
	}()
	if RealServer {
		tc.debug = true
	}
	return tc
}

func (s *Server) Quit() {
	for _, client := range s.clients {
		if client.conn != nil {
			client.drainNoWait()
			client.Send("QUIT")
			client.conn.Close()
		}
	}
	s.server.Quit()
	if RealServer {
		out, err := exec.Command("docker", "kill", "irc-test").CombinedOutput()
		if err != nil {
			s.t.Errorf("error stopping docker instance: \n%v", string(out))
		}
	}
}

func (c *Client) connect(addr string) error {
	retries := 0
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			if retries >= 9 {
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

func (c *Client) Send(line string) {
	if c.err != nil {
		return
	}
	// When testing against a real server, it won't let us flood it and needs to be throttled.
	// https://tools.ietf.org/html/rfc2813#section-5.8
	if RealServer && c.nsent >= 10 {
		c.logf(" z ", "\tthrottle")
		time.Sleep(2500 * time.Millisecond)
	}
	c.logf(" ->", line)
	_, err := c.w.WriteString(line + "\r\n")
	if err != nil {
		c.err = err
		return
	}
	if err := c.w.Flush(); err != nil {
		c.err = err
		return
	}
	c.nsent++
}

func (c *Client) SendMessage(cmd string, params ...string) {
	m := irc.NewMessage(cmd, params...)
	c.Send(m.Encode())
}

func (c *Client) Recv() string {
	if c.err != nil {
		return ""
	}

	timeout := 300 * time.Millisecond
	if RealServer {
		timeout = 5000 * time.Millisecond
	}

	timer := time.NewTimer(timeout)
	select {
	case line := <-c.recvq:
		line = normalizeLine(line)
		c.logf("<- ", line)
		return line
	case <-timer.C:
		c.err = errRecvTimeout
		return "recv timeout"
	}
}

func (c *Client) Drain() *Client {
	if c.err != nil {
		return c
	}
	c.logf(" x ", "\tdraining")
	for {
		_ = c.Recv()
		if c.err != nil {
			if c.err == errRecvTimeout {
				c.err = nil
			}
			c.logf(" x ", "\tdrained")
			return c
		}
	}
}

func (c *Client) drainNoWait() *Client {
	for {
		select {
		case line := <-c.recvq:
			line = normalizeLine(line)
			c.logf("<- ", line)
		default:
			return c
		}
	}
}

func (c *Client) RecvMessage() irc.Message {
	line := c.Recv()
	return irc.DecodeMessage(line)
}

func (c *Client) reader() error {
	scanner := bufio.NewScanner(c.conn)
	for {
		if ok := scanner.Scan(); !ok {
			return scanner.Err()
		}
		line := scanner.Text()
		c.recvq <- line
	}
}

func (c *Client) WaitForAny(expecting []string) irc.Message {
	c.logf(" ! ", "\twait %v", strings.Join(expecting, " or "))
	for {
		m := c.RecvMessage()
		if c.err != nil {
			c.logf(" * ", "\terror %v", c.err)
			return irc.Message{}
		}
		for _, expect := range expecting {
			if m.Cmd == expect {
				c.logf(" . ", "\trecv %v", expect)
				return m
			}
		}
	}
}

func (c *Client) WaitFor(reply string) irc.Message {
	return c.WaitForAny([]string{reply})
}

func (c *Client) Login(nick string, user string) *Client {
	//if RealServer {
	//	c.WaitFor("020")
	//}
	c.Send("NICK " + nick)
	c.Send("USER " + user)
	c.WaitForAny([]string{irc.RplEndOfMotd, irc.ErrNoMotd})
	return c
}

func (c *Client) LoginDefault() *Client {
	c.Login("Batman", "Batman 0 * :Bruce Wayne")
	return c
}

func (c *Client) Join(chname string) *Client {
	c.Send("JOIN " + chname)
	c.WaitFor(irc.RplEndOfNames)
	return c
}

func (c *Client) Err() error {
	return c.err
}

func (c *Client) logf(prefix string, format string, params ...interface{}) {
	tail := fmt.Sprintf(format, params...)
	since := time.Since(c.server.timeStart) / time.Millisecond
	c.t.Logf("%5d %v [%v] %v", since, prefix, c.id, tail)
}

// Replace server specific host info with localhost for testing
func normalizeLine(line string) string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, ":") {
		return line
	}
	parts := strings.Split(line, " ")
	at := strings.Index(parts[0], "@")
	if at < 0 {
		return line
	}
	parts[0] = parts[0][:at] + "@localhost"
	return strings.Join(parts, " ")
}
