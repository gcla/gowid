// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gwutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test1(t *testing.T) {
	str := StringOfLength('x', 10)
	if str != "xxxxxxxxxx" {
		t.Errorf("Failed")
	}
}

func TestHA1(t *testing.T) {
	res1 := HamiltonAllocation([]int{1, 1, 1}, 3)
	assert.Equal(t, res1, []int{1, 1, 1})

	res2 := HamiltonAllocation([]int{1, 1, 1}, 12)
	assert.Equal(t, res2, []int{4, 4, 4})

	res3 := HamiltonAllocation([]int{3, 2, 1}, 12)
	assert.Equal(t, res3, []int{6, 4, 2})

	res4 := HamiltonAllocation([]int{1, 2, 3}, 12)
	assert.Equal(t, res4, []int{2, 4, 6})

	res5 := HamiltonAllocation([]int{10, 5, 1}, 8)
	assert.Equal(t, res5, []int{5, 3, 0})

	res6 := HamiltonAllocation([]int{10, 5, 1}, 0)
	assert.Equal(t, res6, []int{0, 0, 0})
}

func TestOpt1(t *testing.T) {
	opt1 := SomeInt(56)
	assert.Equal(t, "56", fmt.Sprintf("%v", opt1))
	opt1 = NoneInt()
	assert.Equal(t, "None", fmt.Sprintf("%v", opt1))
}

func TestMin1(t *testing.T) {
	assert.Equal(t, 1, Min(1, 2, 3))
	assert.Equal(t, 1, Min(2, 1, 3))
	assert.Equal(t, 1, Min(2, 3, 1))
}

func TestMax1(t *testing.T) {
	assert.Equal(t, 3, Max(1, 2, 3))
	assert.Equal(t, 3, Max(2, 1, 3))
	assert.Equal(t, 3, Max(2, 3, 1))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
