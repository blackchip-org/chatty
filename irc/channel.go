package irc

import (
	"sort"
	"sync"
)

type Channel struct {
	name    string
	topic   string
	status  string
	clients map[string]*Client
	umodes  map[string]UserChanMode
	mutex   sync.RWMutex
}

type UserChanMode string

const (
	UserChan   UserChanMode = ""
	UserChanOp              = "o"
)

var UserChanPrefixes = map[UserChanMode]string{
	UserChan:   "",
	UserChanOp: "@",
}

func NewChannel(name string) *Channel {
	c := &Channel{
		name:    name,
		topic:   "no topic",
		clients: make(map[string]*Client),
		umodes:  make(map[string]UserChanMode),
	}
	return c
}

func (c *Channel) Name() string {
	return c.name
}

func (c *Channel) Topic() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.topic
}

func (c *Channel) Status() string {
	// https://modern.ircdocs.horse/#rplnamreply-353
	return "="
}

func (c *Channel) Nicks() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	nicks := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		prefix := UserChanPrefixes[c.umodes[cli.U.Nick]]
		nicks = append(nicks, prefix+cli.U.Nick)
	}
	sort.Strings(nicks)
	return nicks
}

func (c *Channel) Join(cli *Client) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.clients) == 0 {
		c.umodes[cli.U.Nick] = UserChanOp
	} else {
		c.umodes[cli.U.Nick] = UserChan
	}
	c.clients[cli.U.Nick] = cli
	names := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		cli.Relay(cli.U, JoinCmd, c.name)
		names = append(names, cli.U.Nick)
	}
	return nil
}

func (c *Channel) PrivMsg(src *Client, text string) *Error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, exists := c.clients[src.U.Nick]; !exists {
		return NewError(ErrCannotSendToChan, c.name)
	}
	for _, cli := range c.clients {
		if cli.U.Nick == src.U.Nick {
			continue
		}
		cli.Relay(src.U, PrivMsgCmd, c.name, text)
	}
	return nil
}
