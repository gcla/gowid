// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package cellmod provides a widget that can change the cell data of an inner widget.
package cellmod

import (
	"github.com/gcla/gowid"
)

//======================================================================

type ICellMod interface {
	Transform(gowid.Cell, gowid.Selector) gowid.Cell
}

type Func func(gowid.Cell, gowid.Selector) gowid.Cell

func (f Func) Transform(cell gowid.Cell, focus gowid.Selector) gowid.Cell {
	return f(cell, focus)
}

type IWidget interface {
	gowid.ICompositeWidget
	ICellMod
}

// Widget that adjusts the palette used - if the rendering context provides for a foreground
// color of red (when focused), this widget can provide a map from red -> green to change its
// display
type Widget struct {
	gowid.IWidget
	mod ICellMod
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
}

func New(inner gowid.IWidget, mod ICellMod) *Widget {
	res := &Widget{
		IWidget: inner,
		mod:     mod,
	}

	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}

	var _ gowid.IWidget = res
	var _ gowid.ICompositeWidget = res
	var _ IWidget = res
	return res
}

func Opaque(inner gowid.IWidget) *Widget {
	return New(inner,
		Func(func(c gowid.Cell, focus gowid.Selector) gowid.Cell {
			if !c.HasRune() {
				c = c.WithRune(' ')
			}
			return c
		}))
}

func (w *Widget) String() string {
	return "cellmod"
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(inner gowid.IWidget, app gowid.IApp) {
	w.IWidget = inner
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) Mod() ICellMod {
	return w.mod
}

func (w *Widget) SetMod(mod ICellMod) {
	w.mod = mod
}

func (w *Widget) Transform(c gowid.Cell, focus gowid.Selector) gowid.Cell {
	return w.Mod().Transform(c, focus)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return w.SubWidget().RenderSize(size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return w.SubWidget().RenderSize(size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return gowid.UserInputIfSelectable(w.IWidget, ev, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	c := w.SubWidget().Render(size, focus, app)

	gowid.RangeOverCanvas(c, gowid.CellRangeFunc(func(cell gowid.Cell) gowid.Cell {
		return w.Transform(cell, focus)
	}))

	return c
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
