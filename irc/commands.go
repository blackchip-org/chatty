package irc

const (
	CapCmd     = "CAP"
	CapLsCmd   = "LS"
	CapReqCmd  = "REQ"
	CapEndCmd  = "END"
	JoinCmd    = "JOIN"
	ModeCmd    = "MODE"
	NamesCmd   = "NAMES"
	NickCmd    = "NICK"
	PartCmd    = "PART"
	PassCmd    = "PASS"
	PingCmd    = "PING"
	PongCmd    = "PONG"
	PrivMsgCmd = "PRIVMSG"
	UserCmd    = "USER"
	QuitCmd    = "QUIT"
)

const (
	RplCreated    = "003"
	RplEndOfMotd  = "376"
	RplEndOfNames = "366"
	RplMotdStart  = "375"
	RplMyInfo     = "004"
	RplNameReply  = "353"
	RplNoTopic    = "331"
	RplTopic      = "332"
	RplWelcome    = "001"
	RplYourHost   = "002"
)

var RplText = map[string]string{
	RplEndOfNames: "End of NAMES list.",
	RplNoTopic:    "No topic is set",
}

const (
	ErrAlreadyRegistered = "462"
	ErrCannotSendToChan  = "404"
	ErrChanOpPrivsNeeded = "482"
	ErrInvalidCapCmd     = "410"
	ErrNeedMoreParams    = "461"
	ErrNickNameInUse     = "433"
	ErrNoMotd            = "422"
	ErrNoNickNameGiven   = "431"
	ErrNoSuchChannel     = "403"
	ErrNoSuchNick        = "401"
	ErrNotOnChannel      = "442"
	ErrNotRegistered     = "451"
	ErrUModeUnknownFlag  = "501"
	ErrUnknownMode       = "472"
	ErrUsersDontMatch    = "502"
)

var ErrorText = map[string]string{
	ErrAlreadyRegistered: "Unauthorized command (already registered)",
	ErrCannotSendToChan:  "Cannot send to channel",
	ErrChanOpPrivsNeeded: "You're not channel operator",
	ErrInvalidCapCmd:     "Invalid CAP command",
	ErrNeedMoreParams:    "Not enough parameters",
	ErrNickNameInUse:     "Nickname is already in use",
	ErrNoNickNameGiven:   "No nickname given",
	ErrNoSuchChannel:     "No such channel",
	ErrNoSuchNick:        "No such nick/channel",
	ErrNotOnChannel:      "You're not on that channel",
	ErrNotRegistered:     "You have not registered",
	ErrUModeUnknownFlag:  "Unknown MODE flag",
	ErrUnknownMode:       "is unknown mode char to me",
	ErrUsersDontMatch:    "Cannot change mode for other users",
}
