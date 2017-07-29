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
	c.Send("NICK Batman")
	c.Send("USER Batman 0 * :Bruce Wayne")
	c.WaitFor(irc.RplWelcome)
	if c.Err() != nil {
		t.Fatalf("\n did not get %v", irc.RplWelcome)
	}
}

func TestNickInUse(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.Send("NICK Batman")
	c.Send("USER Batman 0 * :Bruce Wayne")
	c.WaitFor(irc.RplWelcome)

	c2 := s.NewClient()
	c2.Send("NICK Batman")
	c2.Send("USER Batman 0 * :Bruce Wayne")

	want := ":irc.localhost 433 * Batman :Nickname is already in use"
	have := c2.WaitFor(irc.ErrNickNameInUse).Encode()
	if !strings.HasPrefix(have, want) {
		t.Fatalf("\n want: %v \n have: %v \n err:  %v", want, have, c.Err())
	}
}

func TestNoNick(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.Send("NICK")
	c.Send("USER Batman 0 * :Bruce Wayne")

	want := ":irc.localhost 461 * NICK :Not enough parameters"
	have := c.WaitFor(irc.ErrNeedMoreParams).Encode()
	if want != have {
		t.Fatalf("\n want: %v \n have: %v \n err:  %v", want, have, c.Err())
	}
}
