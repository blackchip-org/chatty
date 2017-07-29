package fntest

import (
	"strings"
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/irc/test"
)

func TestRegistration(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.Send("PASS")
	c.Send("NICK bob")
	c.Send("USER bob bob localhost :Bob Mackenzie")
	c.WaitFor(irc.RplWelcome)
	if c.Err() != nil {
		t.Fatalf("\n did not get %v", irc.RplWelcome)
	}
}

func TestNickInUse(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.Send("NICK bob")
	c.Send("USER bob 0 * :Bob Mackenzie")
	c.WaitFor(irc.RplWelcome)

	c2 := s.NewClient()
	c2.Send("NICK bob")
	c2.Send("USER bob 0 * :Bob Mackenzie")

	want := ":irc.localhost 433 * bob :Nickname is already in use"
	have := c2.WaitFor(irc.ErrNickNameInUse).Encode()
	if !strings.HasPrefix(have, want) {
		t.Fatalf("\n want: %v \n have: %v \n err:  %v", want, have, c.Err())
	}
}

func TestNoNick(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.Send("NICK")
	c.Send("USER bob 0 * :Bob Mackenzie")

	want := ":irc.localhost 461 * NICK :Not enough parameters"
	have := c.WaitFor(irc.ErrNeedMoreParams).Encode()
	if want != have {
		t.Fatalf("\n want: %v \n have: %v \n err:  %v", want, have, c.Err())
	}
}
