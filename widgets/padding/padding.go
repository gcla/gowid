// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package padding provides a widget that pads an inner widget on the sides, above and below
package padding

import (
	"errors"
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/tcell"
)

//======================================================================

type IWidget interface {
	gowid.ICompositeWidget
	HAlign() gowid.IHAlignment
	Width() gowid.IWidgetDimension
	VAlign() gowid.IVAlignment
	Height() gowid.IWidgetDimension
}

// Widget renders the wrapped widget with the provided
// width; if the wrapped widget is a box, or the wrapped widget is to be
// packed to a width smaller than specified, the wrapped widget can be
// aligned in the middle, left or right
type Widget struct {
	inner     gowid.IWidget
	vAlign    gowid.IVAlignment
	height    gowid.IWidgetDimension
	hAlign    gowid.IHAlignment
	width     gowid.IWidgetDimension
	opts      Options
	Callbacks *gowid.Callbacks
	gowid.FocusCallbacks
	gowid.SubWidgetCallbacks
}

type Options struct{}

func New(inner gowid.IWidget,
	valign gowid.IVAlignment, height gowid.IWidgetDimension,
	halign gowid.IHAlignment, width gowid.IWidgetDimension,
	opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	res := &Widget{
		inner:     inner,
		vAlign:    valign,
		height:    height,
		hAlign:    halign,
		width:     width,
		opts:      opt,
		Callbacks: gowid.NewCallbacks(),
	}
	//var _ gowid.IWidget = res
	//var _ IWidgetSettable = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("hpad[%v]", w.SubWidget())
}

func (w *Widget) Selectable() bool {
	return w.inner.Selectable()
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.inner
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.inner = wi
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) OnSetAlign(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, gowid.HAlignCB{}, f)
}

func (w *Widget) RemoveOnSetAlign(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, gowid.HAlignCB{}, f)
}

func (w *Widget) HAlign() gowid.IHAlignment {
	return w.hAlign
}

func (w *Widget) SetHAlign(i gowid.IHAlignment, app gowid.IApp) {
	w.hAlign = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.HAlignCB{}, app, w)
}

func (w *Widget) VAlign() gowid.IVAlignment {
	return w.vAlign
}

func (w *Widget) SetVAlign(i gowid.IVAlignment, app gowid.IApp) {
	w.vAlign = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.VAlignCB{}, app, w)
}

func (w *Widget) OnSetHeight(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, gowid.WidthCB{}, f)
}

func (w *Widget) RemoveOnSetHeight(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, gowid.WidthCB{}, f)
}

func (w *Widget) Width() gowid.IWidgetDimension {
	return w.width
}

func (w *Widget) SetWidth(i gowid.IWidgetDimension, app gowid.IApp) {
	w.width = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.WidthCB{}, app, w)
}

func (w *Widget) Height() gowid.IWidgetDimension {
	return w.height
}

