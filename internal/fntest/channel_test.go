package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/tester"
)

func TestJoinChannel(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne")
	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")

	c.Send("JOIN #gotham")
	have := c.WaitFor(irc.RplNameReply).Encode()
	want := ":irc.localhost 353 Batman = #gotham :@Batman"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v \n err:  %v", want, have, c.Err())
	}

	c2.Send("JOIN #gotham")
	have = c2.WaitFor(irc.RplNameReply).Encode()
	want1 := ":irc.localhost 353 Robin = #gotham :@Batman Robin"
	want2 := ":irc.localhost 353 Robin = #gotham :Robin @Batman"
	if want1 != have && want2 != have {
		t.Fatalf("\n want: %v \n or  : %v \n have: %v", want1, want2, have)
	}
}

func TestJoinNoChannel(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.LoginDefault()
	c.Send("JOIN")
	have := c.Recv()
	want := ":irc.localhost 461 Batman JOIN :Not enough parameters"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestMessage(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne")
	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")

	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.RplEndOfNames)
	c.Send("JOIN #gotham")
	c.WaitFor(irc.RplEndOfNames)

	c2.Send("PRIVMSG #gotham :Holy hamburger Batman!")
	have := c.Recv()
	want := ":Robin!~robin@localhost PRIVMSG #gotham :Holy hamburger Batman!"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestTopicSet(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("TOPIC #gotham :Gotham City News")
	c.WaitFor(irc.TopicCmd)

	c.Send("TOPIC #gotham")
	have := c.Recv()
	want := ":irc.localhost 332 Batman #gotham :Gotham City News"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestTopicClear(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("TOPIC #gotham :Gotham City News")
	c.WaitFor(irc.TopicCmd)
	c.Send("TOPIC #gotham :")
	c.WaitFor(irc.TopicCmd)

	c.Send("TOPIC #gotham")
	have := c.Recv()
	want := ":irc.localhost 331 Batman #gotham :No topic is set."
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestPartChannel(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne")
	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")

	c.Send("JOIN #gotham")
	c.WaitFor(irc.RplEndOfNames).Encode()
	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.RplEndOfNames).Encode()
	c.WaitFor(irc.JoinCmd)

	c.Send("PART #gotham :To the Batcave, Robin!")
	have := c.Recv()
	want := ":Batman!~batman@localhost PART #gotham :To the Batcave, Robin!"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}

	have = c2.Recv()
	want = ":Batman!~batman@localhost PART #gotham :To the Batcave, Robin!"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}

	c2.Send("NAMES #gotham")
	have = c2.Recv()
	want = ":irc.localhost 353 Robin = #gotham :Robin"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestQuit(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne")
	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")

	c.Send("JOIN #gotham")
	c.WaitFor(irc.RplEndOfNames)
	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.RplEndOfNames)

	c.Send("QUIT :To the Batcave, Robin!")
	have := c2.Recv()
	want1 := ":Batman!~batman@localhost QUIT :To the Batcave, Robin!"
	want2 := ":Batman!~batman@localhost QUIT :\"To the Batcave, Robin!\""

	if want1 != have && want2 != have {
		t.Fatalf("\n want: %v \n or  : %v \n have: %v", want1, want2, have)
	}

	c2.Send("NAMES #gotham")
	have = c2.Recv()
	want := ":irc.localhost 353 Robin = #gotham :Robin"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}
