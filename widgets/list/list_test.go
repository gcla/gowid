// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package list

import (
	"fmt"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/fixedadapter"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/tcell"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestCanvasListBox1(t *testing.T) {
	widget1 := text.New("a")
	widget2 := text.New("b")
	widget3 := text.New("c")
	walker := NewSimpleListWalker([]gowid.IWidget{widget1, widget2, widget3})
	lb := New(walker)
	canvas1 := lb.Render(gowid.RenderBox{C: 1, R: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is '%s'", canvas1.String())
	// res := strings.Join([]string{""}, "\n")
	// if res != canvas1.ToString() {
	// 	t.Errorf("Failed")
	// }
}

func TestListBox2(t *testing.T) {
	defer gwtest.ClearTestApp()

	ct := &gwtest.CheckBoxTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	widget1 := checkbox.New(false)
	widget2 := checkbox.New(false)
	widget3 := checkbox.New(false)

	widget2.OnClick(ct)

	walker := NewSimpleListWalker([]gowid.IWidget{
		fixedadapter.New(widget1),
		fixedadapter.New(widget2),
		fixedadapter.New(widget3),
	})

	sz := gowid.RenderBox{C: 6, R: 4}
	lb := New(walker)
	c1 := lb.Render(sz, gowid.NotSelected, gwtest.D)
	assert.Equal(t, c1.String(), "[ ]   \n[ ]   \n[ ]   \n      ")

	evsp := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	evlm := tcell.NewEventMouse(1, 1, tcell.Button1, 0)
	evnone := tcell.NewEventMouse(1, 1, tcell.ButtonNone, 0)

	ct.Gotit = false
	lb.UserInput(evsp, sz, gowid.Focused, gwtest.D)
	c1 = lb.Render(sz, gowid.NotSelected, gwtest.D)
	assert.Equal(t, "[X]   \n[ ]   \n[ ]   \n      ", c1.String())
	assert.Equal(t, ct.Gotit, false)

	ct.Gotit = false
	log.Infof("Sending left mouse down at %d,%d", 1, 1)
	lb.UserInput(evlm, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	log.Infof("Sending left mouse up at %d,%d", 1, 1)
	lb.UserInput(evnone, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	c1 = lb.Render(sz, gowid.NotSelected, gwtest.D)
	assert.Equal(t, "[X]   \n[X]   \n[ ]   \n      ", c1.String())
	assert.Equal(t, ct.Gotit, true)

	ct.Gotit = false
	lb.UserInput(evsp, sz, gowid.Focused, gwtest.D)
	c1 = lb.Render(sz, gowid.NotSelected, gwtest.D)
	assert.Equal(t, "[X]   \n[ ]   \n[ ]   \n      ", c1.String())
	assert.Equal(t, ct.Gotit, true)
}

func TestListBox3(t *testing.T) {
	defer gwtest.ClearTestApp()

	ct := &gwtest.CheckBoxTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	widget1 := checkbox.New(false)
	widget2 := checkbox.New(false)
	widget3 := checkbox.New(false)

	widget1.OnClick(ct)

	walker := NewSimpleListWalker([]gowid.IWidget{
		widget1,
		widget2,
		widget3,
	})

	lb := New(walker)
	c1 := lb.Render(gowid.RenderFixed{}, gowid.NotSelected, gwtest.D)
	assert.Equal(t, c1.String(), "[ ]\n[ ]\n[ ]")

	evlm := tcell.NewEventMouse(1, 0, tcell.Button1, 0)
	evnone := tcell.NewEventMouse(1, 0, tcell.ButtonNone, 0)

	ct.Gotit = false
	lb.UserInput(evlm, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	lb.UserInput(evnone, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

}

type focusWidget struct {
	focus    gowid.IWidget
	notfocus gowid.IWidget
}

func (w *focusWidget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	if focus.Focus {
		return w.focus.RenderSize(size, focus, app)
	} else {
		return w.notfocus.RenderSize(size, focus, app)
	}
}

func (w *focusWidget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	if focus.Focus {
		return gowid.Render(w.focus, size, focus, app)
	} else {
		return gowid.Render(w.notfocus, size, focus, app)
	}
}

func (w *focusWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if focus.Focus {
		return gowid.UserInput(w.focus, ev, size, focus, app)
	} else {
		return gowid.UserInput(w.notfocus, ev, size, focus, app)
	}
}

func (w *focusWidget) Selectable() bool {
	return true
}

func TestListBox4(t *testing.T) {
	defer gwtest.ClearTestApp()

	lws := make([]gowid.IWidget, 5)
	for i := 0; i < len(lws); i++ {
		w := text.New(fmt.Sprintf("%d", i))
		lws[i] = &focusWidget{pile.NewFixed(w, w), w}
	}

	lw := NewSimpleListWalker(lws)
	lb := New(lw)

	c1 := lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0\n0\n1\n2\n3\n4", c1.String())

	evpgup := tcell.NewEventKey(tcell.KeyPgUp, ' ', tcell.ModNone)
	evpgdn := tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone)

	// ct.Gotit = false
	lb.UserInput(gwtest.CursorDown(), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0\n1\n1\n2\n3\n4", c1.String())

	for i := 0; i < len(lws); i++ {
		res := lb.UserInput(gwtest.CursorDown(), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
		if i < 3 { // TODO: bug -
			assert.Equal(t, true, res)
		} else {
			assert.Equal(t, false, res)
		}
	}

	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0\n1\n2\n3\n4\n4", c1.String())

	lb.UserInput(evpgup, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0\n0\n1\n2\n3\n4", c1.String())

	lb.UserInput(evpgdn, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0\n1\n2\n3\n4\n4", c1.String())
}

func TestListBox5(t *testing.T) {
	defer gwtest.ClearTestApp()

	lws := make([]gowid.IWidget, 0)
	for i := 0; i < 5; i++ {
		lws = append(lws, &text.Widget1{i})
		lws = append(lws, text.New("-"))
	}

	lw := NewSimpleListWalker(lws)
	lb := New(lw)

	c1 := lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0f\n- \n1 \n- \n2 \n- \n3 \n- \n4 \n- ", c1.String())

	// ct.Gotit = false
	lb.UserInput(gwtest.CursorDown(), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	lb.UserInput(gwtest.CursorDown(), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0 \n- \n1 \n- \n2f\n- \n3 \n- \n4 \n- ", c1.String())

	// Click on a selectable widget, should move
	lb.UserInput(gwtest.ClickAt(1, 0), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	lb.UserInput(gwtest.ClickUpAt(1, 0), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0f\n- \n1 \n- \n2 \n- \n3 \n- \n4 \n- ", c1.String())

	// Click on an unselectable widget, should preserve current state - so no change
	lb.UserInput(gwtest.ClickAt(1, 3), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	lb.UserInput(gwtest.ClickUpAt(1, 3), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0f\n- \n1 \n- \n2 \n- \n3 \n- \n4 \n- ", c1.String())
}

func TestEmptyListBox1(t *testing.T) {
	defer gwtest.ClearTestApp()

	lws := make([]gowid.IWidget, 0)
	lw := NewSimpleListWalker(lws)
	lb := New(lw)

	c1 := lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "", c1.String())

	// ct.Gotit = false
	lb.UserInput(gwtest.CursorDown(), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "", c1.String())

	f := lw.Focus()
	assert.Equal(t, nil, lw.At(f))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
