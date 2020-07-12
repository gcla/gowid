// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package hpadding provides a widget that pads an inner widget on the left and right.
package hpadding

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gdamore/tcell"
)

//======================================================================

type IWidget interface {
	gowid.ICompositeWidget
	Align() gowid.IHAlignment
	Width() gowid.IWidgetDimension
}

// Widget renders the wrapped widget with the provided
// width; if the wrapped widget is a box, or the wrapped widget is to be
// packed to a width smaller than specified, the wrapped widget can be
// aligned in the middle, left or right
type Widget struct {
	gowid.IWidget
	alignment gowid.IHAlignment
	width     gowid.IWidgetDimension
	*gowid.Callbacks
	gowid.FocusCallbacks
	gowid.SubWidgetCallbacks
}

func New(inner gowid.IWidget, alignment gowid.IHAlignment, width gowid.IWidgetDimension) *Widget {
	res := &Widget{
		IWidget:   inner,
		alignment: alignment,
		width:     width,
	}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("hpad[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) OnSetAlign(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, gowid.HAlignCB{}, f)
}

func (w *Widget) RemoveOnSetAlign(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, gowid.HAlignCB{}, f)
}

func (w *Widget) Align() gowid.IHAlignment {
	return w.alignment
}

func (w *Widget) SetAlign(i gowid.IHAlignment, app gowid.IApp) {
	w.alignment = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.HAlignCB{}, app, w)
}

func (w *Widget) OnSetHeight(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, gowid.WidthCB{}, f)
}

func (w *Widget) RemoveOnSetHeight(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, gowid.WidthCB{}, f)
}

func (w *Widget) Width() gowid.IWidgetDimension {
	return w.width
}

func (w *Widget) SetWidth(i gowid.IWidgetDimension, app gowid.IApp) {
	w.width = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.WidthCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func SubWidgetSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	size2 := size
	// If there is a horizontal offset specified, the relative features should reduce the size of the
	// supplied size i.e. it should be relative to the reduced screen size
	switch al := w.Align().(type) {
	case gowid.HAlignLeft:
		switch s := size.(type) {
		case gowid.IRenderBox:
			size2 = gowid.RenderBox{C: s.BoxColumns() - (al.Margin + al.MarginRight), R: s.BoxRows()}
		case gowid.IRenderFlowWith:
			size2 = gowid.RenderFlowWith{C: s.FlowColumns() - (al.Margin + al.MarginRight)}
		default:
		}
	default:
	}

	return gowid.ComputeHorizontalSubSizeUnsafe(size2, w.Width())
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {

	subSize := w.SubWidgetSize(size, focus, app)

	c := w.SubWidget().Render(subSize, focus, app)
	subWidgetMaxColumn := c.BoxColumns()

	var myCols int
	if cols, ok := size.(gowid.IColumns); ok {
		myCols = cols.Columns()
	} else {
		myCols = subWidgetMaxColumn
	}

	if myCols < subWidgetMaxColumn {
		// TODO - bad, mandates trimming on right
		c.TrimRight(myCols)
	} else if myCols > subWidgetMaxColumn {
		switch al := w.Align().(type) {
		case gowid.HAlignRight:
			c.ExtendLeft(gowid.EmptyLine(myCols - subWidgetMaxColumn))
		case gowid.HAlignLeft:
			l := gwutil.Min(al.Margin, myCols-subWidgetMaxColumn)
			r := myCols - (l + subWidgetMaxColumn)
			c.ExtendRight(gowid.EmptyLine(r))
			c.ExtendLeft(gowid.EmptyLine(l))
		default: // middle
			r := (myCols - subWidgetMaxColumn) / 2
			l := myCols - (subWidgetMaxColumn + r)
			c.ExtendRight(gowid.EmptyLine(r))
			c.ExtendLeft(gowid.EmptyLine(l))
		}
	}

	gowid.MakeCanvasRightSize(c, size)

	return c
}

// UserInput will adjust the input event's x coordinate depending on the input size
// and widget alignment. If the input is e.g. IRenderFixed, then no adjustment is
// made.
func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {

	subSize := w.SubWidgetSize(size, focus, app)
	ss := w.SubWidget().RenderSize(subSize, focus, app)
	cols := ss.BoxColumns()

	cols2, ok := size.(gowid.IColumns)

	var xd int

	if ok {
		switch al := w.Align().(type) {
		case gowid.HAlignRight:
			xd = -(cols2.Columns() - cols)
		case gowid.HAlignMiddle:
			r := (cols2.Columns() - cols) / 2
			l := cols2.Columns() - (cols + r)
			xd = -l
		case gowid.HAlignLeft:
			if al.Margin+cols <= cols2.Columns() {
				xd = -al.Margin
			} else {
				xd = -gwutil.Max(0, cols2.Columns()-cols)
			}
		}
	}
	newev := gowid.TranslatedMouseEvent(ev, xd, 0)

	// TODO - don't need to translate event for keyboard event...
	if evm, ok := newev.(*tcell.EventMouse); ok {
		mx, _ := evm.Position()
		if mx >= 0 && mx < cols {
			return gowid.UserInputIfSelectable(w.SubWidget(), newev, subSize, focus, app)
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), newev, subSize, focus, app)
	}

	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
