// Copyright 2021 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package paragraph

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func Test1(t *testing.T) {
	w := New("hello world")
	c := w.Render(gowid.RenderFlowWith{C: 16}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{
		"hello world     ",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello ",
		"world ",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 9}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello    ",
		"world    ",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 5}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello",
		"world",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 4}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hell",
		"o wo",
		"rld ",
	}, "\n")
	assert.Equal(t, res, c.String())

	w = New("hello worldatlarge")
	c = w.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello ",
		"worlda",
		"tlarge",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello wo",
		"rldatlar",
		"ge      ",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 12}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello       ",
		"worldatlarge",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 13}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello        ",
		"worldatlarge ",
	}, "\n")
	assert.Equal(t, res, c.String())

	c = w.Render(gowid.RenderFlowWith{C: 18}, gowid.NotSelected, gwtest.D)
	res = strings.Join([]string{
		"hello worldatlarge",
	}, "\n")
	assert.Equal(t, res, c.String())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
