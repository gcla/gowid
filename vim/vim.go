// Copyright 2020 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package vim provides utilities for parsing and generating vim-like
// keystrokes. This is heavily tailored towards compatibility with key
// events constructed by tcell, for use in terminals.
package vim

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell"
)

//======================================================================

var (
	DefaultEmacsDownKeys = []KeyPress{KeyCtrl('n')}
	DefaultEmacsUpKeys   = []KeyPress{KeyCtrl('p')}
	DefaultVimDownKeys   = []KeyPress{Key('j')}
	DefaultVimUpKeys     = []KeyPress{Key('k')}
	DefaultDownKeys      = []KeyPress{KeyPressDown}
	DefaultUpKeys        = []KeyPress{KeyPressUp}
	AllDownKeys          = append(DefaultDownKeys, append(DefaultEmacsDownKeys, DefaultVimDownKeys...)...)
	AllUpKeys            = append(DefaultUpKeys, append(DefaultEmacsUpKeys, DefaultVimUpKeys...)...)

	DefaultEmacsLeftKeys  = []KeyPress{KeyCtrl('b')}
	DefaultEmacsRightKeys = []KeyPress{KeyCtrl('f')}
	DefaultVimLeftKeys    = []KeyPress{Key('h')}
	DefaultVimRightKeys   = []KeyPress{Key('l')}
	DefaultLeftKeys       = []KeyPress{KeyPressLeft}
	DefaultRightKeys      = []KeyPress{KeyPressRight}
	AllLeftKeys           = append(DefaultLeftKeys, append(DefaultEmacsLeftKeys, DefaultVimLeftKeys...)...)
	AllRightKeys          = append(DefaultRightKeys, append(DefaultEmacsRightKeys, DefaultVimRightKeys...)...)

	ModMapReverse = map[string]tcell.ModMask{
		"C": tcell.ModCtrl,
		"c": tcell.ModCtrl,
		"A": tcell.ModAlt,
		"a": tcell.ModAlt,
		"S": tcell.ModShift,
		"s": tcell.ModShift,
	}

	ModMap = map[tcell.ModMask]string{
		tcell.ModCtrl:  "C",
		tcell.ModAlt:   "A",
		tcell.ModShift: "S",
	}

	SpecialKeyMapReverse = map[string]tcell.Key{
		"<Up>":    tcell.KeyUp,
		"<Down>":  tcell.KeyDown,
		"<Left>":  tcell.KeyLeft,
		"<Right>": tcell.KeyRight,
		"<Enter>": tcell.KeyEnter,
		"<Esc>":   tcell.KeyEscape,
		"<Tab>":   tcell.KeyTab,
		"<Home>":  tcell.KeyHome,
		"<End>":   tcell.KeyEnd,
		"<PgUp>":  tcell.KeyPgUp,
		"<PgDn>":  tcell.KeyPgDn,
		"<F1>":    tcell.KeyF1,
		"<F2>":    tcell.KeyF2,
		"<F3>":    tcell.KeyF3,
		"<F4>":    tcell.KeyF4,
		"<F5>":    tcell.KeyF5,
		"<F6>":    tcell.KeyF6,
		"<F7>":    tcell.KeyF7,
		"<F8>":    tcell.KeyF8,
		"<F9>":    tcell.KeyF9,
		"<F10>":   tcell.KeyF10,
		"<F11>":   tcell.KeyF11,
		"<F12>":   tcell.KeyF12,
	}

	SpecialKeyMap = map[tcell.Key]string{
		tcell.KeyUp:     "<Up>",
		tcell.KeyDown:   "<Down>",
		tcell.KeyLeft:   "<Left>",
		tcell.KeyRight:  "<Right>",
		tcell.KeyEnter:  "<Enter>",
		tcell.KeyEscape: "<Esc>",
		tcell.KeyTab:    "<Tab>",
		tcell.KeyHome:   "<Home>",
		tcell.KeyEnd:    "<End>",
		tcell.KeyPgUp:   "<PgUp>",
		tcell.KeyPgDn:   "<PgDn>",
		tcell.KeyF1:     "<F1>",
		tcell.KeyF2:     "<F2>",
		tcell.KeyF3:     "<F3>",
		tcell.KeyF4:     "<F4>",
		tcell.KeyF5:     "<F5>",
		tcell.KeyF6:     "<F6>",
		tcell.KeyF7:     "<F7>",
		tcell.KeyF8:     "<F8>",
		tcell.KeyF9:     "<F9>",
		tcell.KeyF10:    "<F10>",
		tcell.KeyF11:    "<F11>",
		tcell.KeyF12:    "<F12>",
	}

	KeyPressUp     KeyPress = NewKeyPress(tcell.KeyUp, 0, 0)
	KeyPressDown   KeyPress = NewKeyPress(tcell.KeyDown, 0, 0)
	KeyPressLeft   KeyPress = NewKeyPress(tcell.KeyLeft, 0, 0)
	KeyPressRight  KeyPress = NewKeyPress(tcell.KeyRight, 0, 0)
	KeyPressEnter  KeyPress = NewKeyPress(tcell.KeyEnter, 0, 0)
	KeyPressEscape KeyPress = NewKeyPress(tcell.KeyEscape, 0, 0)
	KeyPressTab    KeyPress = NewKeyPress(tcell.KeyTab, 0, 0)
	KeyPressHome   KeyPress = NewKeyPress(tcell.KeyTab, 0, 0)
	KeyPressEnd    KeyPress = NewKeyPress(tcell.KeyTab, 0, 0)
	KeyPressPgUp   KeyPress = NewKeyPress(tcell.KeyPgUp, 0, 0)
	KeyPressPgDn   KeyPress = NewKeyPress(tcell.KeyPgDn, 0, 0)
	KeyPressF1     KeyPress = NewKeyPress(tcell.KeyF1, 0, 0)
	KeyPressF2     KeyPress = NewKeyPress(tcell.KeyF2, 0, 0)
	KeyPressF3     KeyPress = NewKeyPress(tcell.KeyF3, 0, 0)
	KeyPressF4     KeyPress = NewKeyPress(tcell.KeyF4, 0, 0)
	KeyPressF5     KeyPress = NewKeyPress(tcell.KeyF5, 0, 0)
	KeyPressF6     KeyPress = NewKeyPress(tcell.KeyF6, 0, 0)
	KeyPressF7     KeyPress = NewKeyPress(tcell.KeyF7, 0, 0)
	KeyPressF8     KeyPress = NewKeyPress(tcell.KeyF8, 0, 0)
	KeyPressF9     KeyPress = NewKeyPress(tcell.KeyF9, 0, 0)
	KeyPressF10    KeyPress = NewKeyPress(tcell.KeyF10, 0, 0)
	KeyPressF11    KeyPress = NewKeyPress(tcell.KeyF11, 0, 0)
	KeyPressF12    KeyPress = NewKeyPress(tcell.KeyF12, 0, 0)

	KeyPressF = []KeyPress{
		KeyPressF1,
		KeyPressF2,
		KeyPressF3,
		KeyPressF4,
		KeyPressF5,
		KeyPressF6,
		KeyPressF7,
		KeyPressF8,
		KeyPressF9,
		KeyPressF10,
		KeyPressF11,
		KeyPressF12,
	}
)

