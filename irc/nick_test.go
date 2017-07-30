package irc

import (
	"testing"

	"github.com/blackchip-org/chatty/irc/internal/clock"
)

func TestRegisterNick(t *testing.T) {
	nicks := NewNicks()
	defer nicks.Close()
	if ok := nicks.Register("Batman", &User{}); !ok {
		t.Errorf("wanted to register nick")
	}
}

func TestRegisterNickDup(t *testing.T) {
	nicks := NewNicks()
	defer nicks.Close()
	if ok := nicks.Register("Batman", &User{}); !ok {
		t.Errorf("wanted to register nick")
	}
	if ok := nicks.Register("Batman", &User{}); ok {
		t.Errorf("wanted nick collision")
	}
}

func TestReclaimNick(t *testing.T) {
	user := &User{ID: 42}
	nicks := NewNicks()
	defer nicks.Close()
	nicks.Register("Batman", user)
	nicks.Unregister(user)
	if ok := nicks.Register("Batman", user); !ok {
		t.Errorf("wanted to reclaim nick")
	}
}

func TestNickCooldown(t *testing.T) {
	u1 := &User{ID: 1}
	u2 := &User{ID: 2}
	nicks := NewNicks()
	defer nicks.Close()
	nicks.Register("Batman", &User{ID: 1})
	nicks.Unregister(u1)
	if ok := nicks.Register("Batman", u2); ok {
		t.Errorf("wnated to reject nick on cooldown")
	}
}

func TestNickCooldownExpire(t *testing.T) {
	u1 := &User{ID: 1}
	u2 := &User{ID: 2}
	nicks := NewNicks()
	mockClock := &clock.Mock{}
	nicks.clk = mockClock

	defer nicks.Close()
	nicks.Register("Batman", &User{ID: 1})
	nicks.Unregister(u1)
	mockClock.Add(NickExpireDelay)
	if ok := nicks.Register("Batman", u2); ok {
		t.Errorf("wnated to reuse nick")
	}
}