func (w *Widget) SetHeight(i gowid.IWidgetDimension, app gowid.IApp) {
	w.height = i
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.HeightCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	size2 := size
	// If there is a horizontal offset specified, the relative features should reduce the size of the
	// supplied size i.e. it should be relative to the reduced screen size
	switch al := w.HAlign().(type) {
	case gowid.HAlignLeft:
		switch s := size.(type) {
		case gowid.IRenderBox:
			size2 = gowid.RenderBox{C: s.BoxColumns() - al.Margin, R: s.BoxRows()}
		case gowid.IRenderFlowWith:
			size2 = gowid.RenderFlowWith{C: s.FlowColumns() - al.Margin}
		default:
		}
	default:
	}

	return gowid.ComputeSubSizeUnsafe(size2, w.Width(), w.Height())
	//return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

// func SubWidgetSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
// 	size2 := size
// 	// If there is a horizontal offset specified, the relative features should reduce the size of the
// 	// supplied size i.e. it should be relative to the reduced screen size
// 	switch al := w.Align().(type) {
// 	case gowid.HAlignLeft:
// 		switch s := size.(type) {
// 		case gowid.IRenderBox:
// 			size2 = gowid.RenderBox{C: s.BoxColumns() - al.Margin, R: s.BoxRows()}
// 		case gowid.IRenderFlowWith:
// 			size2 = gowid.RenderFlowWith{C: s.FlowColumns() - al.Margin}
// 		default:
// 		}
// 	default:
// 	}

// 	return gowid.ComputeHorizontalSubSizeUnsafe(size2, w.Width())
// }

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {

	subSize := w.SubWidgetSize(size, focus, app)

	subWidgetCanvas := gowid.Render(w.SubWidget(), subSize, focus, app)
	subWidgetMaxColumn := subWidgetCanvas.BoxColumns()

	var myCols int
	if cols, ok := size.(gowid.IColumns); ok {
		myCols = cols.Columns()
	} else {
		myCols = subWidgetMaxColumn
	}

	if myCols < subWidgetMaxColumn {
		// TODO - bad, mandates trimming on right
		subWidgetCanvas.TrimRight(myCols)
	} else if myCols > subWidgetMaxColumn {
		switch al := w.HAlign().(type) {
		case gowid.HAlignRight:
			subWidgetCanvas.ExtendLeft(gowid.EmptyLine(myCols - subWidgetMaxColumn))
		case gowid.HAlignMiddle:
			r := (myCols - subWidgetMaxColumn) / 2
			l := myCols - (subWidgetMaxColumn + r)
			subWidgetCanvas.ExtendRight(gowid.EmptyLine(r))
			subWidgetCanvas.ExtendLeft(gowid.EmptyLine(l))
		case gowid.HAlignLeft:
			l := gwutil.Min(al.Margin, myCols-subWidgetMaxColumn)
			r := myCols - (l + subWidgetMaxColumn)
			subWidgetCanvas.ExtendRight(gowid.EmptyLine(r))
			subWidgetCanvas.ExtendLeft(gowid.EmptyLine(l))
		default:
			panic(fmt.Errorf("Invalid horizontal alignment setting %v of type %T", al, al))
		}
	}

	maxCol := subWidgetCanvas.BoxColumns()
	subWidgetRows := subWidgetCanvas.BoxRows()
	fill := fill.NewEmpty()
	var rowsToUseInResult int

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
			switch al := w.VAlign().(type) {
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
		panic(fmt.Errorf("Unknown size %v", size))
	}

	switch al := w.VAlign().(type) {
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

	gowid.MakeCanvasRightSize(subWidgetCanvas, size)

	return subWidgetCanvas
}

// UserInput will adjust the input event's x coordinate depending on the input size
// and widget alignment. If the input is e.g. IRenderFixed, then no adjustment is
// made.
func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	//return false
	rSize := gowid.RenderSize(w, size, focus, app)

	subSize := w.SubWidgetSize(size, focus, app)
	ss := w.SubWidget().RenderSize(subSize, focus, app)

	sCols := ss.BoxColumns()
	sRows := ss.BoxRows()

	//cols2, ok := size.(gowid.IColumns)
	cols2 := rSize.BoxColumns()
	rows2 := rSize.BoxRows()

	var xd int

	//if ok {
	switch al := w.HAlign().(type) {
	case gowid.HAlignRight:
		xd = -(cols2 - sCols)
	case gowid.HAlignMiddle:
		r := (cols2 - sCols) / 2
		l := cols2 - (sCols + r)
		xd = -l
	case gowid.HAlignLeft:
		if al.Margin+sCols <= cols2 {
			xd = -al.Margin
		} else {
			xd = -gwutil.Max(0, cols2-sCols)
		}
	}

	var yd int

	switch al := w.VAlign().(type) {
	case gowid.VAlignBottom:
		yd = sRows - rows2
	case gowid.VAlignMiddle:
		yd = (sRows - rows2) / 2
	case gowid.VAlignTop:
		if rows2 > sRows+al.Margin {
			yd = -al.Margin
		} else if rows2 > al.Margin {
			yd = -al.Margin
		} else {
			yd = -(rows2 - 1)
		}
	}

	//}
	newev := gowid.TranslatedMouseEvent(ev, xd, yd)

	// TODO - don't need to translate event for keyboard event...
	if evm, ok := newev.(*tcell.EventMouse); ok {
		mx, transY := evm.Position()
		if mx >= 0 && mx < sCols {
			if transY < sRows && transY >= 0 {
				return gowid.UserInputIfSelectable(w.SubWidget(), newev, subSize, focus, app)
			}
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), newev, subSize, focus, app)
	}

	return false
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
