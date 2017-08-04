package fntest

import (
	"strings"
	"testing"
	"time"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/tester"
)

func TestMessageTooLong(t *testing.T) {
	if tester.RealServer {
		t.Skip("skipping test on real server")
	}
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Joker", "joker 0 * :Jack Nicholson")
	c.Send("PING :" + strings.Repeat("X", irc.MessageMaxLen))
	c.Recv()
	if c.Err() == nil {
		t.Fatalf("expected network drop on message too long")
	}
}

func TestMessageNotTooLong(t *testing.T) {
	if tester.RealServer {
		t.Skip("skipping test on real server")
	}
	s, c := tester.NewServer(t)
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

func TestRegistrationDeadline(t *testing.T) {
	if tester.RealServer {
		t.Skip("skipping test on real server")
	}
	s, c := tester.NewServer(t)
	defer s.Quit()
	c.LoginDefault()
	c.Send("QUIT")

	s.Actual.RegistrationDeadline = 100 * time.Millisecond
	c2 := s.NewClient()
	time.Sleep(150 * time.Millisecond)
	c2.Login("Joker", "joker 0 * :Jack Nicholson")
	if c2.Err() == nil {
		t.Fatal("expected timeout")
	}

	s.Actual.RegistrationDeadline = 100 * time.Millisecond
	c3 := s.NewClient()
	time.Sleep(50 * time.Millisecond)
	c3.Login("Robin", "robin 0 * :Boy Wonder")
	if c3.Err() != nil {
		t.Fatalf("unexpected error: %v", c3.Err())
	}
}
