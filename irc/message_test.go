package irc

import "testing"
import "reflect"

var encodeTests = []struct {
	name   string
	line   string
	prefix string
	cmd    string
	params []string
}{
	{
		"test full message",
		":Macha!~macha@unaffiliated/macha PRIVMSG #botwar :Test response",
		"Macha!~macha@unaffiliated/macha",
		"PRIVMSG",
		[]string{"#botwar", "Test response"},
	},
	{
		"test full message without source",
		"PRIVMSG #botwar :Test response",
		"",
		"PRIVMSG",
		[]string{"#botwar", "Test response"},
	},
	{
		"test missing command",
		":Macha!~macha@unaffiliated/macha",
		"Macha!~macha@unaffiliated/macha",
		"*",
		[]string{},
	},
}

func TestDecodeMessage(t *testing.T) {
	for _, test := range encodeTests {
		t.Run(test.name, func(t *testing.T) {
			m := DecodeMessage(test.line)
			if test.prefix != m.Prefix {
				t.Errorf("expected prefix '%v', got '%v'", test.prefix, m.Prefix)
			}
			if test.cmd != m.Cmd {
				t.Errorf("expected cmd '%v', got '%v'", test.cmd, m.Cmd)
			}
			if !reflect.DeepEqual(test.params, m.Params) {
				t.Errorf("expected args %v, got %v", test.params, m.Params)
			}
		})
	}
}

var decodeTests = []struct {
	name string
	msg  Message
	line string
}{
	{
		"test full message",
		Message{
			Prefix: "Macha!~macha@unaffiliated/macha",
			Cmd:    "PRIVMSG",
			Params: []string{"#botwar", "Test response"},
		},
		":Macha!~macha@unaffiliated/macha PRIVMSG #botwar :Test response",
	},
	{
		"test without source",
		Message{
			Cmd:    "PRIVMSG",
			Params: []string{"#botwar", "Test response"},
		},
		"PRIVMSG #botwar :Test response",
	},
	{
		"test empty message",
		Message{},
		"*",
	},
}

func TestEncodeMessage(t *testing.T) {
	for _, test := range decodeTests {
		t.Run(test.name, func(t *testing.T) {
			line := test.msg.Encode()
			if test.line != line {
				t.Errorf("expecting line '%v', got '%v'", test.line, line)
			}
		})
	}
}
