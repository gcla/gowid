// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package fill provides a widget that can be filled with a styled rune.
package fill

import (
	"fmt"

	"github.com/gcla/gowid"
)

//======================================================================

type ISolidFill interface {
	Cell() gowid.Cell
}

type IWidget interface {
	gowid.IWidget
	ISolidFill
}

type Widget struct {
	cell gowid.Cell
	gowid.RejectUserInput
	gowid.NotSelectable
}

func New(chr rune) *Widget {
	res := &Widget{cell: gowid.CellFromRune(chr)}
	var _ gowid.IWidget = res
	return res
}

func NewEmpty() *Widget {
	return NewSolidFromCell(gowid.Cell{})
}

func NewSolidFromCell(cell gowid.Cell) *Widget {
	res := &Widget{cell: cell}
	var _ gowid.IWidget = res
	return res
}

func (w *Widget) String() string {
	r := ' '
	if w.Cell().HasRune() {
		r = w.Cell().Rune()
	}
	return fmt.Sprintf("fill[%c]", r)
}

func (w *Widget) Cell() gowid.Cell {
	return w.cell
}

func (w *Widget) SetCell(c gowid.Cell, app gowid.IApp) {
	w.cell = c
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func RenderSize(w interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	cols, haveCols := size.(gowid.IColumns)
	rows, haveRows := size.(gowid.IRows)
	switch {
	case haveCols && haveRows:
		return gowid.RenderBox{C: cols.Columns(), R: rows.Rows()}
	case haveCols:
		return gowid.RenderBox{C: cols.Columns(), R: 1}
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}
}

func Render(w ISolidFill, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	// If a col and row not provided, what can I choose??
	cols2, ok := size.(gowid.IColumns)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
	}

	cols := cols2.Columns()
	rows := 1
	if irows, ok2 := size.(gowid.IRows); ok2 {
		rows = irows.Rows()
	}

	fill := w.Cell()
	fillArr := make([]gowid.Cell, cols)
	for i := 0; i < cols; i++ {
		fillArr[i] = fill
	}

	res := gowid.NewCanvas()
	if rows > 0 {
		res.AppendLine(fillArr, false)
		for i := 0; i < rows-1; i++ {
			res.Lines = append(res.Lines, make([]gowid.Cell, 0, 120))
		}
	}
	res.AlignRightWith(w.Cell())

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
