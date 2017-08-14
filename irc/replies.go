package irc

const (
	RplChannelModeIs = "324"
	RplCreated       = "003"
	RplEndOfMotd     = "376"
	RplEndOfNames    = "366"
	RplMotdStart     = "375"
	RplMyInfo        = "004"
	RplNameReply     = "353"
	RplNoTopic       = "331"
	RplTopic         = "332"
	RplWelcome       = "001"
	RplYourHost      = "002"
)

var RplText = map[string]string{
	RplEndOfNames: "End of NAMES list.",
	RplNoTopic:    "No topic is set.",
}
