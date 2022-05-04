// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package list

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/disable"
	"github.com/gcla/gowid/widgets/fixedadapter"
	"github.com/gcla/gowid/widgets/isselected"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
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
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false, time.Now()})
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
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false, time.Now()})
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
		return w.focus.Render(size, focus, app)
	} else {
		return w.notfocus.Render(size, focus, app)
	}
}

func (w *focusWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if focus.Focus {
		return w.focus.UserInput(ev, size, focus, app)
	} else {
		return w.notfocus.UserInput(ev, size, focus, app)
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
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false, time.Now()})
	lb.UserInput(gwtest.ClickUpAt(1, 0), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false, time.Now()})
	c1 = lb.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "0f\n- \n1 \n- \n2 \n- \n3 \n- \n4 \n- ", c1.String())

	// Click on an unselectable widget, should preserve current state - so no change
	lb.UserInput(gwtest.ClickAt(1, 3), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false, time.Now()})
	lb.UserInput(gwtest.ClickUpAt(1, 3), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false, time.Now()})
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

func TestDisabled1(t *testing.T) {
	defer gwtest.ClearTestApp()

	fixed := gowid.RenderFixed{}

	foc := make([]gowid.IWidget, 0)
	notfoc := make([]gowid.IWidget, 0)

	for i := 0; i < 3; i++ {
		txtf := text.New(fmt.Sprintf("f%d", i))
		txtn := text.New(fmt.Sprintf("a%d", i))
		foc = append(foc, txtf)
		notfoc = append(notfoc, txtn)
	}

	lws := make([]gowid.IWidget, 0)
	for i := 0; i < len(foc); i++ {
		lws = append(lws,
			selectable.New(
				isselected.New(notfoc[i], nil, foc[i]),
			),
		)
	}

	lw := NewSimpleListWalker(lws)
	lb := New(lw)

	c1 := lb.Render(fixed, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"f0",
		"a1",
		"a2",
	}, "\n"), c1.String())

	lb.UserInput(gwtest.CursorDown(), fixed, gowid.Focused, gwtest.D)
	c1 = lb.Render(fixed, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"a0",
		"f1",
		"a2",
	}, "\n"), c1.String())

	dis := disable.New(text.New("dd"))

	lws2 := append(lws[0:2], append([]gowid.IWidget{dis}, lws[2:]...)...)
	lw2 := NewSimpleListWalker(lws2)
	lb.SetWalker(lw2, gwtest.D)
	lb.UserInput(gwtest.CursorDown(), fixed, gowid.Focused, gwtest.D)

	c1 = lb.Render(fixed, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"a0",
		"f1",
		"dd",
		"a2",
	}, "\n"), c1.String())

	clickat := func(x, y int) {
		evlm := tcell.NewEventMouse(x, y, tcell.Button1, 0)
		evnone := tcell.NewEventMouse(x, y, tcell.ButtonNone, 0)

		lb.UserInput(evlm, fixed, gowid.Focused, gwtest.D)
		gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false, time.Now()})

		lb.UserInput(evnone, fixed, gowid.Focused, gwtest.D)
		gwtest.D.SetLastMouseState(gowid.MouseState{})
	}

	clickat(1, 1)

	c1 = lb.Render(fixed, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"a0",
		"f1",
		"dd",
		"a2",
	}, "\n"), c1.String())

	clickat(1, 2)

	c1 = lb.Render(fixed, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"a0",
		"f1",
		"dd",
		"a2",
	}, "\n"), c1.String())

	clickat(1, 3)

	c1 = lb.Render(fixed, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join([]string{
		"a0",
		"a1",
		"dd",
		"f2",
	}, "\n"), c1.String())

	fpos := -1

	lb.OnFocusChanged(gowid.WidgetCallback{Name: "cb", WidgetChangedFunction: func(app gowid.IApp, w gowid.IWidget) {
		fpos = lb.Walker().Focus().(ListPos).ToInt()
	}})

	clickat(1, 1)
	assert.Equal(t, 1, fpos)
	clickat(1, 2)
	assert.Equal(t, 1, fpos)
	clickat(1, 3)
	assert.Equal(t, 3, fpos)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
