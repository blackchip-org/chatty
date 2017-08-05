package fntest

import (
	"testing"

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
	have := c2.Recv()
	want := ":Batman!~batman@localhost MODE #gotham +o Robin"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}
