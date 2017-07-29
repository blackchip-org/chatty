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
		":bob!~bob@localhost PRIVMSG #elsinore :Good day, eh?",
		"bob!~bob@localhost",
		"PRIVMSG",
		[]string{"#elsinore", "Good day, eh?"},
	},
	{
		"test full message without source",
		"PRIVMSG #elsinore :Good day, eh?",
		"",
		"PRIVMSG",
		[]string{"#elsinore", "Good day, eh?"},
	},
	{
		"test missing command",
		":bob!~bob@localhost",
		"bob!~bob@localhost",
		"*",
		[]string{},
	},
}

func TestDecodeMessage(t *testing.T) {
	for _, test := range encodeTests {
		t.Run(test.name, func(t *testing.T) {
			m := DecodeMessage(test.line)
			if test.prefix != m.Prefix {
				t.Errorf("\n want: %v \n have: %v", test.prefix, m.Prefix)
			}
			if test.cmd != m.Cmd {
				t.Errorf("\n want: %v \n have: %v", test.cmd, m.Cmd)
			}
			if !reflect.DeepEqual(test.params, m.Params) {
				t.Errorf("\n want: %v \n have: %v", test.params, m.Params)
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
			Prefix: "bob!~bob@localhost",
			Cmd:    "PRIVMSG",
			Params: []string{"#elsinore", "Good day, eh?"},
		},
		":bob!~bob@localhost PRIVMSG #elsinore :Good day, eh?",
	},
	{
		"test always have colon for last parameter",
		Message{
			Prefix: "bob!~bob@localhost",
			Cmd:    "PRIVMSG",
			Params: []string{"#elsinore", "eh?"},
		},
		":bob!~bob@localhost PRIVMSG #elsinore :eh?",
	},
	{
		"test without source",
		Message{
			Cmd:    "PRIVMSG",
			Params: []string{"#elsinore", "Good day, eh?"},
		},
		"PRIVMSG #elsinore :Good day, eh?",
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
