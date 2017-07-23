package irc

import (
	"fmt"
	"log"

	"github.com/blackchip-org/chatty/irc/icmd"
)

type Handler interface {
	HandleCommand(Command) bool
}

type NewHandlerFunc func(*ReplySender) Handler

func NewDefaultHandler(r *ReplySender) Handler {
	return &DefaultHandler{r: r}
}

type DefaultHandler struct {
	r        *ReplySender
	userName string
	userNick string
}

func (h *DefaultHandler) HandleCommand(cmd Command) bool {
	handled := true
	switch cmd.Name {
	case icmd.Cap:
		h.cap(cmd.Params)
	case icmd.Nick:
		h.nick(cmd.Params)
	case icmd.Ping:
		h.ping(cmd.Params)
	case icmd.User:
		h.user(cmd.Params)
	case icmd.Quit:
		h.r.Quit()
	default:
		handled = false
		log.Printf("unhandled message: %+v", cmd)
	}
	return handled
}

func (h *DefaultHandler) cap(params []string) {
	switch params[0] {
	case icmd.CapReq:
		h.r.Send("CAP", "*", "ACK", "multi-prefix")
	case icmd.CapEnd:
		h.welcome()
	}
}

func (h *DefaultHandler) nick(params []string) {
	h.userName = params[0]
	h.r.Target = params[0]
	h.checkHandshake()
}

func (h *DefaultHandler) ping(params []string) {
	h.r.Send("PONG", params[0])
}

func (h *DefaultHandler) user(params []string) {
	h.userNick = params[0]
	h.checkHandshake()
}

// ===============

func (h *DefaultHandler) checkHandshake() error {
	if h.userNick != "" && h.userName != "" {
		h.welcome()
	}
	return nil
}

func (h *DefaultHandler) welcome() {
	h.r.Send(icmd.Welcome, fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.userNick)).
		Send(icmd.YourHost, fmt.Sprintf("Your host is %v running version chatty-0", h.r.ServerName)).
		Send(icmd.MotdStart, "Message of the day!").
		Send(icmd.EndOfMotd)
}
