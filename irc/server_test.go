package irc

import "testing"

func TestHostnameFromAddr(t *testing.T) {
	want := "localhost"
	have := hostnameFromAddr("127.0.0.1:12345")
	if want != have {
		t.Fatalf("\n want: %v \n have: %v", want, have)
	}
}
