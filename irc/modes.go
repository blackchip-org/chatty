package irc

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

type modeChange struct {
	Action string
	Mode   string
	Param  string
}

func parseChanModeChanges(params []string) []modeChange {
	return parseModeChanges(params, chanModesWithArgs)
}

func parseUserModeChanges(params []string) []modeChange {
	return parseModeChanges(params, userModesWithArgs)
}

func parseModeChanges(params []string, hasArg map[string]bool) []modeChange {
	changes := make([]modeChange, 0)
	imode := 0
	iparam := 1
	n := len(params)
	for imode < len(params) {
		modes := params[imode]
		action := ""
		for _, mode := range modes {
			switch mode {
			case '+':
				action = "+"
			case '-':
				action = "-"
			default:
				mc := modeChange{}
				mc.Action = action
				mc.Mode = string(mode)
				actionMode := action + string(mode)
				if _, yes := hasArg[actionMode]; yes && iparam < n {
					mc.Param = params[iparam]
					iparam++
				}
				changes = append(changes, mc)
			}
		}
		imode = iparam
		iparam = imode + 1
	}
	return changes
}

func formatModeChanges(changes []modeChange) []string {
	params := make([]string, 0)
	for _, change := range changes {
		params = append(params, change.Action+change.Mode)
		if change.Param != "" {
			params = append(params, change.Param)
		}
	}
	return params
}
