// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package hpadding

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCanvas26(t *testing.T) {
	widget1i := text.New("1234567890")
	widget1 := New(widget1i, gowid.HAlignLeft{}, gowid.RenderWithUnits{U: 5})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is '%s'", canvas1.String())
	res := strings.Join([]string{"12345   ", "67890   "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestHorizontal1(t *testing.T) {
	w := text.New("abcde")

	h := New(w, gowid.HAlignLeft{}, gowid.RenderFixed{})
	c1 := h.Render(gowid.RenderFlowWith{C: 10}, gowid.Focused, gwtest.D)
	assert.Equal(t, "abcde     ", c1.String())

	h2 := New(w, gowid.HAlignRight{}, gowid.RenderFixed{})
	c2 := h2.Render(gowid.RenderFlowWith{C: 10}, gowid.Focused, gwtest.D)
	assert.Equal(t, "     abcde", c2.String())

	h3 := New(w, gowid.HAlignRight{}, gowid.RenderWithUnits{U: 10})
	c3 := h3.Render(gowid.RenderFlowWith{C: 20}, gowid.Focused, gwtest.D)
	assert.Equal(t, "          abcde     ", c3.String())

	h4 := New(w, gowid.HAlignMiddle{}, gowid.RenderWithUnits{U: 10})
	c4 := h4.Render(gowid.RenderFlowWith{C: 20}, gowid.Focused, gwtest.D)
	assert.Equal(t, "     abcde          ", c4.String())

	h5 := New(w, gowid.HAlignLeft{}, gowid.RenderWithUnits{U: 3})
	c5 := h5.Render(gowid.RenderFlowWith{C: 10}, gowid.Focused, gwtest.D)
	assert.Equal(t, "abc       \nde        ", c5.String())

}

func TestHorizontal2(t *testing.T) {
	w := text.New("abcde")

	var c gowid.ICanvas = w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde")

	c = w.Render(gowid.RenderFlowWith{C: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde")

	c = w.Render(gowid.RenderBox{C: 5, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde")

	w2 := New(w, gowid.HAlignMiddle{}, gowid.RenderFixed{})

	c = w2.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde")

	c = w2.Render(gowid.RenderFlowWith{C: 7}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), " abcde ")

	c = w2.Render(gowid.RenderBox{C: 7, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), " abcde \n       ")

	w3 := New(w, gowid.HAlignRight{}, gowid.RenderFlowWith{C: 7})

	c = w3.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	// Will render with space on the right, because subwidget doesn't inherit alignment
	assert.Equal(t, c.String(), "abcde  ")

	c = w3.Render(gowid.RenderFlowWith{C: 7}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde  ")

	c = w3.Render(gowid.RenderBox{C: 7, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde  \n       ")

	c = w3.Render(gowid.RenderBox{C: 7, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde  \n       ")

	w4 := New(w, gowid.HAlignRight{}, gowid.RenderFlowWith{C: 3})

	c = w4.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abc\nde ")

	c = w4.Render(gowid.RenderFlowWith{C: 7}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "    abc\n    de ")

	c = w4.Render(gowid.RenderBox{C: 7, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "    abc")

	c = w4.Render(gowid.RenderBox{C: 7, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "    abc\n    de \n       ")

	w5 := New(w, gowid.HAlignLeft{}, gowid.RenderWithUnits{U: 7})

	c = w5.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde  ")

	c = w5.Render(gowid.RenderFlowWith{C: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abc\nde ")

	c = w5.Render(gowid.RenderBox{C: 6, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde ")

	c = w5.Render(gowid.RenderBox{C: 3, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abc\nde \n   ")

	c = w5.Render(gowid.RenderBox{C: 7, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abcde  \n       \n       ")

	w6 := New(w, gowid.HAlignLeft{}, gowid.RenderWithUnits{U: 3})

	c = w6.Render(gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), "abc  \nde   \n     ")

}

func TestCheckbox2(t *testing.T) {

	ct := &gwtest.CheckBoxTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	w := checkbox.New(false)
	w.OnClick(ct)

	ev := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)

	w.UserInput(ev, gowid.RenderFixed{}, gowid.Focused, gwtest.D)

	assert.Equal(t, ct.Gotit, true)

	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)

	evlmx1y0 := tcell.NewEventMouse(1, 0, tcell.Button1, 0)
	evnonex1y0 := tcell.NewEventMouse(1, 0, tcell.ButtonNone, 0)

	w.UserInput(evlmx1y0, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w.UserInput(evnonex1y0, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

	w2 := New(w, gowid.HAlignLeft{}, gowid.RenderFixed{})
	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)
	w2.UserInput(evlmx1y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w2.UserInput(evnonex1y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

	w3 := New(w, gowid.HAlignLeft{Margin: 4}, gowid.RenderFixed{})
	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)

	w3.UserInput(evlmx1y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	w3.UserInput(evnonex1y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, ct.Gotit, false)

	evlmx5y0 := tcell.NewEventMouse(5, 0, tcell.ButtonNone, 0)
	evnonex5y0 := tcell.NewEventMouse(5, 0, tcell.ButtonNone, 0)

	w3.UserInput(evlmx5y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w3.UserInput(evnonex5y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)
}

func TestHorizontalPadding1(t *testing.T) {
	w1 := New(fill.New('x'), gowid.HAlignLeft{}, gowid.RenderWithUnits{U: 1})
	c1 := w1.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "x  \nx  \nx  \nx  ")

	w2 := New(fill.New('x'), gowid.HAlignLeft{}, gowid.RenderWithUnits{U: 2})
	c2 := w2.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "xx \nxx \nxx \nxx ")

	w3 := New(fill.New('x'), gowid.HAlignMiddle{}, gowid.RenderWithUnits{U: 1})
	c3 := w3.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c3.String(), " x \n x \n x \n x ")

	w4 := New(fill.New('x'), gowid.HAlignRight{}, gowid.RenderWithUnits{U: 1})
	c4 := w4.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c4.String(), "  x\n  x\n  x\n  x")

	w5 := New(fill.New('x'), gowid.HAlignRight{}, gowid.RenderWithUnits{U: 2})
	c5 := w5.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c5.String(), " xx\n xx\n xx\n xx")

	w6 := New(fill.New('x'), gowid.HAlignRight{}, gowid.RenderWithRatio{0.3})
	c6 := w6.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c6.String(), "  x\n  x\n  x\n  x")

	w7 := New(fill.New('x'), gowid.HAlignRight{}, gowid.RenderWithRatio{0.6})
	c7 := w7.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c7.String(), " xx\n xx\n xx\n xx")

	w8 := New(fill.New('x'), gowid.HAlignRight{}, gowid.RenderWithRatio{8.1})
	c8 := w8.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c8.String(), "xxx\nxxx\nxxx\nxxx")

	w9 := New(fill.New('x'), gowid.HAlignLeft{Margin: 1}, gowid.RenderWithUnits{U: 3})
	c9 := w9.Render(gowid.RenderBox{C: 5, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c9.String(), " xxx \n xxx ")

	w10 := New(fill.New('x'), gowid.HAlignLeft{Margin: 1}, gowid.RenderWithUnits{U: 6})
	c10 := w10.Render(gowid.RenderBox{C: 5, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c10.String(), " xxxx\n xxxx")

	w11 := New(fill.New('x'), gowid.HAlignLeft{Margin: 1}, gowid.RenderWithUnits{U: 4})
	c11 := w11.Render(gowid.RenderBox{C: 5, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c11.String(), " xxxx\n xxxx")

	w12 := New(checkbox.New(false), gowid.HAlignLeft{Margin: 4}, gowid.RenderFixed{})
	c12 := w12.Render(gowid.RenderBox{C: 8, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, c12.String(), "    [ ] ")

	w13 := New(checkbox.New(false), gowid.HAlignLeft{Margin: 7}, gowid.RenderFixed{})
	c13 := w13.Render(gowid.RenderBox{C: 8, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, c13.String(), "     [ ]")

	for _, w := range []gowid.IWidget{w1, w2, w3, w4, w5, w6, w7, w8, w9, w10, w11, w12, w13} {
		gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	}
	for _, w := range []gowid.IWidget{w1, w2, w3, w4, w5, w6, w7, w8, w9, w10, w11, w12, w13} {
		gwtest.RenderFlowManyTimes(t, w, 0, 10)
	}

}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
