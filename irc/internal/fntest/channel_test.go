package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
)

func TestJoinChannel(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()

	c.Login("bob", "bob 0 * :Bob Mackenzie")
	c2 := s.NewClient()
	c2.Login("doug", "doug 0 * :Doug Mackenzie")

	c.Send("JOIN #elsinore")
	got := c.WaitFor(irc.RplNameReply).Encode()
	expected := ":example.com 353 bob = #elsinore bob"
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}

	c2.Send("JOIN #elsinore")
	got = c2.WaitFor(irc.RplNameReply).Encode()
	expected = ":example.com 353 doug = #elsinore :bob doug"
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}

func TestJoinNoChannel(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()

	c.LoginDefault()
	c.Send("JOIN")
	got := c.Recv()
	expected := ":example.com 461 JOIN :Not enough parameters"
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}

func TestMessage(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()

	c.Login("bob", "bob 0 * :Bob Mackenzie")
	c2 := s.NewClient()
	c2.Login("doug", "doug 0 * :Doug Mackenzie")

	c.Send("JOIN #elsinore")
	c.WaitFor(irc.RplEndOfNames)
	c2.Send("JOIN #elsinore")
	c2.WaitFor(irc.RplEndOfNames)

	c.Send("PRIVMSG #elsinore :good day, eh?")
	got := c2.Recv()
	expected := ":bob!bob@localhost PRIVMSG #elsinore :good day, eh?"
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}
