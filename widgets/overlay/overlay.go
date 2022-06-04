// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package overlay is a widget that allows one widget to be overlaid on another.
package overlay

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/padding"
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

// Utility widget, used only to determine if a user input mouse event is within the bounds of
// a widget. This is different from whether or not a widget handles an event. In the case of overlay,
// an overlaid widget may not handle a mouse event, but if it occludes the widget underneath, that
// lower widget should not accept the mouse event (since it ought to be hidden). So the callback
// is expected to set a flag in the composite overlay widget to say the click was within bounds of the
// upper layer.
//
type MouseCheckerWidget struct {
	gowid.IWidget
	ClickWasInBounds func()
}

func (w *MouseCheckerWidget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *MouseCheckerWidget) SetSubWidget(inner gowid.IWidget, app gowid.IApp) {
	w.IWidget = inner
}

func (w *MouseCheckerWidget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return w.SubWidget().RenderSize(size, focus, app)
}

func NewMouseChecker(inner gowid.IWidget, clickWasInBounds func()) *MouseCheckerWidget {
	res := &MouseCheckerWidget{inner, clickWasInBounds}
	var _ gowid.ICompositeWidget = res
	return res
}

func (w *MouseCheckerWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if ev2, ok := ev.(*tcell.EventMouse); ok {
		mx, my := ev2.Position()
		ss := w.RenderSize(size, focus, app)
		if my < ss.BoxRows() && my >= 0 && mx < ss.BoxColumns() && mx >= 0 {
			w.ClickWasInBounds()
		}
	}
	return gowid.UserInputIfSelectable(w.IWidget, ev, size, focus, app)
}

//======================================================================

type IOverlay interface {
	Top() gowid.IWidget
	Bottom() gowid.IWidget
	VAlign() gowid.IVAlignment
	Height() gowid.IWidgetDimension
	HAlign() gowid.IHAlignment
	Width() gowid.IWidgetDimension
	BottomGetsFocus() bool
	TopGetsFocus() bool
	BottomGetsCursor() bool
}

type IWidget interface {
	gowid.IWidget
	IOverlay
}

type IWidgetSettable interface {
	IWidget
	SetTop(gowid.IWidget, gowid.IApp)
	SetBottom(gowid.IWidget, gowid.IApp)
	SetVAlign(gowid.IVAlignment, gowid.IApp)
	SetHeight(gowid.IWidgetDimension, gowid.IApp)
	SetHAlign(gowid.IHAlignment, gowid.IApp)
	SetWidth(gowid.IWidgetDimension, gowid.IApp)
}

// Widget overlays one widget on top of another. The bottom widget
// is rendered without the focus at full size. The bottom widget is
// rendered between a horizontal and vertical padding widget set up with
// the sizes provided.
type Widget struct {
	top       gowid.IWidget
	bottom    gowid.IWidget
	vAlign    gowid.IVAlignment
	height    gowid.IWidgetDimension
	hAlign    gowid.IHAlignment
	width     gowid.IWidgetDimension
	opts      Options
	Callbacks *gowid.Callbacks
}

// For callback registration
type Top struct{}
type Bottom struct{}

type Options struct {
	BottomGetsFocus  bool
	TopGetsNoFocus   bool
	BottomGetsCursor bool
}

func New(top, bottom gowid.IWidget,
	valign gowid.IVAlignment, height gowid.IWidgetDimension,
	halign gowid.IHAlignment, width gowid.IWidgetDimension,
	opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	res := &Widget{
		top:       top,
		bottom:    bottom,
		vAlign:    valign,
		height:    height,
		hAlign:    halign,
		width:     width,
		opts:      opt,
		Callbacks: gowid.NewCallbacks(),
	}
	var _ gowid.IWidget = res
	var _ IWidgetSettable = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("overlay[t=%v,b=%v]", w.top, w.bottom)
}

