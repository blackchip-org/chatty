package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/irc/test"
)

func TestJoinChannel(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()

	c.Login("bob", "bob 0 * :Bob Mackenzie")
	c2 := s.NewClient()
	c2.Login("doug", "doug 0 * :Doug Mackenzie")

	c.Send("JOIN #elsinore")
	have := c.WaitFor(irc.RplNameReply).Encode()
	want := ":irc.localhost 353 bob = #elsinore :@bob"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v \n err:  %v", want, have, c.Err())
	}

	c2.Send("JOIN #elsinore")
	have = c2.WaitFor(irc.RplNameReply).Encode()
	want1 := ":irc.localhost 353 doug = #elsinore :@bob doug"
	want2 := ":irc.localhost 353 doug = #elsinore :doug @bob"
	if want1 != have && want2 != have {
		t.Fatalf("\n want: %v \n or  : %v \n have: %v", want1, want2, have)
	}
}

func TestJoinNoChannel(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()

	c.LoginDefault()
	c.Send("JOIN")
	have := c.Recv()
	want := ":irc.localhost 461 bob JOIN :Not enough parameters"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestMessage(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()

	c.Login("bob", "bob 0 * :Bob Mackenzie")
	c2 := s.NewClient()
	c2.Login("doug", "doug 0 * :Doug Mackenzie")

	c.Send("JOIN #elsinore")
	c.WaitFor(irc.RplEndOfNames)
	c2.Send("JOIN #elsinore")
	c2.WaitFor(irc.RplEndOfNames)

	c.Send("PRIVMSG #elsinore :good day, eh?")
	have := c2.Recv()
	want := ":bob!~bob@localhost PRIVMSG #elsinore :good day, eh?"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}
