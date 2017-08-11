package irc

import (
	"sort"
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
	return c
}

func (c *Chan) Name() string {
	return c.name
}

func (c *Chan) Topic() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.topic
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
		prefix := c.modes.UserPrefix(cli.U.ID)
		nicks = append(nicks, prefix+cli.U.Nick)
	}
	sort.Strings(nicks)
	return nicks
}

func (c *Chan) Join(cli *Client) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.clients) == 0 {
		c.modes.Operators[cli.U.ID] = true
	}
	c.clients[cli.U.ID] = cli
	names := make([]string, 0, len(c.clients))
	for _, cli := range c.clients {
		cli.Relay(cli.U, JoinCmd, c.name)
		names = append(names, cli.U.Nick)
	}
	return nil
}

func (c *Chan) Part(src *Client, reason string) *Error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, exists := c.clients[src.U.ID]
	if !exists {
		return NewError(ErrNotOnChannel)
	}
	for _, cli := range c.clients {
		cli.Relay(src.U, PartCmd, c.name, reason)
	}
	c.remove(src)
	return nil
}

func (c *Chan) PrivMsg(src *Client, text string) *Error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if _, exists := c.clients[src.U.ID]; !exists {
		return NewError(ErrCannotSendToChan, c.name)
	}
	for _, cli := range c.clients {
		if cli.U.Nick == src.U.Nick {
			continue
		}
		cli.Relay(src.U, PrivMsgCmd, c.name, text)
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

func (c *Chan) Mode(src *Client) ChanModeCmds {
	return newChanModeCmds(c, src)
}

func (c *Chan) Quit(src *Client) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.remove(src)
}

func (c *Chan) remove(src *Client) {
	delete(c.modes.Operators, src.U.ID)
	delete(c.modes.Voiced, src.U.ID)
	delete(c.clients, src.U.ID)
}

type ChanModeCmds struct {
	c       *Chan
	src     *Client
	changes []modeChange
}

func newChanModeCmds(c *Chan, src *Client) ChanModeCmds {
	cmd := ChanModeCmds{
		c:       c,
		src:     src,
		changes: make([]modeChange, 0),
	}
	c.mutex.Lock()
	return cmd
}

func (cmd *ChanModeCmds) Oper(action string, name string) *Error {
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
	if !c.modes.Operators[cmd.src.U.ID] {
		return NewError(ErrChanOpPrivsNeeded, c.name)
	}

	// Is a mode change needed?
	_, targetOps := c.modes.Operators[target.U.ID]
	if set == targetOps {
		return nil
	}

	if set {
		c.modes.Operators[target.U.ID] = true
	} else {
		delete(c.modes.Operators, target.U.ID)
	}
	cmd.changes = append(cmd.changes, modeChange{
		Action: action,
		Mode:   ChanModeOper,
		Param:  name,
	})
	return nil
}

func (cmd ChanModeCmds) Done() {
	if len(cmd.changes) > 0 {
		for _, cli := range cmd.c.clients {
			params := append([]string{cmd.c.name}, formatModeChanges(cmd.changes)...)
			m := Message{
				Prefix:   cmd.src.U.Origin(),
				Cmd:      ModeCmd,
				Params:   params,
				NoSpaces: true,
			}
			cli.SendMessage(m)
		}
	}
	cmd.c.mutex.Unlock()
}
