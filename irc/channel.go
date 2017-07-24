package irc

import (
	"sync"
)

type Channel struct {
	Name    string
	Topic   string
	Members map[string]*User
	mutex   sync.RWMutex
}

func NewChannel(name string) *Channel {
	c := &Channel{
		Name:    name,
		Topic:   "no topic",
		Members: make(map[string]*User),
	}
	return c
}

func (c *Channel) Join(u *User) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Members[u.Nick] = u
	names := make([]string, 0, len(c.Members))
	for _, member := range c.Members {
		member.Relay(u, JoinCmd, c.Name)
		names = append(names, member.Nick)
	}
	u.Reply(RplTopic, c.Topic)
	u.Reply(RplNameReply, names...)
	u.Reply(RplEndOfNames)
}
