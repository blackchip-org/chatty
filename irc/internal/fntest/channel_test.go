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
	expected := ":example.com 353 bob bob"
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}

	c2.Send("JOIN #elsinore")
	got = c2.WaitFor(irc.RplNameReply).Encode()
	expected = ":example.com 353 doug bob doug"
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}
