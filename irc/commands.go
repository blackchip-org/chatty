package irc

import (
	"fmt"

	"github.com/blackchip-org/chatty/irc/icmd"
)

func (h *handler) command(m Message) error {
	switch m.Cmd {
	case icmd.Cap:
		return h.cap(m.Args)
	case icmd.Nick:
		return h.nick(m.Args)
	case icmd.Ping:
		return h.ping(m.Args)
	case icmd.User:
		return h.user(m.Args)
	case icmd.Quit:
		return quit
	default:
		h.log.Printf("unhandled message: %+v", m)
	}
	return nil
}

func (h *handler) cap(args Args) error {
	switch args[0] {
	case icmd.CapReq:
		return h.send("CAP", "*", "ACK", "multi-prefix")
	case icmd.CapEnd:
		return h.welcome()
	}
	return nil
}

func (h *handler) nick(args Args) error {
	h.client.Nick = args[0]
	return h.checkHandshake()
}

func (h *handler) ping(args Args) error {
	return h.send("PONG", h.server.Name, args[0])
}

func (h *handler) user(args Args) error {
	h.client.User = args.String()
	return h.checkHandshake()
}

// ===============

func (h *handler) checkHandshake() error {
	if h.client.Nick != "" && h.client.User != "" {
		return h.welcome()
	}
	return nil
}

func (h *handler) welcome() error {
	err := h.send(icmd.Welcome,
		fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.client.Nick))
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
