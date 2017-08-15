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
	modes   map[UserID]*UserModes
}

func newService(name string) *Service {
	s := &Service{
		Name:    name,
		Started: time.Now(),
		chans:   make(map[string]*Chan),
		nicks:   NewNicks(),
		modes:   make(map[UserID]*UserModes),
	}
	return s
}

func (s *Service) Origin() string {
	return s.Name
}

func (s *Service) Chan(name string) (*Chan, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	ch, exists := s.chans[name]
	if !exists {
		return nil, NewError(ErrNoSuchChannel)
	}
	return ch, nil
}

func (s *Service) Login(c *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.modes[c.User.ID] = &UserModes{}
}

// ==== Commands

func (s *Service) Join(c *Client, name string, key string) (*Chan, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.chans[name]
	if !exists {
		ch = NewChan(name, s.nicks)
		s.chans[name] = ch
	}
	err := ch.Join(c, key)
	if err != nil {
		return ch, err
	}
	c.chans[name] = ch
	return ch, nil
}

func (s *Service) Mode(src *Client) *UserModeCmds {
	return newUserModeCmds(s, src)
}

func (s *Service) Nick(c *Client, nick string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if ok := s.nicks.Register(nick, c.User); !ok {
		return NewError(ErrNickNameInUse, nick)
	}
	return nil
}

func (s *Service) Part(c *Client, name string, reason string) error {
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

func (s *Service) PrivMsg(src *Client, dest string, text string) error {
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	notify := make(map[UserID]*Client)
	for _, ch := range src.chans {
		members := ch.Members()
		for _, m := range members {
			if m.User.ID == src.User.ID {
				continue
			}
			notify[m.User.ID] = m
		}
		ch.Quit(src)
	}
	for _, cli := range notify {
		cli.Relay(src.User, QuitCmd, reason)
	}
	src.Quit()
	s.nicks.Unregister(src.User)
	delete(s.modes, src.User.ID)
}

// ===== User Modes

type UserModeCmds struct {
	s       *Service
	src     *Client
	changes []Mode
}

func newUserModeCmds(s *Service, src *Client) *UserModeCmds {
	cmd := &UserModeCmds{
		s:       s,
		src:     src,
		changes: make([]Mode, 0),
	}
	s.mutex.Lock()
	return cmd
}

func (cmd *UserModeCmds) Invisible(action string) error {
	s := cmd.s

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is a mode change needed?
	modes := s.modes[cmd.src.User.ID]
	prev := s.modes[cmd.src.User.ID].Invisible
	if set == prev {
		return nil
	}
	modes.Invisible = set
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   UserModeInvisible,
	})
	return nil
}

func (cmd UserModeCmds) Done() {
	if len(cmd.changes) > 0 {
		params := append([]string{cmd.src.User.Nick}, formatModes(cmd.changes)...)
		m := Message{
			Prefix: cmd.src.User.Nick,
			Cmd:    ModeCmd,
			Params: params,
		}
		cmd.src.SendMessage(m)
	}
	cmd.s.mutex.Unlock()
}
