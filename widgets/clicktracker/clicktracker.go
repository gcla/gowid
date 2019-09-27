// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package clicktracker provides a widget that inverts when the mouse is clicked
// but not yet released.
package clicktracker

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell"
)

//======================================================================

// IWidget is implemented by any widget that contains exactly one
// exposed subwidget (ICompositeWidget), that can distinguish itself
// from another IWidget (via the ID() function), and that can track
// a mouse click prior to a mouse release (simply with a bool flag, in
// this case)
type IWidget interface {
	gowid.ICompositeWidget
	gowid.IClickTracker
	gowid.IIdentity
	ClickPending() bool
}

type Widget struct {
	inner     gowid.IWidget
	Callbacks *gowid.Callbacks
	gowid.SubWidgetCallbacks
	gowid.ClickCallbacks
	gowid.AddressProvidesID
	gowid.IsSelectable
	clickDown bool
}

func New(inner gowid.IWidget) *Widget {
	res := &Widget{
		inner: inner,
	}

	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	res.ClickCallbacks = gowid.ClickCallbacks{CB: &res.Callbacks}

	var _ gowid.IWidget = res
	var _ IWidget = res

	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("clicktracker[%v]", w.SubWidget())
}

func (w *Widget) ClickPending() bool {
	return w.clickDown
}

func (w *Widget) SetClickPending(pending bool) {
	w.clickDown = pending
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.inner
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.inner = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(size, focus, app)
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

//======================================================================

func SubWidgetSize(size interface{}, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return size
}

func RenderSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.RenderSize(w.SubWidget(), size, focus, app)
}

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		switch ev.Buttons() {
		case tcell.Button1, tcell.Button2, tcell.Button3:
			app.SetClickTarget(ev.Buttons(), w)
			w.SetClickPending(true)
		case tcell.ButtonNone:
			w.SetClickPending(false)
		}
	}
	// Never handle the input, always pass it on
	return gowid.UserInput(w.SubWidget(), ev, size, focus, app)
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	res := gowid.Render(w.SubWidget(), gowid.SubWidgetSize(w, size, focus, app), focus, app)

	if w.ClickPending() {
		gowid.RangeOverCanvas(res, gowid.CellRangeFunc(func(c gowid.Cell) gowid.Cell {
			f, b := c.ForegroundColor(), c.BackgroundColor()
			return c.WithBackgroundColor(f).WithForegroundColor(b)
		}))
	}
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
