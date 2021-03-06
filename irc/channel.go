package irc

import (
	"sort"
	"strconv"
	"sync"
)

type Chan struct {
	name    string
	topic   string
	status  string
	nicks   *Nicks
	clients map[UserID]*Client
	modes   *ChanModes
	mutex   sync.RWMutex
}

const (
	ChanPrefixNetwork = "#"
	ChanPrefixLocal   = "&"
)

func HasChanPrefix(chname string) bool {
	if chname == "" {
		return false
	}
	return chname[0] == '#' || chname[0] == '&'
}

func NewChan(name string, nicks *Nicks) *Chan {
	c := &Chan{
		name:    name,
		nicks:   nicks,
		clients: make(map[UserID]*Client),
		modes:   NewChanModes(),
	}
	c.modes.NoExternalMsgs = true
	c.modes.TopicLock = true
	return c
}

func (c *Chan) Name() string {
	return c.name
}

func (c *Chan) Status() string {
	// https://modern.ircdocs.horse/#rplnamreply-353
	return "="
}

func (c *Chan) Names() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	nicks := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		prefix := c.modes.UserPrefix(cli.User.ID)
		nicks = append(nicks, prefix+cli.User.Nick)
	}
	sort.Strings(nicks)
	return nicks
}

func (c *Chan) Join(src *Client, key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.modes.Key != "" && c.modes.Key != key {
		return NewError(ErrBadChannelKey, c.name)
	}
	if c.modes.Limit > 0 && len(c.clients) >= c.modes.Limit {
		return NewError(ErrChannelIsFull, c.name)
	}
	if len(c.clients) == 0 {
		c.modes.Operators[src.User.ID] = true
	}
	c.clients[src.User.ID] = src
	names := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		cli.Relay(src.User, JoinCmd, c.name)
		names = append(names, cli.User.Nick)
	}
	return nil
}

func (c *Chan) Part(src *Client, reason string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, exists := c.clients[src.User.ID]
	if !exists {
		return NewError(ErrNotOnChannel)
	}
	for _, cli := range c.clients {
		cli.Relay(src.User, PartCmd, c.name, reason)
	}
	c.remove(src)
	return nil
}

func (c *Chan) PrivMsg(src *Client, text string) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if _, member := c.clients[src.User.ID]; c.modes.NoExternalMsgs && !member {
		return NewError(ErrCannotSendToChan, c.name)
	}
	if c.modes.Moderated {
		_, oper := c.modes.Operators[src.User.ID]
		_, voiced := c.modes.Voiced[src.User.ID]
		if !oper && !voiced {
			return NewError(ErrCannotSendToChan, c.name)
		}
	}

	for _, cli := range c.clients {
		if cli.User.Nick == src.User.Nick {
			continue
		}
		cli.Relay(src.User, PrivMsgCmd, c.name, text)
	}
	return nil
}

func (c *Chan) Topic(src *Client) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if _, member := c.clients[src.User.ID]; !member {
		return "", NewError(ErrNotOnChannel, c.name)
	}

	return c.topic, nil
}

func (c *Chan) SetTopic(src *Client, topic string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, member := c.clients[src.User.ID]; !member {
		return NewError(ErrNotOnChannel, c.name)
	}
	if _, oper := c.modes.Operators[src.User.ID]; !oper && c.modes.TopicLock {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	c.topic = topic
	for _, client := range c.clients {
		client.Relay(src.User, TopicCmd, c.name, c.topic)
	}
	return nil
}

func (c *Chan) Members() []*Client {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	members := make([]*Client, 0, len(c.clients))
	for _, client := range c.clients {
		members = append(members, client)
	}
	return members
}

func (c *Chan) Mode(src *Client) ([]Mode, error) {
	modes := make([]Mode, 0)
	if c.modes.TopicLock {
		modes = append(modes, Mode{
			Action: "+",
			Char:   ChanModeTopicLock,
		})
	}
	if c.modes.NoExternalMsgs {
		modes = append(modes, Mode{
			Action: "+",
			Char:   ChanModeNoExternalMsgs,
		})
	}
	return modes, nil
}

func (c *Chan) SetMode(src *Client) ChanModeCmds {
	return newChanModeCmds(c, src)
}

func (c *Chan) Quit(src *Client) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.remove(src)
}

func (c *Chan) remove(src *Client) {
	delete(c.modes.Operators, src.User.ID)
	delete(c.modes.Voiced, src.User.ID)
	delete(c.clients, src.User.ID)
}

