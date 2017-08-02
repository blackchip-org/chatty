package irc

// https://www.alien.net.au/irc/chanmodes.html
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

var hasArg = map[string]struct{}{
	"+b": struct{}{},
	"-b": struct{}{},
	"+e": struct{}{},
	"-e": struct{}{},
	"+I": struct{}{},
	"-I": struct{}{},
	"+k": struct{}{},
	"+l": struct{}{},
	"+o": struct{}{},
	"-o": struct{}{},
	"+v": struct{}{},
	"-v": struct{}{},
}

type ChanModes struct {
	Bans           []string
	BanExceptions  []string
	InviteOnly     bool
	InviationMasks []string
	Keylock        string
	Limit          int
	Moderated      bool
	NoExternalMsgs bool
	Operators      map[UserID]bool
	Private        bool
	Secret         bool
	TopicLock      bool
	Voiced         map[UserID]bool
}

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

func parseModeChanges(params []string) []modeChange {
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
