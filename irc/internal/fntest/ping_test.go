package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
)

func TestPing(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()
	c.LoginDefault()
	c.Send(irc.PingCmd)
	m := c.RecvMessage()

	expected := irc.PongCmd
	got := m.Cmd
	if m.Cmd != irc.PongCmd {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}

func TestPingWithParams(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()
	c.LoginDefault()
	c.SendMessage(irc.PingCmd, "foo", "bar")
	m := c.RecvMessage()

	expected := irc.PongCmd
	got := m.Cmd
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}

	expected = "bar"
	got = m.Params[1]
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}