type ChanModeCmds struct {
	c       *Chan
	src     *Client
	changes []Mode
}

func newChanModeCmds(c *Chan, src *Client) ChanModeCmds {
	cmd := ChanModeCmds{
		c:       c,
		src:     src,
		changes: make([]Mode, 0),
	}
	c.mutex.Lock()
	return cmd
}

func (cmd *ChanModeCmds) Ban(action string, who string) error {
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeBan,
		List:   []string{},
	})
	return nil
}

func (cmd *ChanModeCmds) Keylock(action string, key string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	if set {
		// Ignore if no key was sent
		if key == "" {
			return nil
		}
		c.modes.Key = key
	} else {
		c.modes.Key = ""
	}
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeKeylock,
		Param:  c.modes.Key,
	})
	return nil
}

func (cmd *ChanModeCmds) Limit(action string, strlimit string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		// Real server seems to ignore instead of sending an error
		return nil
	}

	if set {
		// Ignore if there isn't a valid limit
		limit, err := strconv.ParseInt(strlimit, 10, 16)
		if err != nil {
			return nil
		}
		if limit < 0 {
			return nil
		}
		c.modes.Limit = int(limit)
	} else {
		strlimit = ""
		c.modes.Limit = 0
	}
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeLimit,
		Param:  strlimit,
	})
	return nil
}

func (cmd *ChanModeCmds) Moderated(action string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	// Is a mode change needed?
	if set == c.modes.Moderated {
		return nil
	}

	c.modes.Moderated = set
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeModerated,
	})
	return nil
}

func (cmd *ChanModeCmds) NoExternalMsgs(action string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	// Is a mode change needed?
	if set == c.modes.NoExternalMsgs {
		return nil
	}

	c.modes.NoExternalMsgs = set
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeNoExternalMsgs,
	})
	return nil
}

func (cmd *ChanModeCmds) Oper(action string, name string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the target user registered?
	user, exists := c.nicks.Get(name)
	if !exists {
		return nil
	}

	// Is the target user in this channel?
	target, yes := c.clients[user.ID]
	if !yes {
		return nil
	}

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	// Is a mode change needed?
	_, targetOps := c.modes.Operators[target.User.ID]
	if set == targetOps {
		return nil
	}

	if set {
		c.modes.Operators[target.User.ID] = true
	} else {
		delete(c.modes.Operators, target.User.ID)
	}
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeOper,
		Param:  name,
	})
	return nil
}

func (cmd *ChanModeCmds) TopicLock(action string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	// Is a mode change needed?
	if set == c.modes.TopicLock {
		return nil
	}

	c.modes.TopicLock = set
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeTopicLock,
	})
	return nil
}

func (cmd *ChanModeCmds) Voice(action string, name string) error {
	c := cmd.c

	// Is the action valid?
	if action != "+" && action != "-" {
		return nil
	}
	set := action == "+"

	// Is the target user registered?
	user, exists := c.nicks.Get(name)
	if !exists {
		return nil
	}

	// Is the target user in this channel?
	target, yes := c.clients[user.ID]
	if !yes {
		return nil
	}

	// Is the user sending the command an operator?
	if !c.modes.Operators[cmd.src.User.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	// Is a mode change needed?
	_, exists = c.modes.Voiced[target.User.ID]
	if set == exists {
		return nil
	}

	if set {
		c.modes.Voiced[target.User.ID] = true
	} else {
		delete(c.modes.Voiced, target.User.ID)
	}
	cmd.changes = append(cmd.changes, Mode{
		Action: action,
		Char:   ChanModeVoice,
		Param:  name,
	})
	return nil
}

func (cmd ChanModeCmds) Done() {
	if len(cmd.changes) > 0 {
		for _, cli := range cmd.c.clients {
			params := append([]string{cmd.c.name}, formatModes(cmd.changes)...)
			// If the only param is the channel name, there is nothing
			// to say
			if len(params) > 1 {
				m := Message{
					Prefix:   cmd.src.User.Origin(),
					Cmd:      ModeCmd,
					Params:   params,
					NoSpaces: true,
				}
				cli.SendMessage(m)
			}
		}
		for _, mode := range cmd.changes {
			switch {
			case mode.Char == ChanModeBan && mode.List != nil:
				cmd.src.Reply(RplEndOfBanList, cmd.c.name)
			}
		}
	}
	cmd.c.mutex.Unlock()
}
