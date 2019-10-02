// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package shadow adds a drop shadow effect to a widget.
package shadow

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell"
)

//======================================================================

type IOffset interface {
	Offset() int
}

type IWidget interface {
	gowid.ICompositeWidget
	IOffset
}

// Widget will render a drop shadow underneath and to the right of the inner widget,
// providing a simple 3D effect.
//
// Offset is the number of lines to extend the drop shadow down - it is
// extended right by 2*Offset because terminal cells aren't square.
//
type Widget struct {
	gowid.IWidget
	offset int // Means y offset, x is 2*y because cells are not squares -
	// we just guess at a reasonable look for a reasonable
	// aspect ratio
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
}

func New(inner gowid.IWidget, offset int) *Widget {
	res := &Widget{
		IWidget: inner,
		offset:  offset,
	}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.ICompositeWidget = res
	var _ IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("shadow[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) Offset() int {
	return w.offset
}

func (w *Widget) SetOffset(x int, app gowid.IApp) {
	w.offset = x
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func UserInput(w gowid.ICompositeWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	subSize := w.SubWidgetSize(size, focus, app)

	if evm, ok := ev.(*tcell.EventMouse); ok {
		ss := w.SubWidget().RenderSize(subSize, focus, app)
		mx, my := evm.Position()
		if my < ss.BoxRows() && my >= 0 && mx < ss.BoxColumns() && mx >= 0 {
			return gowid.UserInputIfSelectable(w.SubWidget(), ev, subSize, focus, app)
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), ev, subSize, focus, app)
	}
	return false
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	newSize := w.SubWidgetSize(size, focus, app)
	innerCanvas := w.SubWidget().Render(newSize, focus, app)

	shadowCanvas := gowid.NewCanvasOfSizeExt(innerCanvas.BoxColumns(), innerCanvas.BoxRows(),
		gowid.MakeCell(' ', gowid.MakeTCellColorExt(tcell.ColorDefault), gowid.MakeTCellColorExt(tcell.ColorBlack), gowid.StyleNone))

	shadowCanvas.ExtendLeft(gowid.EmptyLine(w.Offset() * 2))

	res := gowid.NewCanvasOfSize(shadowCanvas.BoxColumns(), w.Offset())
	res.AppendBelow(shadowCanvas, false, false)
	res.MergeUnder(innerCanvas, 0, 0, false)

	return res
}

func SubWidgetSize(w IOffset, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	var newSize gowid.IRenderSize
	switch sz := size.(type) {
	case gowid.IRenderFixed:
		newSize = gowid.RenderFixed{}
	case gowid.IRenderBox:
		newSize = gowid.RenderBox{C: sz.BoxColumns() - (w.Offset() * 2), R: sz.BoxRows() - w.Offset()}
	case gowid.IRenderFlowWith:
		newSize = gowid.RenderFlowWith{C: sz.FlowColumns() - 2}
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}
	return newSize
}

func RenderSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	ss := w.SubWidgetSize(size, focus, app)
	sdim := w.SubWidget().RenderSize(ss, focus, app)
	return gowid.RenderBox{C: sdim.BoxColumns() + (w.Offset() * 2), R: sdim.BoxRows() + w.Offset()}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
