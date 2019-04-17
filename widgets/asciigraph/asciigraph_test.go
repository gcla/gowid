// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package asciigraph

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	asc "github.com/guptarohit/asciigraph"
	"github.com/stretchr/testify/assert"
)

func TestSolidAsciigraph1(t *testing.T) {

	// Stolen from asciigraph_test.go
	data := []float64{2, 1, 1, 2, -2, 5, 7, 11, 3, 7, 1}
	conf := []asc.Option{}

	w := New(data, conf)

	assert.Panics(t, func() {
		w.Render(gowid.RenderFlowWith{C: 10}, gowid.Focused, gwtest.D)
	})

	c1 := w.Render(gowid.RenderBox{C: 19, R: 14}, gowid.Focused, gwtest.D)

	res := ` 11.00 ┤      ╭╮   
 10.00 ┤      ││   
  9.00 ┼      ││   
  8.00 ┤      ││   
  7.00 ┤     ╭╯│╭╮ 
  6.00 ┤     │ │││ 
  5.00 ┤    ╭╯ │││ 
  4.00 ┤    │  │││ 
  3.00 ┤    │  ╰╯│ 
  2.00 ┼╮ ╭╮│    │ 
  1.00 ┤╰─╯││    ╰ 
  0.00 ┤   ││      
 -1.00 ┤   ││      
 -2.00 ┤   ╰╯      `

	t.Logf("Canvas is\n%v\n", c1.String())

	assert.Equal(t, c1.String(), res)

	c1 = w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), res)

	gwtest.RenderBoxManyTimes(t, w, 0, 20, 0, 20)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
