package irc

import (
	"errors"
	"strings"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
)

type Args []string

type Message struct {
	Source string
	Cmd    string
	Args   Args
}

func NewMessage(cmd string, args ...string) Message {
	return Message{Cmd: cmd, Args: args}
}

func DecodeMessage(line string) (Message, error) {
	m := Message{}
	m.Args = make([]string, 0)
	fields := strings.Split(line, " ")
	if strings.HasPrefix(fields[0], ":") {
		m.Source = fields[0][1:]
		fields = fields[1:]
	}
	if len(fields) == 0 {
		return m, errors.New("no command")
	}
	m.Cmd = fields[0]
	fields = fields[1:]
	for len(fields) > 0 {
		if strings.HasPrefix(fields[0], ":") {
			arg := strings.Join(fields, " ")
			arg = arg[1:] // Remove colon
			m.Args = append(m.Args, arg)
			fields = []string{}
		} else {
			m.Args = append(m.Args, fields[0])
			fields = fields[1:]
		}
	}
	return m, nil
}

func (m Message) validate() error {
	if m.Cmd == "" {
		return ErrInvalidMessage
	}
	// Only the last argument is allowed to have spaces
	if len(m.Args) > 0 {
		for _, arg := range m.Args[:len(m.Args)-1] {
			if strings.Contains(arg, " ") {
				return ErrInvalidMessage
			}
		}
	}
	return nil
}

func (m Message) Encode() (string, error) {
	if err := m.validate(); err != nil {
		return "", err
	}
	fields := make([]string, 0)
	if m.Source != "" {
		fields = append(fields, ":"+m.Source)
	}
	fields = append(fields, m.Cmd)
	if len(m.Args) > 0 {
		fields = append(fields, m.Args.String())
	}
	return strings.Join(fields, " "), nil
}

func (a Args) String() string {
	n := len(a)
	if n == 0 {
		return ""
	}
	last := n - 1
	parts := make([]string, 0)
	for i, arg := range a {
		if i == last && strings.Contains(arg, " ") {
			parts = append(parts, ":"+arg)
		} else {
			parts = append(parts, arg)
		}
	}
	return strings.Join(parts, " ")
}
