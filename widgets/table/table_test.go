// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package table

import (
	"sort"
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type MyTable struct {
	rows [][]gowid.IWidget
	hor  bool
	ver  bool
	wid  []gowid.IWidgetDimension
}

type MyHeader struct {
	widgets []gowid.IWidget
}

func (t MyTable) Columns() int {
	return 3
}

func (t MyTable) Rows() int {
	return len(t.rows)
}

func (t MyTable) HeaderWidgets() []gowid.IWidget {
	return nil
}

func (t MyTable) CellWidgets(row RowId) []gowid.IWidget {
	if row >= 0 && int(row) < t.Rows() {
		return t.rows[int(row)]
	} else {
		return nil
	}
}

func (t MyTable) Comparators() []ICompare {
	return nil
}

func (t MyTable) RowIdentifier(row int) (RowId, bool) {
	return RowId(row), true
}

type MyTableWithHeader struct {
	MyTable
	MyHeader
}

func (t MyTableWithHeader) HeaderWidgets() []gowid.IWidget {
	return t.MyHeader.widgets
}

func (t MyTable) VerticalSeparator() gowid.IWidget {
	if t.ver {
		return fill.New('|')
	}
	return nil
}

func (t MyTable) HorizontalSeparator() gowid.IWidget {
	if t.hor {
		return divider.NewAscii()
	}
	return nil
}

func (t MyTable) HeaderSeparator() gowid.IWidget {
	if t.hor {
		return divider.NewAscii()
	}
	return nil
}

func (t MyTable) Widths() []gowid.IWidgetDimension {
	return t.wid
}

var _ IModel = MyTable{}

//======================================================================

func makew(txt string) *selectable.Widget {
	return selectable.New(text.New(txt))
}

func TestEmptyTable1(t *testing.T) {
	model := MyTable{
		rows: [][]gowid.IWidget{},
		ver:  true,
	}

	sz := gowid.RenderFlowWith{C: 19}

	w1 := New(model)
	c1 := w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "", c1.String())

	_, err := w1.FocusXY()
	assert.Error(t, err)

	cbcalled := false

	w1.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		assert.Equal(t, w, w1)
		cbcalled = true
	}})

	model.rows = append(model.rows, []gowid.IWidget{makew("w1r0"), makew("w2r0"), makew("w3r0")})

	_, err = w1.FocusXY()
	assert.Error(t, err)
	assert.Equal(t, false, cbcalled)

	w1.SetModel(model, gwtest.D)

	xy, err := w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, Coords{0, 0}, xy)
	assert.Equal(t, true, cbcalled)
}

func TestTable1(t *testing.T) {
	model := MyTable{
		rows: [][]gowid.IWidget{
			{makew("w1r0"), makew("w2r0"), makew("w3r0")},
			{makew("w1r1"), makew("w2r1"), makew("w3r1")},
		},
		ver: true,
	}

	sz := gowid.RenderFlowWith{C: 19}

	w1 := New(model)
	c1 := w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "|w1r0 |w2r0 |w3r0 |\n|w1r1 |w2r1 |w3r1 |", c1.String())

	model.hor = true
	w1 = New(model)

	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "-------------------\n|w1r0 |w2r0 |w3r0 |\n-------------------\n|w1r1 |w2r1 |w3r1 |\n-------------------", c1.String())

	headers := MyHeader{
		widgets: []gowid.IWidget{
			makew("col0"), makew("col1"), makew("col2"),
		},
	}

	modelh := MyTableWithHeader{
		MyTable:  model,
		MyHeader: headers,
	}

	w1 = New(modelh)
	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
-------------------
|col0 |col1 |col2 |
-------------------
|w1r0 |w2r0 |w3r0 |
-------------------
|w1r1 |w2r1 |w3r1 |
-------------------
`[1:], "\n"), c1.String())

	widths := []gowid.IWidgetDimension{
		gowid.RenderWithUnits{6},
		gowid.RenderWithUnits{5},
		gowid.RenderWithUnits{4},
	}
	modelh.wid = widths
	w1 = New(modelh)
	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
-------------------
|col0  |col1 |col2|
-------------------
|w1r0  |w2r0 |w3r0|
-------------------
|w1r1  |w2r1 |w3r1|
-------------------
`[1:], "\n"), c1.String())

	assert.Equal(t, true, w1.Focus().Equal(Position(0)))
	xy, err := w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 0, xy.Column)
	assert.Equal(t, 0, xy.Row)

	evr := gwtest.CursorRight()

	w1.UserInput(evr, sz, gowid.Focused, gwtest.D)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 1, xy.Column)
	assert.Equal(t, 0, xy.Row)

	evd := gwtest.CursorDown()

	w1.UserInput(evd, sz, gowid.Focused, gwtest.D)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 1, xy.Column)
	assert.Equal(t, 1, xy.Row)

	evu := gwtest.CursorUp()

	w1.UserInput(evr, sz, gowid.Focused, gwtest.D)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 1, xy.Row)
	w1.UserInput(evu, sz, gowid.Focused, gwtest.D)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 0, xy.Row)

	cbcalled := false

	w1.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		assert.Equal(t, w, w1)
		cbcalled = true
	}})

	evmdown := tcell.NewEventMouse(1, 1, tcell.WheelDown, 0)

	w1.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	w1.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 2, xy.Row)
	w1.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	assert.Equal(t, false, cbcalled)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 2, xy.Row)

	w1.SetFocusXY(gwtest.D, Coords{Column: 1, Row: 0})
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 1, xy.Column)
	assert.Equal(t, 0, xy.Row)

	w1.SetFocusXY(gwtest.D, Coords{Column: 1, Row: 1})
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 1, xy.Column)
	assert.Equal(t, 1, xy.Row)

	w1.SetFocusXY(gwtest.D, Coords{Column: 2, Row: 2})
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 2, xy.Row)

	evlmx14y5 := tcell.NewEventMouse(14, 5, tcell.Button1, 0)
	evnonex14y5 := tcell.NewEventMouse(14, 5, tcell.ButtonNone, 0)

	w1.UserInput(evlmx14y5, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w1.UserInput(evnonex14y5, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 2, xy.Row)

	evmup := tcell.NewEventMouse(1, 1, tcell.WheelUp, 0)
	w1.UserInput(evmup, sz, gowid.Focused, gwtest.D)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 1, xy.Row)
	w1.UserInput(evmdown, sz, gowid.Focused, gwtest.D)
	xy, err = w1.FocusXY()
	assert.NoError(t, err)
	assert.Equal(t, 2, xy.Column)
	assert.Equal(t, 2, xy.Row)
}

