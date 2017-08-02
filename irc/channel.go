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
	modes   *ChanModes
	mutex   sync.RWMutex
}

const (
	ChanPrefixNetwork = "#"
	ChanPrefixLocal   = "&"
)

func HasChanPrefix(chname string) bool {
	if chname == "" {
		return false
	}
	return chname[0] == '#' || chname[0] == '&'
}

func NewChan(name string) *Chan {
	c := &Chan{
		name:    name,
		clients: make(map[UserID]*Client),
		modes:   NewChanModes(),
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
		prefix := c.modes.UserPrefix(cli.U.ID)
		nicks = append(nicks, prefix+cli.U.Nick)
	}
	sort.Strings(nicks)
	return nicks
}

func (c *Chan) Join(cli *Client) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.clients) == 0 {
		c.modes.Operators[cli.U.ID] = true
	}
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
	delete(c.modes.Operators, src.U.ID)
	delete(c.modes.Voiced, src.U.ID)
	delete(c.clients, src.U.ID)
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
