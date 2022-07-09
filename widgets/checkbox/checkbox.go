// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package checkbox provides a widget which can be checked or unchecked.
package checkbox

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/button"
)

//======================================================================

type IChecked interface {
	button.IDecoratedAround
	button.IDecoratedMiddle
	IsChecked() bool
}

type IWidget interface {
	gowid.IWidget
	IChecked
}

//======================================================================

type Decoration struct {
	button.Decoration
	Middle string
}

func (b *Decoration) MiddleDec() string {
	return b.Middle
}

func (w *Decoration) SetMiddleDec(dec string, app gowid.IApp) {
	w.Middle = dec
}

//======================================================================

type Widget struct {
	checked   bool
	Callbacks *gowid.Callbacks
	gowid.ClickCallbacks
	Decoration
	gowid.AddressProvidesID
	gowid.IsSelectable
}

func New(isChecked bool) *Widget {
	cb := gowid.NewCallbacks()
	res := &Widget{
		checked:        isChecked,
		Callbacks:      cb,
		ClickCallbacks: gowid.ClickCallbacks{CB: &cb},
		Decoration:     Decoration{button.Decoration{"[", "]"}, "X"},
	}
	var _ gowid.IWidget = res
	return res
}

func NewDecorated(isChecked bool, decoration Decoration) *Widget {
	cb := gowid.NewCallbacks()
	res := &Widget{
		checked:        isChecked,
		Callbacks:      cb,
		ClickCallbacks: gowid.ClickCallbacks{CB: &cb},
		Decoration:     decoration,
	}
	var _ gowid.IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("checkbox[%s]", gwutil.If(w.IsChecked(), "X", " ").(string))
}

func (w *Widget) IsChecked() bool {
	return w.checked
}

func (w *Widget) SetChecked(app gowid.IApp, val bool) {
	w.setChecked(app, val)
}

func (w *Widget) setChecked(app gowid.IApp, val bool) {
	w.checked = val
	gowid.RunWidgetCallbacks(*w.CB, gowid.ClickCB{}, app, w)
}

func (w *Widget) Click(app gowid.IApp) {
	if app.GetMouseState().NoButtonClicked() || app.GetMouseState().LeftIsClicked() {
		w.setChecked(app, !w.IsChecked())
	}
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.RenderBox{C: len(w.LeftDec()) + len(w.MiddleDec()) + len(w.RightDec()), R: 1}
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {

	if _, ok := size.(gowid.IRenderFixed); !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFixed"})
	}

	return Render(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if _, ok := size.(gowid.IRenderFixed); !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFixed"})
	}
	return button.UserInput(w, ev, size, focus, app)
}

//======================================================================

func Render(w IChecked, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	line := make([]gowid.Cell, 0)
	line = append(line, gowid.CellsFromString(w.LeftDec())...)
	if w.IsChecked() {
		line = append(line, gowid.CellsFromString(w.MiddleDec())...)
	} else {
		line = append(line, gowid.CellsFromString(gwutil.StringOfLength(' ', len(w.MiddleDec())))...)
	}
	line = append(line, gowid.CellsFromString(w.RightDec())...)

	res := gowid.NewCanvasWithLines([][]gowid.Cell{line})
	res.SetCursorCoords(len(w.LeftDec())+(len(w.MiddleDec())/2), 0)

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
