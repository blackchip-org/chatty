package irc

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type UserID uint64

type User struct {
	ID         UserID
	Nick       string
	Name       string
	Host       string
	RealHost   string
	FullName   string
	ServerName string
}

type Client struct {
	U          *User
	conn       net.Conn
	mutex      sync.RWMutex
	err        error
	registered bool
	sendq      chan Message
	chans      map[string]*Chan
}

var (
	nextID = UserID(1)
	mutex  sync.Mutex
)

func newClientUser(conn net.Conn, server *Server) *Client {
	mutex.Lock()
	c := &Client{
		U: &User{
			ID:         nextID,
			ServerName: server.Name,
			Host:       server.Name,
			RealHost:   hostnameFromAddr(conn.RemoteAddr().String()),
		},
		conn:  conn,
		sendq: make(chan Message, queueMaxLen),
		chans: make(map[string]*Chan),
	}
	conn.SetDeadline(time.Now().Add(server.RegistrationDeadline))
	nextID++
	mutex.Unlock()
	return c
}

func (c *Client) Send(cmd string, params ...string) *Client {
	m := Message{Prefix: c.U.ServerName, Cmd: cmd, Params: params}
	c.send(m)
	return c
}

func (c *Client) Reply(cmd string, params ...string) *Client {
	m := Message{Prefix: c.U.ServerName, Target: c.U.Nick, Cmd: cmd, Params: params}
	c.send(m)
	return c
}

func (c *Client) Relay(o Origin, cmd string, params ...string) *Client {
	m := Message{Prefix: o.Origin(), Cmd: cmd, Params: params}
	c.send(m)
	return c
}

func (c *Client) SendError(err *Error) *Client {
	nick := "*"
	if c.U.Nick != "" {
		nick = c.U.Nick
	}
	m := Message{Prefix: c.U.ServerName, Target: nick, Cmd: err.Numeric, Params: err.Params}
	c.send(m)
	return c
}

func (c *Client) SetRegistered() {
	c.registered = true
	c.conn.SetDeadline(time.Time{})
}

func (c *Client) send(m Message) {
	if c.err != nil {
		return
	}
	select {
	case c.sendq <- m:
		return
	default:
		c.err = errors.New("send queue full")
	}
}

func (c *Client) Quit() {
	c.err = quit
}

func (u User) Origin() string {
	nick := "*"
	name := "*"
	host := "*"

	if u.Nick != "" {
		nick = u.Nick
	}
	if u.Name != "" {
		name = u.Name
	}
	if u.Host != "" {
		host = u.Host
	}
	// FIXME: Not sure why ircd-irc2 returns a tilde in the name
	return fmt.Sprintf("%s!~%s@%s", nick, name, host)
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