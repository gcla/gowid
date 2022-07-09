// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package button

import (
	"strings"
	"testing"
	"time"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestButton1(t *testing.T) {
	tw := text.New("click")
	w := New(tw, Options{
		Decoration:       NormalDecoration,
		DoubleClickDelay: 1 * time.Second,
	})

	ct := &gwtest.ButtonTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)
	dct := &gwtest.ButtonTester{Gotit: false}

	w.OnClick(ct)

	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "<click>")

	w.Click(gwtest.D)
	assert.Equal(t, ct.Gotit, true)

	sz := gowid.RenderFlowWith{C: 10}
	cled1 := tcell.NewEventMouse(1, 0, tcell.Button1, 0)
	cleu1 := tcell.NewEventMouse(1, 0, tcell.ButtonNone, 0)

	w.OnDoubleClick(dct)

	ct.Gotit = false
	dct.Gotit = false
	assert.Equal(t, ct.Gotit, false)
	assert.Equal(t, dct.Gotit, false)

	w.UserInput(cled1, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{MouseLeftClicked: true, MouseLastClickedTime: cleu1.When().Add(-500 * time.Millisecond)})
	w.UserInput(cleu1, sz, gowid.Focused, gwtest.D)

	assert.Equal(t, ct.Gotit, false)
	assert.Equal(t, dct.Gotit, true)
	ct.Gotit = false
	dct.Gotit = false

	w.UserInput(cled1, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{MouseLeftClicked: true, MouseLastClickedTime: cleu1.When().Add(-2 * time.Second)})
	w.UserInput(cleu1, sz, gowid.Focused, gwtest.D)

	assert.Equal(t, ct.Gotit, true)
	assert.Equal(t, dct.Gotit, false)

	ct.Gotit = false
	dct.Gotit = false
	assert.Equal(t, ct.Gotit, false)
	w.RemoveOnClick(ct)
	w.Click(gwtest.D)
	assert.Equal(t, ct.Gotit, false)

	gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	gwtest.RenderFlowManyTimes(t, w, 0, 20)
}

func TestButton2(t *testing.T) {
	w1a := text.New("1.2")
	w1 := NewBare(w1a)
	c1 := w1.Render(gowid.RenderFlowWith{C: 3}, gowid.NotSelected, gwtest.D)
	assert.Equal(t, strings.Join([]string{"1.2"}, "\n"), c1.String())
}

func TestCanvas13(t *testing.T) {
	widget1a := text.New("hello world")
	widget1 := New(widget1a)
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 13}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget13 is %v", widget1)
	log.Infof("Canvas13 is %s", canvas1.String())
	res := strings.Join([]string{"<hello world>"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas14(t *testing.T) {
	widget1a := text.New("hello world")
	widget1 := New(widget1a)
	canvas1 := widget1.Render(gowid.RenderBox{C: 13, R: 1}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget14 is %v", widget1)
	log.Infof("Canvas14 is %s", canvas1.String())
	res := strings.Join([]string{"<hello world>"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas15(t *testing.T) {
	widget1a := text.New("helloworld")
	widget1 := New(widget1a)
	canvas1 := widget1.Render(gowid.RenderBox{C: 7, R: 2}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget15 is %v", widget1)
	log.Infof("Canvas15 is %s", canvas1.String())
	res := strings.Join([]string{"<hello>", "<world>"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
