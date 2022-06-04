// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package fixedadaptor provides a widget that will render a fixed widget when
// supplied with a box context.
package fixedadapter

import (
	"fmt"

	"github.com/gcla/gowid"
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

// Wraps a Fixed widget and turns it into a Box widget. If rendered in a Fixed
// context, render as normal. If rendered in a Box context, render as a Fixed
// widget, then either truncate or grow the resulting canvas to meet the
// box size requirement.
//
type Widget struct {
	gowid.IWidget
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
}

func New(inner gowid.IWidget) *Widget {
	res := &Widget{
		IWidget: inner,
	}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.IWidget = res
	var _ gowid.IComposite = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("fixedadapter[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func RenderSize(w gowid.IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func SubWidgetSize(w interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return gowid.RenderFixed{}
}

func Render(w gowid.IComposite, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	res := w.SubWidget().Render(SubWidgetSize(w, size, focus, app), focus, app)

	cols, ok := size.(gowid.IColumns)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
	}

	// Make sure that if we're rendered as a box, we have enough rows.
	gowid.FixCanvasHeight(res, size)

	res.ExtendRight(gowid.EmptyLine(cols.Columns() - res.BoxColumns()))

	return res
}

// Ensure that a valid mouse interaction with a flow widget will result in a
// mouse interaction with the subwidget
func UserInput(w gowid.ICompositeWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if evm, ok := ev.(*tcell.EventMouse); ok {
		box := RenderSize(w, size, focus, app)
		mx, my := evm.Position()
		if (my < box.BoxRows() && my >= 0) && (mx < box.BoxColumns() && mx >= 0) {
			return gowid.UserInputIfSelectable(w.SubWidget(), ev, SubWidgetSize(w, size, focus, app), focus, app)
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), ev, SubWidgetSize(w, size, focus, app), focus, app)
	}
	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
