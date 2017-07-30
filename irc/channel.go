package irc

import (
	"sort"
	"sync"
)

type Chan struct {
	name    string
	topic   string
	status  string
	clients map[UserID]*Client
	umodes  map[UserID]UserChanModes
	mutex   sync.RWMutex
}

type UserChanModes struct {
	Op    bool
	Voice bool
}

func (u UserChanModes) Prefix() string {
	switch {
	case u.Op:
		return "@"
	case u.Voice:
		return "+"
	}
	return ""
}

const (
	ChanModeOp    = "o"
	ChanModeVoice = "v"
)

func NewChan(name string) *Chan {
	c := &Chan{
		name:    name,
		topic:   "no topic",
		clients: make(map[UserID]*Client),
		umodes:  make(map[UserID]UserChanModes),
	}
	return c
}

func (c *Chan) Name() string {
	return c.name
}

func (c *Chan) Topic() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.topic
}

func (c *Chan) Status() string {
	// https://modern.ircdocs.horse/#rplnamreply-353
	return "="
}

func (c *Chan) Nicks() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	nicks := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		prefix := c.umodes[cli.U.ID].Prefix()
		nicks = append(nicks, prefix+cli.U.Nick)
	}
	sort.Strings(nicks)
	return nicks
}

func (c *Chan) Join(cli *Client) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	umodes := UserChanModes{}
	if len(c.clients) == 0 {
		umodes.Op = true
	}
	c.umodes[cli.U.ID] = umodes
	c.clients[cli.U.ID] = cli
	names := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		cli.Relay(cli.U, JoinCmd, c.name)
		names = append(names, cli.U.Nick)
	}
	return nil
}

func (c *Chan) Part(src *Client, reason string) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, exists := c.clients[src.U.ID]
	if !exists {
		return NewError(ErrNotOnChannel)
	}
	for _, cli := range c.clients {
		cli.Relay(src.U, PartCmd, c.name, reason)
	}
	delete(c.clients, src.U.ID)
	delete(c.umodes, src.U.ID)
	return nil
}

func (c *Chan) PrivMsg(src *Client, text string) *Error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, exists := c.clients[src.U.ID]; !exists {
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

func (c *Chan) Members() []*Client {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	members := make([]*Client, 0, len(c.clients))
	for _, client := range c.clients {
		members = append(members, client)
	}
	return members
}
