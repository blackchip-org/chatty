package fntest

import (
	"testing"

	"github.com/blackchip-org/chatty/irc"
	"github.com/blackchip-org/chatty/tester"
)

// ===== Keylock

func TestModeKeylockFail(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Drain()
	c.Send("MODE #gotham +k swordfish")

	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")
	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.ErrBadChannelKey)
	if c2.Err() != nil {
		t.Errorf("unexpected error: %v", c2.Err())
	}
}

func TestModeKeylockSuccess(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Drain()
	c.Send("MODE #gotham +k swordfish")

	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")
	c2.Send("JOIN #gotham swordfish")
	c2.WaitFor(irc.RplEndOfNames)
	if c2.Err() != nil {
		t.Errorf("unexpected error: %v", c2.Err())
	}
}

func TestModeKeylockNoAction(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham k")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeKeylockNotOper(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Joker", "joker 0 * :The Joker").Join("#gotham")

	c2.Send("MODE #gotham +k swordfish")
	c2.WaitFor(irc.ErrChanOpPrivsNeeded)
	if c2.Err() != nil {
		t.Errorf("unexpected error: %v", c2.Err())
	}
}

func TestModeKeylockNoModeChange(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham -k")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

// ===== Limit

func TestModeLimitFail(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham +l 1")

	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")
	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.ErrChannelIsFull)
	if c2.Err() != nil {
		t.Errorf("unexpected error: %v", c2.Err())
	}
}

func TestModeLimitSuccess(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham +l 1")

	c2 := s.NewClient()
	c2.Login("Robin", "robin 0 * :Boy Wonder")
	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.ErrChannelIsFull)

	c.Drain()
	c.Send("MODE #gotham -l")

	c2.Drain()
	c2.Send("JOIN #gotham")
	c2.WaitFor(irc.RplEndOfNames)
	if c2.Err() != nil {
		t.Errorf("unexpected error: %v", c2.Err())
	}
}

func TestModeLimitNegative(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham +l -123")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeLimitNoAction(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham l")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeLimitNotOper(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Joker", "joker 0 * :The Joker").Join("#gotham")

	c2.Send("MODE #gotham +l 10")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeLimitNoModeChange(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham -l")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

// ==== Moderated
func TestModeModerated(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Send("MODE #gotham +m")
	c.WaitFor(irc.ModeCmd)

	c2.Drain()
	c2.Send("PRIVMSG #gotham :Can you hear me now?")
	c2.WaitFor(irc.ErrCannotSendToChan)
	if c2.Err() != nil {
		t.Fatalf("unexpected error: %v", c2.Err())
	}
}

func TestModeModeratedWithVoice(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Send("MODE #gotham +m")
	c.WaitFor(irc.ModeCmd)
	c.Send("MODE #gotham +v Robin")
	c.WaitFor(irc.ModeCmd)

	c2.Send("PRIVMSG #gotham :Can you hear me now?")
	c.WaitFor(irc.PrivMsgCmd)
	if c.Err() != nil {
		t.Fatalf("unexpected error: %v", c.Err())
	}
}

// ==== Oper

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

func TestModeRemoveOper(t *testing.T) {
	s, c := tester.NewServer(t)
	c2 := s.NewClient()
	defer s.Quit()

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

func TestModeOperNoAction(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham o Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeOperNoUser(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham +o Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeOperNoUserInChannel(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#batcave")

	c.Send("MODE #gotham +o Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeOperNotOper(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()
	c3 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")
	c3.Login("Joker", "joker 0 * :The Joker").Join("#gotham")

	c3.Drain()
	c3.Send("MODE #gotham +o Robin")
	c3.WaitFor(irc.ErrChanOpPrivsNeeded)

	c3.Send("NAMES #gotham")
	have := AnyOf(c3.Recv(), "@Batman", "Robin", "Joker")
	want := ":irc.localhost 353 X = #gotham :X X X"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestModeOperNoModeChange(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Send("MODE #gotham -o Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

// ==== Voice

func TestModeVoice(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Drain()
	c.Send("MODE #gotham +v Robin")
	{
		have := c2.Recv()
		want := ":Batman!~batman@localhost MODE #gotham +v Robin"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
	c.Drain()
	c.Send("NAMES #gotham")
	{
		have := AnyOf(c.Recv(), "@Batman", "+Robin")
		want := ":irc.localhost 353 Batman = #gotham :X X"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
}

func TestModeRemoveVoice(t *testing.T) {
	s, c := tester.NewServer(t)
	c2 := s.NewClient()
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Send("MODE #gotham +v Robin")
	c.WaitFor(irc.ModeCmd)
	c.Send("MODE #gotham -v Robin")
	{
		have := c.Recv()
		want := ":Batman!~batman@localhost MODE #gotham -v Robin"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
	c.Drain()
	c.Send("NAMES #gotham")
	{
		have := AnyOf(c.Recv(), "@Batman", "Robin")
		want := ":irc.localhost 353 Batman = #gotham :X X"
		if want != have {
			t.Fatalf("\n want: %v \n have: %v", want, have)
		}
	}
}

func TestModeVoiceNoAction(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham v Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeVoiceNoUser(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c.Send("MODE #gotham +v Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeVoiceNoUserInChannel(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#batcave")

	c.Send("MODE #gotham +v Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}

func TestModeVoiceNotOper(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()
	c3 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")
	c3.Login("Joker", "joker 0 * :The Joker").Join("#gotham")

	c3.Drain()
	c3.Send("MODE #gotham +v Robin")
	c3.WaitFor(irc.ErrChanOpPrivsNeeded)

	c3.Send("NAMES #gotham")
	have := AnyOf(c3.Recv(), "@Batman", "Robin", "Joker")
	want := ":irc.localhost 353 X = #gotham :X X X"
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}

func TestModeVoiceNoModeChange(t *testing.T) {
	s, c := tester.NewServer(t)
	defer s.Quit()
	c2 := s.NewClient()

	c.Login("Batman", "batman 0 * :Bruce Wayne").Join("#gotham")
	c2.Login("Robin", "robin 0 * :Boy Wonder").Join("#gotham")

	c.Send("MODE #gotham -v Robin")
	c.Send("PING hello")
	c.WaitFor(irc.PongCmd)
	if c.Err() != nil {
		t.Errorf("\n want: no error \n have: %v", c.Err())
	}
}
