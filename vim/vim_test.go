// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package vim

import (
	"fmt"
	"strings"
	"testing"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

type keytest struct {
	str string
	key KeyPress
}

func TestVim1(t *testing.T) {

	for _, kt := range []keytest{
		{"<Up>", KeyPressUp},
		{"<Down>", KeyPressDown},
		{"Z", NewSimpleKeyPress('Z')},
		{"z", NewSimpleKeyPress('z')},
		{"<C-f>", NewKeyPress(tcell.KeyRune, 'f', tcell.ModCtrl)},
		{"<A-|>", NewKeyPress(tcell.KeyRune, '|', tcell.ModAlt)},
		{"<S-\">", NewKeyPress(tcell.KeyRune, '"', tcell.ModShift)},
		{"<S-`>", NewKeyPress(tcell.KeyRune, '`', tcell.ModShift)},
		{"`", NewSimpleKeyPress('`')},
		{"<Space>", NewSimpleKeyPress(' ')},
		{"<Esc>", KeyPressEscape},
		{"<Right>", KeyPressRight},
		{"<PgDn>", KeyPressPgDn},
		{"<Esc>", KeyPressEscape},
		{"<F4>", KeyPressF4},
		{"<F12>", KeyPressF12},
		{"<Lt>", NewKeyPress(tcell.KeyRune, '<', 0)},
	} {
		res := VimStringToKeys(kt.str)
		assert.Equal(t, 1, len(res))
		assert.Equal(t, kt.key, res[0])

		str := fmt.Sprintf("%v", kt.key)
		assert.Equal(t, kt.str, str)

		if strings.Contains(kt.str, "<") {
			res = VimStringToKeys(strings.ToLower(kt.str))
			assert.Equal(t, 1, len(res))
			assert.Equal(t, kt.key, res[0])

			str = fmt.Sprintf("%v", kt.key)
			assert.Equal(t, kt.str, str)
		}
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
