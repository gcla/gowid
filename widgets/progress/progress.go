// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package progress provides a simple progress bar.
package progress

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/hpadding"
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
	// Progress returns the number of units completed.
	Progress() int
	// Target returns the number of units required overall.
	Target() int
	// Normal is used to render the incomplete part of the progress bar.
	Normal() gowid.ICellStyler
	// Complete is used to render the complete part of the progress bar.
	Complete() gowid.ICellStyler
}

// For callback registration
type ProgressCB struct{}
type TargetCB struct{}

// Widget is the concrete type of a progressbar widget.
type Widget struct {
	Current, Done    int
	normal, complete gowid.ICellStyler
	Callbacks        *gowid.Callbacks
	gowid.RejectUserInput
	gowid.NotSelectable
}

// Options is used for passing arguments to the progressbar initializer, New().
type Options struct {
	Normal, Complete gowid.ICellStyler
	Target, Current  int
}

// New will return an initialized progressbar Widget/
func New(args Options) *Widget {
	if args.Target == 0 {
		args.Target = 100
	}
	res := &Widget{
		Current:   args.Current,
		Done:      args.Target,
		normal:    args.Normal,
		complete:  args.Complete,
		Callbacks: gowid.NewCallbacks(),
	}
	var _ IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("progress")
}

func (w *Widget) Text() string {
	var percent int
	if w.Done == 0 {
		percent = 100
	} else {
		percent = gwutil.Min(100, gwutil.Max(0, w.Current*100/w.Done))
	}
	return fmt.Sprintf("%d %%", percent)
}

func (w *Widget) OnSetProgress(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ProgressCB{}, f)
}

func (w *Widget) RemoveOnSetProgress(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ProgressCB{}, f)
}

func (w *Widget) OnSetTarget(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, TargetCB{}, f)
}

func (w *Widget) RemoveOnSetTarget(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, TargetCB{}, f)
}

func (w *Widget) SetProgress(app gowid.IApp, current int) {
	w.Current = current
	if w.Current > w.Done {
		w.Current = w.Done
	} else if w.Current < 0 {
		w.Current = 0
	}
	gowid.RunWidgetCallbacks(w.Callbacks, ProgressCB{}, app, w)
}

func (w *Widget) SetTarget(app gowid.IApp, target int) {
	w.Done = target
	if w.Done < 0 {
		w.Done = 0
	}
	if w.Current > w.Done {
		w.Current = w.Done
	}
	gowid.RunWidgetCallbacks(w.Callbacks, TargetCB{}, app, w)
}

func (w *Widget) Progress() int {
	return w.Current
}

func (w *Widget) Target() int {
	return w.Done
}

func (w *Widget) Normal() gowid.ICellStyler {
	return w.normal
}

func (w *Widget) Complete() gowid.ICellStyler {
	return w.complete
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

	barCanvas :=
		styled.New(
			text.New(gwutil.StringOfLength(' ', cols)),
			w.Normal(),
		).Render(

			gowid.RenderBox{C: cols, R: 1}, gowid.NotSelected, app)

	fnorm, _, _ := w.Normal().GetStyle(app)
	percentStyle := gowid.MakePaletteEntry(fnorm, gowid.NoColor{})

	fcomp, bcomp, scomp := w.Complete().GetStyle(app)
	fcompCol := gowid.IColorToTCell(fcomp, gowid.ColorNone, app.GetColorMode())
	bcompCol := gowid.IColorToTCell(bcomp, gowid.ColorNone, app.GetColorMode())

	cur, done := w.Progress(), w.Target()
	var cutoff int
	if done == 0 {
		cutoff = cols
	} else {
		cutoff = (cur * cols) / done
	}
	for i := 0; i < cutoff; i++ {
		barCanvas.SetCellAt(i, 0, barCanvas.CellAt(i, 0).WithForegroundColor(fcompCol).WithBackgroundColor(bcompCol).WithStyle(scomp))
	}

	percent := hpadding.New(
		styled.New(
			text.New(w.Text()),
			percentStyle,
		),
		gowid.HAlignMiddle{}, gowid.RenderFixed{},
	)
	percentCanvas := percent.Render(gowid.RenderBox{C: cols, R: 1}, gowid.NotSelected, app)
	barCanvas.MergeUnder(percentCanvas, 0, 0, false)

	return barCanvas
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
