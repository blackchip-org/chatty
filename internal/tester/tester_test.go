package tester

import "testing"

var lineTests = []struct {
	name  string
	input string
	want  string
}{
	{
		"translate",
		":bob!~bob@172.17.0.1 PRIVMSG #elsinore :good day, eh?",
		":bob!~bob@localhost PRIVMSG #elsinore :good day, eh?",
	},
	{
		"no prefix",
		"PRIVMSG #elsinore :good day, eh?",
		"PRIVMSG #elsinore :good day, eh?",
	},
	{
		"no host",
		":bob PRIVMSG #elsinore :good day, eh?",
		":bob PRIVMSG #elsinore :good day, eh?",
	},
}

func TestNormalizeLine(t *testing.T) {
	for _, lt := range lineTests {
		t.Run(lt.name, func(t *testing.T) {
			have := normalizeLine(lt.input)
			if lt.want != have {
				t.Errorf("\n want: %v \n have: %v", lt.want, have)
			}
		})
	}
}
