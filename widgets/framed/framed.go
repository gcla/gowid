// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package framed provides a widget that draws a frame around an inner widget.
package framed

import (
	"fmt"
	"runtime"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
)

//======================================================================

type FrameRunes struct {
	Tl, Tr, Bl, Br rune
	T, B, L, R     rune
}

var (
	AsciiFrame      = FrameRunes{'-', '-', '-', '-', '-', '-', '|', '|'}
	UnicodeFrame    = FrameRunes{'┏', '┓', '┗', '┛', '━', '━', '┃', '┃'}
	UnicodeAltFrame = FrameRunes{'▛', '▜', '▙', '▟', '▀', '▄', '▌', '▐'}
	SpaceFrame      = FrameRunes{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
)

func init() {
	if runtime.GOOS == "windows" {
		UnicodeFrame = FrameRunes{'┌', '┐', '└', '┘', '―', '―', '|', '|'}
		UnicodeAltFrame = UnicodeFrame
	}
}

type IFramed interface {
	Opts() Options
}

type IWidget interface {
	gowid.ICompositeWidget
	IFramed
}

type Widget struct {
	gowid.IWidget // Embed for Selectable method
	Params        Options
	Callbacks     *gowid.Callbacks
	gowid.SubWidgetCallbacks
}

type Options struct {
	Frame       FrameRunes
	Title       string
	TitleWidget gowid.IWidget
	Style       gowid.ICellStyler
}

// For callback identification
type Title struct{}

func New(inner gowid.IWidget, opts ...Options) *Widget {
	cb := gowid.NewCallbacks()
	var opt Options
	if len(opts) == 0 {
		opt = Options{
			Frame: AsciiFrame,
		}
	} else {
		opt = opts[0]
	}

	res := &Widget{
		IWidget:            inner,
		Params:             opt,
		Callbacks:          cb,
		SubWidgetCallbacks: gowid.SubWidgetCallbacks{ICallbacks: cb},
	}
	var _ gowid.IWidget = res
	var _ IWidget = res
	return res
}

func NewUnicode(inner gowid.IWidget) *Widget {
	params := Options{
		Frame: UnicodeFrame,
	}
	return New(inner, params)
}

func NewUnicodeAlt(inner gowid.IWidget) *Widget {
	params := Options{
		Frame: UnicodeAltFrame,
	}
	return New(inner, params)
}

func NewSpace(inner gowid.IWidget) *Widget {
	params := Options{
		Frame: SpaceFrame,
	}
	return New(inner, params)
}

func (w *Widget) String() string {
	return fmt.Sprintf("framed[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(wi gowid.IWidget, app gowid.IApp) {
	w.IWidget = wi
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) OnSetTitle(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w, Title{}, f)
}

func (w *Widget) RemoveOnSetAlign(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w, Title{}, f)
}

// Call from Render thread
func (w *Widget) SetTitle(title string, app gowid.IApp) {
	w.Params.Title = title
	w.Params.TitleWidget = nil
	gowid.RunWidgetCallbacks(w, Title{}, app, w)
}

func (w *Widget) GetTitle() string {
	return w.Params.Title
}

func (w *Widget) SetTitleWidget(widget gowid.IWidget, app gowid.IApp) {
	w.Params.TitleWidget = widget
	gowid.RunWidgetCallbacks(w, Title{}, app, w)
}

func (w *Widget) GetTitleWidget() gowid.IWidget {
	return w.Params.TitleWidget
}

func (w *Widget) Opts() Options {
	return w.Params
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
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

//======================================================================

func RenderSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	ss := w.SubWidgetSize(size, focus, app)
	sdim := w.SubWidget().RenderSize(ss, focus, app)
	return gowid.RenderBox{C: sdim.BoxColumns() + 2, R: sdim.BoxRows() + 2}
}

