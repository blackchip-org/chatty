package irc

import (
	"bytes"
	"sort"
)

type UserChanModes struct {
	Op    bool
	Voice bool
}

func (u UserChanModes) Prefix() string {
	switch {
	case u.Op:
		return "@"
	case u.Voice:
		return "+"
	}
	return ""
}

const (
	ChanModeOp    = "o"
	ChanModeVoice = "v"
)

type UserModes struct {
	Away      bool
	Invisible bool
	Op        bool
	LocalOp   bool
}

const (
	UserModeAway      = "a"
	UserModeInvisible = "i"
	UserModeOp        = "o"
	UserModeLocalOp   = "O"
)

type ModeChanges struct {
	Set   []string
	Clear []string
	valid map[string]bool // duplicate checks
}

func NewModeChanges() *ModeChanges {
	return &ModeChanges{
		Set:   make([]string, 0),
		Clear: make([]string, 0),
		valid: make(map[string]bool),
	}
}

func (m *ModeChanges) Normalize() *ModeChanges {
	nset := make([]string, 0)
	sort.Strings(m.Set)
	for _, v := range m.Set {
		_, ok := m.valid[v]
		if ok {
			nset = append(nset, v)
		}
	}
	nclear := make([]string, 0)
	sort.Strings(m.Clear)
	for _, v := range m.Clear {
		_, ok := m.valid[v]
		if ok {
			nclear = append(nclear, v)
		}
	}
	m.Set = nset
	m.Clear = nclear
	return m
}

func (m ModeChanges) String() string {
	var buf bytes.Buffer
	if len(m.Set) > 0 {
		buf.WriteString("+")
		for _, c := range m.Set {
			buf.WriteString(c)
		}
	}
	if len(m.Clear) > 0 {
		buf.WriteString("-")
		for _, c := range m.Clear {
			buf.WriteString(c)
		}
	}
	return buf.String()
}

func ParseModeChanges(param string) *ModeChanges {
	action := "+"
	mc := NewModeChanges()
	for _, char := range param {
		switch {
		case char == '+':
			action = "+"
		case char == '-':
			action = "-"
		default:
			mode := string(char)
			if action == "+" {
				mc.Set = append(mc.Set, string(char))
			} else {
				mc.Clear = append(mc.Clear, string(char))
			}
			_, exists := mc.valid[mode]
			if !exists {
				mc.valid[mode] = true
			} else {
				delete(mc.valid, mode)
			}
		}
	}
	return mc
}
