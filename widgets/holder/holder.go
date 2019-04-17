// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package holder provides a widget that wraps an inner widget, and allows it to be easily swapped for another.
package holder

import (
	"fmt"

	"github.com/gcla/gowid"
)

//======================================================================

// Widget is the gowid analog of urwid's WidgetWrap.
type Widget struct {
	gowid.IWidget
	gowid.SubWidgetCallbacks
}

func New(w gowid.IWidget) *Widget {
	res := &Widget{
		IWidget:            w,
		SubWidgetCallbacks: gowid.SubWidgetCallbacks{ICallbacks: gowid.NewCallbacks()},
	}
	var _ gowid.IWidget = res
	var _ gowid.ICompositeWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("holder[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return w.SubWidget().RenderSize(size, focus, app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
