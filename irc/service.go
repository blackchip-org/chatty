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
