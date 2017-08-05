package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/tester"
)

func TestModeOper(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Drain()
	c.Send("MODE #gotham +o Robin")
	{
		have := c2.Recv()
		want := ":Batman!~batman@localhost MODE #gotham +o Robin"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
	c.Drain()
	c.Send("NAMES #gotham")
	{
		have := AnyOf(c.Recv(), "@Batman", "@Robin")
		want := ":irc.localhost 353 Batman = #gotham :X X"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
}

func TestModeDeOper(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")
	c2.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")

	c.Send("MODE #gotham +o Batman")
	c.WaitFor(irc.ModeCmd)
	c2.Send("MODE #gotham -o Robin")
	c2.Recv() // +o Batman
	{
		have := c2.Recv()
		want := ":Batman!~batman@localhost MODE #gotham -o Robin"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
	c2.Send("NAMES #gotham")
	{
		have := AnyOf(c2.Recv(), "@Batman", "Robin")
		want := ":irc.localhost 353 Batman = #gotham :X X"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}

}
