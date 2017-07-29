package irc

import (
	"fmt"
	"log"
	"strings"
)

type Handler interface {
	Handle(Command) (bool, error)
}

type NewHandlerFunc func(*Service, *User) Handler

func NewDefaultHandler(s *Service, u *User) Handler {
	return &DefaultHandler{
		s: s,
		u: u,
	}
}

type DefaultHandler struct {
	s *Service
	u *User
}

func (h *DefaultHandler) Handle(cmd Command) (bool, error) {
	handled := true
	switch cmd.Name {
	case CapCmd:
		h.cap(cmd.Params)
	case JoinCmd:
		h.join(cmd.Params)
	case NickCmd:
		h.nick(cmd.Params)
	case PassCmd:
		// ignore
	case PingCmd:
		h.ping(cmd.Params)
	case PrivMsgCmd:
		h.privMsg(cmd.Params)
	case UserCmd:
		h.user(cmd.Params)
	case QuitCmd:
		h.u.Quit()
	default:
		handled = false
		log.Printf("unhandled message: %+v", cmd)
	}
	return handled, h.u.err
}

func (h *DefaultHandler) cap(params []string) {
	switch params[0] {
	case CapReqCmd:
		h.u.Reply("CAP", "*", "ACK", "multi-prefix")
	case CapEndCmd:
		h.welcome()
	}
}

func (h *DefaultHandler) join(params []string) {
	if len(params) == 0 {
		h.u.SendError(NewError(ErrNeedMoreParams, JoinCmd))
		return
	}
	name := params[0]
	ch, err := h.s.Join(h.u, name)
	if err != nil {
		h.u.SendError(err)
		return
	}
	h.u.Reply(RplTopic, ch.Topic())
	nicks := strings.Join(ch.Nicks(), " ")
	h.u.Reply(RplNameReply, ch.Status(), ch.Name(), nicks)
	h.u.Reply(RplEndOfNames, ch.Name())
}

func (h *DefaultHandler) nick(params []string) {
	if len(params) != 1 {
		h.u.SendError(NewError(ErrNeedMoreParams, NickCmd))
		return
	}
	nick := params[0]
	if err := h.s.Nick(h.u, nick); err != nil {
		h.u.SendError(err)
		return
	}
	h.checkHandshake()
}

func (h *DefaultHandler) ping(params []string) {
	if len(params) == 0 {
		h.u.SendError(NewError(ErrNeedMoreParams, PingCmd))
		return
	}
	outparams := append([]string{h.s.Prefix()}, params...)
	fmt.Printf("*** PREFIX: %v\n", h.s.Prefix())
	fmt.Printf("*** PARAMS: %v\n", outparams)
	h.u.Send(PongCmd, outparams...)
}

func (h *DefaultHandler) privMsg(params []string) {
	target := params[0]
	text := params[1]
	h.s.PrivMsg(h.u, target, text)
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
	h.u.Reply(RplWelcome, fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.u.Nick)).
		Reply(RplYourHost, fmt.Sprintf("Your host is %v running version chatty-0", h.s.Prefix())).
		Reply(RplMotdStart, "Message of the day!").
		Reply(RplEndOfMotd)
}
