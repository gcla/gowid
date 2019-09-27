// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package vpadding provides a widget that pads an inner widget on the top and bottom.
package vpadding

import (
	"errors"
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gdamore/tcell"
)

//======================================================================

type IVerticalPadding interface {
	Align() gowid.IVAlignment
	Height() gowid.IWidgetDimension
}

type IWidget interface {
	gowid.ICompositeWidget
	IVerticalPadding
}

// Widget wraps a widget and aligns it vertically according to the supplied arguments. The wrapped
// widget can be aligned to the top, bottom or middle, and can be provided with a specific height in #lines.
//
type Widget struct {
	gowid.IWidget
	alignment gowid.IVAlignment
	height    gowid.IWidgetDimension
	*gowid.Callbacks
	gowid.FocusCallbacks
	gowid.SubWidgetCallbacks
}

func New(inner gowid.IWidget, alignment gowid.IVAlignment, height gowid.IWidgetDimension) *Widget {
	res := &Widget{
		IWidget:   inner,
		alignment: alignment,
		height:    height,
	}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}

	var _ gowid.IWidget = res

	return res
}

func NewBox(inner gowid.IWidget, rows int) *Widget {
	return New(inner, gowid.VAlignTop{}, gowid.RenderWithUnits{U: rows})
}

func (w *Widget) String() string {
	return fmt.Sprintf("vpad[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) OnSetAlign(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, gowid.VAlignCB{}, f)
}

func (w *Widget) RemoveOnSetAlign(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, gowid.VAlignCB{}, f)
}

func (w *Widget) Align() gowid.IVAlignment {
	return w.alignment
}

func (w *Widget) SetAlign(i gowid.IVAlignment, app gowid.IApp) {
	w.alignment = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.VAlignCB{}, app, w)
}

func (w *Widget) OnSetHeight(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, gowid.HeightCB{}, f)
}

func (w *Widget) RemoveOnSetHeight(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, gowid.HeightCB{}, f)
}

func (w *Widget) Height() gowid.IWidgetDimension {
	return w.height
}

func (w *Widget) SetHeight(i gowid.IWidgetDimension, app gowid.IApp) {
	w.height = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.HeightCB{}, app, w)
}

