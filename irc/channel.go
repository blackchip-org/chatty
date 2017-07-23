package irc

import (
	"sync"

	"github.com/blackchip-org/chatty/irc/msg"
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
		member.Relay(u, msg.Join, c.Name)
		names = append(names, member.Nick)
	}
	u.Reply(msg.Topic, c.Topic)
	u.Reply(msg.NameReply, names...)
	u.Reply(msg.EndOfNames)
}
