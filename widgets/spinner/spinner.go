// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package spinner provides a simple themable spinner.
package spinner

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
)

//======================================================================

// IWidget - if your widget implements progress.IWidget, you will be able to render it using the
// progress.Render() function.
//
type IWidget interface {
	gowid.IWidget
	// Text should return the string to be displayed inside the progress bar e.g. "50%"
	Text() string
	// Enabled returns true if spinner is spinning
	Enabled() bool
	// SetEnabled enables or disables the animation/spinning
	SetEnabled(bool, gowid.IApp)
	// Index returns the index at which to start drawing from the spinner characters
	Index() int
	// SpinnerLen returns the length of the set of chars used to draw the spinner
	SpinnerLen() int
	// Styler is used to render the incomplete part of the progress bar.
	Styler() gowid.ICellStyler
}

// Widget is the concrete type of a progressbar widget.
type Widget struct {
	enabled   bool
	label     string
	idx       int
	ticker    *time.Ticker
	stopChan  chan struct{}
	styler    gowid.ICellStyler
	Callbacks *gowid.Callbacks
	gowid.RejectUserInput
	gowid.NotSelectable
}

type ChangeStateCB struct{}

//var wave []rune = []rune("▁▃▄▅▆▇█▇▆▅▄▃")
//var wave []rune = []rune("◢■◤")

var wave []rune

func init() {
	if runtime.GOOS == "windows" {
		wave = []rune("▲     ")
	} else {
		wave = []rune("◤ ◢")
	}
}

// Options is used for passing arguments to the progressbar initializer, New().
type Options struct {
	Label  string
	Styler gowid.ICellStyler
}

// New will return an initialized spinner
func New(args Options) *Widget {
	res := &Widget{
		label:     args.Label,
		styler:    args.Styler,
		Callbacks: gowid.NewCallbacks(),
	}
	var _ IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("spinner")
}

func (w *Widget) Text() string {
	return w.label
}

func (w *Widget) Index() int {
	return w.idx
}

func (w *Widget) SpinnerLen() int {
	return len(wave)
}

func (w *Widget) OnChangeState(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ChangeStateCB{}, f)
}

func (w *Widget) RemoveOnChangeState(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ChangeStateCB{}, f)
}

func (w *Widget) Enabled() bool {
	return w.enabled
}

func (w *Widget) Update() {
	w.idx -= 1
	if w.idx < 0 {
		w.idx = len(wave) - 1
	}
}

func (w *Widget) SetEnabled(enabled bool, app gowid.IApp) {
	cur := w.enabled
	w.enabled = enabled

	if enabled != cur {
		gowid.RunWidgetCallbacks(w.Callbacks, ChangeStateCB{}, app, w)
	}
}

func (w *Widget) Styler() gowid.ICellStyler {
	return w.styler
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

// Render will render a progressbar IWidget.
func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	flow, isFlow := size.(gowid.IRenderFlowWith)
	if !isFlow {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFlowWith"})
	}
	cols := flow.FlowColumns()

	display := make([]rune, cols)
	wi := w.Index()
	for i := 0; i < cols; i++ {
		display[i] = wave[wi]
		wi += 1
		if wi == w.SpinnerLen() {
			wi = 0
		}
	}
	barCanvas := gowid.Render(
		styled.New(
			text.New(string(display)),
			w.Styler(),
		),
		gowid.RenderBox{C: cols, R: 1}, gowid.NotSelected, app,
	)

	return barCanvas
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