// SubWidgetSize returns the size that will be passed down to the
// subwidget's Render(), based on the size passed to the current widget.
// If this widget is rendered in a Flow context and the vertical height
// specified is in Units, then the subwidget is rendered in a Box content
// with Units-number-of-rows. This gives the subwidget an opportunity to
// render to fill the space given to it, rather than risking truncation. If
// the subwidget cannot render in Box mode, then wrap it in a
// FlowToBoxWidget first.
//
func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func SubWidgetSize(w IVerticalPadding, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	size2 := size
	// If there is a vertical offset specified, the relative features should reduce the size of the
	// supplied size i.e. it should be relative to the reduced screen size
	switch al := w.Align().(type) {
	case gowid.VAlignTop:
		switch s := size.(type) {
		case gowid.IRenderBox:
			size2 = gowid.RenderBox{C: s.BoxColumns(), R: s.BoxRows() - al.Margin}
		}
	}
	// rows := -1
	// switch ss := w.Height().(type) {
	// case gowid.IRenderWithUnits:
	// 	rows = ss.Units()
	// }

	return gowid.ComputeVerticalSubSizeUnsafe(size2, w.Height(), -1, -1)
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	var subWidgetCanvas gowid.ICanvas
	var rowsToUseInResult int

	subSize := w.SubWidgetSize(size, focus, app)
	subWidgetCanvas = gowid.Render(w.SubWidget(), subSize, focus, app)
	subWidgetRows := subWidgetCanvas.BoxRows()

	// Compute number of rows to use in final canvas
	switch sz := size.(type) {
	case gowid.IRenderBox:
		rowsToUseInResult = sz.BoxRows()
	case gowid.IRenderFlowWith:
		switch w.Height().(type) {
		case gowid.IRenderFlow, gowid.IRenderFixed, gowid.IRenderWithUnits:
			rowsToUseInResult = subWidgetRows
		default:
			panic(fmt.Errorf("Height spec %v cannot be used in flow mode for %T", w.Height(), w))
		}
	case gowid.IRenderFixed:
		switch w.Height().(type) {
		case gowid.IRenderFlow, gowid.IRenderFixed:
			rowsToUseInResult = subWidgetRows
			switch al := w.Align().(type) {
			case gowid.VAlignTop:
				rowsToUseInResult += al.Margin
			}
		case gowid.IRenderWithUnits:
			rowsToUseInResult = w.Height().(gowid.IRenderWithUnits).Units()
		default:
			panic(fmt.Errorf("This spec %v of type %T cannot be used in flow mode for %T",
				w.Height(), w.Height(), w))
		}
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}

	maxCol := subWidgetCanvas.BoxColumns()
	fill := fill.NewEmpty()

	switch al := w.Align().(type) {
	case gowid.VAlignBottom:
		if rowsToUseInResult > subWidgetRows {
			fc := gowid.Render(fill, gowid.RenderBox{C: maxCol, R: rowsToUseInResult - subWidgetRows}, gowid.NotSelected, app)
			fc.AppendBelow(subWidgetCanvas, true, false)
			subWidgetCanvas = fc
		} else {
			subWidgetCanvas.Truncate(rowsToUseInResult-subWidgetRows, 0)
		}
	case gowid.VAlignMiddle:
		if rowsToUseInResult > subWidgetRows {
			topl := (rowsToUseInResult - subWidgetRows) / 2
			bottoml := rowsToUseInResult - (topl + subWidgetRows)
			fc1 := gowid.Render(fill, gowid.RenderBox{C: maxCol, R: topl}, gowid.NotSelected, app)
			fc2 := gowid.Render(fill, gowid.RenderBox{C: maxCol, R: bottoml}, gowid.NotSelected, app)

			fc1.AppendBelow(subWidgetCanvas, true, false)
			subWidgetCanvas = fc1
			subWidgetCanvas.AppendBelow(fc2, false, false)
		} else {
			topl := (subWidgetRows - rowsToUseInResult) / 2
			bottoml := subWidgetRows - (rowsToUseInResult + topl)
			subWidgetCanvas.Truncate(topl, bottoml)
		}
	case gowid.VAlignTop:
		if rowsToUseInResult > subWidgetRows+al.Margin {
			topl := al.Margin
			bottoml := rowsToUseInResult - (topl + subWidgetRows)
			fc1 := gowid.Render(fill, gowid.RenderBox{C: maxCol, R: topl}, gowid.NotSelected, app)
			fc2 := gowid.Render(fill, gowid.RenderBox{C: maxCol, R: bottoml}, gowid.NotSelected, app)
			fc1.AppendBelow(subWidgetCanvas, true, false)
			subWidgetCanvas = fc1
			subWidgetCanvas.AppendBelow(fc2, false, false)

		} else if rowsToUseInResult > al.Margin {
			topl := al.Margin
			bottoml := subWidgetRows - (rowsToUseInResult - al.Margin)

			subWidgetCanvas.Truncate(0, bottoml)
			fc1 := gowid.Render(fill, gowid.RenderBox{C: maxCol, R: topl}, gowid.NotSelected, app)
			fc1.AppendBelow(subWidgetCanvas, true, false)
			subWidgetCanvas = fc1
		} else {
			topl := rowsToUseInResult
			subWidgetCanvas = gowid.Render(fill, gowid.RenderBox{C: maxCol, R: topl}, gowid.NotSelected, app)
		}

	default:
		panic(errors.New("Invalid vertical alignment setting"))
	}

	// The embedded widget might have rendered with a different width
	gowid.MakeCanvasRightSize(subWidgetCanvas, size)

	return subWidgetCanvas
}

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {

	subSize := w.SubWidgetSize(size, focus, app)
	subWidgetBox := w.SubWidget().RenderSize(subSize, focus, app)
	subWidgetRows := subWidgetBox.BoxRows()

	myBox := w.RenderSize(size, focus, app)
	rowsToUseInResult := myBox.BoxRows()

	var yd int

	switch al := w.Align().(type) {
	case gowid.VAlignBottom:
		yd = subWidgetRows - rowsToUseInResult
	case gowid.VAlignMiddle:
		yd = (subWidgetRows - rowsToUseInResult) / 2
	case gowid.VAlignTop:
		if rowsToUseInResult > subWidgetRows+al.Margin {
			yd = -al.Margin
		} else if rowsToUseInResult > al.Margin {
			yd = -al.Margin
		} else {
			yd = -(rowsToUseInResult - 1)
		}
	}

	// Note that yd will be less than zero, so this translates upwards
	transEv := gowid.TranslatedMouseEvent(ev, 0, yd)

	if evm, ok := transEv.(*tcell.EventMouse); ok {
		_, transY := evm.Position()
		if transY < subWidgetRows && transY >= 0 {
			return gowid.UserInputIfSelectable(w.SubWidget(), transEv, subSize, focus, app)
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), transEv, subSize, focus, app)
	}
	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
