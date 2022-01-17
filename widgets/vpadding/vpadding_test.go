// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package vpadding

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestVerticalPadding1(t *testing.T) {
	w1 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithUnits{U: 2})
	c1 := w1.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "   \nxxx\nxxx\n   ")

	w2 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.5})
	c2 := w2.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "   \nxxx\nxxx\n   ")

	w3 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.75})
	c3 := w3.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c3.String(), "xxx\nxxx\nxxx\n   ")

	w4 := New(fill.New('x'), gowid.VAlignTop{}, gowid.RenderWithRatio{0.75})
	c4 := w4.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c4.String(), "xxx\nxxx\nxxx\n   ")

	w5 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.8})
	c5 := w5.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c5.String(), "xxx\nxxx\nxxx\n   ")

	w6 := New(fill.New('x'), gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.88})
	c6 := w6.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c6.String(), "xxx\nxxx\nxxx\nxxx")

	w7 := New(fill.New('x'), gowid.VAlignTop{1}, gowid.RenderWithUnits{U: 3})
	c7 := w7.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c7.String(), "   \nxxx\nxxx\nxxx")

	w8 := New(fill.New('x'), gowid.VAlignTop{2}, gowid.RenderWithUnits{U: 3})
	c8 := w8.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c8.String(), "   \n   \nxxx\nxxx")

	w9 := New(fill.New('x'), gowid.VAlignTop{2}, gowid.RenderWithUnits{U: 3})
	c9 := w9.Render(gowid.RenderBox{C: 3, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c9.String(), "   \n   \nxxx")

	w10 := New(fill.New('x'), gowid.VAlignTop{2}, gowid.RenderWithUnits{U: 4})
	c10 := w10.Render(gowid.RenderBox{C: 3, R: 8}, gowid.Focused, gwtest.D)
	assert.Equal(t, c10.String(), "   \n   \nxxx\nxxx\nxxx\nxxx\n   \n   ")

	w11 := New(fill.New('x'), gowid.VAlignBottom{}, gowid.RenderWithUnits{U: 3})
	c11 := w11.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c11.String(), "   \nxxx\nxxx\nxxx")

	p1 := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			IWidget: text.New("111"),
			D:       gowid.RenderFixed{},
		},
		&gowid.ContainerWidget{
			IWidget: text.New("222"),
			D:       gowid.RenderFixed{},
		},
		&gowid.ContainerWidget{
			IWidget: text.New("333"),
			D:       gowid.RenderFixed{},
		},
		&gowid.ContainerWidget{
			IWidget: text.New("444"),
			D:       gowid.RenderFixed{},
		},
	})

	w12 := New(p1, gowid.VAlignBottom{}, gowid.RenderWithUnits{U: 3})
	c12 := w12.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c12.String(), "   \n111\n222\n333")

	w13 := New(p1, gowid.VAlignBottom{}, gowid.RenderWithUnits{U: 3})
	c13 := w13.Render(gowid.RenderBox{C: 3, R: 2}, gowid.Focused, gwtest.D)
	assert.Equal(t, c13.String(), "111\n222")

	for _, w := range []gowid.IWidget{w1, w2, w3, w4, w5, w6, w7, w8, w9, w10} {
		gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	}
	for _, w := range []gowid.IWidget{w1, w7, w8, w9, w10} {
		gwtest.RenderFlowManyTimes(t, w, 0, 10)
	}

}

func TestCanvas17(t *testing.T) {
	widget1i := text.New("line 1line 2line 3")
	widget1 := NewBox(widget1i, 2)
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget17 is %v", widget1)
	log.Infof("Canvas17 is %s", canvas1.String())
	res := strings.Join([]string{"line 1", "line 2"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas18(t *testing.T) {
	widget1i := text.New("line 1 line 2 line 3")
	widget1 := NewBox(widget1i, 5)
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget18 is %v", widget1)
	log.Infof("Canvas18 is %s", canvas1.String())
	res := strings.Join([]string{"line 1", " line ", "2 line", " 3    ", "      "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}

	gwtest.RenderBoxManyTimes(t, widget1, 0, 10, 0, 10)
	gwtest.RenderFlowManyTimes(t, widget1, 0, 10)
	gwtest.RenderFixedDoesNotPanic(t, widget1)
}

func TestCheckbox3(t *testing.T) {

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

	sz := gowid.RenderBox{C: 5, R: 3}
	w2 := New(w, gowid.VAlignTop{}, gowid.RenderFixed{})
	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)
	w2.UserInput(evlmx1y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w2.UserInput(evnonex1y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

	w3 := New(w, gowid.VAlignBottom{}, gowid.RenderFixed{})
	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)

	w3.UserInput(evlmx1y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w3.UserInput(evnonex1y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, false)

	evlmx1y2 := tcell.NewEventMouse(1, 2, tcell.ButtonNone, 0)

	w3.UserInput(evlmx1y2, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w3.UserInput(evnonex1y0, gowid.RenderBox{C: 10, R: 1}, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

	w4 := New(w, gowid.VAlignTop{1}, gowid.RenderFixed{})
	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)

	w4.UserInput(evlmx1y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w4.UserInput(evnonex1y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, false)

	evlmx1y1 := tcell.NewEventMouse(1, 1, tcell.Button1, 0)
	evnonex1y1 := tcell.NewEventMouse(1, 1, tcell.ButtonNone, 0)

	w4.UserInput(evlmx1y1, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w4.UserInput(evnonex1y1, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{})
	assert.Equal(t, ct.Gotit, true)

}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
