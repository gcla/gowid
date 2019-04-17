// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package isselected provides a widget that acts differently if selected (or focused)
package isselected

import (
	"github.com/gcla/gowid"
)

//======================================================================

type Widget struct {
	Not      gowid.IWidget // Must not be nil
	Selected gowid.IWidget // If nil, then Not used
	Focused  gowid.IWidget // If nil, then Selected used
}

var _ gowid.IWidget = (*Widget)(nil)

func New(w1, w2, w3 gowid.IWidget) *Widget {
	return &Widget{
		Not:      w1,
		Selected: w2,
		Focused:  w3,
	}
}

func (w *Widget) pick(focus gowid.Selector) gowid.IWidget {
	if focus.Focus {
		if w.Focused == nil && w.Selected == nil {
			return w.Not
		} else if w.Focused == nil {
			return w.Selected
		} else {
			return w.Focused
		}
	} else if focus.Selected {
		if w.Selected == nil {
			return w.Not
		} else {
			return w.Selected
		}
	} else {
		return w.Not
	}
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.RenderSize(w.pick(focus), size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return gowid.Render(w.pick(focus), size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return gowid.UserInput(w.pick(focus), ev, size, focus, app)
}

// TODO - this isn't right. Should Selectable be conditioned on focus?
func (w *Widget) Selectable() bool {
	return w.Not.Selectable()
}

//======================================================================

// For uses that require IComposite
type WidgetExt struct {
	*Widget
}

func NewExt(w1, w2, w3 gowid.IWidget) *WidgetExt {
	return &WidgetExt{New(w1, w2, w3)}
}

var _ gowid.IWidget = (*WidgetExt)(nil)
var _ gowid.IComposite = (*WidgetExt)(nil)

// Return Focused because UserInput operations that change state
// will apply when the widget is in focus - so this is likely the
// one we want. But this looks like a rich source of bugs...
func (w *WidgetExt) SubWidget() gowid.IWidget {
	return w.pick(gowid.Focused)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