var keyExp *regexp.Regexp

func init() {
	// This crazy-looking regexp will parse correctly formatted vim keys. See vim_test.go for examples. The groups
	// allow easy extraction of the key pieces of syntax.
	keyExp = regexp.MustCompile(`(<(?P<mod>[CSAcsa])-((?P<modchar>[A-Za-z0-9!@#$%^&*()\[\]\/\-_+=~"':;<>,.?|` + "`" + `])|(?P<modspecial>(?i)space(?-i)))>|(?P<char>[A-Za-z0-9!@#$%^&*()\[\]\/\-_+=~"':;>,.?|` + "`" + `])|<(?P<up>(?i)Up(?-i))>|<(?P<down>(?i)Down(?-i))>|<(?P<left>(?i)Left(?-i))>|<(?P<right>(?i)Right(?-i))>|<(?P<esc>(?i)Esc(?-i))>|<(?P<cr>(?i)CR(?-i))>|<(?P<return>(?i)Return(?-i))>|<(?P<enter>(?i)Enter(?-i))>|<(?P<space>(?i)Space(?-i))>|<(?P<lt>(?i)lt(?-i))>|<(?P<bs>(?i)BS(?-i))>|<(?P<tab>(?i)Tab(?-i))>|<(?P<home>(?i)Home(?-i))>|<(?P<end>(?i)End(?-i))>|<(?P<pgup>(?i)PgUp(?-i))>|<(?P<pgdn>(?i)PgDn(?-i))>|<([fF])(?P<f>[1-9]|(1[0-2]))>)`)
}

