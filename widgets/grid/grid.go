// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package grid allows widgets to be arranged in rows and columns.
package grid

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/vim"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/hpadding"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

type IGrid interface {
	gowid.IFindNextSelectable
	gowid.IFocus
	SubWidgets() []gowid.IWidget
	GenerateWidgets(size gowid.IRenderSize, attrs gowid.IRenderContext) (pile.IWidget, int)
	Width() int
	HSep() int
	VSep() int
	HAlign() gowid.IHAlignment
	Wrap() bool
	KeyIsUp(*tcell.EventKey) bool
	KeyIsDown(*tcell.EventKey) bool
	KeyIsLeft(*tcell.EventKey) bool
	KeyIsRight(*tcell.EventKey) bool
}

type IWidget interface {
	gowid.IWidget
	IGrid
}

type Width struct{}
type Align struct{}
type VSepCB struct{}
type HSepCB struct{}

// align sets the alignment of the group within the leftover space in the row.
type Widget struct {
	widgets []gowid.IWidget
	width   int
	hSep    int
	vSep    int
	align   gowid.IHAlignment
	focus   int // -1 means nothing selectable
	wrap    bool
	options Options
	*gowid.Callbacks
	gowid.SubWidgetsCallbacks
	gowid.FocusCallbacks
}

type Options struct {
	StartPos  int
	Wrap      bool
	DownKeys  []vim.KeyPress
	UpKeys    []vim.KeyPress
	LeftKeys  []vim.KeyPress
	RightKeys []vim.KeyPress
}

func New(widgets []gowid.IWidget, width int, hSep int, vSep int, align gowid.IHAlignment, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = Options{
			StartPos: -1,
		}
	}
	if opt.DownKeys == nil {
		opt.DownKeys = vim.AllDownKeys
	}
	if opt.UpKeys == nil {
		opt.UpKeys = vim.AllUpKeys
	}
	if opt.LeftKeys == nil {
		opt.LeftKeys = vim.AllLeftKeys
	}
	if opt.RightKeys == nil {
		opt.RightKeys = vim.AllRightKeys
	}
	res := &Widget{
		widgets: widgets,
		width:   width,
		hSep:    hSep,
		vSep:    vSep,
		align:   align,
		focus:   -1,
		options: opt,
	}
	res.SubWidgetsCallbacks = gowid.SubWidgetsCallbacks{CB: &res.Callbacks}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}
	if opt.StartPos >= 0 {
		res.focus = gwutil.Min(opt.StartPos, len(widgets)-1)
	} else {
		res.focus, _ = res.FindNextSelectable(1, res.Wrap())
	}

	var _ gowid.IWidget = res
	var _ gowid.ICompositeMultiple = res
	var _ IWidget = res

	return res
}

func (w *Widget) String() string {
	cols := make([]string, len(w.widgets))
	for i := 0; i < len(cols); i++ {
		cols[i] = fmt.Sprintf("%v", w.widgets[i])
	}
	return fmt.Sprintf("grid[%s]", strings.Join(cols, ","))
}

func (w *Widget) Width() int {
	return w.width
}

func (w *Widget) SetWidth(i int, app gowid.IApp) {
	w.width = i
	gowid.RunWidgetCallbacks(w.Callbacks, Width{}, app, w)
}

func (w *Widget) HSep() int {
	return w.hSep
}

func (w *Widget) SetHSep(i int, app gowid.IApp) {
	w.hSep = i
	gowid.RunWidgetCallbacks(w.Callbacks, HSepCB{}, app, w)
}

func (w *Widget) VSep() int {
	return w.vSep
}

func (w *Widget) SetVSep(i int, app gowid.IApp) {
	w.vSep = i
	gowid.RunWidgetCallbacks(w.Callbacks, VSepCB{}, app, w)
}

func (w *Widget) HAlign() gowid.IHAlignment {
	return w.align
}

func (w *Widget) SetHAlign(i gowid.IHAlignment, app gowid.IApp) {
	w.align = i
	gowid.RunWidgetCallbacks(w.Callbacks, Align{}, app, w)
}

func (w *Widget) SubWidgets() []gowid.IWidget {
	return gowid.CopyWidgets(w.widgets)
}

func (w *Widget) SetSubWidgets(widgets []gowid.IWidget, app gowid.IApp) {
	w.widgets = widgets
	gowid.RunWidgetCallbacks(w.Callbacks, gowid.SubWidgetsCB{}, app, w)
}

