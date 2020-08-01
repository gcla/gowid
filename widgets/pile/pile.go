// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package pile provides a widget for organizing other widgets in a vertical stack.
package pile

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/vim"
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
	RenderBoxMaker(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp, fn IPileBoxMaker) ([]gowid.IRenderBox, []gowid.IRenderSize)
	Wrap() bool
	KeyIsUp(*tcell.EventKey) bool
	KeyIsDown(*tcell.EventKey) bool
}

type Widget struct {
	widgets []gowid.IContainerWidget
	focus   int // -1 means nothing selectable
	prefRow int // caches the last set prefered row. Passes it on if widget hasn't changed focus
	opt     Options
	*gowid.Callbacks
	gowid.AddressProvidesID
	gowid.FocusCallbacks
	gowid.SubWidgetsCallbacks
}

type Options struct {
	StartRow         int
	Wrap             bool
	DoNotSetSelected bool // Whether or not to set the focus.Selected field for the selected child
	DownKeys         []vim.KeyPress
	UpKeys           []vim.KeyPress
}

var _ gowid.IWidget = (*Widget)(nil)
var _ IWidget = (*Widget)(nil)
var _ gowid.ICompositeMultipleDimensions = (*Widget)(nil)
var _ gowid.ICompositeMultipleWidget = (*Widget)(nil)

func New(widgets []gowid.IContainerWidget, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = Options{
			StartRow: -1,
		}
	}
	if opt.DownKeys == nil {
		opt.DownKeys = vim.AllDownKeys
	}
	if opt.UpKeys == nil {
		opt.UpKeys = vim.AllUpKeys
	}

	res := &Widget{
		widgets: widgets,
		focus:   -1,
		prefRow: -1,
		opt:     opt,
	}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}
	res.SubWidgetsCallbacks = gowid.SubWidgetsCallbacks{CB: &res.Callbacks}
	if opt.StartRow >= 0 {
		res.focus = gwutil.Min(opt.StartRow, len(widgets)-1)
	} else {
		res.focus, _ = res.FindNextSelectable(1, res.opt.Wrap)
	}

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
	rows := make([]string, len(w.widgets))
	for i := 0; i < len(rows); i++ {
		rows[i] = fmt.Sprintf("%v", w.widgets[i])
	}
	return fmt.Sprintf("pile[%s]", strings.Join(rows, ","))
}

// Tries to set at required index, will choose first selectable from there
func (w *Widget) SetFocus(app gowid.IApp, i int) {
	oldpos := w.focus
	w.focus = gwutil.Min(gwutil.Max(i, 0), len(w.widgets)-1)
	w.prefRow = -1 // moved, so pass on real focus from now on
	if oldpos != w.focus {
		gowid.RunWidgetCallbacks(w.Callbacks, gowid.FocusCB{}, app, w)
	}
}

func (w *Widget) Wrap() bool {
	return w.opt.Wrap
}

// Tries to set at required index, will choose first selectable from there
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

