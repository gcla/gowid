// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package padding

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

type renderRatioUpTo struct {
	gowid.RenderWithRatio
	max int
}

func (s renderRatioUpTo) MaxUnits() int {
	return s.max
}

func ratioupto(w float64, max int) renderRatioUpTo {
	return renderRatioUpTo{gowid.RenderWithRatio{R: w}, max}
}

func TestPadding1(t *testing.T) {
	var w gowid.IWidget
	var c gowid.ICanvas

	w = New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithUnits{U: 2}, gowid.HAlignMiddle{}, gowid.RenderWithUnits{U: 2})
	c = w.Render(gowid.RenderBox{C: 4, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, "    \n xx \n xx \n    ", c.String())

	w = New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithUnits{U: 2}, gowid.HAlignMiddle{}, gowid.RenderWithRatio{R: 0.5})
	c = w.Render(gowid.RenderBox{C: 4, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, "    \n xx \n xx \n    ", c.String())

	w = New(fill.New('x'), gowid.VAlignMiddle{}, ratioupto(0.5, 1), gowid.HAlignMiddle{}, gowid.RenderWithUnits{U: 2})
	c = w.Render(gowid.RenderBox{C: 4, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, "    \n xx \n    \n    ", c.String())

	w = New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithUnits{U: 2}, gowid.HAlignMiddle{}, ratioupto(0.5, 1))
	c = w.Render(gowid.RenderBox{C: 4, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, "    \n  x \n  x \n    ", c.String())

	w = New(text.New("foo"), gowid.VAlignMiddle{}, gowid.RenderFixed{}, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	c = w.Render(gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, "     \n foo \n     ", c.String())

	w = New(framed.New(text.New("foo")), gowid.VAlignMiddle{}, gowid.RenderFixed{}, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	c = w.Render(gowid.RenderBox{C: 7, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, "       \n ----- \n |foo| \n ----- \n       ", c.String())

	w = New(framed.New(text.New("foo")), gowid.VAlignTop{}, gowid.RenderFixed{}, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	c = w.Render(gowid.RenderBox{C: 7, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, " ----- \n |foo| \n ----- \n       \n       ", c.String())

	w = New(framed.New(text.New("foo")), gowid.VAlignBottom{}, gowid.RenderFixed{}, gowid.HAlignRight{}, gowid.RenderFixed{})
	c = w.Render(gowid.RenderBox{C: 7, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, "       \n       \n  -----\n  |foo|\n  -----", c.String())

	ct := &gwtest.ButtonTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	btn := button.NewBare(text.New("foo"))
	btn.OnClick(ct)

	w = New(btn, gowid.VAlignMiddle{}, gowid.RenderFixed{}, gowid.HAlignMiddle{}, gowid.RenderFixed{})
	c = w.Render(gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, "     \n foo \n     ", c.String())

	btn.Click(gwtest.D)
	assert.Equal(t, ct.Gotit, true)
	ct.Gotit = false

	ev31 := tcell.NewEventMouse(3, 1, tcell.Button1, 0)
	evnone31 := tcell.NewEventMouse(3, 1, tcell.ButtonNone, 0)

	w.UserInput(ev31, gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w.UserInput(evnone31, gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

	// w2 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.5})
	// c2 := w2.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c2.String(), "   \nxxx\nxxx\n   ")

	// w3 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.75})
	// c3 := w3.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c3.String(), "xxx\nxxx\nxxx\n   ")

	// w4 := New(fill.New('x'), gowid.VAlignTop{}, gowid.RenderWithRatio{0.75})
	// c4 := w4.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c4.String(), "xxx\nxxx\nxxx\n   ")

	// w5 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.8})
	// c5 := w5.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c5.String(), "xxx\nxxx\nxxx\n   ")

	// w6 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.88})
	// c6 := w6.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c6.String(), "xxx\nxxx\nxxx\nxxx")

	// w7 := New(fill.New('x'), gowid.VAlignTop{1}, gowid.RenderWithUnits{U: 3})
	// c7 := w7.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c7.String(), "   \nxxx\nxxx\nxxx")

	// w8 := New(fill.New('x'), gowid.VAlignTop{2}, gowid.RenderWithUnits{U: 3})
	// c8 := w8.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c8.String(), "   \n   \nxxx\nxxx")

	// w9 := New(fill.New('x'), gowid.VAlignTop{2}, gowid.RenderWithUnits{U: 3})
	// c9 := w9.Render(gowid.RenderBox{C: 3, R: 3}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c9.String(), "   \n   \nxxx")

	// w10 := New(fill.New('x'), gowid.VAlignTop{2}, gowid.RenderWithUnits{U: 4})
	// c10 := w10.Render(gowid.RenderBox{C: 3, R: 8}, gowid.Focused, gwtest.D)
	// assert.Equal(t, c10.String(), "   \n   \nxxx\nxxx\nxxx\nxxx\n   \n   ")

	// for _, w := range []gowid.IWidget{w1, w2, w3, w4, w5, w6, w7, w8, w9, w10} {
	// 	gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	// }
	// for _, w := range []gowid.IWidget{w1, w7, w8, w9, w10} {
	// 	gwtest.RenderFlowManyTimes(t, w, 0, 10)
	// }

}

// func TestCanvas17(t *testing.T) {
// 	widget1i := text.New("line 1line 2line 3")
// 	widget1 := NewBox(widget1i, 2)
// 	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
// 	log.Infof("Widget17 is %v", widget1)
// 	log.Infof("Canvas17 is %s", canvas1.String())
// 	res := strings.Join([]string{"line 1", "line 2"}, "\n")
// 	if res != canvas1.String() {
// 		t.Errorf("Failed")
// 	}
// }

// func TestCanvas18(t *testing.T) {
// 	widget1i := text.New("line 1 line 2 line 3")
// 	widget1 := NewBox(widget1i, 5)
// 	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
// 	log.Infof("Widget18 is %v", widget1)
// 	log.Infof("Canvas18 is %s", canvas1.String())
// 	res := strings.Join([]string{"line 1", " line ", "2 line", " 3    ", "      "}, "\n")
// 	if res != canvas1.String() {
// 		t.Errorf("Failed")
// 	}

// 	gwtest.RenderBoxManyTimes(t, widget1, 0, 10, 0, 10)
// 	gwtest.RenderFlowManyTimes(t, widget1, 0, 10)
// 	gwtest.RenderFixedDoesNotPanic(t, widget1)
// }

// func TestCheckbox3(t *testing.T) {

// 	ct := &gwtest.CheckBoxTester{Gotit: false}
// 	assert.Equal(t, ct.Gotit, false)

// 	w := checkbox.New(false)
// 	w.OnClick(ct)

// 	ev := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)

// 	w.UserInput(ev, gowid.RenderFixed{}, gowid.Focused, gwtest.D)

// 	assert.Equal(t, ct.Gotit, true)

// 	ct.Gotit = false
// 	assert.Equal(t, ct.Gotit, false)

// 	evlmx1y0 := tcell.NewEventMouse(1, 0, tcell.Button1, 0)
// 	evnonex1y0 := tcell.NewEventMouse(1, 0, tcell.ButtonNone, 0)

// 	sz := gowid.RenderBox{C: 5, R: 3}
// 	w2 := New(w, gowid.VAlignTop{}, gowid.RenderFixed{})
// 	ct.Gotit = false
// 	assert.Equal(t, ct.Gotit, false)
// 	w2.UserInput(evlmx1y0, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
// 	w2.UserInput(evnonex1y0, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{})
// 	assert.Equal(t, ct.Gotit, true)

// 	w3 := New(w, gowid.VAlignBottom{}, gowid.RenderFixed{})
// 	ct.Gotit = false
// 	assert.Equal(t, ct.Gotit, false)

// 	w3.UserInput(evlmx1y0, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
// 	w3.UserInput(evnonex1y0, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{})
// 	assert.Equal(t, ct.Gotit, false)

// 	evlmx1y2 := tcell.NewEventMouse(1, 2, tcell.ButtonNone, 0)

// 	w3.UserInput(evlmx1y2, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
// 	w3.UserInput(evnonex1y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{})
// 	assert.Equal(t, ct.Gotit, true)

// 	w4 := New(w, gowid.VAlignTop{1}, gowid.RenderFixed{})
// 	ct.Gotit = false
// 	assert.Equal(t, ct.Gotit, false)

// 	w4.UserInput(evlmx1y0, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
// 	w4.UserInput(evnonex1y0, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{})
// 	assert.Equal(t, ct.Gotit, false)

// 	evlmx1y1 := tcell.NewEventMouse(1, 1, tcell.Button1, 0)
// 	evnonex1y1 := tcell.NewEventMouse(1, 1, tcell.ButtonNone, 0)

// 	w4.UserInput(evlmx1y1, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
// 	w4.UserInput(evnonex1y1, sz, gowid.Focused, gwtest.D)
// 	gwtest.D.SetLastMouseState(gowid.MouseState{})
// 	assert.Equal(t, ct.Gotit, true)

// }

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
