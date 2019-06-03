// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gowid

import (
	"testing"

	"github.com/gcla/tcell"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
)

func TestColor1(t *testing.T) {
	c, _ := MakeRGBColorExtSafe(0, 0, 0)
	i2a, _ := c.ToTCellColor(Mode256Colors)
	i2 := i2a.ToTCell()
	if i2 != 17-1 {
		t.Errorf("Failed")
	}
}

func TestColor2(t *testing.T) {
	c := NewUrwidColor("dark red")
	i2a, _ := c.ToTCellColor(Mode256Colors)
	i2 := i2a.ToTCell()
	if i2 != 2-1 {
		t.Errorf("Failed")
	}
}

func TestColor3(t *testing.T) {
	c := MakeGrayColor("g#ff")
	if c.Val != 255 {
		t.Errorf("Failed")
	}
}

func TestColor4(t *testing.T) {
	c := MakeGrayColor("g99")
	if c.Val != 99 {
		t.Errorf("Failed")
	}
}

func TestColor5(t *testing.T) {
	c := MakeGrayColor("g100")
	if c.Val != 100 {
		t.Errorf("Failed")
	}
	if v, _ := c.ToTCellColor(Mode256Colors); v.ToTCell() != 232 {
		t.Errorf("Failed")
	}
}

func TestColor6(t *testing.T) {
	c := MakeGrayColor("g3")
	if c.Val != 3 {
		t.Errorf("Failed")
	}
	if v, _ := c.ToTCellColor(Mode256Colors); v.ToTCell() != 233 {
		t.Errorf("Failed")
	}
}

func TestColor7(t *testing.T) {
	c := MakeGrayColor("g0")
	if c.Val != 0 {
		t.Errorf("Failed")
	}
	if v, _ := c.ToTCellColor(Mode256Colors); v.ToTCell() != 17 {
		t.Errorf("Failed")
	}
}

func TestColorLookup1(t *testing.T) {
	res := makeColorLookup([]int{0, 7, 9}, 10)
	if deep.Equal(res, []int{0, 0, 0, 0, 1, 1, 1, 1, 1, 2}) != nil {
		t.Errorf("Failed")
	}
}

func TestIntScale1(t *testing.T) {
	if intScale(0x7, 0x10, 0x10000) != 0x7777 {
		t.Errorf("Failed val was %d", intScale(0x7, 0x10, 0x10000))
	}
	if intScale(0x5f, 0x100, 0x10) != 6 {
		t.Errorf("Failed val was %d", intScale(0x5f, 0x100, 0x10))
	}
	if intScale(2, 6, 101) != 40 {
		t.Errorf("Failed")
	}
	if intScale(1, 3, 4) != 2 {
		t.Errorf("Failed")
	}
}

func TestStringColor1(t *testing.T) {
	col1, _ := MakeRGBColorSafe("#12f")
	col2, _ := MakeRGBColorExtSafe(1*16, 2*16, 15*16)
	if deep.Equal(col1, col2) != nil {
		t.Errorf("Failed")
	}
}

func TestStringColor2(t *testing.T) {
	col1, _ := MakeRGBColorSafe("#12fgogogog")
	col2, _ := MakeRGBColorExtSafe(1*16, 2*16, 15*16)
	if deep.Equal(col1, col2) == nil {
		t.Errorf("Failed")
	}
}

func TestStringColor3(t *testing.T) {
	_, err := MakeRGBColorSafe("#34g")
	if err == nil {
		t.Errorf("Failed")
	}
}

func TestGray881(t *testing.T) {
	c := MakeGrayColor("g100")
	v, _ := c.ToTCellColor(Mode88Colors)
	assert.Equal(t, v.ToTCell(), tcell.Color(80))
}

func TestDefault1(t *testing.T) {
	c, _ := MakeColorSafe("default")
	v, _ := c.ToTCellColor(Mode256Colors)
	assert.Equal(t, v.ToTCell(), tcell.Color(-1))
}

func TestTCell1(t *testing.T) {
	c, _ := MakeColorSafe("maroon")
	v, _ := c.ToTCellColor(Mode256Colors)
	assert.Equal(t, v.ToTCell(), tcell.ColorMaroon)
}

func TestTCell2(t *testing.T) {
	c, _ := MakeTCellColorSafe(tcell.ColorMaroon)
	v, _ := c.ToTCellColor(Mode256Colors)
	assert.Equal(t, v.ToTCell(), tcell.ColorMaroon)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