// Tries to set at required index, will choose first selectable from there
func (w *Widget) SetFocus(app gowid.IApp, i int) {
	old := w.focus
	w.focus = gwutil.Min(gwutil.Max(i, 0), len(w.widgets)-1)
	if old != w.focus {
		gowid.RunWidgetCallbacks(w.Callbacks, gowid.FocusCB{}, app, w)
	}
}

func (w *Widget) Wrap() bool {
	return w.options.Wrap
}

// Tries to set at required index, will choose first selectable from there
func (w *Widget) Focus() int {
	return w.focus
}

func (w *Widget) Selectable() bool {
	for _, widget := range w.widgets {
		if widget.Selectable() {
			return true
		}
	}
	return false
}

func (w *Widget) FindNextSelectable(dir gowid.Direction, wrap bool) (int, bool) {
	return gowid.FindNextSelectableWidget(w.widgets, w.focus, dir, wrap)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) GenerateWidgets(size gowid.IRenderSize, attrs gowid.IRenderContext) (pile.IWidget, int) {
	return GenerateWidgets(w, size, attrs)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) KeyIsUp(evk *tcell.EventKey) bool {
	return vim.KeyIn(evk, w.options.UpKeys)
}

func (w *Widget) KeyIsDown(evk *tcell.EventKey) bool {
	return vim.KeyIn(evk, w.options.DownKeys)
}

func (w *Widget) KeyIsLeft(evk *tcell.EventKey) bool {
	return vim.KeyIn(evk, w.options.LeftKeys)
}

func (w *Widget) KeyIsRight(evk *tcell.EventKey) bool {
	return vim.KeyIn(evk, w.options.RightKeys)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	pile, _ := w.GenerateWidgets(size, app)
	res := pile.Render(size, focus, app)
	return res
}

// Scroll sequentially through the widgets on mouse scroll events or key up/down
func UserInput(w IGrid, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	subfocus := w.Focus()
	if !focus.Focus {
		subfocus = -1
	}

	forChild := false

	var gen pile.IWidget
	var cols int

	scrollDown := false
	scrollUp := false
	scrollRight := false
	scrollLeft := false

	if evm, ok := ev.(*tcell.EventMouse); ok {
		// If it's a scroll event, then handle it in the grid, not by delegating to the
		// construction of piles and columns
		if evm.Buttons()&(tcell.WheelDown|tcell.WheelUp|tcell.WheelLeft|tcell.WheelRight) == 0 {
			gen, cols = w.GenerateWidgets(size, app)
			forChild = gowid.UserInputIfSelectable(gen, ev, size, focus, app)
			if evm.Buttons() == tcell.Button1 && forChild {
				pileidx := gen.Focus()
				if pileidx != -1 {
					// if ForChild, then it can't be any of the in between text widgets used to pad
					// things, so don't worry about checking the casts
					rowx, _ := gen.SubWidgets()[pileidx].(*gowid.ContainerWidget)
					rowy, _ := rowx.SubWidget().(*hpadding.Widget)
					row, _ := rowy.IWidget.(*columns.Widget)
					colidx := row.Focus()
					w.SetFocus(app, (((pileidx+1)/2)*cols)+((colidx+1)/2))
				}
			}
		}
	} else {
		if subfocus != -1 {
			forChild = gowid.UserInputIfSelectable(w.SubWidgets()[w.Focus()], ev, size, focus, app)
		}
	}

	if forChild {
		return true
	} else {
		newfocus := -1

		if evk, ok := ev.(*tcell.EventKey); ok {
			switch {
			case w.KeyIsRight(evk):
				next, ok := w.FindNextSelectable(1, w.Wrap())
				if ok {
					w.SetFocus(app, next)
					return true
				}
			case w.KeyIsLeft(evk):
				next, ok := w.FindNextSelectable(-1, w.Wrap())
				if ok {
					w.SetFocus(app, next)
					return true
				}
			case w.KeyIsDown(evk):
				scrollDown = true
			case w.KeyIsUp(evk):
				scrollUp = true
			}
		} else if ev2, ok := ev.(*tcell.EventMouse); ok {
			switch ev2.Buttons() {
			case tcell.WheelDown:
				scrollDown = true
			case tcell.WheelUp:
				scrollUp = true
			case tcell.WheelRight:
				scrollRight = true
			case tcell.WheelLeft:
				scrollLeft = true
			}
		}

		if scrollDown || scrollUp || scrollRight || scrollLeft {
			if gen == nil {
				_, cols = w.GenerateWidgets(size, app)
			}

			if scrollDown {
				for i := w.Focus() + cols; i < len(w.SubWidgets()); i += cols {
					if w.SubWidgets()[i].Selectable() {
						newfocus = i
						break
					}
				}
			}
			if scrollUp {
				for i := w.Focus() - cols; i >= 0; i -= cols {
					if w.SubWidgets()[i].Selectable() {
						newfocus = i
						break
					}
				}
			}
			if scrollRight {
				i := ((w.Focus() / cols) * cols) + gwutil.Min(w.Focus()+1, cols-1)
				j := ((w.Focus() / cols) * cols) + (cols - 1)
				for ; i <= j; i++ {
					if w.SubWidgets()[i].Selectable() {
						newfocus = i
						break
					}
				}
			}
			if scrollLeft {
				i := ((w.Focus() / cols) * cols) + gwutil.Max((w.Focus()%cols)-1, 0)
				j := ((w.Focus() / cols) * cols)
				for ; i >= j; i-- {
					if w.SubWidgets()[i].Selectable() {
						newfocus = i
						break
					}
				}
			}
			if newfocus != -1 {
				w.SetFocus(app, newfocus)
				return true
			}
		}

		return false
	}
}

