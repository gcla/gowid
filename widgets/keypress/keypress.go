// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package keypress provides a widget which responds to keyboard input.
package keypress

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gdamore/tcell"
)

//======================================================================

type ICustomKeys interface {
	CustomSelectKeys() bool
	SelectKeys() []gowid.IKey // can't be nil
}

type KeyPressFunction func(app gowid.IApp, widget gowid.IWidget, key gowid.IKey)

func (f KeyPressFunction) Changed(app gowid.IApp, widget gowid.IWidget, data ...interface{}) {
	k := data[0].(gowid.IKey)
	f(app, widget, k)
}

// WidgetCallback is a simple struct with a name field for IIdentity and
// that embeds a WidgetChangedFunction to be issued as a callback when a widget
// property changes.
type WidgetCallback struct {
	Name interface{}
	KeyPressFunction
}

func MakeCallback(name interface{}, fn KeyPressFunction) WidgetCallback {
	return WidgetCallback{
		Name:             name,
		KeyPressFunction: fn,
	}
}

func (f WidgetCallback) ID() interface{} {
	return f.Name
}

// IWidget is implemented by any widget that contains exactly one
// exposed subwidget (ICompositeWidget) and that is decorated on its left
// and right (IDecoratedAround).
type IWidget interface {
	gowid.ICompositeWidget
}

type Options struct {
	Keys []gowid.IKey
}

type Widget struct {
	inner     gowid.IWidget
	opts      Options
	Callbacks *gowid.Callbacks
	gowid.SubWidgetCallbacks
	gowid.KeyPressCallbacks
	gowid.IsSelectable
}

func New(inner gowid.IWidget, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	cb := gowid.NewCallbacks()
	res := &Widget{
		inner:              inner,
		opts:               opt,
		Callbacks:          cb,
		SubWidgetCallbacks: gowid.SubWidgetCallbacks{ICallbacks: cb},
		KeyPressCallbacks:  gowid.KeyPressCallbacks{ICallbacks: cb},
	}

	var _ gowid.IWidget = res
	var _ gowid.ICompositeWidget = res
	var _ IWidget = res
	var _ ICustomKeys = res

	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("keypress[%v]", w.SubWidget())
}

func (w *Widget) KeyPress(key gowid.IKey, app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.KeyPressCB{}, app, w, key)
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.inner
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.inner = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
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

func (w *Widget) CustomSelectKeys() bool {
	return true
}

func (w *Widget) SelectKeys() []gowid.IKey {
	return w.opts.Keys
}

//======================================================================

func SubWidgetSize(w gowid.ICompositeWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return size
}

func RenderSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.RenderSize(w.SubWidget(), size, focus, app)
}

type IKeyPresser interface {
	gowid.IKeyPress
	ICustomKeys
	gowid.IComposite
}

func UserInput(w IKeyPresser, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := false
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if w.CustomSelectKeys() {
			for _, k := range w.SelectKeys() {
				if gowid.KeysEqual(k, ev) {
					w.KeyPress(ev, app)
					res = true
					break
				}
			}
		}
	}
	if !res {
		res = gowid.UserInput(w.SubWidget(), ev, size, focus, app)
	}

	return res
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return gowid.Render(w.SubWidget(), size, focus, app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
