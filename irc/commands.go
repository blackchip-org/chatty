package irc

import "github.com/blackchip-org/chatty/irc/icmd"

func (h *handler) command(m Message) error {
	switch m.Cmd {
	case icmd.Cap:
		return h.cap(m.Args)
	case icmd.Ping:
		return h.ping()
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
		return h.send(icmd.Welcome, "Welcome to the IRC server")
	}
	return nil
}

func (h *handler) ping() error {
	return h.send("PONG")
}
