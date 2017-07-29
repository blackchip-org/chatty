package irc

import (
	"sort"
	"sync"
)

type Channel struct {
	name    string
	topic   string
	status  string
	members map[string]*User
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
		members: make(map[string]*User),
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
	nicks := make([]string, 0, len(c.members))
	for _, u := range c.members {
		prefix := UserChanPrefixes[c.umodes[u.Nick]]
		nicks = append(nicks, prefix+u.Nick)
	}
	sort.Strings(nicks)
	return nicks
}

func (c *Channel) Join(u *User) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.members) == 0 {
		c.umodes[u.Nick] = UserChanOp
	} else {
		c.umodes[u.Nick] = UserChan
	}
	c.members[u.Nick] = u
	names := make([]string, 0, len(c.members))
	for _, member := range c.members {
		member.Relay(u, JoinCmd, c.name)
		names = append(names, member.Nick)
	}
	return nil
}

func (c *Channel) PrivMsg(u *User, text string) *Error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, exists := c.members[u.Nick]; !exists {
		return NewError(ErrCannotSendToChan, c.name)
	}
	for _, member := range c.members {
		if member.Nick == u.Nick {
			continue
		}
		member.Relay(u, PrivMsgCmd, c.name, text)
	}
	return nil
}
