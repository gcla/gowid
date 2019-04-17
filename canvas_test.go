// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package gowid

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterfaces1(t *testing.T) {
	var _ io.Writer = (*Canvas)(nil)
}

func TestInterfaces5(t *testing.T) {
	var _ IComposite = (*App)(nil)
}

func TestCanvas19(t *testing.T) {
	canvas := NewCanvas()
	c := canvas.BoxColumns()
	r := canvas.BoxRows()
	if c != 0 || r != 0 {
		t.Errorf("Failed")
	}
	r1 := CellsFromString("abc")
	r2 := CellsFromString("12")
	canvas.AppendLine(r1, false)
	canvas.AppendLine(r2, false)
	c = canvas.BoxColumns()
	r = canvas.BoxRows()
	if c != 3 || r != 2 {
		t.Errorf("Failed c is %d r is %d", c, r)
	}
	cs := canvas.String()
	if cs != "abc\n12 " {
		t.Errorf("Failed cs is %v", cs)
	}
	canvas.AlignRight()
	cs = canvas.String()
	if cs != "abc\n12 " {
		t.Errorf("Failed")
	}
	canvas2 := NewCanvas()
	r21 := CellsFromString(" X Z")
	r22 := CellsFromString("Y2")
	canvas2.AppendLine(r21, false)
	canvas2.AppendLine(r22, false)
	canvas.MergeUnder(canvas2, 0, 0, false)
	cs = canvas.String()
	if cs != "aXc\nY2 " {
		t.Errorf("Failed")
	}
	assert.Equal(t, canvas.BoxColumns(), 3)
	var n int
	var err error
	n, err = canvas.Write([]byte{'1', '2', '3', 'Q'})
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, canvas.String(), "123\nQ2 ")
	n, err = canvas.Write([]byte{'5', '\n'})
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, canvas.String(), "5  \nQ2 ")
	n, err = canvas.Write([]byte{0xe2, 0x98, 0xa0, '\n'})
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, canvas.String(), "☠  \nQ2 ")

	n, err = canvas.Write([]byte{'1', '2', '\n', 'R'})
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, canvas.String(), "12 \nR2 ")
}

type MyString string

func (s MyString) Tester() int {
	return len(s)
}

type FooType interface {
	Tester() int
}

func MyTestFn(f FooType) {
}

func TestCanvas1(t *testing.T) {
	f := MyString("xyz")
	MyTestFn(f)
	assert.Equal(t, f.Tester(), 3)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
