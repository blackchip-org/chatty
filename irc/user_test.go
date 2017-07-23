package irc

import "testing"

var prefixTests = []struct {
	testName string
	nick     string
	name     string
	host     string
	prefix   string
}{
	{"full", "john", "doe", "example.com", "john!doe@example.com"},
	{"empty", "", "", "", "*!*@*"},
}

func TestUser(t *testing.T) {
	for _, test := range prefixTests {
		t.Run(test.testName, func(t *testing.T) {
			u := &User{Nick: test.nick, Name: test.name, Host: test.host}
			if test.prefix != u.Prefix() {
				t.Errorf("expecting '%v', got '%v'", test.prefix, u.Prefix())
			}
		})
	}
}
