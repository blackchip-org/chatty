package irc

import (
	"reflect"
	"strings"
	"testing"
)

func mc(action string, char string, param string) Mode {
	return Mode{
		Action: action,
		Char:   char,
		Param:  param,
	}
}

var parseModeTests = []struct {
	modes   string
	changes []Mode
}{
	{"i", []Mode{
		mc("", "i", ""),
	}},
	{"+i", []Mode{
		mc("+", "i", ""),
	}},
	{"-i", []Mode{
		mc("-", "i", ""),
	}},
	{"+nt", []Mode{
		mc("+", "n", ""),
		mc("+", "t", ""),
	}},
	{"+n +t", []Mode{
		mc("+", "n", ""),
		mc("+", "t", ""),
	}},
	{"+n-ti", []Mode{
		mc("+", "n", ""),
		mc("-", "t", ""),
		mc("-", "i", ""),
	}},
	{"+nbI foo bar", []Mode{
		mc("+", "n", ""),
		mc("+", "b", "foo"),
		mc("+", "I", "bar"),
	}},
	{"+bnI foo bar", []Mode{
		mc("+", "b", "foo"),
		mc("+", "n", ""),
		mc("+", "I", "bar"),
	}},
	{"+b foo +b bar", []Mode{
		mc("+", "b", "foo"),
		mc("+", "b", "bar"),
	}},
}

func TestChanParseModes(t *testing.T) {
	for _, test := range parseModeTests {
		t.Run(test.modes, func(t *testing.T) {
			want := test.changes
			have := parseChanModes(strings.Split(test.modes, " "))
			if !reflect.DeepEqual(want, have) {
				t.Errorf("\n want: %v \n have: %v", want, have)
			}
		})
	}
}
