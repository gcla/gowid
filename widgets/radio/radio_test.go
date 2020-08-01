// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package radio

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestRadioButton1(t *testing.T) {
	rbgroup := make([]IWidget, 0)
	rb1 := New(&rbgroup)
	rb2 := New(&rbgroup)
	rb3 := New(&rbgroup)

	assert.Equal(t, rb1.IsChecked(), true)
	assert.Equal(t, rb2.IsChecked(), false)
	assert.Equal(t, rb3.IsChecked(), false)

	c1 := rb1.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "(X)")
	c2 := rb2.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "( )")

	ct1 := &RadioButtonTester{State: true}
	assert.Equal(t, ct1.State, true)
	rb1.OnClick(ct1)
	ct2 := &RadioButtonTester{State: false}
	assert.Equal(t, ct2.State, false)
	rb2.OnClick(ct2)

	rb2.Click(gwtest.D)
	assert.Equal(t, rb1.IsChecked(), false)
	assert.Equal(t, ct1.State, false)
	assert.Equal(t, rb2.IsChecked(), true)
	assert.Equal(t, ct2.State, true)
	assert.Equal(t, rb3.IsChecked(), false)

	fixed := gowid.RenderFixed{}

	ccols := []gowid.IContainerWidget{
		&gowid.ContainerWidget{rb1, fixed},
		&gowid.ContainerWidget{rb2, fixed},
		&gowid.ContainerWidget{rb3, fixed},
	}

	cols := columns.New(ccols)

	c3 := cols.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c3.String(), "( )(X)( )")
	cpos := c3.CursorCoords()
	cx, cy := cpos.X, cpos.Y
	assert.Equal(t, c3.CursorEnabled(), true)
	assert.Equal(t, cx, 1)
	assert.Equal(t, cy, 0)
	assert.Equal(t, c3.String(), "( )(X)( )")

	ev := gwtest.CursorRight()

	cols.UserInput(ev, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	c3 = cols.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	cx = c3.CursorCoords().X

	assert.Equal(t, cx, 4)

	gwtest.RenderBoxManyTimes(t, cols, 0, 20, 0, 20)
	gwtest.RenderFlowManyTimes(t, cols, 0, 20)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
