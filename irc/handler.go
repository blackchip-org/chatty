package irc

import (
	"fmt"
	"log"

	"github.com/blackchip-org/chatty/irc/msg"
)

type Handler interface {
	Handle(Command) (bool, error)
}

type NewHandlerFunc func(*Server, *User) Handler

func NewDefaultHandler(s *Server, u *User) Handler {
	return &DefaultHandler{s: s, u: u}
}

type DefaultHandler struct {
	s *Server
	u *User
}

func (h *DefaultHandler) Handle(cmd Command) (bool, error) {
	handled := true
	switch cmd.Name {
	case msg.Cap:
		h.cap(cmd.Params)
	case msg.Join:
		h.join(cmd.Params)
	case msg.Nick:
		h.nick(cmd.Params)
	case msg.Pass:
		// ignore
	case msg.Ping:
		h.ping(cmd.Params)
	case msg.User:
		h.user(cmd.Params)
	case msg.Quit:
		h.u.Quit()
	default:
		handled = false
		log.Printf("unhandled message: %+v", cmd)
	}
	return handled, h.u.err
}

func (h *DefaultHandler) cap(params []string) {
	switch params[0] {
	case msg.CapReq:
		h.u.Reply("CAP", "*", "ACK", "multi-prefix")
	case msg.CapEnd:
		h.welcome()
	}
}

func (h *DefaultHandler) join(params []string) {
	if _, err := h.s.JoinChannel(h.u, params[0]); err != nil {
		h.u.SendError(err)
	}
}

func (h *DefaultHandler) nick(params []string) {
	if len(params) != 1 {
		h.u.SendError(NewError(msg.ErrNoNickNameGiven))
		return
	}
	h.u.Nick = params[0]
	h.checkHandshake()
}

func (h *DefaultHandler) ping(params []string) {
	h.u.Send("PONG", params[0])
}

func (h *DefaultHandler) user(params []string) {
	h.u.Name = params[0]
	h.u.FullName = params[3]
	h.checkHandshake()
}

// ===============

func (h *DefaultHandler) checkHandshake() error {
	if h.u.Nick != "" && h.u.Name != "" {
		h.welcome()
	}
	return nil
}

func (h *DefaultHandler) welcome() {
	h.u.Reply(msg.Welcome, fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.u.Nick)).
		Reply(msg.YourHost, fmt.Sprintf("Your host is %v running version chatty-0", h.s.Name)).
		Reply(msg.MotdStart, "Message of the day!").
		Reply(msg.EndOfMotd)
}
