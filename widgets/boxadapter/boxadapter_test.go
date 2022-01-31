// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package boxadapter

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/list"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestBoxadapter1(t *testing.T) {
	w := edit.New(edit.Options{Caption: "", Text: "aaaaaaaaaaaaaaaaaaaa"})
	w2 := edit.New(edit.Options{Caption: "", Text: "bbbbbbbbbbbbbbbbbbbb"})

	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaa", c1.String())

	c1 = w.Render(gowid.RenderFlowWith{C: 6}, gowid.Focused, gwtest.D)
	assert.Equal(t, "aaaaaa\naaaaaa\naaaaaa\naa    ", c1.String())

	walker := list.NewSimpleListWalker([]gowid.IWidget{w, w2})
	lb := list.New(walker)

	c1 = lb.Render(gowid.RenderFlowWith{C: 6}, gowid.Focused, gwtest.D)
	assert.Equal(t, "aaaaaa\naaaaaa\naaaaaa\naa    \nbbbbbb\nbbbbbb\nbbbbbb\nbb    ", c1.String())

	evx := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
	evlmx1y0 := tcell.NewEventMouse(1, 0, tcell.Button1, 0)
	evlmx1y4 := tcell.NewEventMouse(1, 4, tcell.Button1, 0)
	evlmx2y2 := tcell.NewEventMouse(2, 2, tcell.Button1, 0)
	evnonex1y0 := tcell.NewEventMouse(1, 0, tcell.ButtonNone, 0)
	evnonex1y4 := tcell.NewEventMouse(1, 4, tcell.ButtonNone, 0)
	evnonex2y2 := tcell.NewEventMouse(2, 2, tcell.ButtonNone, 0)

	w.UserInput(evlmx1y0, gowid.RenderFlowWith{C: 21}, gowid.Focused, gwtest.D)
	w.UserInput(evnonex1y0, gowid.RenderFlowWith{C: 21}, gowid.Focused, gwtest.D)
	w.UserInput(evx, gowid.RenderFlowWith{C: 21}, gowid.Focused, gwtest.D)
	c1 = w.Render(gowid.RenderFlowWith{C: 21}, gowid.Focused, gwtest.D)
	assert.Equal(t, "axaaaaaaaaaaaaaaaaaaa", c1.String())

	sz := gowid.RenderFlowWith{C: 6}
	c1 = lb.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"axaaaa",
		"aaaaaa",
		"aaaaaa",
		"aaa   ",
		"bbbbbb",
		"bbbbbb",
		"bbbbbb",
		"bb    ",
	}, "\n"), c1.String())

	lb.UserInput(evlmx1y4, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	lb.UserInput(evnonex1y4, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	lb.UserInput(evx, sz, gowid.Focused, gwtest.D)

	c1 = lb.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"axaaaa",
		"aaaaaa",
		"aaaaaa",
		"aaa   ",
		"bxbbbb",
		"bbbbbb",
		"bbbbbb",
		"bbb   ",
	}, "\n"), c1.String())

	bw := New(w, 2)
	bw2 := New(w2, 2)

	cw := bw.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"axaaaa",
		"aaaaaa",
	}, "\n"), cw.String())

	walker2 := list.NewSimpleListWalker([]gowid.IWidget{bw, bw2})
	lb2 := list.New(walker2)

	c2 := lb2.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"axaaaa",
		"aaaaaa",
		"bxbbbb",
		"bbbbbb",
	}, "\n"), c2.String())

	lb2.UserInput(evlmx2y2, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	lb2.UserInput(evnonex2y2, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	lb2.UserInput(evx, sz, gowid.Focused, gwtest.D)

	c2 = lb2.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"axaaaa",
		"aaaaaa",
		"bxxbbb",
		"bbbbbb",
	}, "\n"), c2.String())

}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