// Can't support RenderFixed{} because I need to know how many columns so I can roll over widgets
// to the next line.
//
func GenerateWidgets(w IGrid, size gowid.IRenderSize, attrs gowid.IRenderContext) (pile.IWidget, int) {
	focusIdx := w.Focus()
	cols2, isColumns := size.(gowid.IColumns)
	if !isColumns {
		panic(errors.New("GridFlow widget must not be rendered in Fixed mode."))
	}
	cols := cols2.Columns()

	// TODO - what about when cols < vsep?
	numInRow := (cols - w.HSep()) / (w.Width() + w.HSep())
	wWidth := numInRow * w.Width()
	if numInRow > 0 {
		wWidth += (numInRow - 1) * w.HSep()
	}

	pileWidgets := make([]gowid.IContainerWidget, 0)

	curInRow := 0
	hSepWidget := &gowid.ContainerWidget{text.New(gwutil.StringOfLength(' ', w.HSep())), gowid.RenderWithUnits{U: w.HSep()}}
	hBlankWidget := &gowid.ContainerWidget{text.New(gwutil.StringOfLength(' ', w.Width())), gowid.RenderWithUnits{U: w.Width()}}
	vBlank := text.New("")
	vBlankWidget := vpadding.New(vBlank, gowid.VAlignTop{}, gowid.RenderWithUnits{U: w.VSep()})
	curRow := make([]gowid.IContainerWidget, 0)

	todo := 0
	rowFocusIdx := -1
	pileFocusIdx := -1
	var fakeApp gowid.IApp
	if len(w.SubWidgets()) > 0 {
		todo = (((len(w.SubWidgets()) - 1) / numInRow) + 1) * numInRow
	}
	firstRow := true
	for i := 0; i < todo; i++ {
		if i >= len(w.SubWidgets()) {
			curRow = append(curRow, hBlankWidget)
		} else {
			if i == focusIdx {
				rowFocusIdx = len(curRow)
			}
			curRow = append(curRow, &gowid.ContainerWidget{w.SubWidgets()[i], gowid.RenderWithUnits{U: w.Width()}})
		}
		curInRow += 1
		if curInRow == numInRow {
			cols := columns.New(curRow)
			alignedColumns := hpadding.New(cols, w.HAlign(), gowid.RenderWithUnits{U: wWidth})
			if !firstRow {
				pileWidgets = append(pileWidgets, &gowid.ContainerWidget{vBlankWidget, gowid.RenderFlow{}})
			}
			pileWidgets = append(pileWidgets, &gowid.ContainerWidget{alignedColumns, gowid.RenderFlow{}})
			if rowFocusIdx != -1 {
				// TODO: wrong wrong wrong
				// It's not necessarily an IApp e.g. when called from Render - but it's ok because I haven't set
				// any callbacks
				cols.SetFocus(fakeApp, rowFocusIdx)
				rowFocusIdx = -1
				pileFocusIdx = len(pileWidgets) - 1
			}
			firstRow = false
			curInRow = 0
			curRow = make([]gowid.IContainerWidget, 0)
		} else {
			curRow = append(curRow, hSepWidget)
		}
	}

	pile := pile.New(pileWidgets)
	if pileFocusIdx != -1 {
		pile.SetFocus(fakeApp, pileFocusIdx)
	}

	return pile, numInRow
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
