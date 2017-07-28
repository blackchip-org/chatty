package irc

import "sync"

type Service struct {
	name     string
	mutex    sync.RWMutex
	channels map[string]*Channel
	nicks    map[string]*User
}

func newService() *Service {
	s := &Service{
		channels: make(map[string]*Channel),
		nicks:    make(map[string]*User),
	}
	return s
}

func (s *Service) Prefix() string {
	return s.name
}

func (s *Service) Join(u *User, name string) (*Channel, *Error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ch, exists := s.channels[name]
	if !exists {
		ch = NewChannel(name)
		s.channels[name] = ch
	}
	err := ch.Join(u)
	return ch, err
}

func (s *Service) Nick(u *User, nick string) *Error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, exists := s.nicks[nick]; exists {
		return NewError(ErrNickNameInUse, nick)
	}
	delete(s.nicks, u.Nick)
	s.nicks[nick] = u
	u.Nick = nick
	return nil
}

/*
func (s *Service) PrivMsg(src *User, dest string, text string) *Error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if strings.HasPrefix(dest, "#") {
		ch, ok := s.channels[dest]
		if !ok {
			return NewError(ErrNoSuchNick, dest)
		}
		if !ch.IsMember(src.Nick) {
			return NewError(ErrCannotSendToChan, dest)
		}
	}
}
*/
