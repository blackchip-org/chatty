package irc

import (
	"strings"

	"github.com/blackchip-org/chatty/irc/msg"
)

type Message struct {
	Prefix string
	Cmd    string
	Target string
	Params []string
}

func NewMessage(cmd string, params ...string) Message {
	return Message{Cmd: cmd, Params: params}
}

func DecodeMessage(line string) Message {
	m := Message{}
	m.Params = make([]string, 0)
	fields := strings.Split(line, " ")
	if strings.HasPrefix(fields[0], ":") {
		m.Prefix = fields[0][1:]
		fields = fields[1:]
	}
	if len(fields) == 0 {
		m.Cmd = "*"
	} else {
		m.Cmd = fields[0]
		fields = fields[1:]
	}
	for len(fields) > 0 {
		if strings.HasPrefix(fields[0], ":") {
			arg := strings.Join(fields, " ")
			arg = arg[1:] // Remove colon
			m.Params = append(m.Params, arg)
			fields = []string{}
		} else {
			m.Params = append(m.Params, fields[0])
			fields = fields[1:]
		}
	}
	return m
}

func (m Message) Encode() string {
	fields := make([]string, 0)
	if m.Prefix != "" {
		fields = append(fields, ":"+m.Prefix)
	}
	cmd := m.Cmd
	if cmd == "" {
		cmd = "*"
	}
	fields = append(fields, cmd)
	if m.Target != "" && m.Cmd[0] >= '0' && m.Cmd[0] <= '9' {
		fields = append(fields, m.Target)
	}
	for i, param := range m.Params {
		if i == len(m.Params)-1 && strings.Contains(param, " ") {
			param = ":" + param
		}
		fields = append(fields, param)
	}
	return strings.Join(fields, " ")
}

func (m Message) String() string {
	return m.Encode()
}

type Command struct {
	Name   string
	Params []string
}

type Error struct {
	Numeric string
	Params  []string
}

func (e Error) Error() string {
	return strings.Join(e.Params, " ")
}

func NewError(numeric string, params ...string) *Error {
	if text, ok := msg.ErrorText[numeric]; ok {
		params = append(params, text)
	}
	return &Error{Numeric: numeric, Params: params}
}