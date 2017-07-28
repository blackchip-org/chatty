package irc

import "testing"

func TestHostnameFromAddr(t *testing.T) {
	expected := "localhost"
	got := hostnameFromAddr("127.0.0.1:12345")
	if expected != got {
		t.Fatalf("\n expected: %v \n got: %v", expected, got)
	}
}
