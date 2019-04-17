// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package vscroll

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/gwutil"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestVerticalSplits(t *testing.T) {
	x, y, z := 1, 1, 1
	splits := gwutil.HamiltonAllocation([]int{x, y, z}, 3)
	assert.Equal(t, []int{1, 1, 1}, splits)

	x, y, z = 5, 5, 5
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 3)
	assert.Equal(t, []int{1, 1, 1}, splits)

	x, y, z = 5, 5, 5
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 6)
	assert.Equal(t, []int{2, 2, 2}, splits)

	x, y, z = 2, 4, 6
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 6)
	assert.Equal(t, []int{1, 2, 3}, splits)

	x, y, z = 0, 3, 6
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 6)
	assert.Equal(t, []int{0, 2, 4}, splits)

	x, y, z = 1, 3, 8
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 6)
	assert.Equal(t, []int{1, 1, 4}, splits)

	x, y, z = 1, 3, 16
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 6)
	assert.Equal(t, []int{0, 1, 5}, splits)

	x, y, z = 1, 3, 16
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 0)
	assert.Equal(t, []int{0, 0, 0}, splits)

	x, y, z = 18, 14, 643
	splits = gwutil.HamiltonAllocation([]int{x, y, z}, 12)
	assert.Equal(t, []int{0, 0, 12}, splits)
}

func TestVerticalScrollbar(t *testing.T) {
	w := New()
	c1 := w.Render(gowid.RenderBox{C: 2, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "^^\n  \n##\n  \nvv")

	assert.Panics(t, func() {
		w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	})

}

func TestVerticalScrollbar2(t *testing.T) {
	w := New()
	c1 := w.Render(gowid.RenderBox{C: 1, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), strings.Join([]string{"^", " ", "#", " ", "v"}, "\n"))

	w.Top = 0
	w.Middle = 1
	w.Bottom = 1
	c1 = w.Render(gowid.RenderBox{C: 1, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), strings.Join([]string{"^", "#", "#", " ", "v"}, "\n"))

	// w.Top = 1
	// w.Middle = 1
	// w.Bottom = 97
	// c1 = w.Render(gowid.RenderBox{C: 1, R: 5}, true, gwtest.D)
	// assert.Equal(t, c1.String(), strings.Join([]string{"^", "#", "#", " ", "v"}, "\n"))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
