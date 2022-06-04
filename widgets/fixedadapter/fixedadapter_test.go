// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package fixedadapter

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/stretchr/testify/assert"
)

func TestBoxify1(t *testing.T) {
	w := checkbox.New(false)

	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "[ ]", c1.String())

	w2 := New(w)
	c2 := w2.Render(gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, "[ ]  \n     \n     ", c2.String())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
