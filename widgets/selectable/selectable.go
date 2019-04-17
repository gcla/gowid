// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package selectable provides a widget that forces its inner widget to be selectable.
package selectable

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
	Callbacks *gowid.Callbacks
	gowid.SubWidgetCallbacks
	isSelectable bool
}

func New(w gowid.IWidget) *Widget {
	return NewWith(w, true)
}

func NewSelectable(w gowid.IWidget) *Widget {
	return NewWith(w, true)
}

func NewUnselectable(w gowid.IWidget) *Widget {
	return NewWith(w, false)
}

func NewWith(w gowid.IWidget, isSelectable bool) *Widget {
	cb := gowid.NewCallbacks()
	res := &Widget{
		IWidget:            w,
		Callbacks:          cb,
		SubWidgetCallbacks: gowid.SubWidgetCallbacks{ICallbacks: cb},
		isSelectable:       isSelectable,
	}
	var _ gowid.ICompositeWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("selectable[%v]", w.SubWidget())
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
	return w.isSelectable
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