func SubWidgetSize(w interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	var newSize gowid.IRenderSize
	switch sz := size.(type) {
	case gowid.IRenderFixed:
		newSize = gowid.RenderFixed{}
	case gowid.IRenderBox:
		newSize = gowid.RenderBox{C: gwutil.Max(sz.BoxColumns()-2, 0), R: gwutil.Max(sz.BoxRows()-2, 0)}
	case gowid.IRenderFlowWith:
		newSize = gowid.RenderFlowWith{C: gwutil.Max(sz.FlowColumns()-2, 0)}
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}
	return newSize
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	res := gowid.NewCanvas()
	tmp := gowid.NewCanvas()
	newSize := w.SubWidgetSize(size, focus, app)

	innerCanvas := gowid.Render(w.SubWidget(), newSize, focus, app)
	innerLines := innerCanvas.BoxRows()
	maxCol := innerCanvas.BoxColumns()

	frame := w.Opts().Frame
	empty := FrameRunes{}
	if frame == empty {
		frame = AsciiFrame
	}

	var tophor, bottomhor, leftver, rightver gowid.Cell
	tophor = gowid.CellFromRune(frame.T)
	bottomhor = gowid.CellFromRune(frame.B)
	leftver = gowid.CellFromRune(frame.L)
	rightver = gowid.CellFromRune(frame.R)
	if w.Opts().Style != nil {
		f, _, _ := w.Opts().Style.GetStyle(app)
		fc := gowid.IColorToTCell(f, gowid.ColorNone, app.GetColorMode())
		tophor = tophor.WithForegroundColor(fc)
		bottomhor = bottomhor.WithForegroundColor(fc)
		leftver = leftver.WithForegroundColor(fc)
		rightver = rightver.WithForegroundColor(fc)
	}

	titleWidget := w.Opts().TitleWidget
	if titleWidget == nil && w.Opts().Title != "" {
		titleWidget = text.New(" " + w.Opts().Title + " ")
	}

	leftverCanvas := gowid.NewCanvas()
	rightverCanvas := gowid.NewCanvas()
	leftverLine := make([]gowid.Cell, 0)
	rightverLine := make([]gowid.Cell, 0)
	leftverLine = append(leftverLine, leftver)
	rightverLine = append(rightverLine, rightver)
	for i := 0; i < innerLines; i++ {
		leftverCanvas.AppendLine(leftverLine, false)
		rightverCanvas.AppendLine(rightverLine, false)
		tmp.AppendLine(make([]gowid.Cell, 0), false)
	}

	tophorArr := make([]gowid.Cell, 0)
	bottomhorArr := make([]gowid.Cell, 0)
	for i := 0; i < maxCol+2; i++ {
		tophorArr = append(tophorArr, tophor)
		bottomhorArr = append(bottomhorArr, bottomhor)
	}

	tmp.AppendRight(leftverCanvas, false)
	tmp.AppendRight(innerCanvas, true)
	tmp.AppendRight(rightverCanvas, false)

	res.AppendLine(tophorArr, false)
	res.AppendBelow(tmp, true, false)
	res.AppendLine(bottomhorArr, false)

	res.Lines[0][0] = res.Lines[0][0].WithRune(frame.Tl)
	res.Lines[0][len(res.Lines[0])-1] = res.Lines[0][len(res.Lines[0])-1].WithRune(frame.Tr)

	resl := res.BoxRows()
	res.Lines[resl-1][0] = res.Lines[resl-1][0].WithRune(frame.Bl)
	res.Lines[resl-1][len(res.Lines[0])-1] = res.Lines[resl-1][len(res.Lines[0])-1].WithRune(frame.Br)

	if titleWidget != nil {
		titleCanvas := gowid.Render(titleWidget, gowid.RenderFixed{}, gowid.NotSelected, app)
		res.MergeUnder(titleCanvas, 2, 0, false)
	}

	return res
}

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	subSize := w.SubWidgetSize(size, focus, app)
	newev := gowid.TranslatedMouseEvent(ev, -1, -1)

	if _, ok := ev.(*tcell.EventMouse); ok {
		ss := w.SubWidget().RenderSize(subSize, focus, app)
		newev2, _ := newev.(*tcell.EventMouse) // gcla tcell todo - clumsy
		mx, my := newev2.Position()
		if my < ss.BoxRows() && my >= 0 && mx < ss.BoxColumns() && mx >= 0 {
			return gowid.UserInputIfSelectable(w.SubWidget(), newev, subSize, focus, app)
		}
	} else {
		return gowid.UserInputIfSelectable(w.SubWidget(), newev, subSize, focus, app)
	}
	return false
}

//======================================================================

type FrameIfSelectedForCopy struct{}

var _ gowid.IClipboardSelected = FrameIfSelectedForCopy{}

func (r FrameIfSelectedForCopy) AlterWidget(w gowid.IWidget, app gowid.IApp) gowid.IWidget {
	return New(w)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
