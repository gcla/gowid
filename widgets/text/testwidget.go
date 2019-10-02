// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package text provides a text field widget.
package text

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell"
)

//======================================================================

type Widget1 struct {
	I int
}

func (w *Widget1) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	if focus.Focus {
		return New(fmt.Sprintf("%df", w.I)).RenderSize(size, focus, app)
	} else {
		return New(fmt.Sprintf("%d ", w.I)).RenderSize(size, focus, app)
	}
}

func (w *Widget1) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	if focus.Focus {
		return New(fmt.Sprintf("%df", w.I)).Render(size, focus, app)
	} else {
		return New(fmt.Sprintf("%d ", w.I)).Render(size, focus, app)
	}
}

func (w *Widget1) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {

	switch ev := ev.(type) {
	case *tcell.EventKey:
		if focus.Focus {
			return New(fmt.Sprintf("%df", w.I)).UserInput(ev, size, focus, app)
		} else {
			return New(fmt.Sprintf("%d ", w.I)).UserInput(ev, size, focus, app)
		}
	case *tcell.EventMouse:
		// Take all mouse input so I can test clicking in different columns changing the focus
		return true
	}

	return false
}

func (w *Widget1) Selectable() bool {
	return true
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
