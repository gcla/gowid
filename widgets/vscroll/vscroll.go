// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package vscroll provides a vertical scrollbar widget with mouse support. See the editor
// demo for more.
package vscroll

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

type VerticalScrollbarRunes struct {
	Up, Down, Space, Handle rune
}

var (
	VerticalScrollbarAsciiRunes   = VerticalScrollbarRunes{'^', 'v', ' ', '#'}
	VerticalScrollbarUnicodeRunes = VerticalScrollbarRunes{'▲', '▼', ' ', '█'}
)

//======================================================================

type ClickUp struct{}
type ClickDown struct{}
type ClickAbove struct{}
type ClickBelow struct{}
type RightClick struct{}

//======================================================================

type IVerticalScrollbar interface {
	GetTop() int
	GetMiddle() int
	GetBottom() int
	ClickUp(app gowid.IApp)
	ClickDown(app gowid.IApp)
	ClickAbove(app gowid.IApp)
	ClickBelow(app gowid.IApp)
	GetRunes() VerticalScrollbarRunes
}

type IRightMouseClick interface {
	RightClick(frac float32, app gowid.IApp)
}

type IWidget interface {
	gowid.IWidget
	IVerticalScrollbar
}

type Widget struct {
	Top       int
	Middle    int
	Bottom    int
	Runes     VerticalScrollbarRunes
	Callbacks *gowid.Callbacks
	gowid.IsSelectable
}

func New() *Widget {
	return NewWithChars(VerticalScrollbarAsciiRunes)
}

func NewUnicode() *Widget {
	return NewWithChars(VerticalScrollbarUnicodeRunes)
}

func NewWithChars(runes VerticalScrollbarRunes) *Widget {
	return NewExt(runes)
}

func NewExt(runes VerticalScrollbarRunes) *Widget {
	return &Widget{
		Top:       -1,
		Middle:    -1,
		Bottom:    -1,
		Runes:     runes,
		Callbacks: gowid.NewCallbacks(),
	}
}

func (w *Widget) String() string {
	return fmt.Sprintf("vscroll[t=%d,m=%d,b=%d]", w.GetTop(), w.GetMiddle(), w.GetBottom())
}

func (w *Widget) GetTop() int {
	return w.Top
}

func (w *Widget) GetMiddle() int {
	return w.Middle
}

func (w *Widget) GetBottom() int {
	return w.Bottom
}

func (w *Widget) OnClickBelow(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ClickBelow{}, f)
}

func (w *Widget) RemoveOnClickBelow(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ClickBelow{}, f)
}

func (w *Widget) OnClickAbove(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ClickAbove{}, f)
}

func (w *Widget) RemoveOnClickAbove(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ClickAbove{}, f)
}

func (w *Widget) OnRightClick(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, RightClick{}, f)
}

func (w *Widget) RemoveOnRightClick(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, RightClick{}, f)
}

func (w *Widget) OnClickDownArrow(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ClickDown{}, f)
}

func (w *Widget) RemoveOnClickDownArrow(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ClickDown{}, f)
}

func (w *Widget) OnClickUpArrow(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ClickUp{}, f)
}

func (w *Widget) RemoveOnClickUpArrow(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ClickUp{}, f)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) GetRunes() VerticalScrollbarRunes {
	return w.Runes
}

func (w *Widget) ClickUp(app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, ClickUp{}, app, w)
}

func (w *Widget) ClickDown(app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, ClickDown{}, app, w)
}

func (w *Widget) ClickAbove(app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, ClickAbove{}, app, w)
}

func (w *Widget) ClickBelow(app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, ClickBelow{}, app, w)
}

func (w *Widget) RightClick(frac float32, app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, RightClick{}, app, w, frac)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func RenderSize(w interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	cols, haveCols := size.(gowid.IColumns)
	rows, haveRows := size.(gowid.IRows)
	switch {
	case haveCols && haveRows:
		return gowid.RenderBox{C: cols.Columns(), R: rows.Rows()}
	case haveCols:
		return gowid.RenderBox{C: cols.Columns(), R: 1}
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}
}

