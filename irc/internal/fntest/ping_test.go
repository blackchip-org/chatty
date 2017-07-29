package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/irc/test"
)

func TestPingNoParams(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.LoginDefault()
	c.Send(irc.PingCmd)
	have := c.Recv()
	want := ":irc.localhost 461 bob PING :Not enough parameters"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestPingWithParams(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()
	c.LoginDefault()
	c.SendMessage(irc.PingCmd, "LAG1501295043420757")
	m := c.RecvMessage()

	want := ":irc.localhost PONG irc.localhost :LAG1501295043420757"
	have := m.Encode()
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}
