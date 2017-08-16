package irc

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	User       *User
	ServerName string
	conn       net.Conn
	mutex      sync.RWMutex
	err        error
	registered bool
	sendq      chan Message

	password string
	chans    map[string]*Chan
}

func newClientUser(conn net.Conn, server *Server) *Client {
	host := server.Name
	realHost := hostnameFromAddr(conn.RemoteAddr().String())

	c := &Client{
		User:       newUser(host, realHost),
		ServerName: server.Name,
		conn:       conn,
		sendq:      make(chan Message, queueMaxLen),
		chans:      make(map[string]*Chan),
	}
	conn.SetDeadline(time.Now().Add(server.RegistrationDeadline))
	return c
}

func (c *Client) Send(cmd string, params ...string) *Client {
	m := Message{Prefix: c.ServerName, Cmd: cmd, Params: params}
	c.SendMessage(m)
	return c
}

func (c *Client) Reply(cmd string, params ...string) *Client {
	text, exists := RplText[cmd]
	if exists {
		params = append(params, text)
	}
	m := Message{
		Prefix: c.ServerName,
		Target: c.User.Nick,
		Cmd:    cmd,
		Params: params,
	}
	c.SendMessage(m)
	return c
}

func (c *Client) Relay(o Origin, cmd string, params ...string) *Client {
	m := Message{Prefix: o.Origin(), Cmd: cmd, Params: params}
	c.SendMessage(m)
	return c
}

func (c *Client) SendError(err error) *Client {
	var numeric string
	var params []string

	if ircErr, ok := err.(*Error); ok {
		numeric = ircErr.Numeric
		params = ircErr.Params
	} else {
		numeric = "000"
		params = []string{err.Error()}
	}

	nick := "*"
	if c.User.Nick != "" {
		nick = c.User.Nick
	}
	m := Message{
		Prefix: c.ServerName,
		Target: nick,
		Cmd:    numeric,
		Params: params,
	}
	c.SendMessage(m)
	return c
}

func (c *Client) SetRegistered() {
	c.registered = true
	c.conn.SetDeadline(time.Time{})
}

func (c *Client) SendMessage(m Message) {
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
	c.err = Quit
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
