package irc

import (
	"strings"
	"sync"
)

type Service struct {
	name     string
	mutex    sync.RWMutex
	channels map[string]*Channel
	nicks    map[string]*User
}

func newService(name string) *Service {
	s := &Service{
		channels: make(map[string]*Channel),
		nicks:    make(map[string]*User),
		name:     name,
	}
	return s
}

func (s *Service) Prefix() string {
	return s.name
}

func (s *Service) Join(c *Client, name string) (*Channel, *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.channels[name]
	if !exists {
		ch = NewChannel(name)
		s.channels[name] = ch
	}
	err := ch.Join(c)
	return ch, err
}

func (s *Service) Nick(c *Client, nick string) *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exists := s.nicks[nick]; exists {
		return NewError(ErrNickNameInUse, nick)
	}
	delete(s.nicks, c.U.Nick)
	s.nicks[nick] = c.U
	c.U.Nick = nick
	return nil
}

func (s *Service) PrivMsg(src *Client, dest string, text string) *Error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if strings.HasPrefix(dest, "#") {
		ch, ok := s.channels[dest]
		if !ok {
			return NewError(ErrNoSuchNick, dest)
		}
		return ch.PrivMsg(src, text)
	}
	return nil
}