// KeyPress represents a gowid keypress. It's a tcell.EventKey without the time
// of the keypress.
type KeyPress gowid.Key

func KeyCtrl(r rune) KeyPress {
	return KeyPress(gowid.MakeKeyExt2(tcell.ModCtrl, tcell.KeyRune, r))
}

// KeyPressFromTcell converts a *tcell.EventKey to a KeyPress. This can then be
// serialized to a vim-style keypress e.g. <C-s>
func KeyPressFromTcell(k *tcell.EventKey) KeyPress {
	mod := k.Modifiers()
	tk := k.Key()
	ch := k.Rune()
	if tk >= tcell.KeyCtrlA && tk <= tcell.KeyCtrlZ {
		ch = rune(int(tk) + int('a') - 1)
		tk = tcell.KeyRune
	} else {
		switch tk {
		case tcell.KeyCtrlSpace:
			ch = ' '
			tk = tcell.KeyRune
		case tcell.KeyCtrlLeftSq:
			ch = '['
			tk = tcell.KeyRune
		case tcell.KeyCtrlRightSq:
			ch = ']'
			tk = tcell.KeyRune
		case tcell.KeyCtrlCarat:
			ch = '^'
			tk = tcell.KeyRune
		case tcell.KeyCtrlUnderscore:
			ch = '_'
			tk = tcell.KeyRune
		case tcell.KeyCtrlBackslash:
			ch = '\\'
			tk = tcell.KeyRune
		}
	}
	return KeyPress(gowid.MakeKeyExt2(mod, tk, ch))
}

func NewSimpleKeyPress(ch rune) KeyPress {
	return NewKeyPress(tcell.KeyRune, ch, 0)
}

func Key(ch rune) KeyPress {
	return NewKeyPress(tcell.KeyRune, ch, 0)
}

func NewKeyPress(k tcell.Key, ch rune, mod tcell.ModMask) KeyPress {
	if k == tcell.KeyRune && (ch < ' ' || ch == 0x7f) {
		// Turn specials into proper key codes.  This is for
		// control characters and the DEL.
		k = tcell.Key(ch)
		if mod == tcell.ModNone && ch < ' ' {
			switch tcell.Key(ch) {
			case tcell.KeyBackspace, tcell.KeyTab, tcell.KeyEsc, tcell.KeyEnter:
				// these keys are directly typeable without CTRL
			default:
				// most likely entered with a CTRL keypress
				mod = tcell.ModCtrl
			}
		}
	}
	return KeyPress(gowid.MakeKeyExt2(mod, k, ch))
}

