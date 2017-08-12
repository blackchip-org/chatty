package irc

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Handler interface {
	Handle(Command) (bool, error)
}

type NewHandlerFunc func(*Service, *Client) Handler

func NewDefaultHandler(s *Service, c *Client) Handler {
	return &DefaultHandler{
		s: s,
		c: c,
	}
}

type DefaultHandler struct {
	s *Service
	c *Client
}

var prereg = map[string]bool{
	PassCmd: true,
	NickCmd: true,
	UserCmd: true,
	CapCmd:  true,
}

func (h *DefaultHandler) Handle(cmd Command) (bool, error) {
	handled := true

	if !h.c.registered {
		allowed := prereg[cmd.Name]
		if !allowed {
			h.c.SendError(NewError(ErrNotRegistered))
			return true, nil
		}
	}

	switch cmd.Name {
	case CapCmd:
		h.cap(cmd.Params)
	case JoinCmd:
		h.join(cmd.Params)
	case ModeCmd:
		h.mode(cmd.Params)
	case NamesCmd:
		h.names(cmd.Params)
	case NickCmd:
		h.nick(cmd.Params)
	case PartCmd:
		h.part(cmd.Params)
	case PassCmd:
		h.pass(cmd.Params)
	case PingCmd:
		h.ping(cmd.Params)
	case PrivMsgCmd:
		h.privMsg(cmd.Params)
	case UserCmd:
		h.user(cmd.Params)
	case QuitCmd:
		h.quit(cmd.Params)
	default:
		handled = false
		log.Printf("unhandled message: %+v", cmd)
	}
	return handled, h.c.err
}

func (h *DefaultHandler) cap(params []string) {
	if len(params) == 0 {
		h.c.SendError(NewError(ErrNeedMoreParams))
		return
	}
	capcmd := params[0]
	switch capcmd {
	case CapLsCmd:
		h.c.Reply(CapCmd, CapLsCmd)
	case CapReqCmd:
		h.c.Reply("CAP", "*", "ACK", "multi-prefix")
	case CapEndCmd:
		h.welcome()
	default:
		h.c.SendError(NewError(ErrInvalidCapCmd, capcmd))
	}
}

func (h *DefaultHandler) join(params []string) {
	if len(params) == 0 {
		h.c.SendError(NewError(ErrNeedMoreParams, JoinCmd))
		return
	}
	name := params[0]
	key := ""
	if len(params) > 1 {
		key = params[1]
	}
	ch, err := h.s.Join(h.c, name, key)
	if err != nil {
		h.c.SendError(err)
		return
	}
	topic := ch.Topic()
	if topic == "" {
		h.c.Reply(RplNoTopic)
	} else {
		h.c.Reply(RplTopic, topic)
	}
	h.names([]string{name})
}

func (h *DefaultHandler) mode(params []string) {
	if len(params) == 0 {
		h.c.SendError(NewError(ErrNeedMoreParams, ModeCmd))
		return
	}
	if HasChanPrefix(params[0]) {
		h.modeChan(params)
		return
	}
	h.modeUser(params)
}

func (h *DefaultHandler) modeChan(params []string) {
	if len(params) < 2 {
		h.c.SendError(NewError(ErrNeedMoreParams, ModeCmd))
		return
	}
	chname, params := params[0], params[1:]
	ch, err := h.s.Chan(chname)
	if err != nil {
		h.c.SendError(err)
		return
	}
	requests := parseChanModeChanges(params)
	cmds := ch.Mode(h.c)
	for _, req := range requests {
		var err *Error
		switch req.Mode {
		case ChanModeKeylock:
			err = cmds.Keylock(req.Action, req.Param)
		case ChanModeLimit:
			err = cmds.Limit(req.Action, req.Param)
		case ChanModeOper:
			err = cmds.Oper(req.Action, req.Param)
		default:
			err = NewError(ErrUnknownMode, req.Mode, ch.name)
		}
		if err != nil {
			h.c.SendError(err)
			continue
		}
	}
	cmds.Done()
}

