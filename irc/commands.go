package irc

const (
	CapCmd    = "CAP"
	CapReqCmd = "REQ"
	CapEndCmd = "END"
	JoinCmd   = "JOIN"
	NickCmd   = "NICK"
	PassCmd   = "PASS"
	PingCmd   = "PING"
	PongCmd   = "PONG"
	UserCmd   = "USER"
	QuitCmd   = "QUIT"
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

const (
	ErrCannotSendToChan = "404"
	ErrNeedMoreParams   = "461"
	ErrNickNameInUse    = "433"
	ErrNoNickNameGiven  = "431"
	ErrNoSuchNick       = "401"
)

var ErrorText = map[string]string{
	ErrCannotSendToChan: "Cannot send to channel",
	ErrNeedMoreParams:   "Not enough parameters",
	ErrNickNameInUse:    "Nickname is already in use",
	ErrNoNickNameGiven:  "No nickname given",
	ErrNoSuchNick:       "No such nick/channel",
}
