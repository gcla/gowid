// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package null provides a widget which does nothing and renders nothing.
package null

import (
	"fmt"

	"github.com/gcla/gowid"
)

//======================================================================

type Widget struct{}

func New() *Widget {
	res := &Widget{}
	var _ gowid.IWidget = res
	return res
}

func (w *Widget) Selectable() bool {
	return false
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return false
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	cols, haveCols := size.(gowid.IColumns)
	rows, haveRows := size.(gowid.IRows)
	switch {
	case haveCols && haveRows:
		return gowid.RenderBox{C: cols.Columns(), R: rows.Rows()}
	case haveCols:
		return gowid.RenderBox{C: cols.Columns(), R: 0}
	default:
		return gowid.RenderBox{C: 0, R: 0}
	}
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	res := gowid.NewCanvasOfSize(0, 0)
	gowid.MakeCanvasRightSize(res, size)
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("null")
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
