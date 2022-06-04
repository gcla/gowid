// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package button provides a clickable widget which can be decorated.
package button

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

// IDecoratedAround is the interface for any type that provides
// "decoration" on its left and right side e.g. for buttons, something like
// "<" and ">".
type IDecoratedAround interface {
	LeftDec() string
	RightDec() string
}

// IDecoratedMiddle is implemented by any type that provides "decoration"
// in the middle of its render, such as a 'x' or a '-' symbol on a checked
// button.
type IDecoratedMiddle interface {
	MiddleDec() string
}

//======================================================================

// Decoration is a simple struct that implements IDecoratedAround.
type Decoration struct {
	Left  string
	Right string
}

func (b *Decoration) LeftDec() string {
	return b.Left
}

func (b *Decoration) RightDec() string {
	return b.Right
}

func (w *Decoration) SetLeftDec(dec string, app gowid.IApp) {
	w.Left = dec
}

func (w *Decoration) SetRightDec(dec string, app gowid.IApp) {
	w.Right = dec
}

var (
	BareDecoration   = Decoration{Left: "", Right: ""}
	NormalDecoration = Decoration{Left: "<", Right: ">"}
	AltDecoration    = Decoration{Left: "[", Right: "]"}
)

//======================================================================

type ICustomKeys interface {
	CustomSelectKeys() bool
	SelectKeys() []gowid.IKey // can't be nil
}

// IWidget is implemented by any widget that contains exactly one
// exposed subwidget (ICompositeWidget) and that is decorated on its left
// and right (IDecoratedAround).
type IWidget interface {
	gowid.ICompositeWidget
	IDecoratedAround
}

type Options struct {
	Decoration
	SelectKeysProvided bool
	SelectKeys         []gowid.IKey
}

type Widget struct {
	inner gowid.IWidget
	opts  Options
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
	gowid.ClickCallbacks
	*Decoration
	gowid.AddressProvidesID
	gowid.IsSelectable
}

func New(inner gowid.IWidget, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		// Make the default have visible decorators, if none are provided explicitly.
		opt.Decoration = NormalDecoration
	}

	res := &Widget{
		inner: inner,
		opts:  opt,
	}

	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	res.ClickCallbacks = gowid.ClickCallbacks{CB: &res.Callbacks}

	res.Decoration = &res.opts.Decoration

	var _ gowid.IWidget = res
	var _ gowid.ICompositeWidget = res
	var _ IWidget = res
	var _ ICustomKeys = res

	return res
}

func NewAlt(inner gowid.IWidget) *Widget {
	return New(inner, Options{
		Decoration: AltDecoration,
	})
}

func NewBare(inner gowid.IWidget) *Widget {
	return New(inner, Options{
		Decoration: BareDecoration,
	})
}

func NewDecorated(inner gowid.IWidget, decoration Decoration) *Widget {
	return New(inner, Options{
		Decoration: decoration,
	})
}

func (w *Widget) String() string {
	return fmt.Sprintf("button[%v]", w.SubWidget())
}

func (w *Widget) Click(app gowid.IApp) {
	// No button clicked means a key was pressed
	if app.GetMouseState().NoButtonClicked() || app.GetMouseState().LeftIsClicked() {
		gowid.RunWidgetCallbacks(w.Callbacks, gowid.ClickCB{}, app, w)
	}
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.inner
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.inner = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) SetLeftDec(dec string, app gowid.IApp) {
	w.Decoration.Left = dec
}

func (w *Widget) SetRightDec(dec string, app gowid.IApp) {
	w.Decoration.Right = dec
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
	return w.opts.SelectKeysProvided
}

func (w *Widget) SelectKeys() []gowid.IKey {
	return w.opts.SelectKeys
}

//======================================================================

func SubWidgetSize(w IWidget, size interface{}, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	cols, haveCols := size.(gowid.IColumns)
	rows, haveRows := size.(gowid.IRows)
	switch {
	case haveCols && haveRows:
		return gowid.RenderBox{C: gwutil.Max(0, cols.Columns()-(len(w.LeftDec())+len(w.RightDec()))), R: rows.Rows()}
	case haveCols:
		return gowid.RenderFlowWith{C: gwutil.Max(0, cols.Columns()-(len(w.LeftDec())+len(w.RightDec())))}
	default:
		return gowid.RenderFixed{}
	}
}

func RenderSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	innerSize := w.SubWidgetSize(size, focus, app)
	innerRendered := w.SubWidget().RenderSize(innerSize, focus, app)
	boxHeight := innerRendered.BoxRows()
	boxWidth := innerRendered.BoxColumns() + len(w.LeftDec()) + len(w.RightDec())
	if bsz, ok := size.(gowid.IColumns); ok {
		if bsz.Columns() < boxWidth {
			boxWidth = bsz.Columns()
		}
	}
	return gowid.RenderBox{boxWidth, boxHeight}
}

type IClickableIdentityWidget interface {
	gowid.IClickableWidget
	gowid.IIdentity
}

func UserInput(w IClickableIdentityWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := false
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		switch ev.Buttons() {
		case tcell.Button1, tcell.Button2, tcell.Button3:
			app.SetClickTarget(ev.Buttons(), w)
			res = true
		case tcell.ButtonNone:
			if !app.GetLastMouseState().NoButtonClicked() {
				clickit := false
				app.ClickTarget(func(k tcell.ButtonMask, v gowid.IIdentityWidget) {
					if v != nil && v.ID() == w.ID() {
						clickit = true
					}
				})
				if clickit {
					w.Click(app)
					res = true
				}
			}
		}
	case *tcell.EventKey:
		if wk, ok := w.(ICustomKeys); ok && wk.CustomSelectKeys() {
			for _, k := range wk.SelectKeys() {
				if gowid.KeysEqual(k, ev) {
					w.Click(app)
					res = true
					break
				}
			}
		} else {
			if ev.Key() == tcell.KeyEnter || (ev.Key() == tcell.KeyRune && ev.Rune() == ' ') {
				w.Click(app)
				res = true
			}
		}
	default:
		if wc, ok := w.(gowid.IComposite); ok {
			res = wc.SubWidget().UserInput(ev, size, focus, app)
		}
	}
	return res
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	newSize := w.SubWidgetSize(size, focus, app)

	res := w.SubWidget().Render(newSize, focus, app)
	leftClicker := gowid.CellsFromString(w.LeftDec())
	rightClicker := gowid.CellsFromString(w.RightDec())
	res.ExtendLeft(leftClicker)
	res.ExtendRight(rightClicker)
	gowid.MakeCanvasRightSize(res, size)
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
