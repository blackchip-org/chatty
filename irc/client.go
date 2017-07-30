package irc

import (
	"errors"
	"fmt"
	"sync"
)

type UserID uint64

type User struct {
	ID         UserID
	Nick       string
	Name       string
	Host       string
	FullName   string
	ServerName string
}

type Client struct {
	U     *User
	mutex sync.RWMutex
	err   error
	sendq chan Message
}

var (
	nextID = UserID(1)
	mutex  sync.Mutex
)

func NewClientUser(serverName string, host string) *Client {
	mutex.Lock()
	c := &Client{
		U: &User{
			ID:         nextID,
			ServerName: serverName,
			Host:       host,
		},
		sendq: make(chan Message, queueMaxLen),
	}
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

func (c *Client) Relay(source Source, cmd string, params ...string) *Client {
	m := Message{Prefix: source.Prefix(), Cmd: cmd, Params: params}
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

func (u User) Prefix() string {
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
