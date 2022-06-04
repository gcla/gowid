// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package keypress

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/text"
	"github.com/stretchr/testify/assert"
)

//======================================================================

type KeyPressTester struct {
	Gotit bool
}

func (f *KeyPressTester) Changed(gowid.IApp, gowid.IWidget, ...interface{}) {
	f.Gotit = true
}

func (f *KeyPressTester) ID() interface{} { return "foo" }

//======================================================================

func TestKey1(t *testing.T) {
	tw := text.New("hitq")
	w := New(tw, Options{
		Keys: []gowid.IKey{gowid.MakeKey('q')},
	})

	ct := &KeyPressTester{Gotit: false}
	assert.Equal(t, ct.Gotit, false)

	w.OnKeyPress(ct)

	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "hitq")

	UserInput(w, gwtest.KeyEvent('q'), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	// w.Click(gwtest.D)
	assert.Equal(t, ct.Gotit, true)
	ct.Gotit = false

	// most widgets will funnel user input to the "focus" widget - so this should
	// never be necessary. But if you built something that passed input to both
	// widgets when only one had focus, you'd presumably want this to fire
	UserInput(w, gwtest.KeyEvent('q'), gowid.RenderFixed{}, gowid.NotSelected, gwtest.D)
	assert.Equal(t, ct.Gotit, true)
	ct.Gotit = false

	UserInput(w, gwtest.KeyEvent('r'), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, ct.Gotit, false)
	ct.Gotit = false

	cbCalled := false
	w.OnKeyPress(WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget, k gowid.IKey) {
		assert.Equal(t, true, gowid.KeysEqual(k, gowid.MakeKey('q')))
		cbCalled = true
	}})

	UserInput(w, gwtest.KeyEvent('q'), gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, ct.Gotit, true)
	assert.Equal(t, true, cbCalled)
	ct.Gotit = false

	// assert.Equal(t, ct.Gotit, false)
	// w.RemoveOnClick(ct)
	// w.Click(gwtest.D)
	// assert.Equal(t, ct.Gotit, false)

	gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	gwtest.RenderFlowManyTimes(t, w, 0, 20)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
