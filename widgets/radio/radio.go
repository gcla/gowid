// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package radio provides radio button widgets where one can be selected among many.
package radio

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/checkbox"
)

//======================================================================

type IWidget interface {
	gowid.IWidget
	gowid.ICallbacks
	IsChecked() bool
	Group() *[]IWidget
	SetStateInternal(selected bool)
}

type Widget struct {
	Selected  bool
	group     *[]IWidget
	Callbacks *gowid.Callbacks
	gowid.ClickCallbacks
	checkbox.Decoration
	gowid.AddressProvidesID
	gowid.IsSelectable
}

// If the group supplied is empty, this radio button will be marked as selected, regardless
// of the isChecked parameter.
func New(group *[]IWidget) *Widget {
	cb := gowid.NewCallbacks()
	res := &Widget{
		Selected:       false,
		group:          group,
		Callbacks:      cb,
		ClickCallbacks: gowid.ClickCallbacks{ICallbacks: cb},
		Decoration:     checkbox.Decoration{button.Decoration{"(", ")"}, "X"},
	}
	res.initRadioButton(group)

	var _ IWidget = res

	return res
}

func NewDecorated(group *[]IWidget, decoration checkbox.Decoration) *Widget {
	cb := gowid.NewCallbacks()
	res := &Widget{
		Selected:       false,
		group:          group,
		Callbacks:      cb,
		ClickCallbacks: gowid.ClickCallbacks{ICallbacks: cb},
		Decoration:     decoration,
	}
	res.initRadioButton(group)

	var _ gowid.IWidget = res

	return res
}

func (w *Widget) initRadioButton(group *[]IWidget) {
	*group = append(*group, w)
	if len(*group) == 1 {
		w.SetStateInternal(true)
	}
}

func (w *Widget) String() string {
	return fmt.Sprintf("radio[%s]", gwutil.If(w.IsChecked(), "X", " ").(string))
}

func (w *Widget) Select(app gowid.IApp) {
	Select(w, app)
}

func (w *Widget) Group() *[]IWidget {
	return w.group
}

// Don't ensure consistency of other widgets, but do issue callbacks for
// state change. TODO - need to do callbacks here to capture
// losing selection
func (w *Widget) SetStateInternal(selected bool) {
	w.Selected = selected
}

func (w *Widget) IsChecked() bool {
	return w.Selected
}

func (w *Widget) Click(app gowid.IApp) {
	if app.GetMouseState().NoButtonClicked() || app.GetMouseState().LeftIsClicked() {
		w.Select(app)
	}
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	if _, ok := size.(gowid.IRenderFixed); !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFixed"})
	}
	return gowid.RenderBox{C: len(w.LeftDec()) + len(w.RightDec()) + len(w.MiddleDec()), R: 1}
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	if _, ok := size.(gowid.IRenderFixed); !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFixed"})
	}

	res := checkbox.Render(w, size, focus, app)

	return res
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return button.UserInput(w, ev, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func Select(w IWidget, app gowid.IApp) {
	cur := w.IsChecked()
	if !cur {
		for _, w2 := range *w.Group() {
			if w != w2 && w2.IsChecked() {
				w2.SetStateInternal(false)
				gowid.RunWidgetCallbacks(w2, gowid.ClickCB{}, app, w2)
				break
			}
		}
		w.SetStateInternal(true)
		gowid.RunWidgetCallbacks(w, gowid.ClickCB{}, app, w)
	}
}

//======================================================================

// This is here to avoid import cycles
type RadioButtonTester struct {
	State bool
}

func (f *RadioButtonTester) Changed(app gowid.IApp, w gowid.IWidget, data ...interface{}) {
	rb := w.(*Widget)
	f.State = rb.Selected
}

func (f *RadioButtonTester) ID() interface{} { return "bar" }

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
