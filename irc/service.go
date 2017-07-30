package irc

import (
	"strings"
	"sync"
)

type Service struct {
	name  string
	mutex sync.RWMutex
	chans map[string]*Chan
	nicks *Nicks
}

func newService(name string) *Service {
	s := &Service{
		chans: make(map[string]*Chan),
		nicks: NewNicks(),
		name:  name,
	}
	return s
}

func (s *Service) Prefix() string {
	return s.name
}

func (s *Service) Join(c *Client, name string) (*Chan, *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.chans[name]
	if !exists {
		ch = NewChan(name)
		s.chans[name] = ch
	}
	err := ch.Join(c)
	return ch, err
}

func (s *Service) Nick(c *Client, nick string) *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if ok := s.nicks.Register(nick, c.U); !ok {
		return NewError(ErrNickNameInUse, nick)
	}
	return nil
}

func (s *Service) PrivMsg(src *Client, dest string, text string) *Error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if strings.HasPrefix(dest, "#") {
		ch, ok := s.chans[dest]
		if !ok {
			return NewError(ErrNoSuchNick, dest)
		}
		return ch.PrivMsg(src, text)
	}
	return nil
}
