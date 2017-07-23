package irc

import (
	"errors"
	"fmt"
)

type User struct {
	Nick       string
	Name       string
	Host       string
	FullName   string
	ServerName string
	Registered bool

	err   error
	sendq chan Message
}

func (u *User) Send(cmd string, params ...string) *User {
	m := Message{Prefix: u.ServerName, Cmd: cmd, Params: params}
	u.send(m)
	return u
}

func (u *User) Reply(cmd string, params ...string) *User {
	m := Message{Prefix: u.ServerName, Target: u.Nick, Cmd: cmd, Params: params}
	u.send(m)
	return u
}

func (u *User) Relay(source Source, cmd string, params ...string) *User {
	m := Message{Prefix: source.Prefix(), Cmd: cmd, Params: params}
	u.send(m)
	return u
}

func (u *User) SendError(err *Error) *User {
	m := Message{Prefix: u.ServerName, Cmd: err.Numeric, Params: err.Params}
	u.send(m)
	return u
}

func (u *User) send(m Message) {
	if u.err != nil {
		return
	}
	select {
	case u.sendq <- m:
		return
	default:
		u.err = errors.New("send queue full")
	}
}

func (u *User) Quit() {
	u.err = quit
}

func (u *User) Prefix() string {
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
	return fmt.Sprintf("%s!%s@%s", nick, name, host)
}