//======================================================================

func TestTable2(t *testing.T) {
	csv := strings.TrimSuffix(`
1,c,-2
3,a,1.2
2,b,3.4
`[1:], "\n")

	sz := gowid.RenderFlowWith{C: 13}
	// Implements ITable
	t1 := NewCsvModel(strings.NewReader(csv), false, SimpleOptions{
		Style: StyleOptions{
			VerticalSeparator: fill.New('|'),
		},
	})
	w1 := New(t1)
	c1 := w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|1  |c  |-2 |
|3  |a  |1.2|
|2  |b  |3.4|
`[1:], "\n"), c1.String())

	assert.Equal(t, []int{0, 1, 2}, t1.SortOrder)
	assert.Equal(t, []int{0, 1, 2}, t1.InvSortOrder)

	logrus.Infof("cols is %d", t1.Columns())
	logrus.Infof("rows is %d", len(t1.Data))
	logrus.Infof("sort order is %v", t1.SortOrder)
	logrus.Infof("comp is %v", t1.Comparators)

	sorter := &SimpleTableByColumn{
		SimpleModel: t1,
		Column:      0,
	}
	sort.Sort(sorter)

	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|1  |c  |-2 |
|2  |b  |3.4|
|3  |a  |1.2|
`[1:], "\n"), c1.String())

	assert.Equal(t, []int{0, 2, 1}, t1.SortOrder)
	assert.Equal(t, []int{0, 2, 1}, t1.InvSortOrder)

	t1.Comparators[2] = FloatCompare{}
	sorter = &SimpleTableByColumn{
		SimpleModel: t1,
		Column:      2,
	}
	sort.Sort(sorter)

	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|1  |c  |-2 |
|3  |a  |1.2|
|2  |b  |3.4|
`[1:], "\n"), c1.String())

	assert.Equal(t, []int{0, 1, 2}, t1.SortOrder)
	assert.Equal(t, []int{0, 1, 2}, t1.InvSortOrder)

	sort.Sort(sort.Reverse(sorter))

	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|2  |b  |3.4|
|3  |a  |1.2|
|1  |c  |-2 |
`[1:], "\n"), c1.String())

	assert.Equal(t, []int{2, 1, 0}, t1.SortOrder)
	assert.Equal(t, []int{2, 1, 0}, t1.InvSortOrder)

	t1.Layout.Widths = []gowid.IWidgetDimension{
		gowid.RenderWithUnits{1},
		gowid.RenderWithUnits{2},
		gowid.RenderWithUnits{3},
	}
	w1 = New(t1)
	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|2|b |3.4|   
|3|a |1.2|   
|1|c |-2 |   
`[1:], "\n"), c1.String())

}

//======================================================================

func TestTable3(t *testing.T) {
	csv := strings.TrimSuffix(`
aaaaaaaaaaa,1
bbbbbbbbbbb,2
`[1:], "\n")

	// Implements ITable
	t1 := NewCsvModel(strings.NewReader(csv), false, SimpleOptions{
		Style: StyleOptions{
			VerticalSeparator: fill.New('|'),
		},
		Layout: LayoutOptions{
			Widths: []gowid.IWidgetDimension{
				gowid.RenderWithWeight{1},
				gowid.RenderWithUnits{1},
			},
		},
	})
	w1 := New(t1)
	c1 := w1.Render(gowid.RenderFlowWith{C: 15}, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|aaaaaaaaaaa|1|
|bbbbbbbbbbb|2|
`[1:], "\n"), c1.String())

	c1 = w1.Render(gowid.RenderFlowWith{C: 12}, gowid.Focused, gwtest.D)
	assert.Equal(t,
		strings.TrimSuffix(`
|aaaaaaaa|1|
|aaa     | |
|bbbbbbbb|2|
|bbb     | |
`[1:], "\n"), c1.String())

}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