func (k KeyPress) String() string {
	gk := gowid.Key(k)
	if gk.Key() == tcell.KeyRune {
		if mod, ok := ModMap[gk.Modifiers()]; ok {
			if gk.Rune() == ' ' {
				return fmt.Sprintf("<%s-space>", mod)
			} else {
				return fmt.Sprintf("<%s-%c>", mod, gk.Rune())
			}
		} else {
			if gk.Rune() == '<' {
				return "<Lt>"
			} else if gk.Rune() == ' ' {
				return "<Space>"
			} else {
				return string(gk.Rune())
			}
		}
	} else if str, ok := SpecialKeyMap[gk.Key()]; ok {
		return str
	} else {
		return "<Unknown>"
	}
}

// KeySequence is an array of KeyPress. The KeySequence type allows
// the sequence to be serialized the way vim would do it e.g. <C-s>abc<Esc>
type KeySequence []KeyPress

func (ks KeySequence) String() string {
	var res string
	for _, kp := range ks {
		res = res + kp.String()
	}
	return res
}

// VimStringToKeys converts e.g. <C-s>abc<Esc> into a sequence of KeyPress
func VimStringToKeys(input string) KeySequence {
	matches := keyExp.FindAllStringSubmatch(input, -1)
	results := make([]map[string]string, len(matches))
	for j, _ := range matches {
		results[j] = make(map[string]string)
		for i, name := range keyExp.SubexpNames() {
			if i != 0 && name != "" {
				results[j][name] = matches[j][i]
			}
		}
	}

	res := make(KeySequence, 0)

	for _, result := range results {
		if str, ok := result["up"]; ok && str != "" {
			res = append(res, KeyPressUp)
		} else if str, ok := result["down"]; ok && str != "" {
			res = append(res, KeyPressDown)
		} else if str, ok := result["left"]; ok && str != "" {
			res = append(res, KeyPressLeft)
		} else if str, ok := result["right"]; ok && str != "" {
			res = append(res, KeyPressRight)
		} else if str, ok := result["cr"]; ok && str != "" {
			res = append(res, KeyPressEnter)
		} else if str, ok := result["return"]; ok && str != "" {
			res = append(res, KeyPressEnter)
		} else if str, ok := result["enter"]; ok && str != "" {
			res = append(res, KeyPressEnter)
		} else if str, ok := result["esc"]; ok && str != "" {
			res = append(res, KeyPressEscape)
		} else if str, ok := result["tab"]; ok && str != "" {
			res = append(res, KeyPressTab)
		} else if str, ok := result["home"]; ok && str != "" {
			res = append(res, KeyPressHome)
		} else if str, ok := result["end"]; ok && str != "" {
			res = append(res, KeyPressEnd)
		} else if str, ok := result["pgup"]; ok && str != "" {
			res = append(res, KeyPressPgUp)
		} else if str, ok := result["pgdn"]; ok && str != "" {
			res = append(res, KeyPressPgDn)
		} else if str, ok := result["f"]; ok && str != "" {
			i, _ := strconv.Atoi(str)
			res = append(res, KeyPressF[i-1])
		} else if str, ok := result["lt"]; ok && str != "" {
			res = append(res, NewSimpleKeyPress('<'))
		} else if str, ok := result["space"]; ok && str != "" {
			res = append(res, NewSimpleKeyPress(' '))
		} else if str, ok := result["char"]; ok && str != "" {
			res = append(res, NewSimpleKeyPress(rune(str[0])))
		} else if str, ok := result["modchar"]; ok && str != "" {
			// regexp guarantees ModMask lookup is safe
			res = append(res, NewKeyPress(tcell.KeyRune, rune(str[0]), ModMapReverse[result["mod"]]))
		} else if str, ok := result["modspecial"]; ok && str != "" {
			// regexp guarantees ModMask lookup is safe
			switch strings.ToLower(str) {
			case "space":
				res = append(res, NewKeyPress(tcell.KeyRune, ' ', ModMapReverse[result["mod"]]))
			}
		}
	}

	return res
}

func KeyIn(k *tcell.EventKey, keys []KeyPress) bool {
	kp := KeyPressFromTcell(k)
	for i, _ := range keys {
		if kp == keys[i] {
			return true
		}
	}
	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
