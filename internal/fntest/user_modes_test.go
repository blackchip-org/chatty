package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/internal/tester"

)

func TestModeInvisible(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE Batman +i")

	have := c.Recv()
	want := ":Batman MODE Batman :+i"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestModeNoChange(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE Batman -i")
	c.Send("PING ping")

	have := c.Recv()
	want := ":irc.localhost PONG irc.localhost :ping"

	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}
