// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package columns provides a widget for organizing other widgets in columns.
package columns

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gdamore/tcell"
)

//======================================================================

type IWidget interface {
	gowid.ICompositeMultipleWidget
	gowid.ISettableDimensions
	gowid.ISettableSubWidgets
	gowid.IFindNextSelectable
	gowid.IPreferedPosition
	gowid.ISelectChild
	gowid.IIdentity
	WidgetWidths(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []int
	Wrap() bool
}

type Widget struct {
	widgets []gowid.IContainerWidget
	focus   int // -1 means nothing selectable
	prefCol int // caches the last set prefered col. Passes it on if widget hasn't changed focus
	opt     Options
	*gowid.Callbacks
	gowid.AddressProvidesID
	gowid.SubWidgetsCallbacks
	gowid.FocusCallbacks
}

type Options struct {
	StartColumn      int  // column that gets initial focus
	Wrap             bool // whether or not to wrap from last column to first with movement operations
	DoNotSetSelected bool // Whether or not to set the focus.Selected field for the selected child
}

func New(widgets []gowid.IContainerWidget, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = Options{
			StartColumn: -1,
		}
	}
	res := &Widget{
		widgets: widgets,
		focus:   -1,
		prefCol: -1,
		opt:     opt,
	}
	res.SubWidgetsCallbacks = gowid.SubWidgetsCallbacks{CB: &res.Callbacks}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}

	if opt.StartColumn >= 0 {
		res.focus = gwutil.Min(opt.StartColumn, len(widgets)-1)
	} else {
		res.focus, _ = res.FindNextSelectable(1, res.Wrap())
	}

	var _ gowid.IWidget = res
	var _ IWidget = res
	var _ gowid.ICompositeMultipleDimensions = res
	var _ gowid.ICompositeMultipleWidget = res

	return res
}

//func Simple(ws ...gowid.IWidget) *Widget {
func NewFlow(ws ...interface{}) *Widget {
	return NewWithDim(gowid.RenderFlow{}, ws...)
}

func NewFixed(ws ...interface{}) *Widget {
	return NewWithDim(gowid.RenderFixed{}, ws...)
}

func NewWithDim(method gowid.IWidgetDimension, ws ...interface{}) *Widget {
	cws := make([]gowid.IContainerWidget, len(ws))
	for i := 0; i < len(ws); i++ {
		if cw, ok := ws[i].(gowid.IContainerWidget); ok {
			cws[i] = cw
		} else {
			cws[i] = &gowid.ContainerWidget{
				IWidget: ws[i].(gowid.IWidget),
				D:       method,
			}
		}
	}
	return New(cws)
}

func (w *Widget) SelectChild(f gowid.Selector) bool {
	return !w.opt.DoNotSetSelected && f.Selected
}

func (w *Widget) String() string {
	cols := make([]string, len(w.widgets))
	for i := 0; i < len(cols); i++ {
		cols[i] = fmt.Sprintf("%v", w.widgets[i])
	}
	return fmt.Sprintf("columns[%s]", strings.Join(cols, ","))
}

func (w *Widget) SetFocus(app gowid.IApp, i int) {
	old := w.focus
	w.focus = gwutil.Min(gwutil.Max(i, 0), len(w.widgets)-1)
	w.prefCol = -1 // moved, so pass on real focus from now on
	if old != w.focus {
		gowid.RunWidgetCallbacks(w.Callbacks, gowid.FocusCB{}, app, w)
	}
}

func (w *Widget) Wrap() bool {
	return w.opt.Wrap
}

func (w *Widget) Focus() int {
	return w.focus
}

func (w *Widget) SubWidgets() []gowid.IWidget {
	res := make([]gowid.IWidget, len(w.widgets))
	for i, iw := range w.widgets {
		res[i] = iw
	}
	return res
}

func (w *Widget) SetSubWidgets(widgets []gowid.IWidget, app gowid.IApp) {
	ws := make([]gowid.IContainerWidget, len(widgets))
	for i, iw := range widgets {
		if iwc, ok := iw.(gowid.IContainerWidget); ok {
			ws[i] = iwc
		} else {
			ws[i] = &gowid.ContainerWidget{IWidget: iw, D: gowid.RenderFlow{}}
		}
	}
	oldFocus := w.Focus()
	w.widgets = ws
	w.SetFocus(app, oldFocus)
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetsCB{}, app, w)
}

