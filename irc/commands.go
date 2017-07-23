package irc

import (
	"fmt"
	"log"

	"github.com/blackchip-org/chatty/irc/icmd"
)

func (h *handler) command(m Message) error {
	switch m.Cmd {
	case icmd.Cap:
		return h.cap(m.Params)
	case icmd.Nick:
		return h.nick(m.Params)
	case icmd.Ping:
		return h.ping(m.Params)
	case icmd.User:
		return h.user0(m.Params)
	case icmd.Quit:
		return quit
	default:
		log.Printf("unhandled message: %+v", m)
	}
	return nil
}

func (h *handler) cap(params []string) error {
	switch params[0] {
	case icmd.CapReq:
		return h.send("CAP", "*", "ACK", "multi-prefix")
	case icmd.CapEnd:
		return h.welcome()
	}
	return nil
}

func (h *handler) nick(params []string) error {
	h.user.Nick = params[0]
	return h.checkHandshake()
}

func (h *handler) ping(params []string) error {
	return h.send("PONG", h.server.Name, params[0])
}

func (h *handler) user0(params []string) error {
	h.user.Name = params[0]
	return h.checkHandshake()
}

// ===============

func (h *handler) checkHandshake() error {
	if h.user.Nick != "" && h.user.Name != "" {
		return h.welcome()
	}
	return nil
}

func (h *handler) welcome() error {
	err := h.send(icmd.Welcome,
		fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.user.Nick))
	if err != nil {
		return err
	}
	err = h.send(icmd.YourHost,
		fmt.Sprintf("Your host is %v running version chatty-0", h.server.Name))
	if err != nil {
		return err
	}
	err = h.send(icmd.MotdStart, "Message of the day!")
	if err != nil {
		return err
	}
	err = h.send(icmd.EndOfMotd)
	if err != nil {
		return err
	}
	return nil
}
