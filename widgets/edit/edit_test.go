// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package edit

import (
	"io"
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/stretchr/testify/assert"
)

func TestEdit1(t *testing.T) {
	w := New(Options{Caption: "hi: ", Text: "hello world"})
	sz := gowid.RenderFlowWith{C: 20}
	c1 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "hi: hello world     ")

	tset := false
	w.OnTextSet(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		tset = true
	}})
	cset := false
	w.OnCaptionSet(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		cset = true
	}})
	pset := false
	w.OnCursorPosSet(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		pset = true
	}})

	_, err := io.Copy(&Writer{w, gwtest.D}, strings.NewReader("goodbye everyone"))
	assert.NoError(t, err)
	assert.Equal(t, tset, true)
	tset = false
	c2 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "hi: goodbye everyone")
	assert.Equal(t, w.CursorEnabled(), true)
	assert.Equal(t, c2.CursorEnabled(), true)

	_, err = io.Copy(&Writer{w, gwtest.D}, strings.NewReader("multi\nline\ntest"))
	assert.NoError(t, err)
	assert.Equal(t, tset, true)
	tset = false
	c3 := w.Render(gowid.RenderFlowWith{C: 10}, gowid.Focused, gwtest.D)
	assert.Equal(t, c3.String(), "hi: multi \nline      \ntest      ")

	w.SetCaption("bye", gwtest.D)
	assert.Equal(t, w.Caption(), "bye")
	assert.Equal(t, cset, true)

	assert.Equal(t, w.CursorPos(), 0)
	w.SetCursorPos(1, gwtest.D)
	assert.Equal(t, w.CursorPos(), 1)
	assert.Equal(t, pset, true)
	x := len(w.Text())
	w.SetCursorPos(500, gwtest.D)
	assert.Equal(t, w.CursorPos(), x)

	// Make sure no crashes!
	for i := 0; i < 100; i++ {
		w.UserInput(gwtest.CursorDown(), gowid.RenderBox{C: 20, R: 5}, gowid.Focused, gwtest.D)
	}

}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
