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
	ErrNickNameInUse   = "433"
	ErrNoNickNameGiven = "431"
)

var ErrorText = map[string]string{
	ErrNickNameInUse:   "Nickname is already in use",
	ErrNoNickNameGiven: "No nickname given",
}
