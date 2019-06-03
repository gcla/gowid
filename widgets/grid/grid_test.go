// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package grid

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/tcell"
	"github.com/stretchr/testify/assert"
)

func TestGridFlow(t *testing.T) {

	btns := make([]gowid.IWidget, 0)
	clicks := make([]*gwtest.ButtonTester, 0)

	for i := 0; i < 9; i++ {
		btn := button.New(text.New("abc"))
		click := &gwtest.ButtonTester{Gotit: false}
		btn.OnClick(click)
		btns = append(btns, btn)
		clicks = append(clicks, click)
	}

	gf := New(btns, 5, 1, 1, gowid.HAlignMiddle{})

	st1 := "  <abc> <abc> <abc> \n                    \n  <abc> <abc> <abc> \n                    \n  <abc> <abc> <abc> "

	sz := gowid.RenderBox{C: 20, R: 5}
	c := gf.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), st1)

	c2 := gf.Render(gowid.RenderFlowWith{C: 20}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), st1)

	assert.Equal(t, gf.Focus(), 0)

	evspace := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	evmdown := tcell.NewEventMouse(1, 1, tcell.WheelDown, 0)
	evmup := tcell.NewEventMouse(1, 1, tcell.WheelUp, 0)
	evmright := tcell.NewEventMouse(1, 1, tcell.WheelRight, 0)
	evmleft := tcell.NewEventMouse(1, 1, tcell.WheelLeft, 0)

	cbcalled := false

	gf.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		assert.Equal(t, w, gf)
		cbcalled = true
	}})

	gf.UserInput(gwtest.CursorRight(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evspace, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, gf.Focus())
	assert.Equal(t, clicks[1].Gotit, true)
	assert.Equal(t, false, cbcalled)
	cbcalled = false

	clicks[1].Gotit = false

	gf.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 4, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 7, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 7, gf.Focus())
	assert.Equal(t, false, cbcalled)
	cbcalled = false

	gf.UserInput(evmup, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 4, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmup, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmup, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, gf.Focus())
	assert.Equal(t, false, cbcalled)
	cbcalled = false

	gf.UserInput(evmright, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 2, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmright, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 2, gf.Focus())
	assert.Equal(t, false, cbcalled)
	cbcalled = false

	gf.UserInput(gwtest.CursorDown(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 5, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false

	gf.UserInput(evmleft, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 4, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmleft, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 3, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(evmleft, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 3, gf.Focus())
	assert.Equal(t, false, cbcalled)
	cbcalled = false

	gf.UserInput(gwtest.CursorLeft(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 2, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(gwtest.CursorLeft(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(gwtest.CursorLeft(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 0, gf.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	gf.UserInput(gwtest.CursorLeft(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 0, gf.Focus())
	assert.Equal(t, false, cbcalled)
	cbcalled = false

}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
