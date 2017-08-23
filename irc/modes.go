package irc

import "strings"

const (
	ModeGrant  = "+"
	ModeRevoke = "-"
)

// https://www.alien.net.au/irc/chanmodes.html
type ChanModes struct {
	Bans           []string
	BanExceptions  []string
	InviteOnly     bool
	InviationMasks []string
	Key            string
	Limit          int
	Moderated      bool
	NoExternalMsgs bool
	Operators      map[UserID]bool
	Private        bool
	Secret         bool
	TopicLock      bool
	Voiced         map[UserID]bool
}

const (
	ChanModeBan            = "b"
	ChanModeBanException   = "e"
	ChanModeInviteOnly     = "i"
	ChanModeInvitationMask = "I"
	ChanModeKeylock        = "k"
	ChanModeLimit          = "l"
	ChanModeModerated      = "m"
	ChanModeNoExternalMsgs = "n"
	ChanModeOper           = "o"
	ChanModePrivate        = "p"
	ChanModeSecret         = "s"
	ChanModeTopicLock      = "t"
	ChanModeVoice          = "v"
)

var chanModesWithArgs = map[string]bool{
	"+b": true, "-b": true,
	"+e": true, "-e": true,
	"+I": true, "-I": true,
	"+k": true,
	"+l": true,
	"+o": true, "-o": true,
	"+v": true, "-v": true,
}

type UserModes struct {
	Away           bool
	Invisible      bool
	GlobalOperator bool
	LocalOperator  bool
}

const (
	UserModeAway           = "a"
	UserModeInvisible      = "i"
	UserModeGlobalOperator = "o"
	UserModeLocalOperator  = "O"
)

var userModesWithArgs = map[string]bool{}

func NewChanModes() *ChanModes {
	return &ChanModes{
		Operators: make(map[UserID]bool),
		Voiced:    make(map[UserID]bool),
	}
}

func (c ChanModes) UserPrefix(id UserID) string {
	prefix := ""
	if _, yes := c.Voiced[id]; yes {
		prefix = "+"
	}
	if _, yes := c.Operators[id]; yes {
		prefix = "@"
	}
	return prefix
}

type Mode struct {
	Action string
	Char   string
	Param  string
	List   []string
}

func parseChanModes(params []string) []Mode {
	return parseModes(params, chanModesWithArgs)
}

func parseUserModes(params []string) []Mode {
	return parseModes(params, userModesWithArgs)
}

func parseModes(params []string, hasArg map[string]bool) []Mode {
	result := make([]Mode, 0)
	imode := 0
	iparam := 1
	n := len(params)
	for imode < len(params) {
		chars := params[imode]
		action := ""
		for _, char := range chars {
			switch char {
			case '+':
				action = "+"
			case '-':
				action = "-"
			default:
				mode := Mode{
					Action: action,
					Char:   string(char),
				}
				actionMode := action + string(char)
				if _, yes := hasArg[actionMode]; yes && iparam < n {
					mode.Param = params[iparam]
					iparam++
				}
				result = append(result, mode)
			}
		}
		imode = iparam
		iparam = imode + 1
	}
	return result
}

func formatModes(modes []Mode) []string {
	chars := make([]string, 0)
	params := make([]string, 0)
	action := ""
	query := false
	for _, mode := range modes {
		// If this is a mode query for a list, skip as it wil be sent
		// as separate messages later.
		if mode.List != nil {
			query = true
			continue
		}
		if action != mode.Action {
			action = mode.Action
			chars = append(chars, action)
		}
		chars = append(chars, mode.Char)
		if mode.Param != "" {
			params = append(params, mode.Param)
		}
	}
	strchars := strings.Join(chars, "")
	if strchars == "" && query {
		return []string{}
	}
	if strchars == "" {
		strchars = "+"
	}
	return append([]string{strchars}, params...)
}
