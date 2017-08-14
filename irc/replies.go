package irc

const (
	RplChannelModeIs = "324"
	RplCreated       = "003"
	RplEndOfMotd     = "376"
	RplEndOfNames    = "366"
	RplEndOfWho      = "315"
	RplMotdStart     = "375"
	RplMyInfo        = "004"
	RplNameReply     = "353"
	RplNoTopic       = "331"
	RplTopic         = "332"
	RplWelcome       = "001"
	RplWhoReply      = "352"
	RplYourHost      = "002"
)

var RplText = map[string]string{
	RplEndOfNames: "End of NAMES list.",
	RplEndOfWho:   "End of WHO list.",
	RplNoTopic:    "No topic is set.",
}