// TODO - widen each line to same width
func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) RenderedSubWidgetsSizes(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.IRenderBox {
	res, _ := RenderedChildrenSizes(w, size, focus, focusIdx, app)
	return res
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) RenderSubWidgets(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.ICanvas {
	return RenderSubwidgets(w, size, focus, focusIdx, app)
}

//
// TODO - widen each line to same width
// gcdoc - the fn argument is used to return either canvases or sizes, depending on whether
// the caller is rendering, or rendering subsizes
func (w *Widget) RenderBoxMaker(size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp, fn IPileBoxMaker) ([]gowid.IRenderBox, []gowid.IRenderSize) {
	return RenderBoxMaker(w, size, focus, focusIdx, app, fn)
}

// SubWidgetSize is the size that should be used to render a child widget, based on the size used to render the parent.
func (w *Widget) SubWidgetSize(size gowid.IRenderSize, newY int, sub gowid.IWidget, dim gowid.IWidgetDimension) gowid.IRenderSize {
	return gowid.ComputeVerticalSubSizeUnsafe(size, dim, -1, newY)
}

func (w *Widget) GetPreferedPosition() gwutil.IntOption {
	f := w.prefRow
	if f == -1 {
		f = w.Focus()
	}
	if f == -1 {
		return gwutil.NoneInt()
	} else {
		return gwutil.SomeInt(f)
	}
}

func (w *Widget) SetPreferedPosition(rows int, app gowid.IApp) {
	row := gwutil.Min(gwutil.Max(rows, 0), len(w.widgets)-1)
	pref := row
	rowLeft := row - 1
	rowRight := row
	for rowLeft >= 0 || rowRight < len(w.widgets) {
		if rowRight < len(w.widgets) && w.widgets[rowRight].Selectable() {
			w.SetFocus(app, rowRight)
			break
		} else {
			rowRight++
		}
		if rowLeft >= 0 && w.widgets[rowLeft].Selectable() {
			w.SetFocus(app, rowLeft)
			break
		} else {
			rowLeft--
		}
	}
	w.prefRow = pref // Save it. Pass it on if widget doesn't change col before losing focus.
}

func (w *Widget) KeyIsUp(evk *tcell.EventKey) bool {
	return vim.KeyIn(evk, w.opt.UpKeys)
}

func (w *Widget) KeyIsDown(evk *tcell.EventKey) bool {
	return vim.KeyIn(evk, w.opt.DownKeys)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {

	subfocus := w.Focus()
	// An array of IRenderBoxes
	ss, ss2 := RenderedChildrenSizes(w, size, focus, subfocus, app)
	forChild := false

	subs := w.SubWidgets()

	focusEvent := func(selectable bool) {
		srows := 0
		for i := 0; i < subfocus; i++ {
			srows += ss[i].BoxRows()
		}

		subSize := ss2[subfocus]
		if selectable {
			forChild = subs[subfocus].UserInput(gowid.TranslatedMouseEvent(ev, 0, -srows), subSize, focus, app)
		} else {
			forChild = gowid.UserInputIfSelectable(subs[subfocus], gowid.TranslatedMouseEvent(ev, 0, -srows), subSize, focus, app)
		}
	}

	if evm, ok := ev.(*tcell.EventMouse); ok {
		switch evm.Buttons() {
		case tcell.WheelUp, tcell.WheelDown:
			// A wheel up/down event applied to a pile should be sent to the focus widget. Consider a pile
			// which contains a list. If the pile focus isn't the list (say it's the previous widget), then
			// the wheel event will invoke the scroll-down code below (forChild == false), putting the list
			// in focus. If the focus is on the list, then forChild == true and the pile won't handle it further,
			// but the list will scroll as intended. If we scroll back to the top of the list, then the list
			// will return false for handle (forChild == false) and we'll invoke the scroll-up code below
			// to scroll to the previous pile focus.
			if subfocus != -1 {
				focusEvent(true)
			}
		default:
			// Don't try to compute
			if subfocus == -1 {
				break
			}

			// A left click sets focus if the widget is selectable and would take the mouse input; but
			// if I don't filter by click, then moving the mouse over another widget would shift focus
			// automatically, which is not usually what's wanted.
			_, my := evm.Position()
			curY := 0
		Loop:
			for i, c := range ss {
				if my < curY+c.BoxRows() && my >= curY {
					subSize := ss2[i]
					forChild = subs[i].UserInput(gowid.TranslatedMouseEvent(ev, 0, -curY), subSize, focus.SelectIf(w.SelectChild(focus) && i == subfocus), app)

					// Give the child focus if (a) it's selectable, and (b) if this is the click up corresponding
					// to a previous click down on this pile widget.
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
				}
				curY += c.BoxRows()
			}
		}
	} else {
		if subfocus != -1 {
			focusEvent(false)
		}
	}

	res := forChild

	if !forChild && w.Focus() != -1 { // e.g. if none of the subwidgets are selectable
		res = true
		scrollDown := false
		scrollUp := false

		if evk, ok := ev.(*tcell.EventKey); ok {
			switch {
			case w.KeyIsDown(evk):
				scrollDown = true
			case w.KeyIsUp(evk):
				scrollUp = true
			default:
				res = false
			}
		} else if ev2, ok := ev.(*tcell.EventMouse); ok {
			switch ev2.Buttons() {
			case tcell.WheelDown:
				scrollDown = true
			case tcell.WheelUp:
				scrollUp = true
			default:
				res = false
			}
		} else {
			res = false
		}

		if scrollUp || scrollDown {

			curw := subs[w.Focus()]
			prefPos := gowid.PrefPosition(curw)

			if scrollUp {
				res = gowid.ChangeFocus(w, -1, w.Wrap(), app)
			} else {
				res = gowid.ChangeFocus(w, 1, w.Wrap(), app)
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

func RenderSize(w gowid.ICompositeMultipleWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	subfocus := w.Focus()
	sizes := w.RenderedSubWidgetsSizes(size, focus, subfocus, app)

	maxcol := 0
	maxrow := 0

	for _, sz := range sizes {
		maxcol = gwutil.Max(maxcol, sz.BoxColumns())
		maxrow += sz.BoxRows()
	}

	if sz, ok := size.(gowid.IRenderBox); ok {
		maxrow = gwutil.Min(maxrow, sz.BoxRows())
	}

	return gowid.RenderBox{maxcol, maxrow}
}

// Heights can be:
// - Pack - use what you need
// - Units - fixed number of rows (RenderBox)
// - Weight - divide up
//
// What if height is bigger than rows we have? Then we chop.
// What if height is bigger before weights taken into consideration? They are zero
// What if the Pile is rendered as a RenderFlow? Then you can't specify any weighted widgets

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {

	subfocus := w.Focus()
	// if !focus.Focus {
	// 	subfocus = -1
	// }
	canvases := w.RenderSubWidgets(size, focus, subfocus, app)

	rows, ok := size.(gowid.IRows)
	haveMaxRow := ok

	res := gowid.NewCanvas()
	trim := false

	for i := 0; i < len(canvases); i++ {
		// Can be nil if weighted widgets were included but there wasn't enough space
		if canvases[i] != nil {
			// make sure each canvas uses up the width it is alloted - so if I render
			// a pile of width 20, and put it in a column, the next column starts at 21
			// TODO - remember which one has focus
			res.AppendBelow(canvases[i], i == subfocus, false)
			if haveMaxRow && res.BoxRows() >= rows.Rows() {
				trim = true
				break
			}
		}
	}

	if trim {
		res.Truncate(0, res.BoxRows()-rows.Rows())
	}

	if haveMaxRow && res.BoxRows() < rows.Rows() {
		gowid.AppendBlankLines(res, rows.Rows()-res.BoxRows())
	}

	if cols, ok := size.(gowid.IColumns); ok {
		res.ExtendRight(gowid.EmptyLine(cols.Columns() - res.BoxColumns()))
	}

	return res
}

func RenderedChildrenSizes(w IWidget, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) ([]gowid.IRenderBox, []gowid.IRenderSize) {
	fn2 := BoxMakerFunc(func(w gowid.IWidget, subSize gowid.IRenderSize, focus gowid.Selector, subApp gowid.IApp) gowid.IRenderBox {
		return w.RenderSize(subSize, focus, subApp)
	})
	res := make([]gowid.IRenderBox, 0)
	resSS := make([]gowid.IRenderSize, 0)
	sts, stsSS := w.RenderBoxMaker(size, focus, focusIdx, app, fn2)
	res = append(res, sts...)
	resSS = append(resSS, stsSS...)
	return res, resSS
}

func RenderSubwidgets(w IWidget, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp) []gowid.ICanvas {
	fn1 := BoxMakerFunc(func(w gowid.IWidget, subSize gowid.IRenderSize, focus gowid.Selector, subApp gowid.IApp) gowid.IRenderBox {
		return w.Render(subSize, focus, subApp)
	})

	canvases, _ := w.RenderBoxMaker(size, focus, focusIdx, app, fn1)
	res := make([]gowid.ICanvas, len(canvases))
	for i := 0; i < len(canvases); i++ {
		if canvases[i] != nil {
			res[i] = canvases[i].(gowid.ICanvas)
		}
	}
	return res
}

// TODO - make this an interface
type IPileBoxMaker interface {
	MakeBox(gowid.IWidget, gowid.IRenderSize, gowid.Selector, gowid.IApp) gowid.IRenderBox
}

type BoxMakerFunc func(gowid.IWidget, gowid.IRenderSize, gowid.Selector, gowid.IApp) gowid.IRenderBox

func (f BoxMakerFunc) MakeBox(w gowid.IWidget, s gowid.IRenderSize, b gowid.Selector, c gowid.IApp) gowid.IRenderBox {
	return f(w, s, b, c)
}

func RenderBoxMaker(w IWidget, size gowid.IRenderSize, focus gowid.Selector, focusIdx int, app gowid.IApp, fn IPileBoxMaker) ([]gowid.IRenderBox, []gowid.IRenderSize) {
	dims := w.Dimensions()

	_, ok1 := size.(gowid.IRenderFlowWith)
	_, ok2 := size.(gowid.IRenderFixed)
	weightWidgets := 0
	if ok1 || ok2 {
		for _, ww := range dims {
			if _, ok := ww.(gowid.IRenderWithWeight); ok {
				weightWidgets++
				if weightWidgets > 1 {
					panic(fmt.Errorf("Pile is rendered as Flow/Fixed %v of type %T so cannot contain more than one Weight widget",
						size, size))
				}
			}
		}
	}

	subs := w.SubWidgets()
	wlen := len(subs)
	res := make([]gowid.IRenderBox, wlen)
	resSS := make([]gowid.IRenderSize, wlen)

	heights := make([]int, wlen)
	ineligible := make([]bool, wlen)

	// So I know what is initialized later and what isn't (since 0 can be a legit row height)
	for i := 0; i < wlen; i++ {
		heights[i] = -1
	}

	rowsUsed := 0
	totalWeight := 0

	// Render all fixed first. This will determine the maximum width which can then be
	// supplied as an advisory parameter to unrendered subwidgets. Any that are specified
	// as RenderFlow{} can then be rendered to the max width
	maxcol := -1
	for i := 0; i < wlen; i++ {
		subSize, err := gowid.ComputeVerticalSubSize(size, dims[i], -1, -1)
		if err == nil {
			if _, ok := subSize.(gowid.IRenderFixed); ok {
				// only do if subsize is fixed
				resSS[i] = subSize
				res[i] = fn.MakeBox(subs[i], subSize, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
				heights[i] = res[i].BoxRows()
				rowsUsed += heights[i]
				if res[i].BoxColumns() > maxcol {
					maxcol = res[i].BoxColumns()
				}
			}
		}
	}

	//
	// Do the packed and fixed height widgets first. Then after, divvy up the rest
	// among the weighted widgets
	//
	for i := 0; i < wlen; i++ {
		// TODO - remember which one has focus
		if res[i] == nil {
			subSize, err := gowid.ComputeVerticalSubSize(size, dims[i], maxcol, -1)
			if err == nil {
				resSS[i] = subSize
				res[i] = fn.MakeBox(subs[i], subSize, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
				heights[i] = res[i].BoxRows()
				rowsUsed += heights[i]
			} else {
				if w2, ok := dims[i].(gowid.IRenderWithWeight); !ok {
					panic(fmt.Errorf("Unsupported dimension %T of type %T for widget %v - %v",
						dims[i], dims[i], subs[i], err))
				} else {
					// It must be weighted
					totalWeight += w2.Weight()
				}
			}
		}
	}

	//
	// Now divide up remaining space
	//

	// Track the last height row, so I can adjust for floating errors
	lasti := -1

	if box, ok := size.(gowid.IRenderBox); ok {
		rowsToDivideUp := box.BoxRows() - rowsUsed
		rowsLeft := rowsToDivideUp
		for {
			if rowsLeft == 0 {
				break
			}
			doneone := false
			totalWeight = 0
			for i := 0; i < wlen; i++ {
				if w2, ok := dims[i].(gowid.IRenderWithWeight); ok && !ineligible[i] {
					totalWeight += w2.Weight()
				}
			}
			rowsToDivideUp = rowsLeft
			for i := 0; i < wlen; i++ {
				if w2, ok := dims[i].(gowid.IRenderWithWeight); ok && !ineligible[i] {
					rows := int(((float32(w2.Weight()) / float32(totalWeight)) * float32(rowsToDivideUp)) + 0.5)

					if max, ok := dims[i].(gowid.IRenderMaxUnits); ok {
						if rows > max.MaxUnits() {
							rows = max.MaxUnits()
							ineligible[i] = true // this one is done
						}
					}

					if rows > rowsLeft {
						rows = rowsLeft
					}
					if rows > 0 {
						if heights[i] == -1 {
							heights[i] = 0
						}
						heights[i] += rows

						rowsLeft -= rows
						lasti = i
						doneone = true
					}
				}
			}
			if !doneone {
				break
			}
		}
		if lasti != -1 && rowsLeft > 0 {
			heights[lasti] += rowsLeft
		}
		// Now actually render
		for i := 0; i < wlen; i++ {
			if _, ok := dims[i].(gowid.IRenderWithWeight); ok {
				ss := gowid.RenderBox{box.BoxColumns(), heights[i]}
				resSS[i] = ss
				res[i] = fn.MakeBox(subs[i], ss, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
			}
		}
	} else {
		// FlowWith and Fixed
		for i := 0; i < wlen; i++ {
			// Should only be one!
			if _, ok := dims[i].(gowid.IRenderWithWeight); ok {
				resSS[i] = size
				res[i] = fn.MakeBox(subs[i], size, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
			}
		}
	}

	zbox := gowid.RenderBox{0, 0}
	for i := 0; i < wlen; i++ {
		if res[i] == nil {
			resSS[i] = zbox
			res[i] = fn.MakeBox(subs[i], zbox, focus.SelectIf(w.SelectChild(focus) && i == focusIdx), app)
		}
	}

	return res, resSS
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
