package irc

import (
	"fmt"

	"github.com/blackchip-org/chatty/internal/counter"
)

type UserID uint64

type User struct {
	ID       UserID
	Nick     string
	Name     string
	Host     string
	RealHost string
	FullName string
}

func newUser(host string, realHost string) *User {
	u := &User{
		ID:       UserID(counter.Next()),
		Host:     host,
		RealHost: realHost,
	}
	return u
}

func (u User) Origin() string {
	nick := "*"
	name := "*"
	host := "*"

	if u.Nick != "" {
		nick = u.Nick
	}
	if u.Name != "" {
		name = u.Name
	}
	if u.Host != "" {
		host = u.Host
	}
	// FIXME: Not sure why ircd-irc2 returns a tilde in the name
	return fmt.Sprintf("%s!~%s@%s", nick, name, host)
}