func (h *DefaultHandler) modeUser(params []string) {
	if len(params) < 2 {
		h.c.SendError(NewError(ErrNeedMoreParams, ModeCmd))
		return
	}
	nick := params[0]
	if nick != h.c.U.Nick {
		h.c.SendError(NewError(ErrUsersDontMatch))
	}
	requests := parseUserModeChanges(params[1:])
	cmds := h.s.Mode(h.c)
	for _, req := range requests {
		var err *Error
		switch req.Mode {
		case UserModeInvisible:
			err = cmds.Invisible(req.Action)
		default:
			err = NewError(ErrUnknownMode, req.Mode)
		}
		if err != nil {
			h.c.SendError(err)
			continue
		}
	}
	cmds.Done()
}

func (h *DefaultHandler) names(params []string) {
	if len(params) == 0 {
		h.c.Send(RplEndOfNames)
		return
	}
	chname := params[0]
	ch, err := h.s.Chan(chname)
	if err != nil {
		h.c.SendError(err)
		return
	}
	nicks := strings.Join(ch.Names(), " ")
	h.c.Reply(RplNameReply, ch.Status(), ch.Name(), nicks)
	h.c.Reply(RplEndOfNames, ch.Name())
}

func (h *DefaultHandler) nick(params []string) {
	if len(params) != 1 {
		h.c.SendError(NewError(ErrNeedMoreParams, NickCmd))
		return
	}
	nick := params[0]
	if err := h.s.Nick(h.c, nick); err != nil {
		h.c.SendError(err)
		return
	}
	h.checkHandshake()
}

func (h *DefaultHandler) part(params []string) {
	if len(params) < 1 {
		h.c.SendError(NewError(ErrNeedMoreParams, PartCmd))
	}
	chname := params[0]
	reason := ""
	if len(params) >= 2 {
		reason = params[1]
	}
	if err := h.s.Part(h.c, chname, reason); err != nil {
		h.c.SendError(err)
	}
}

func (h *DefaultHandler) pass(params []string) {
	if h.c.registered {
		h.c.SendError(NewError(ErrAlreadyRegistered))
		return
	}
}

func (h *DefaultHandler) ping(params []string) {
	if len(params) == 0 {
		h.c.SendError(NewError(ErrNeedMoreParams, PingCmd))
		return
	}
	outparams := append([]string{h.s.Origin()}, params...)
	h.c.Send(PongCmd, outparams...)
}

func (h *DefaultHandler) privMsg(params []string) {
	target := params[0]
	text := params[1]
	h.s.PrivMsg(h.c, target, text)
}

func (h *DefaultHandler) quit(params []string) {
	reason := ""
	if len(params) > 0 {
		reason = params[0]
	}
	h.s.Quit(h.c, reason)
}

func (h *DefaultHandler) user(params []string) {
	if h.c.registered {
		h.c.SendError(NewError(ErrAlreadyRegistered))
		return
	}
	if len(params) != 4 {
		h.c.SendError(NewError(ErrNeedMoreParams, UserCmd))
		return
	}
	h.c.U.Name = params[0]
	h.c.U.FullName = params[3]
	h.checkHandshake()
}

// ===============

func (h *DefaultHandler) checkHandshake() error {
	if h.c.U.Nick != "" && h.c.U.Name != "" {
		h.c.SetRegistered()
		h.s.Login(h.c)
		h.welcome()
	}
	return nil
}

func (h *DefaultHandler) welcome() {
	h.c.Reply(RplWelcome, fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.c.U.Nick)).
		Reply(RplYourHost, fmt.Sprintf("Your host is %v running version %v", h.s.Origin(), Version)).
		Reply(RplCreated, fmt.Sprintf("This server was started on %v", h.s.Started.Format(time.RFC1123))).
		SendError(NewError(ErrNoMotd, "No MOTD set"))
}
