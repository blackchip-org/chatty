package irc

import (
	"reflect"
	"testing"
)

func TestNormalize(t *testing.T) {
	var tests = []struct {
		name  string
		set   []string
		clear []string
		have  *ModeChanges
	}{
		{
			"standard",
			[]string{"a", "b", "c"},
			[]string{"d", "e", "f"},
			ParseModeChanges("+abc-def").Normalize(),
		},
		{
			"sorted",
			[]string{"a", "b", "c"},
			[]string{"d", "e", "f"},
			ParseModeChanges("+bca-dfe").Normalize(),
		},
		{
			"cancel",
			[]string{"a", "b"},
			[]string{"d", "e"},
			ParseModeChanges("+abc-dec").Normalize(),
		},
		{
			"cancel all",
			[]string{},
			[]string{},
			ParseModeChanges("+abc-cba").Normalize(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want := test.set
			have := test.have.Set
			if !reflect.DeepEqual(want, have) {
				t.Errorf("\n for:  Set \n want: %v \n have: %v", want, have)
			}
			want = test.clear
			have = test.have.Clear
			if !reflect.DeepEqual(want, have) {
				t.Errorf("\n for:  Clear \n want: %v \n have: %v", want, have)
			}
		})
	}
}