func (w *Widget) Dimensions() []gowid.IWidgetDimension {
	res := make([]gowid.IWidgetDimension, len(w.widgets))
	for i, iw := range w.widgets {
		res[i] = iw.Dimension()
	}
	return res
}

func (w *Widget) SetDimensions(dimensions []gowid.IWidgetDimension, app gowid.IApp) {
	for i, id := range dimensions {
		w.widgets[i].SetDimension(id)
	}
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.DimensionsCB{}, app, w)
}

func (w *Widget) Selectable() bool {
	return gowid.SelectableIfAnySubWidgetsAre(w)
}

func (w *Widget) FindNextSelectable(dir gowid.Direction, wrap bool) (int, bool) {
	return gowid.FindNextSelectableFrom(w, w.Focus(), dir, wrap)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

// RenderSize computes the size of this widget when it renders. This is
// done by computing the sizes of each subwidget, then arranging them the
// same way that Render() does.
//
func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) RenderSubWidgets(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.ICanvas {
	return RenderSubWidgets(w, size, focus, focusIdx, app)
}

func (w *Widget) RenderedSubWidgetsSizes(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.IRenderBox {
	return RenderedSubWidgetsSizes(w, size, focus, focusIdx, app)
}

// Return a slice of ints representing the width in columns for each of the subwidgets to be rendered
// in this context given the size argument.
//
func (w *Widget) WidgetWidths(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []int {
	return WidgetWidths(w, size, focus, focusIdx, app)
}

// Construct the context in which each subwidget will be rendered. It's important to
// preserve the type of context e.g. a subwidget may only support being rendered in a
// fixed context. The newX parameter is the width the subwidget will have within the
// context of the Columns widget.
//
func (w *Widget) SubWidgetSize(size gowid.IRenderSize, newX int, sub gowid.IWidget, dim gowid.IWidgetDimension) gowid.IRenderSize {
	return SubWidgetSize(size, newX, dim)
}

func (w *Widget) GetPreferedPosition() gwutil.IntOption {
	f := w.prefCol
	if f == -1 {
		f = w.Focus()
	}
	if f == -1 {
		return gwutil.NoneInt()
	} else {
		return gwutil.SomeInt(f)
	}
}

func (w *Widget) SetPreferedPosition(cols int, app gowid.IApp) {
	col := gwutil.Min(gwutil.Max(cols, 0), len(w.widgets)-1)
	pref := col
	colLeft := col - 1
	colRight := col
	for colLeft >= 0 || colRight < len(w.widgets) {
		if colRight < len(w.widgets) && w.widgets[colRight].Selectable() {
			w.SetFocus(app, colRight)
			break
		} else {
			colRight++
		}
		if colLeft >= 0 && w.widgets[colLeft].Selectable() {
			w.SetFocus(app, colLeft)
			break
		} else {
			colLeft--
		}
	}
	w.prefCol = pref // Save it. Pass it on if widget doesn't change col before losing focus.
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func SubWidgetSize(size gowid.IRenderSize, newX int, dim gowid.IWidgetDimension) gowid.IRenderSize {
	var subSize gowid.IRenderSize

	switch sz := size.(type) {
	case gowid.IRenderFixed:
		switch dim.(type) {
		case gowid.IRenderBox:
			subSize = dim
		default:
			subSize = gowid.RenderFixed{}
		}
	case gowid.IRenderBox:
		switch dim.(type) {
		case gowid.IRenderFixed:
			subSize = gowid.RenderFixed{}
		case gowid.IRenderFlow:
			subSize = gowid.RenderFlowWith{C: newX}
		case gowid.IRenderWithUnits, gowid.IRenderWithWeight:
			subSize = gowid.RenderBox{C: newX, R: sz.BoxRows()}
		default:
			subSize = gowid.RenderBox{C: newX, R: sz.BoxRows()}
		}
	case gowid.IRenderFlowWith:
		switch dim.(type) {
		case gowid.IRenderFixed:
			subSize = gowid.RenderFixed{}
		case gowid.IRenderFlow, gowid.IRenderWithUnits, gowid.IRenderWithWeight, gowid.IRenderRelative:
			// The newX argument is already computed to be the right number of cols for the subwidget
			subSize = gowid.RenderFlowWith{C: newX}
		default:
			panic(gowid.DimensionError{Size: size, Dim: dim})
		}
	default:
		panic(gowid.DimensionError{Size: size, Dim: dim})
	}
	return subSize
}

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := true
	subfocus := w.Focus()

	subSizes := w.WidgetWidths(size, focus, subfocus, app)

	dims := w.Dimensions()
	subs := w.SubWidgets()

	forChild := false

	if subfocus != -1 {
		if evm, ok := ev.(*tcell.EventMouse); ok {
			curX := 0
			mx, _ := evm.Position()
		Loop:
			for i, c := range subSizes {
				if mx < curX+c && mx >= curX {
					subSize := w.SubWidgetSize(size, c, subs[i], dims[i])
					forChild = subs[i].UserInput(gowid.TranslatedMouseEvent(ev, -curX, 0), subSize, focus.SelectIf(w.SelectChild(focus) && i == subfocus), app)

					// Give the child focus if (a) it's selectable, and (b) if this is the click up corresponding
					// to a previous click down on this columns widget.
					switch evm.Buttons() {
					case tcell.Button1, tcell.Button2, tcell.Button3:
						app.SetClickTarget(evm.Buttons(), w)
					case tcell.ButtonNone:
						if !app.GetLastMouseState().NoButtonClicked() {
							if subs[i].Selectable() {
								clickit := false
								app.ClickTarget(func(k tcell.ButtonMask, v gowid.IIdentityWidget) {
									if v != nil && v.ID() == w.ID() {
										clickit = true
									}
								})
								if clickit {
									w.SetFocus(app, i)
								}
							}
						}
						break Loop
					}
					break
				}
				curX += c
			}
		} else {
			subC := subSizes[subfocus] // guaranteed to be a box
			subSize := w.SubWidgetSize(size, subC, subs[subfocus], dims[subfocus])
			forChild = gowid.UserInputIfSelectable(subs[w.Focus()], ev, subSize, focus, app)
		}
	}

	if !forChild && w.Focus() != -1 {
		res = false
		if evk, ok := ev.(*tcell.EventKey); ok {

			curw := subs[w.Focus()]
			prefPos := gowid.PrefPosition(curw)

			switch evk.Key() {
			case tcell.KeyRight, tcell.KeyCtrlF:
				res = Scroll(w, 1, w.Wrap(), app)
			case tcell.KeyLeft, tcell.KeyCtrlB:
				res = Scroll(w, -1, w.Wrap(), app)
			}

			if !prefPos.IsNone() {
				// New focus widget
				curw = subs[w.Focus()]
				gowid.SetPrefPosition(curw, prefPos.Val(), app)
			}

		}
	}

	return res
}

type IFocusSelectable interface {
	gowid.IFocus
	gowid.IFindNextSelectable
}

func Scroll(w IFocusSelectable, dir gowid.Direction, wrap bool, app gowid.IApp) bool {
	res := false
	next, ok := w.FindNextSelectable(dir, wrap)
	if ok {
		w.SetFocus(app, next)
		res = true
	}
	return res
}

type ICompositeMultipleDimensionsExt interface {
	gowid.ICompositeMultipleDimensions
	gowid.ISelectChild
}

func WidgetWidths(w ICompositeMultipleDimensionsExt, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []int {
	return widgetWidthsExt(w, w.SubWidgets(), w.Dimensions(), size, focus, focusIdx, app)
}

// Precompute dims and subs
func widgetWidthsExt(w gowid.ISelectChild, subs []gowid.IWidget, dims []gowid.IWidgetDimension, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []int {
	lenw := len(subs)

	res := make([]int, lenw)
	helper := make([]bool, lenw)

	haveColsTotal := false
	var colsTotal int
	if _, ok := size.(gowid.IRenderFixed); !ok {
		cols, ok := size.(gowid.IColumns)
		if !ok {
			panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
		}
		colsTotal = cols.Columns()
		haveColsTotal = true
	}

	colsUsed := 0
	totalWeight := 0

	trunc := func(x *int) {
		if haveColsTotal && colsUsed+*x > colsTotal {
			*x = colsTotal - colsUsed
		}
	}

	// First, render the widgets whose width is known
	for i := 0; i < lenw; i++ {
		// This doesn't support IRenderFlow. That type comes with an associated width e.g.
		// "Flow with 25 columns". We don't have any way to apportion those columns amongst
		// the overall width for the widget.
		switch w2 := dims[i].(type) {
		case gowid.IRenderFixed:
			c := gowid.RenderSize(subs[i], gowid.RenderFixed{}, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
			res[i] = c.BoxColumns()
			trunc(&res[i])
			colsUsed += res[i]
			helper[i] = true
		case gowid.IRenderBox:
			res[i] = w2.BoxColumns()
			trunc(&res[i])
			colsUsed += res[i]
			helper[i] = true
		case gowid.IRenderFlowWith:
			res[i] = w2.FlowColumns()
			trunc(&res[i])
			colsUsed += res[i]
			helper[i] = true
		case gowid.IRenderWithUnits:
			res[i] = w2.Units()
			trunc(&res[i])
			colsUsed += res[i]
			helper[i] = true
		case gowid.IRenderRelative:
			cols, ok := size.(gowid.IColumns)
			if !ok {
				panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
			}
			res[i] = int((w2.Relative() * float64(cols.Columns())) + 0.5)
			trunc(&res[i])
			colsUsed += res[i]
			helper[i] = true
		case gowid.IRenderWithWeight:
			// widget must be weighted
			totalWeight += w2.Weight()
			helper[i] = false
		default:
			panic(gowid.DimensionError{Size: size, Dim: w2})
		}
	}

	var colsLeft int
	var colsToDivideUp int
	if haveColsTotal {
		colsToDivideUp = colsTotal - colsUsed
		colsLeft = colsToDivideUp
	}

	// Now, divide up the remaining space among the weight columns
	lasti := -1
	for {
		if colsLeft == 0 {
			break
		}
		doneone := false
		totalWeight = 0
		for i := 0; i < lenw; i++ {
			if w2, ok := dims[i].(gowid.IRenderWithWeight); ok && !helper[i] {
				totalWeight += w2.Weight()
			}
		}
		colsToDivideUp = colsLeft
		for i := 0; i < lenw; i++ {
			// Can only be weight here if !helper[i] ; but not sufficient for it to be eligible
			if !helper[i] {
				cols := int(((float32(dims[i].(gowid.IRenderWithWeight).Weight()) / float32(totalWeight)) * float32(colsToDivideUp)) + 0.5)
				if max, ok := dims[i].(gowid.IRenderMaxUnits); ok {
					if cols > max.MaxUnits() {
						cols = max.MaxUnits()
						helper[i] = true // this one is done
					}
				}
				if cols > colsLeft {
					cols = colsLeft
				}
				if cols > 0 {
					if res[i] == -1 {
						res[i] = 0
					}
					res[i] += cols
					colsLeft -= cols
					lasti = i
					doneone = true
				}
			}
		}
		if !doneone {
			break
		}
	}
	if lasti != -1 && colsLeft > 0 {
		res[lasti] += colsLeft
	}

	return res
}

func RenderSize(w gowid.ICompositeMultipleWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	subfocus := w.Focus()
	sizes := w.RenderedSubWidgetsSizes(size, focus, subfocus, app)

	maxcol, maxrow := 0, 0

	for _, sz := range sizes {
		maxcol += sz.BoxColumns()
		maxrow = gwutil.Max(maxrow, sz.BoxRows())
	}

	if cols, ok := size.(gowid.IColumns); ok {
		maxcol = cols.Columns()
		if rows, ok2 := size.(gowid.IRows); ok2 {
			maxrow = rows.Rows()
		}
	}

	return gowid.RenderBox{maxcol, maxrow}
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	res := gowid.NewCanvas()
	subfocus := w.Focus()
	canvases := w.RenderSubWidgets(size, focus, subfocus, app)

	subs := w.SubWidgets()

	// Assemble subcanvases into final canvas
	for i := 0; i < len(subs); i++ {
		diff := res.BoxRows() - canvases[i].BoxRows()
		if diff > 0 {
			fill := fill.NewEmpty()
			fc := fill.Render(gowid.RenderBox{canvases[i].BoxColumns(), diff}, gowid.NotSelected, app)
			canvases[i].AppendBelow(fc, false, false)
		} else if diff < 0 {
			fill := fill.NewEmpty()
			fc := fill.Render(gowid.RenderBox{res.BoxColumns(), -diff}, gowid.NotSelected, app)
			res.AppendBelow(fc, false, false)
		}
		res.AppendRight(canvases[i], i == subfocus)
	}

	if cols, ok := size.(gowid.IColumns); ok {
		res.ExtendRight(gowid.EmptyLine(cols.Columns() - res.BoxColumns()))
		if rows, ok2 := size.(gowid.IRenderBox); ok2 && res.BoxRows() < rows.BoxRows() {
			gowid.AppendBlankLines(res, rows.BoxRows()-res.BoxRows())
		}
	}

	gowid.MakeCanvasRightSize(res, size)

	return res
}

var AllChildrenMaxDimension = fmt.Errorf("All columns widgets were rendered Max, so there is no max height to use.")

// RenderSubWidgets returns an array of canvases for each of the subwidgets, rendering them
// with in the context of a column with the provided size and focus.
func RenderSubWidgets(w IWidget, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.ICanvas {
	subs := w.SubWidgets()
	dims := w.Dimensions()
	l := len(subs)
	canvases := make([]gowid.ICanvas, l)

	if l == 0 {
		return canvases
	}

	weights := w.WidgetWidths(size, focus, focusIdx, app)

	maxes := make([]int, 0, l)
	ssizes := make([]gowid.IRenderSize, 0, l)
	curMax := -1

	for i := 0; i < l; i++ {
		subSize := w.SubWidgetSize(size, weights[i], subs[i], dims[i])
		if _, ok := dims[i].(gowid.IRenderMax); ok {
			maxes = append(maxes, i)
			ssizes = append(ssizes, subSize)
		} else {
			canvases[i] = subs[i].Render(subSize, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
			if canvases[i].BoxRows() > curMax {
				curMax = canvases[i].BoxRows()
			}
		}
	}

	if curMax == -1 {
		panic(AllChildrenMaxDimension)
	}

	for j := 0; j < len(maxes); j++ {
		i := maxes[j]
		var mss gowid.IRenderSize = ssizes[j]
		switch css := mss.(type) {
		case gowid.IRenderFlowWith:
			mss = gowid.MakeRenderBox(css.FlowColumns(), curMax)
		case gowid.IRenderBox:
			mss = gowid.MakeRenderBox(css.BoxColumns(), curMax)
		default:
		}
		canvases[i] = subs[i].Render(mss, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
	}

	return canvases
}

// RenderedSubWidgetsSizes returns an array of boxes that bound each of the subwidgets as they
// would be rendered with the given size and focus.
func RenderedSubWidgetsSizes(w IWidget, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.IRenderBox {
	subs := w.SubWidgets()
	dims := w.Dimensions()
	l := len(subs)

	res := make([]gowid.IRenderBox, l)
	weights := w.WidgetWidths(size, focus, focusIdx, app)

	maxes := make([]int, 0, l)
	ssizes := make([]int, 0, l)
	curMax := -1

	for i := 0; i < l; i++ {
		subSize := w.SubWidgetSize(size, weights[i], subs[i], dims[i])

		if _, ok := dims[i].(gowid.IRenderMax); ok {
			maxes = append(maxes, i)
			ssizes = append(ssizes, weights[i])
		} else {
			c := subs[i].RenderSize(subSize, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
			res[i] = gowid.RenderBox{weights[i], c.BoxRows()}

			if res[i].BoxRows() > curMax {
				curMax = res[i].BoxRows()
			}
		}
	}

	if curMax == -1 {
		panic(AllChildrenMaxDimension)
	}

	for j := 0; j < len(maxes); j++ {
		res[maxes[j]] = gowid.RenderBox{ssizes[j], curMax}
	}

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