func (w *Widget) BottomGetsCursor() bool {
	return w.opts.BottomGetsCursor
}

func (w *Widget) BottomGetsFocus() bool {
	return w.opts.BottomGetsFocus
}

func (w *Widget) TopGetsFocus() bool {
	return !w.opts.TopGetsNoFocus
}

func (w *Widget) Top() gowid.IWidget {
	return w.top
}

func (w *Widget) SetTop(w2 gowid.IWidget, app gowid.IApp) {
	w.top = w2
	gowid.RunWidgetCallbacks(w.Callbacks, Top{}, app, w)
}

func (w *Widget) Bottom() gowid.IWidget {
	return w.bottom
}

func (w *Widget) SetBottom(w2 gowid.IWidget, app gowid.IApp) {
	w.bottom = w2
	gowid.RunWidgetCallbacks(w.Callbacks, Bottom{}, app, w)
}

func (w *Widget) VAlign() gowid.IVAlignment {
	return w.vAlign
}

func (w *Widget) SetVAlign(valign gowid.IVAlignment, app gowid.IApp) {
	w.vAlign = valign
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.VAlignCB{}, app, w)
}

func (w *Widget) Height() gowid.IWidgetDimension {
	return w.height
}

func (w *Widget) SetHeight(height gowid.IWidgetDimension, app gowid.IApp) {
	w.height = height
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.HeightCB{}, app, w)
}

func (w *Widget) HAlign() gowid.IHAlignment {
	return w.hAlign
}

func (w *Widget) SetHAlign(halign gowid.IHAlignment, app gowid.IApp) {
	w.hAlign = halign
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.HAlignCB{}, app, w)
}

func (w *Widget) Width() gowid.IWidgetDimension {
	return w.width
}

func (w *Widget) SetWidth(width gowid.IWidgetDimension, app gowid.IApp) {
	w.width = width
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.WidthCB{}, app, w)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return w.bottom.RenderSize(size, gowid.NotSelected, app)
}

func (w *Widget) Selectable() bool {
	return (w.top != nil && w.top.Selectable()) || w.bottom.Selectable()
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) SubWidget() gowid.IWidget {
	if w.opts.BottomGetsFocus {
		return w.bottom
	} else {
		return w.top
	}
}

func (w *Widget) SetSubWidget(inner gowid.IWidget, app gowid.IApp) {
	if w.opts.BottomGetsFocus {
		w.bottom = inner
	} else {
		w.top = inner
	}
}

//======================================================================

func UserInput(w IOverlay, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := false
	notOccluded := true

	if w.Top() == nil {
		res = gowid.UserInputIfSelectable(w.Bottom(), ev, size, focus, app)
	} else {
		top := NewMouseChecker(w.Top(), func() {
			notOccluded = false
		})
		p := padding.New(top, w.VAlign(), w.Height(), w.HAlign(), w.Width())

		res = gowid.UserInputIfSelectable(p, ev, size, focus, app)
		if !res {
			_, ok1 := ev.(*tcell.EventKey)
			_, ok2 := ev.(*tcell.EventMouse)
			_, ok3 := ev.(*tcell.EventPaste)
			if notOccluded && (ok1 || ok2 || ok3) {
				res = gowid.UserInputIfSelectable(w.Bottom(), ev, size, focus, app)
			}
		}
	}
	return res
}

func Render(w IOverlay, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	bfocus := focus.And(w.BottomGetsFocus())
	tfocus := focus.And(w.TopGetsFocus())

	bottomC := w.Bottom().Render(size, bfocus, app)
	if w.Top() == nil {
		return bottomC
	} else {
		bottomC2 := bottomC.Duplicate()
		p2 := padding.New(w.Top(), w.VAlign(), w.Height(), w.HAlign(), w.Width())
		topC := p2.Render(size, tfocus, app)
		bottomC2.MergeUnder(topC, 0, 0, w.BottomGetsCursor())
		return bottomC2
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
