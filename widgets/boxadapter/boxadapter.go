// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code
// is governed by the MIT license that can be found in the LICENSE file.

// Package boxadapter provides a widget that will allow a box widget to be used
// in a flow context. Based on urwid's BoxAdapter - http://urwid.org/reference/widget.html#boxadapter.
package boxadapter

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell"
)

//======================================================================

type Widget struct {
	gowid.IWidget
	rows int
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
}

type IBoxAdapter interface {
	Rows() int
}

type IBoxAdapterWidget interface {
	gowid.ICompositeWidget
	IBoxAdapter
}

func New(inner gowid.IWidget, rows int) *Widget {
	res := &Widget{
		IWidget: inner,
		rows:    rows,
	}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.IWidget = res
	var _ gowid.IComposite = res
	var _ IBoxAdapter = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("boxadapter[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) Rows() int {
	return w.rows
}

func (w *Widget) SetRows(rows int, app gowid.IApp) {
	w.rows = rows
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

//======================================================================

// SubWidgetSize is the same as RenderSize for this widget - the inner widget will
// be rendered as a box with the specified number of columns and the widget's
// set number of rows.
func SubWidgetSize(w IBoxAdapter, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return RenderSize(w, size, focus, app)
}

func RenderSize(w IBoxAdapter, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	switch sz := size.(type) {
	case gowid.IRenderFlowWith:
		return gowid.RenderBox{
			C: sz.FlowColumns(),
			R: w.Rows(),
		}
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFlow"})
	}
}

func Render(w IBoxAdapterWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	rsize := RenderSize(w, size, focus, app)
	res := gowid.Render(w.SubWidget(), rsize, focus, app)

	return res
}

// Ensure that a valid mouse interaction with a flow widget will result in a
// mouse interaction with the subwidget
func UserInput(w IBoxAdapterWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if _, ok := size.(gowid.IRenderFlowWith); !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFlow"})
	}

	box := RenderSize(w, size, focus, app)

	if evm, ok := ev.(*tcell.EventMouse); ok {
		_, my := evm.Position()
		if my < box.BoxRows() && my >= 0 {
			return gowid.UserInputIfSelectable(w.SubWidget(), ev, box, focus, app)
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), ev, box, focus, app)
	}

	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
