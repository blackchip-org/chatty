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

// ==== Commands

func (s *Service) Join(c *Client, name string) (*Chan, *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.chans[name]
	if !exists {
		ch = NewChan(name)
		s.chans[name] = ch
	}
	err := ch.Join(c)
	if err != nil {
		return ch, err
	}
	c.chans[name] = ch
	return ch, nil
}

func (s *Server) ModeUserSet(c *Client, mode string) (bool, *Error) {
	switch mode {
	case UserModeAway:
		// The flag 'a' SHALL NOT be toggled by the user using the MODE
		// command, instead use of the AWAY command is REQUIRED.
		return false, nil
	case UserModeInvisible:
		changed := c.modes.Invisible != true
		c.modes.Invisible = true
		return changed, nil
	case UserModeOp, UserModeLocalOp:
		// If a user attempts to make themselves an operator using the "+o" or
		// "+O" flag, the attempt SHOULD be ignored as users could bypass the
		// authentication mechanisms of the OPER command.
		return false, nil
	}
	return false, NewError(ErrUModeUnknownFlag, mode)
}

func (s *Service) ModeUserClear(c *Client, mode string) (bool, *Error) {
	switch mode {
	case UserModeAway:
		// The flag 'a' SHALL NOT be toggled by the user using the MODE
		// command, instead use of the AWAY command is REQUIRED.
		return false, nil
	case UserModeInvisible:
		changed := c.modes.Invisible != false
		c.modes.Invisible = false
		return changed, nil
	case UserModeOp, UserModeLocalOp:
		// There is no restriction, however, on anyone `deopping' themselves
		// (using "-o" or "-O").
		changed := c.modes.Op != false
		c.modes.Op = false
		c.modes.LocalOp = false
		return changed, nil
	}
	return false, NewError(ErrUModeUnknownFlag, mode)
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
	}
	for _, cli := range notify {
		cli.Relay(src.U, QuitCmd, reason)
	}
	src.Quit()
}
