// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package checkbox

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/text"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestButton1(t *testing.T) {
	tw := text.New("click")
	w := button.New(tw)

	ct := &gwtest.ButtonTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	w.OnClick(ct)

	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "<click>")

	w.Click(gwtest.D)
	assert.Equal(t, ct.Gotit, true)

	ct.Gotit = false
	assert.Equal(t, ct.Gotit, false)
	w.RemoveOnClick(ct)
	w.Click(gwtest.D)
	assert.Equal(t, ct.Gotit, false)

}

func TestCheckbox1(t *testing.T) {
	w := New(false)

	ct := &gwtest.CheckBoxTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	w.OnClick(ct)

	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "[ ]")
	assert.Equal(t, w.IsChecked(), false)

	w.SetChecked(gwtest.D, true)
	assert.Equal(t, w.IsChecked(), true)
	assert.Equal(t, ct.Gotit, true)
	ct.Gotit = false
	// WRONG Shouldn't issue another callback since state didn't change
	w.SetChecked(gwtest.D, true)
	assert.Equal(t, ct.Gotit, true)
	assert.Equal(t, w.IsChecked(), true)

	c2 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "[X]")
	assert.Equal(t, w.IsChecked(), true)

	assert.Panics(t, func() {
		w.Render(gowid.RenderFlowWith{C: 5}, gowid.Focused, gwtest.D)
	})

	assert.Panics(t, func() {
		w.Render(gowid.RenderBox{C: 3, R: 1}, gowid.Focused, gwtest.D)
	})

}

var (
	cb1 int
	cb2 int
)

func testCallback1(app gowid.IApp, w gowid.IWidget) {
	cb1++
}

func testCallback2(app gowid.IApp, w gowid.IWidget) {
	cb2++
}

func TestCallbacks(t *testing.T) {
	cbs := gowid.NewCallbacks()
	assert.Equal(t, cb1, 0)
	assert.Equal(t, cb2, 0)
	gowid.AddWidgetCallback(cbs, "test", gowid.WidgetCallback{"cb1", testCallback1})
	dummy := New(false)
	gowid.RunWidgetCallbacks(cbs, "test", gwtest.D, dummy)
	assert.Equal(t, cb1, 1)
	gowid.RemoveWidgetCallback(cbs, "test", gowid.CallbackID{Name: "cb1"})
	gowid.RunWidgetCallbacks(cbs, "test", gwtest.D, dummy)
	assert.Equal(t, cb1, 1)
	cb1 = 0
	assert.Equal(t, cb1, 0)
	gowid.AddWidgetCallback(cbs, "test", gowid.WidgetCallback{123, testCallback1})
	gowid.AddWidgetCallback(cbs, "test", gowid.WidgetCallback{123, testCallback2})
	gowid.RunWidgetCallbacks(cbs, "test2", gwtest.D, dummy)
	assert.Equal(t, cb1, 0)
	assert.Equal(t, cb2, 0)
	gowid.RunWidgetCallbacks(cbs, "test", gwtest.D, dummy)
	assert.Equal(t, cb1, 1)
	assert.Equal(t, cb2, 1)
	gowid.RemoveWidgetCallback(cbs, "test2", gowid.CallbackID{Name: 123})
	gowid.RunWidgetCallbacks(cbs, "test", gwtest.D, dummy)
	assert.Equal(t, cb1, 2)
	assert.Equal(t, cb2, 2)
	gowid.RemoveWidgetCallback(cbs, "test", gowid.CallbackID{Name: "xx2"})
	gowid.RunWidgetCallbacks(cbs, "test", gwtest.D, dummy)
	assert.Equal(t, cb1, 3)
	assert.Equal(t, cb2, 3)
	gowid.RemoveWidgetCallback(cbs, "test", gowid.CallbackID{Name: 123})
	gowid.RunWidgetCallbacks(cbs, "test", gwtest.D, dummy)
	assert.Equal(t, cb1, 3)
	assert.Equal(t, cb2, 3)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
