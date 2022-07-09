// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package disable provides a widget that forces its inner widget to be disable (or enabled).
package disable

import (
	"fmt"

	"github.com/gcla/gowid"
)

//======================================================================

// If you would like a non-selectable widget like TextWidget to be selectable
// in some context, wrap it in Widget
//
type Widget struct {
	gowid.IWidget
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
	isDisabled bool
}

func New(w gowid.IWidget) *Widget {
	return NewWith(w, true)
}

func NewDisabled(w gowid.IWidget) *Widget {
	return NewWith(w, true)
}

func NewEnabled(w gowid.IWidget) *Widget {
	return NewWith(w, false)
}

func NewWith(w gowid.IWidget, isDisabled bool) *Widget {
	res := &Widget{
		IWidget:    w,
		isDisabled: isDisabled,
	}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.ICompositeWidget = res
	return res
}

func (w *Widget) Enable() {
	w.isDisabled = false
}

func (w *Widget) Disable() {
	w.isDisabled = true
}

func (w *Widget) Set(val bool) {
	w.isDisabled = val
}

func (w *Widget) String() string {
	return fmt.Sprintf("disabled[d=%v,%v]", w.isDisabled, w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return size
}

func (w *Widget) Selectable() bool {
	return !w.isDisabled && w.SubWidget().Selectable()
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if w.isDisabled {
		return false
	}
	return w.SubWidget().UserInput(ev, size, focus, app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
