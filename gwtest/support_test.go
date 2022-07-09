// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gwtest

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	fx := gowid.RenderFixed{}
	t1 := edit.New(edit.Options{Text: "foo"})
	ct1 := &gowid.ContainerWidget{IWidget: t1, D: fx}

	c1 := columns.New([]gowid.IContainerWidget{ct1, ct1, ct1})
	cc1 := &gowid.ContainerWidget{IWidget: c1, D: fx}
	c1.SetFocus(D, 2)
	assert.Equal(t, 2, c1.Focus())

	c2 := columns.New([]gowid.IContainerWidget{ct1, ct1})
	c2.SetFocus(D, 1)
	assert.Equal(t, 1, c2.Focus())

	w3 := styled.New(c2, gowid.MakeForeground(gowid.ColorBlack))
	cw3 := &gowid.ContainerWidget{IWidget: w3, D: fx}

	p1 := pile.New([]gowid.IContainerWidget{ct1, cc1, cw3})
	assert.Equal(t, 0, p1.Focus())
	assert.Equal(t, gowid.FocusPath(p1), []interface{}{0})

	p1.SetFocus(D, 1)
	assert.Equal(t, 1, p1.Focus())
	assert.Equal(t, gowid.FocusPath(p1), []interface{}{1, 2})

	p1.SetFocus(D, 2)
	assert.Equal(t, 2, p1.Focus())
	assert.Equal(t, gowid.FocusPath(p1), []interface{}{2, 1})

	r := gowid.SetFocusPath(p1, []interface{}{0, 4, 5}, D)
	assert.Equal(t, false, r.Succeeded)
	assert.Equal(t, 1, r.FailedLevel)
	assert.Equal(t, 0, p1.Focus())

	p1.SetFocus(D, 2)
	assert.Equal(t, 2, p1.Focus())

	r = gowid.SetFocusPath(p1, []interface{}{0}, D)
	assert.Equal(t, true, r.Succeeded)
	assert.Equal(t, 0, p1.Focus())

	c1.SetFocus(D, 2)
	assert.Equal(t, 2, c1.Focus())

	r = gowid.SetFocusPath(p1, []interface{}{1}, D)
	assert.Equal(t, true, r.Succeeded)
	assert.Equal(t, 1, p1.Focus())
	assert.Equal(t, 2, c1.Focus())

	r = gowid.SetFocusPath(p1, []interface{}{1, 0}, D)
	assert.Equal(t, true, r.Succeeded)
	assert.Equal(t, 1, p1.Focus())
	assert.Equal(t, 0, c1.Focus())

	c2.SetFocus(D, 1)
	assert.Equal(t, 1, c2.Focus())

	r = gowid.SetFocusPath(p1, []interface{}{2, 0}, D)
	assert.Equal(t, true, r.Succeeded)
	assert.Equal(t, 2, p1.Focus())
	assert.Equal(t, 0, c2.Focus())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