func UserInput(w IVerticalScrollbar, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if ev2, ok := ev.(*tcell.EventMouse); ok {
		switch ev2.Buttons() {
		case tcell.Button1, tcell.Button3:
			b3 := (ev2.Buttons() == tcell.Button3)
			t, m, b := w.GetTop(), w.GetMiddle(), w.GetBottom()
			rows := 1
			if box, ok := size.(gowid.IRenderBox); ok {
				rows = box.BoxRows()
			}
			//t, m, b = granularizeSplits(t, m, b, gwutil.Max(0, rows-2))
			splits := gwutil.HamiltonAllocation([]int{t, m, b}, gwutil.Max(0, rows-2))
			// Make sure that the "handle" in the middle is always at least 1 row tall
			if splits[1] == 0 {
				fixSplit(1, 0, 2, &splits)
			}
			// Make sure that unless we're at the top, there is a space to click to go closer to the top
			// and vice-versa for the bottom
			if t != 0 && splits[0] == 0 {
				fixSplit(0, 1, 2, &splits)
			}
			if b != 0 && splits[2] == 0 {
				fixSplit(2, 0, 1, &splits)
			}
			_, y := ev2.Position()
			res := false
			switch b3 {
			case true:
				if w, ok := w.(IRightMouseClick); ok {
					var frac float32
					switch {
					case y == 0:
						frac = 0.0
					case y <= splits[2]+splits[1]+splits[0]:
						if rows > 2 {
							frac = float32(y-1) / float32(rows-2)
						}
					default:
						frac = 1.0
					}
					w.RightClick(frac, app)
					res = true
				}
			case false:
				switch {
				case y == 0:
					w.ClickUp(app)
					res = true
				case y <= splits[0]:
					w.ClickAbove(app)
					res = true
				case y <= splits[1]+splits[0]:
				case y <= splits[2]+splits[1]+splits[0]:
					w.ClickBelow(app)
					res = true
				default:
					w.ClickDown(app)
					res = true
				}
			}
			return res
		default:
			return false
		}
	} else {
		return false
	}
}

func fixSplit(i int, o1, o2 int, splits *[]int) {
	if (*splits)[o1] > (*splits)[o2] {
		if (*splits)[o1] > 2 {
			(*splits)[i]++
			(*splits)[o1]--
		}
	} else {
		if (*splits)[o2] > 2 {
			(*splits)[i]++
			(*splits)[o2]--
		}
	}
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {

	// If a col and row not provided, what can I choose??
	cl, isCols := size.(gowid.IColumns)
	if !isCols {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IColumns"})
	}

	cols := cl.Columns()
	rows := 1
	if rs, isRows := size.(gowid.IRows); isRows {
		rows = rs.Rows()
	}

	t, m, b := w.GetTop(), w.GetMiddle(), w.GetBottom()
	splits := gwutil.HamiltonAllocation([]int{t, m, b}, gwutil.Max(0, rows-2))
	// Make sure that the "handle" in the middle is always at least 1 row tall
	if splits[1] == 0 {
		fixSplit(1, 0, 2, &splits)
	}
	// Make sure that unless we're at the top, there is a space to click to go closer to the top
	// and vice-versa for the bottom
	if t != 0 && splits[0] == 0 {
		fixSplit(0, 1, 2, &splits)
	}
	if b != 0 && splits[2] == 0 {
		fixSplit(2, 0, 1, &splits)
	}

	fill := gowid.CellFromRune(w.GetRunes().Handle)
	fillArr := make([]gowid.Cell, 0)
	blank := gowid.CellFromRune(w.GetRunes().Space)
	blankArr := make([]gowid.Cell, 0)
	for i := 0; i < cols; i++ {
		fillArr = append(fillArr, fill)
		blankArr = append(blankArr, blank)
	}

	resblankabove := gowid.NewCanvas()
	resblankbelow := gowid.NewCanvas()
	resfill := gowid.NewCanvas()

	// [2, 4, 7]
	dup1 := make([]gowid.Cell, cols)
	copy(dup1, blankArr)
	resblankabove.Lines = append(resblankabove.Lines, dup1)

	var dup []gowid.Cell
	for i := 0; i < rows-2; i++ {
		if i < splits[0] {
			dup = make([]gowid.Cell, cols)
			copy(dup, blankArr)
			resblankabove.Lines = append(resblankabove.Lines, dup)
		} else if i < splits[1]+splits[0] {
			dup = make([]gowid.Cell, cols)
			copy(dup, fillArr)
			resfill.Lines = append(resfill.Lines, dup)
		} else {
			dup = make([]gowid.Cell, cols)
			copy(dup, blankArr)
			resblankbelow.Lines = append(resblankbelow.Lines, dup)
		}
	}

	dup2 := make([]gowid.Cell, cols)
	copy(dup2, blankArr)
	resblankbelow.Lines = append(resblankbelow.Lines, dup2)

	resblankabove.AlignRight()
	resfill.AlignRightWith(gowid.MakeCell(
		'#',
		gowid.MakeTCellColorExt(tcell.ColorDefault),
		gowid.MakeTCellColorExt(tcell.ColorDefault),
		gowid.StyleNone))
	resblankbelow.AlignRight()

	res := gowid.NewCanvas()
	res.AppendBelow(resblankabove, false, false)
	res.AppendBelow(resfill, false, false)
	res.AppendBelow(resblankbelow, false, false)

	l := len(res.Lines)
	if l > 0 {
		for i := 0; i < len(res.Lines[0]); i++ {
			res.Lines[0][i] = res.Lines[0][i].WithRune(w.GetRunes().Up)
		}
		for i := 0; i < len(res.Lines[0]); i++ {
			res.Lines[l-1][i] = res.Lines[l-1][i].WithRune(w.GetRunes().Down)
		}
	}

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
