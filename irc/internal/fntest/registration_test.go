package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
)

func TestRegistration(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()
	c.Send("NICK bob")
	c.Send("USER bob 0 * :Bob Mackenzie")
	expected := ":example.com 001 bob :Welcome to the Internet Relay Chat Network bob"
	got := c.Recv()
	if expected != got {
		t.Fatalf("\n expected - %v \n got - %v", expected, got)
	}
}

func TestNickInUse(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()
	c.Send("NICK bob")
	c.Send("USER bob 0 * :Bob Mackenzie")
	c.WaitFor(irc.RplWelcome)

	c2 := s.NewClient()
	c2.Send("NICK bob")
	c2.Send("USER bob 0 * :Bob Mackenzie")

	expected := ":example.com 433 bob :Nickname is already in use"
	got := c2.Recv()
	if expected != got {
		t.Fatalf("\n expected - %v \n got - %v", expected, got)
	}
}

func TestNoNick(t *testing.T) {
	s, c := irc.NewTestServer()
	defer s.Quit()
	c.Send("NICK")
	c.Send("USER bob 0 * :Bob Mackenzie")

	expected := ":example.com 431 :No nickname given"
	got := c.Recv()
	if expected != got {
		t.Fatalf("\n expected - %v \n got - %v", expected, got)
	}
}
