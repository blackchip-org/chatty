package irc

import (
	"sync"
)

type Channel struct {
	name    string
	topic   string
	members map[string]*User
	mutex   sync.RWMutex
}

func NewChannel(name string) *Channel {
	c := &Channel{
		name:    name,
		topic:   "no topic",
		members: make(map[string]*User),
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

func (c *Channel) Nicks() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	nicks := make([]string, 0, len(c.members))
	for _, u := range c.members {
		nicks = append(nicks, u.Nick)
	}
	return nicks
}

func (c *Channel) Join(u *User) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.members[u.Nick] = u
	names := make([]string, 0, len(c.members))
	for _, member := range c.members {
		member.Relay(u, JoinCmd, c.name)
		names = append(names, member.Nick)
	}
	return nil
}

func (c *Channel) IsMember(nick string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	_, exists := c.members[nick]
	return exists
}
