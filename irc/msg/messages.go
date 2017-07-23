package msg

const (
	Cap    = "CAP"
	CapReq = "REQ"
	CapEnd = "END"
	Join   = "JOIN"
	Nick   = "NICK"
	Pass   = "PASS"
	Ping   = "PING"
	User   = "USER"
	Quit   = "QUIT"
)

const (
	Created    = "003"
	EndOfMotd  = "376"
	EndOfNames = "366"
	MotdStart  = "375"
	MyInfo     = "004"
	NameReply  = "353"
	NoTopic    = "331"
	Topic      = "332"
	Welcome    = "001"
	YourHost   = "002"
)

const (
	ErrNickNameInUse   = "433"
	ErrNoNickNameGiven = "431"
)

var ErrorText = map[string]string{
	ErrNickNameInUse:   "Nickname is already in use",
	ErrNoNickNameGiven: "No nickname given",
}
