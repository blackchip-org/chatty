package fntest

import (
	"strings"
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/irc/test"
)

func TestMessageTooLong(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()

	c.Login("Joker", "joker 0 * :Jack Nicholson")
	c.Send("PING :" + strings.Repeat("X", irc.MessageMaxLen))
	c.Recv()
	if c.Err() == nil {
		t.Fatalf("expected network drop on message too long")
	}
	if !strings.Contains(c.Err().Error(), "reset by peer") {
		t.Fatalf("unexpected error: %v", c.Err())
	}
}

func TestMessageNotTooLong(t *testing.T) {
	s, c := test.NewServer(t)
	defer s.Quit()

	c.Login("Joker", "joker 0 * :Jack Nicholson")
	c.Send("PING :" + strings.Repeat("X", irc.MessageMaxLen/2))
	c.Recv()
	c.Send("PING :" + strings.Repeat("X", irc.MessageMaxLen/2))
	c.Recv()
	c.Send("PING :" + strings.Repeat("X", irc.MessageMaxLen/2))
	c.Recv()
	if c.Err() != nil {
		t.Fatalf("unexpected error: %v", c.Err())
	}
}
