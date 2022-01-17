// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package edit

import (
	"io"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func evclick(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.Button1, 0)
}

func evunclick(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.ButtonNone, 0)
}

func TestType1(t *testing.T) {
	w := New(Options{Caption: "", Text: "hi: 现在 abc"})
	sz := gowid.RenderFlowWith{C: 15}
	c1 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "hi: 现在 abc   ", c1.String())

	evq := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)

	w.SetCursorPos(0, gwtest.D)
	w.UserInput(evq, sz, gowid.Focused, gwtest.D)
	c1 = w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "qhi: 现在 abc  ", c1.String())

	w.SetCursorPos(6, gwtest.D)
	w.UserInput(evq, sz, gowid.Focused, gwtest.D)
	c1 = w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "qhi: 现q在 abc ", c1.String())
}

func TestRender1(t *testing.T) {
	w := New(Options{Caption: "", Text: "abcde现fgh"})
	sz := gowid.RenderFlowWith{C: 6}
	c1 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "abcde \n现fgh ", c1.String())
}

func TestType2(t *testing.T) {
	w := New(Options{Caption: "", Text: "hi:  abc"})
	sz := gowid.RenderFlowWith{C: 15}
	c1 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "hi:  abc       ", c1.String())

	evq := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)

	w.SetCursorPos(0, gwtest.D)
	w.UserInput(evq, sz, gowid.Focused, gwtest.D)
	c1 = w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "qhi:  abc      ", c1.String())
}

func TestMove1(t *testing.T) {
	w := New(Options{Caption: "hi: ", Text: "now\n\nis the time"})
	sz := gowid.RenderFlowWith{C: 12}
	c1 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "hi: now     \n            \nis the time ", c1.String())

	w.SetCursorPos(0, gwtest.D)
	assert.Equal(t, 0, w.CursorPos())
	w.UserInput(gwtest.CursorDown(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 4, w.CursorPos())
	w.UserInput(gwtest.CursorRight(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 5, w.CursorPos())
	w.UserInput(gwtest.CursorLeft(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 4, w.CursorPos())
	w.UserInput(gwtest.CursorDown(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 5, w.CursorPos())
}

func TestLong1(t *testing.T) {
	w := New(Options{Caption: "现: ", Text: "现在是hetimeforallgoodmentocometotheaid\n\nofthe"})
	sz := gowid.RenderFlowWith{C: 12}
	c1 := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "现: 现在是he\ntimeforallgo\nodmentocomet\notheaid     \n            \nofthe       ", c1.String())

	clickat := func(x, y int) {
		w.UserInput(evclick(x, y), sz, gowid.Focused, gwtest.D)
		gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
		w.UserInput(evunclick(x, y), sz, gowid.Focused, gwtest.D)
		gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	}

	clickat(4, 0)
	assert.Equal(t, 0, w.CursorPos())
	w.SetCursorPos(1, gwtest.D)
	assert.Equal(t, 1, w.CursorPos())
	x := utf8.RuneCountInString(w.Text())
	w.SetCursorPos(500, gwtest.D)
	assert.Equal(t, x, w.CursorPos())

	clickat(11, 0)
	assert.Equal(t, 4, w.CursorPos())

	clickat(0, 1)
	assert.Equal(t, 5, w.CursorPos())

	w.UserInput(gwtest.CursorLeft(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 4, w.CursorPos())

	w.UserInput(gwtest.CursorDown(), sz, gowid.Focused, gwtest.D)
	assert.Equal(t, 16, w.CursorPos())
}

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
