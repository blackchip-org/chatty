package irc

import "strings"

type Error struct {
	Numeric string
	Params  []string
}

func (e Error) Error() string {
	return strings.Join(e.Params, " ")
}

func NewError(numeric string, params ...string) error {
	if text, ok := ErrorText[numeric]; ok {
		params = append(params, text)
	}
	return &Error{Numeric: numeric, Params: params}
}

const (
	ErrAlreadyRegistered = "462"
	ErrBadChannelKey     = "475"
	ErrCannotSendToChan  = "404"
	ErrChannelIsFull     = "471"
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
	ErrBadChannelKey:     "Cannot join channel (+k)",
	ErrCannotSendToChan:  "Cannot send to channel",
	ErrChannelIsFull:     "Cannot join channel (+l)",
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
