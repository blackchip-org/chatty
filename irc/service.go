package irc

import (
	"strings"
	"sync"
	"time"
)

type Service struct {
	Name    string
	Started time.Time
	mutex   sync.RWMutex
	chans   map[string]*Chan
	nicks   *Nicks
}

func newService(name string) *Service {
	s := &Service{
		Name:    name,
		Started: time.Now(),
		chans:   make(map[string]*Chan),
		nicks:   NewNicks(),
	}
	return s
}

func (s *Service) Origin() string {
	return s.Name
}

func (s *Service) Chan(name string) (*Chan, *Error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ch, exists := s.chans[name]
	if !exists {
		return nil, NewError(ErrNoSuchChannel)
	}
	return ch, nil
}

// ==== Commands

func (s *Service) Join(c *Client, name string) (*Chan, *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.chans[name]
	if !exists {
		ch = NewChan(name, s.nicks)
		s.chans[name] = ch
	}
	err := ch.Join(c)
	if err != nil {
		return ch, err
	}
	c.chans[name] = ch
	return ch, nil
}

func (s *Service) Nick(c *Client, nick string) *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if ok := s.nicks.Register(nick, c.U); !ok {
		return NewError(ErrNickNameInUse, nick)
	}
	return nil
}

func (s *Service) Part(c *Client, name string, reason string) *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.chans[name]
	if !exists {
		return NewError(ErrNoSuchChannel, name)
	}
	err := ch.Part(c, reason)
	if err != nil {
		return err
	}
	delete(c.chans, name)
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

func (s *Service) Quit(src *Client, reason string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	notify := make(map[UserID]*Client)
	for _, ch := range src.chans {
		members := ch.Members()
		for _, m := range members {
			if m.U.ID == src.U.ID {
				continue
			}
			notify[m.U.ID] = m
		}
		ch.Quit(src)
	}
	for _, cli := range notify {
		cli.Relay(src.U, QuitCmd, reason)
	}
	src.Quit()
	s.nicks.Unregister(src.U)
}
