package irc

import (
	"context"
	"sync"
	"time"

	"github.com/blackchip-org/chatty/internal/clock"
)

const (
	NickMaxLen       = 40
	NickExpireDelay  = 10 * time.Minute
	NickReapInterval = 1 * time.Minute
)

type nickHistory struct {
	user User
	seen time.Time
}

type Nicks struct {
	active map[string]User
	prev   map[string]nickHistory
	mutex  sync.RWMutex
	clk    clock.C
	cancel context.CancelFunc
}

func NewNicks() *Nicks {
	n := &Nicks{
		active: make(map[string]User),
		prev:   make(map[string]nickHistory),
		clk:    clock.Real{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	n.cancel = cancel

	go func() {
		t := time.NewTimer(NickReapInterval)
		for {
			select {
			case <-t.C:
				n.reap()
			case <-ctx.Done():
				return
			}
		}
	}()
	return n
}

func (n *Nicks) Close() {
	n.cancel()
}

func (n *Nicks) Register(nick string, u *User) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if len(nick) > NickMaxLen {
		nick = nick[:NickMaxLen]
	}
	if !n.canRegister(nick, u) {
		return false
	}
	delete(n.prev, nick)
	n.active[nick] = *u
	u.Nick = nick
	return true
}

func (n *Nicks) Unregister(u *User) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	delete(n.active, u.Nick)
}

func (n *Nicks) Get(name string) (User, bool) {
	u, ok := n.active[name]
	return u, ok
}

func (n *Nicks) canRegister(nick string, u *User) bool {
	// Cannot register if the nick is already active
	_, exists := n.active[nick]
	if exists {
		return false
	}

	// Okay if the nick was not previously used
	was, exists := n.prev[nick]
	if !exists {
		return true
	}

	// Okay if the nick has expired
	if was.seen.Sub(n.clk.Now()) > NickExpireDelay {
		return true
	}

	// Okay if the same user is trying to reclaim
	if was.user.ID == u.ID {
		return true
	}

	return false
}

func (n *Nicks) reap() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	now := n.clk.Now()
	for nick, history := range n.prev {
		if history.seen.Sub(now) > NickExpireDelay {
			delete(n.prev, nick)
		}
	}
}
