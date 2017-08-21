package irc

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blackchip-org/chatty/internal/passwd"
	"github.com/boltdb/bolt"
)

type Handler interface {
	Handle(Command) error
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

func (h *DefaultHandler) Handle(cmd Command) error {
	if !h.c.registered {
		allowed := prereg[cmd.Name]
		if !allowed {
			h.c.SendError(NewError(ErrNotRegistered))
			return h.c.err
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
	case TopicCmd:
		h.topic(cmd.Params)
	case UserCmd:
		h.user(cmd.Params)
	case QuitCmd:
		h.quit(cmd.Params)
	case WhoCmd:
		h.who(cmd.Params)
	default:
		log.Printf("unhandled message: %+v", cmd)
	}
	return h.c.err
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
	_, err := h.s.Join(h.c, name, key)
	if err != nil {
		h.c.SendError(err)
		return
	}
	h.topic([]string{name})
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
	if len(params) < 1 {
		h.c.SendError(NewError(ErrNeedMoreParams, ModeCmd))
		return
	}
	chname := params[0]
	ch, err := h.s.Chan(chname)
	if err != nil {
		h.c.SendError(err)
		return
	}

	if len(params) == 1 {
		modes, err := ch.Mode(h.c)
		if err != nil {
			h.c.SendError(err)
			return
		}
		fmodes := formatModes(modes)
		rparams := append([]string{ch.name}, fmodes...)
		message := Message{
			Prefix:   h.c.ServerName,
			Target:   h.c.User.Nick,
			Cmd:      RplChannelModeIs,
			Params:   rparams,
			NoSpaces: true,
		}
		h.c.SendMessage(message)
		return
	}

	params = params[1:]
	requests := parseChanModes(params)
	cmds := ch.SetMode(h.c)
	for _, req := range requests {
		var err error
		switch req.Char {
		case ChanModeBan:
			err = cmds.Ban(req.Action, req.Param)
		case ChanModeKeylock:
			err = cmds.Keylock(req.Action, req.Param)
		case ChanModeLimit:
			err = cmds.Limit(req.Action, req.Param)
		case ChanModeModerated:
			err = cmds.Moderated(req.Action)
		case ChanModeNoExternalMsgs:
			err = cmds.NoExternalMsgs(req.Action)
		case ChanModeTopicLock:
			err = cmds.TopicLock(req.Action)
		case ChanModeOper:
			err = cmds.Oper(req.Action, req.Param)
		case ChanModeVoice:
			err = cmds.Voice(req.Action, req.Param)
		default:
			err = NewError(ErrUnknownMode, req.Char, ch.name)
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
	if nick != h.c.User.Nick {
		h.c.SendError(NewError(ErrUsersDontMatch))
	}
	requests := parseUserModes(params[1:])
	cmds := h.s.Mode(h.c)
	for _, req := range requests {
		var err error
		switch req.Char {
		case UserModeInvisible:
			err = cmds.Invisible(req.Action)
		default:
			err = NewError(ErrUnknownMode, req.Char)
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
	if len(params) == 0 {
		h.c.SendError(NewError(ErrNeedMoreParams, PingCmd))
		return
	}
	h.c.password = params[0]
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
	err := h.s.PrivMsg(h.c, target, text)
	if err != nil {
		h.c.SendError(err)
	}
}

func (h *DefaultHandler) topic(params []string) {
	if len(params) == 0 {
		h.c.Send(ErrNeedMoreParams, TopicCmd)
		return
	}

	chname := params[0]
	ch, err := h.s.Chan(chname)
	if err != nil {
		h.c.SendError(err)
		return
	}

	if len(params) == 1 {
		topic, err := ch.Topic(h.c)
		if err != nil {
			h.c.SendError(err)
			return
		}
		if topic == "" {
			h.c.Reply(RplNoTopic, ch.name)
		} else {
			h.c.Reply(RplTopic, ch.name, topic)
		}
	} else {
		topic := params[1]
		err := ch.SetTopic(h.c, topic)
		if err != nil {
			h.c.SendError(err)
			return
		}
	}
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
	h.c.User.Name = params[0]
	h.c.User.FullName = params[3]
	h.checkHandshake()
}

// http://chi.cs.uchicago.edu/chirc/assignment3.html#who
// Only channels at the moment
func (h *DefaultHandler) who(params []string) {
	if len(params) < 1 {
		h.c.SendError(NewError(ErrNeedMoreParams, WhoCmd))
		return
	}
	chname := params[0]
	ch, err := h.s.Chan(chname)
	if err != nil {
		h.c.SendError(err)
		return
	}
	members := ch.Members()
	for _, member := range members {
		avail := "H"
		op := ""
		prefix := ch.modes.UserPrefix(member.User.ID)
		params := []string{
			ch.name,
			"~" + member.User.Name,
			member.User.Host,
			h.c.ServerName, // FIXME
			member.User.Nick,
			avail + op + prefix,
			"0 " + member.User.FullName,
		}
		h.c.Reply(RplWhoReply, params...)
	}
	h.c.Reply(RplEndOfWho, ch.name)
}

// ===============

func (h *DefaultHandler) checkHandshake() error {
	if h.c.User.Nick != "" && h.c.User.Name != "" {
		if err := h.canRegister(); err != nil {
			return err
		}
		h.c.SetRegistered()
		h.s.Login(h.c)
		h.welcome()
		return nil
	}
	return nil
}

func (h *DefaultHandler) canRegister() error {
	err := h.s.db.View(func(tx *bolt.Tx) error {
		bpass := tx.Bucket(BucketConfig).Get(ConfigPass)
		if bpass == nil {
			return nil
		}
		bsalt := tx.Bucket(BucketConfig).Get(ConfigSalt)
		if bsalt == nil {
			return errors.New("no salt")
		}
		if !bytes.Equal(bpass, passwd.Encode([]byte(h.c.password), bsalt)) {
			h.c.SendError(NewError(ErrPasswordMismatch))
			return errors.New("invalid password")
		}
		return nil
	})
	return err
}

func (h *DefaultHandler) welcome() {
	log.Printf("[%v] is %v", h.c.conn.RemoteAddr(), h.c.User.Nick)
	h.c.Reply(RplWelcome, fmt.Sprintf("Welcome to the Internet Relay Chat Network %v", h.c.User.Nick)).
		Reply(RplYourHost, fmt.Sprintf("Your host is %v running version %v", h.s.Origin(), Version)).
		Reply(RplCreated, fmt.Sprintf("This server was started on %v", h.s.Started.Format(time.RFC1123))).
		SendError(NewError(ErrNoMotd, "No MOTD set"))
}
