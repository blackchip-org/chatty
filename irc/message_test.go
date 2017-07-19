package irc

import "testing"
import "reflect"

var encodeTests = []struct {
	name   string
	line   string
	source string
	cmd    string
	args   Args
	err    bool
}{
	{
		"test full message",
		":Macha!~macha@unaffiliated/macha PRIVMSG #botwar :Test response",
		"Macha!~macha@unaffiliated/macha",
		"PRIVMSG",
		Args{"#botwar", "Test response"},
		false,
	},
	{
		"test full message without source",
		"PRIVMSG #botwar :Test response",
		"",
		"PRIVMSG",
		Args{"#botwar", "Test response"},
		false,
	},
	{
		"test missing command",
		":Macha!~macha@unaffiliated/macha",
		"Macha!~macha@unaffiliated/macha",
		"",
		Args{},
		true,
	},
}

func TestDecodeMessage(t *testing.T) {
	for _, test := range encodeTests {
		t.Run(test.name, func(t *testing.T) {
			m, err := DecodeMessage(test.line)
			if test.source != m.Source {
				t.Errorf("expected source '%v', got '%v'", test.source, m.Source)
			}
			if test.cmd != m.Cmd {
				t.Errorf("expected cmd '%v', got '%v'", test.cmd, m.Cmd)
			}
			if !reflect.DeepEqual(test.args, m.Args) {
				t.Errorf("expected args %v, got %v", test.args, m.Args)
			}
			hasError := err != nil
			if test.err != hasError {
				t.Errorf("expected err %v, got %v (%v)", test.err, hasError, err)
			}
		})
	}
}

var decodeTests = []struct {
	name string
	msg  Message
	line string
	err  bool
}{
	{
		"test full message",
		Message{
			Source: "Macha!~macha@unaffiliated/macha",
			Cmd:    "PRIVMSG",
			Args:   Args{"#botwar", "Test response"},
		},
		":Macha!~macha@unaffiliated/macha PRIVMSG #botwar :Test response",
		false,
	},
	{
		"test without source",
		Message{
			Cmd:  "PRIVMSG",
			Args: Args{"#botwar", "Test response"},
		},
		"PRIVMSG #botwar :Test response",
		false,
	},
	{
		"test empty message",
		Message{},
		"",
		true,
	},
	{
		"test invalid args",
		Message{
			Cmd:  "PRIVMSG",
			Args: Args{"one two", "three four"},
		},
		"",
		true,
	},
}

func TestEncodeMessage(t *testing.T) {
	for _, test := range decodeTests {
		t.Run(test.name, func(t *testing.T) {
			line, err := test.msg.Encode()
			if test.line != line {
				t.Errorf("expecting line '%v', got '%v'", test.line, line)
			}
			hasError := err != nil
			if test.err != hasError {
				t.Errorf("expected err %v, got %v (%v)", test.err, hasError, err)
			}
		})
	}
}
