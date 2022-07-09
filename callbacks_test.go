// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gowid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	cb := NewCallbacks()

	x := 1
	cb.RunCallbacks("test1", 1)
	assert.Equal(t, 1, x)

	cb.AddCallback("test2", Callback{"addit", CallbackFunction(func(args ...interface{}) {
		y := args[0].(int)
		x = x + y
	})})
	cb.RunCallbacks("test1", 1)
	assert.Equal(t, 1, x)
	cb.RunCallbacks("test2", 1)
	assert.Equal(t, 2, x)
	cb.RunCallbacks("test2", 2)
	assert.Equal(t, 4, x)

	cb.AddCallback("test2", Callback{"addit100", CallbackFunction(func(args ...interface{}) {
		y := args[0].(int)
		x = x + (y * 100)
	})})

	cb.RunCallbacks("test2", 3)
	assert.Equal(t, 307, x)

	assert.Equal(t, false, cb.RemoveCallback("test2bad", CallbackID{"addit100"}))
	assert.Equal(t, false, cb.RemoveCallback("test2", CallbackID{"addit100bad"}))
	assert.Equal(t, true, cb.RemoveCallback("test2", CallbackID{"addit100"}))

	cb.RunCallbacks("test2", 8)
	assert.Equal(t, 315, x)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
