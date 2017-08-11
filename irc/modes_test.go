package irc

import (
	"reflect"
	"strings"
	"testing"
)

func mc(action string, mode string, param string) modeChange {
	return modeChange{
		Action: action,
		Mode:   mode,
		Param:  param,
	}
}

var parseModeTests = []struct {
	modes   string
	changes []modeChange
}{
	{"i", []modeChange{
		mc("", "i", ""),
	}},
	{"+i", []modeChange{
		mc("+", "i", ""),
	}},
	{"-i", []modeChange{
		mc("-", "i", ""),
	}},
	{"+nt", []modeChange{
		mc("+", "n", ""),
		mc("+", "t", ""),
	}},
	{"+n +t", []modeChange{
		mc("+", "n", ""),
		mc("+", "t", ""),
	}},
	{"+n-ti", []modeChange{
		mc("+", "n", ""),
		mc("-", "t", ""),
		mc("-", "i", ""),
	}},
	{"+nbI foo bar", []modeChange{
		mc("+", "n", ""),
		mc("+", "b", "foo"),
		mc("+", "I", "bar"),
	}},
	{"+bnI foo bar", []modeChange{
		mc("+", "b", "foo"),
		mc("+", "n", ""),
		mc("+", "I", "bar"),
	}},
	{"+b foo +b bar", []modeChange{
		mc("+", "b", "foo"),
		mc("+", "b", "bar"),
	}},
}

func TestChanParseModes(t *testing.T) {
	for _, test := range parseModeTests {
		t.Run(test.modes, func(t *testing.T) {
			want := test.changes
			have := parseChanModeChanges(strings.Split(test.modes, " "))
			if !reflect.DeepEqual(want, have) {
				t.Errorf("\n want: %v \n have: %v", want, have)
			}
		})
	}
}
