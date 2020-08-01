// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package list provides a widget displaying a vertical list of widgets with one in focus and support for previous and next.
package list

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/vim"
	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
)

//======================================================================

// IWalkerPosition is satisfied by any struct with an Equal() method that can
// determine whether it's at the "same" position as another or is more advanced
// than another.
type IWalkerPosition interface {
	Equal(IWalkerPosition) bool
	GreaterThan(IWalkerPosition) bool
}

// For most simple uses
type IBoundedWalkerPosition interface {
	IWalkerPosition
	ToInt() int
}

// WalkerMoved is a simple struct used to hold the result of a Next() or
// Previous() call on an IWalker. I find its use more convenient than
// returning multiple values because you can use it inline with other
// expressions.
type WalkerIndex struct {
	Widget gowid.IWidget
	Pos    IWalkerPosition
}

// Remove HaveFocus because the Next and Previous APIs have to
// return an gowid.IWidget anyway - so just standardize on assuming a nil interface
// means invalid
type IWalker interface {
	At(pos IWalkerPosition) gowid.IWidget
	Focus() IWalkerPosition
	SetFocus(pos IWalkerPosition, app gowid.IApp)
	Next(pos IWalkerPosition) IWalkerPosition
	Previous(pos IWalkerPosition) IWalkerPosition
}

// IBoundedWalker is implemented by an IWalker type that knows the maximum length
// of its underlying data set. It must return IBoundedWalkerPosition rather than only
// IWalkerPosition.
type IBoundedWalker interface {
	IWalker
	Length() int
}

// IWalkerHome is any type that supports being able to provide a first position.
type IWalkerHome interface {
	First() IWalkerPosition // nil possible if empty
}

// IWalkerEnd is any type that supports being able to provide a last position. This
// is used to support the End key on the keyboard.
type IWalkerEnd interface {
	Last() IWalkerPosition // nil possible if empty
}

//======================================================================

type WidgetIsUnboundedError struct {
	Type interface{}
}

var _ error = WidgetIsUnboundedError{}

func (e WidgetIsUnboundedError) Error() string {
	return fmt.Sprintf("Widget does not use IBoundedWalker - %v of type %T", e.Type, e.Type)
}

var BadState = fmt.Errorf("Broken state in list widget")

//======================================================================

type ListPos int

func (l ListPos) ToInt() int {
	return int(l)
}

func (l ListPos) Equal(other IWalkerPosition) bool {
	switch o := other.(type) {
	case ListPos:
		return o == l
	default:
		panic(gowid.InvalidTypeToCompare{LHS: l, RHS: other})
	}
}

func (l ListPos) GreaterThan(other IWalkerPosition) bool {
	switch o := other.(type) {
	case ListPos:
		return l > o
	default:
		panic(gowid.InvalidTypeToCompare{LHS: l, RHS: other})
	}
}

type SimpleListWalker struct {
	Widgets []gowid.IWidget
	focus   ListPos
}

var _ IBoundedWalker = (*SimpleListWalker)(nil)
var _ IWalkerHome = (*SimpleListWalker)(nil)

func NewSimpleListWalker(widgets []gowid.IWidget) *SimpleListWalker {
	res := &SimpleListWalker{
		Widgets: widgets,
		focus:   -1,
	}

	pos, _ := gowid.FindNextSelectableWidget(widgets, -1, 1, false)
	res.focus = ListPos(pos)
	// If nothing is selectable, choose the first, and we'll scroll like a browser
	if res.focus == -1 && len(widgets) > 0 {
		res.focus = 0
	}
	return res
}

func (w *SimpleListWalker) First() IWalkerPosition {
	if len(w.Widgets) == 0 {
		return nil
	}
	return ListPos(0)
}

func (w *SimpleListWalker) Last() IWalkerPosition {
	if len(w.Widgets) == 0 {
		return nil
	}
	return ListPos(len(w.Widgets) - 1)
}

func (w *SimpleListWalker) Length() int {
	return len(w.Widgets)
}

func (w *SimpleListWalker) At(pos IWalkerPosition) gowid.IWidget {
	var res gowid.IWidget
	ipos := int(pos.(ListPos))
	if ipos >= 0 && ipos < len(w.Widgets) {
		res = w.Widgets[ipos]
	}
	return res
}

func (w *SimpleListWalker) Focus() IWalkerPosition {
	return w.focus
}

func (w *SimpleListWalker) SetFocus(focus IWalkerPosition, app gowid.IApp) {
	w.focus = focus.(ListPos)
}

func (w *SimpleListWalker) Next(ipos IWalkerPosition) IWalkerPosition {
	pos := ipos.(ListPos)
	if int(pos) == len(w.Widgets)-1 {
		return ListPos(-1)
	} else {
		return pos + 1
	}
}

func (w *SimpleListWalker) Previous(ipos IWalkerPosition) IWalkerPosition {
	pos := ipos.(ListPos)
	if pos-1 == -1 {
		return ListPos(-1)
	} else {
		return pos - 1
	}
}

//======================================================================

type IListFns interface {
	RenderSubwidgets(gowid.IRenderSize, gowid.Selector, gowid.IApp) ([]SubRenders, SubRenders, []SubRenders)
	Walker() IWalker
	SetWalker(IWalker, gowid.IApp)
}

type IWidget interface {
	gowid.IWidget
	IListFns
}

type Widget struct {
	walker IWalker
	// This says how many lines to cut from the top of the widget rendered at the top of the listbox.
	// It might be too big to be rendered fully in the space.
	st      state
	options Options
	gowid.AddressProvidesID
	*gowid.Callbacks
	gowid.FocusCallbacks
	gowid.IsSelectable
}

type Options struct {
	//SelectedStyle gowid.ICellStyler // apply a style to the selected widget - orthogonal to focus styling
	DownKeys []vim.KeyPress
	UpKeys   []vim.KeyPress
}

type IndexedWidget struct {
	*Widget
	walker IBoundedWalker
}

type state struct {
	linesOffTop           int // used only if focus widget has more lines than can be displayed
	topToBottomRatio      float32
	topToBottomRatioValid bool // Means denominator is 0 if true i.e. at the bottom
}

func New(walker IWalker, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.DownKeys == nil {
		opt.DownKeys = vim.AllDownKeys
	}
	if opt.UpKeys == nil {
		opt.UpKeys = vim.AllUpKeys
	}
	res := &Widget{
		walker:  walker,
		options: opt,
	}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}
	res.goToTop()

	var _ gowid.IWidget = res

	return res
}

func NewBounded(walker IBoundedWalker, opts ...Options) *IndexedWidget {
	res := New(walker, opts...)
	return &IndexedWidget{
		Widget: res,
		walker: walker,
	}
}

func (w *Widget) String() string {
	cur := w.Walker().Focus()
	return fmt.Sprintf("list[pos=%v,f=%v]", cur, w.walker.At(cur))
}

func (w *Widget) Walker() IWalker {
	return w.walker
}

func (w *IndexedWidget) Walker() IWalker {
	return w.walker
}

func (w *Widget) SetWalker(l IWalker, app gowid.IApp) {
	w.walker = l
}

func (w *IndexedWidget) SetWalker(l IWalker, app gowid.IApp) {
	w.walker = l.(IBoundedWalker)
	w.Widget.SetWalker(l, app)
}

func (w *Widget) State() interface{} {
	return w.st
}

func (w *Widget) SetState(st interface{}, app gowid.IApp) {
	if state, ok := st.(state); !ok {
		panic(BadState)
	} else {
		w.st = state
	}
}

func (w *Widget) GoToTop(app gowid.IApp) {
	w.goToTop()
}

func (w *Widget) goToTop() {
	w.st.topToBottomRatioValid = true
	w.st.topToBottomRatio = 0
	w.st.linesOffTop = 0
}

func (w *Widget) GoToBottom(app gowid.IApp) {
	w.st.topToBottomRatioValid = false
}

func (w *Widget) GoToMiddle(app gowid.IApp) {
	w.st.topToBottomRatioValid = true
	w.st.topToBottomRatio = 0.5
	w.st.linesOffTop = 0
}

func (w *Widget) AtTop() bool {
	return w.st.topToBottomRatioValid && gwutil.AlmostEqual(float64(w.st.topToBottomRatio), 0.0)
}

func (w *Widget) AtBottom() bool {
	return !w.st.topToBottomRatioValid
}

func (w *Widget) InMiddle() bool {
	return w.st.topToBottomRatioValid && gwutil.AlmostEqual(float64(w.st.topToBottomRatio), 0.5)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return SubWidgetSize(w, size, focus, app)
}

func (w *Widget) CalculateOnScreen(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) (int, int, int, error) {
	return CalculateOnScreen(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

type SubRenders struct {
	Widget          gowid.IWidget
	Position        IWalkerPosition
	Canvas          gowid.ICanvas
	FullCanvasLines int
}

// IsChopped is a utility function for a SubRender struct that returns true if the canvas returned for this
// widget is smaller than the full size rendered (i.e. that it has been adjusted vertically)
func (r *SubRenders) IsChopped() bool {
	return r.Canvas.BoxRows() < r.FullCanvasLines
}

// RenderSubWidgets starts at the focus widget, rendering it, and returning the result as middle, a SubRenders
// struct. This tells the caller information about the widget rendered, including the full number of lines
// that would've been used if the provided render size had been large enough (this information tells the
// caller that the whole widget isn't displayed). After rendering the middle widget, the function renders
// Previous and Next widgets until the space above the middle widget and the space below the middle widget is
// filled.
func (w *Widget) RenderSubwidgets(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) (top []SubRenders, middle SubRenders, bottom []SubRenders) {
	cols, haveCols := size.(gowid.IColumns)
	rows, haveRows := size.(gowid.IRows)

	top = make([]SubRenders, 0)
	bottom = make([]SubRenders, 0)

	cur := w.Walker().Focus()
	curPos := cur
	curWidget := w.walker.At(curPos)

	if curWidget == nil {
		middle = SubRenders{nil, nil, gowid.NewCanvas(), 0}
	} else {
		var linesNeeded int
		haveLinesNeeded := haveRows
		if haveLinesNeeded {
			linesNeeded = rows.Rows()
		}
		var c gowid.ICanvas
		//foobar := styled.New(curWidget, gowid.MakeStyledAs(gowid.StyleReverse))
		var curToRender gowid.IWidget = curWidget
		if haveCols {
			//c = gowid.Render(curWidget, gowid.RenderFlowWith{C: cols.Columns()}, focus, app)
			c = curToRender.Render(gowid.RenderFlowWith{C: cols.Columns()}, focus, app)
		} else {
			//c = gowid.Render(curWidget, gowid.RenderFixed{}, focus, app)
			c = curToRender.Render(gowid.RenderFixed{}, focus, app)
		}
		creallines := c.BoxRows()
		middle = SubRenders{curWidget, curPos, c, creallines}

		// If the focus widget just rendered has more rows than the required size provided, then...
		if haveLinesNeeded && (c.BoxRows() > linesNeeded) {
			chopOffTop := w.st.linesOffTop
			// We don't chop off so much that it brings the next widget into view
			if c.BoxRows()-chopOffTop < linesNeeded {
				chopOffTop = c.BoxRows() - linesNeeded
			}
			c.Truncate(chopOffTop, c.BoxRows()-(linesNeeded+chopOffTop))
			middle = SubRenders{curWidget, curPos, c, creallines}
		} else {
			middle = SubRenders{curWidget, curPos, c, c.BoxRows()}
			upPos := curPos
			downPos := curPos
			var topLinesNeeded, bottomLinesNeeded int
			if haveLinesNeeded {
				if w.st.topToBottomRatioValid {
					topLinesNeeded = gwutil.RoundFloatToInt(float32(linesNeeded) * w.st.topToBottomRatio)
					bottomLinesNeeded = linesNeeded - (topLinesNeeded + c.BoxRows())
					if bottomLinesNeeded < 0 {
						topLinesNeeded -= -bottomLinesNeeded // take away from the top enough to bring the current widget into full display if possible
						bottomLinesNeeded = 0
						if topLinesNeeded < 0 {
							topLinesNeeded = 0
						}
					}
				} else {
					topLinesNeeded = linesNeeded - c.BoxRows()
				}
			}
			var upWidget, downWidget gowid.IWidget
			for {
				if haveLinesNeeded && (topLinesNeeded <= 0) {
					break
				}
				up := w.Walker().Previous(upPos)
				upPos = up
				upWidget = w.Walker().At(upPos)
				//upWidget, upPos = up.Widget, up.Pos
				if upWidget == nil {
					bottomLinesNeeded += topLinesNeeded
					break
				} else {
					var upC gowid.ICanvas
					if haveCols {
						upC = upWidget.Render(gowid.RenderFlowWith{C: cols.Columns()}, gowid.NotSelected, app)
					} else {
						upC = upWidget.Render(gowid.RenderFixed{}, gowid.NotSelected, app)
					}
					upreallines := upC.BoxRows()
					if haveLinesNeeded {
						if upreallines > topLinesNeeded {
							upC.Truncate(upreallines-topLinesNeeded, 0)
						}
						topLinesNeeded -= upC.BoxRows()
					}
					top = append(top, SubRenders{upWidget, upPos, upC, upreallines})
				}
			}
			for {
				if haveLinesNeeded && (bottomLinesNeeded <= 0) {
					break
				}
				down := w.Walker().Next(downPos)
				downPos = down
				downWidget = w.Walker().At(downPos)
				//downWidget, downPos = down.Widget, down.Pos
				if downWidget == nil {
					break
				} else {
					var downC gowid.ICanvas
					if haveCols {
						downC = downWidget.Render(gowid.RenderFlowWith{C: cols.Columns()}, gowid.NotSelected, app)
					} else {
						downC = downWidget.Render(gowid.RenderFixed{}, gowid.NotSelected, app)
					}
					downreallines := downC.BoxRows()
					if haveLinesNeeded {
						if downreallines > bottomLinesNeeded {
							downC.Truncate(0, downreallines-bottomLinesNeeded)
						}
						bottomLinesNeeded -= downC.BoxRows()
					}
					bottom = append(bottom, SubRenders{downWidget, downPos, downC, downreallines})
				}
			}
		}
	}
	return
}

func SubWidgetSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	switch sz := size.(type) {
	case gowid.IRenderBox:
		return gowid.RenderFlowWith{C: sz.BoxColumns()}
	case gowid.IRenderFlowWith:
		return sz
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}
}

func CalculateOnScreen(w IListFns, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) (int, int, int, error) {
	aboveMiddle, middle, belowMiddle := w.RenderSubwidgets(size, focus, app)
	mc := 0
	tc := 0
	bc := 0
	if len(aboveMiddle) > 0 {
		if pos, ok := aboveMiddle[len(aboveMiddle)-1].Position.(IBoundedWalkerPosition); ok {
			tc = pos.ToInt()
			// mc must not be nil, because if we rendered widgets above, then certainly there is a focus widget
			mc = len(aboveMiddle) + len(belowMiddle) + 1
		} else {
			return -1, -1, -1, errors.WithStack(WidgetIsUnboundedError{Type: aboveMiddle[len(aboveMiddle)-1].Position})
		}
	} else {
		// e.g. focus is at the top of the screen, so nothing to render above
		if middle.Widget != nil {
			// If == nil, this could be because there is no focus widget currently e.g. an empty list
			if pos, ok := middle.Position.(IBoundedWalkerPosition); ok {
				tc = pos.ToInt()
				mc = len(belowMiddle) + 1
			} else {
				return -1, -1, -1, errors.WithStack(WidgetIsUnboundedError{Type: middle.Position})
			}
		}
	}
	//allif len(bottom) > 0 {
	if i, ok := w.Walker().(IBoundedWalker); ok {
		bc = i.Length() - (mc + tc)
	} else {
		return -1, -1, -1, errors.WithStack(WidgetIsUnboundedError{Type: w.Walker()})
	}
	//}
	return tc, mc, bc, nil
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	rows, haveRows := size.(gowid.IRows)

	top, middle, bottom := w.RenderSubwidgets(size, focus, app)

	topC := gowid.NewCanvas()
	bottomC := gowid.NewCanvas()
	for i := len(top); i > 0; i-- {
		topC.AppendBelow(top[i-1].Canvas, false, false)
	}
	for _, ic := range bottom {
		bottomC.AppendBelow(ic.Canvas, false, false)
	}
	topC.AppendBelow(middle.Canvas, true, false)
	topC.AppendBelow(bottomC, false, false)

	if haveRows && (topC.BoxRows() < rows.Rows()) {
		gowid.AppendBlankLines(topC, rows.Rows()-topC.BoxRows())
	}

	return topC
}

func calcPrefPosition(curw gowid.IWidget) gwutil.IntOption {
	// Repeatedly unpack composite widgets until I have to stop. Look as I unpack for something that
	// exports a prefered column API. The widget might be ContainerWidget/StyledWidget/...
	var prefCol gwutil.IntOption
	for {
		if icol, ok := curw.(gowid.IPreferedPosition); ok {
			prefCol = icol.GetPreferedPosition()
			break
		}
		if curw2, ok2 := curw.(gowid.IComposite); ok2 {
			curw = curw2.SubWidget()
		} else {
			break
		}
	}
	return prefCol
}

func setPrefPosition(app gowid.IApp, curw gowid.IWidget, prefCol int) {
	for {
		if icol, ok := curw.(gowid.IPreferedPosition); ok {
			icol.SetPreferedPosition(prefCol, app)
			break
		}
		if curw2, ok2 := curw.(gowid.IComposite); ok2 {
			curw = curw2.SubWidget()
		} else {
			break
		}
	}
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := false
	rows, haveRows := size.(gowid.IRows)
	cols, haveCols := size.(gowid.IColumns)

	var numLinesToUse, numColumnsToUse int
	if haveRows && haveCols {
		numLinesToUse = rows.Rows()
		numColumnsToUse = cols.Columns()
	}

	sumSubRenders := func(renders []SubRenders) int {
		res := 0
		for _, r := range renders {
			res += r.FullCanvasLines
		}
		return res
	}

	var top, bottom, all []SubRenders
	var middle SubRenders
	initTMB := false
	initTopMiddleBottom := func() {
		if !initTMB {
			top, middle, bottom = w.RenderSubwidgets(size, focus, app)
			initTMB = true
		}
	}

	initLSR := false
	midIndex := -1

	initListOfSubRenders := func() int {
		if !initLSR {
			all = make([]SubRenders, len(top)+len(bottom)+1)
			j := 0
			for i := len(top); i > 0; i, j = i-1, j+1 {
				all[j] = top[i-1]
			}
			// Remember which one has focus
			midIndex = j
			all[j] = middle
			j++
			for i := 0; i < len(bottom); i, j = i+1, j+1 {
				all[j] = bottom[i]
			}
			initLSR = true
		}
		return midIndex
	}

	calculateScreenLines := func() {
		if numLinesToUse == 0 {
			for _, r := range all {
				numLinesToUse += r.Canvas.BoxRows()
				numColumnsToUse = gwutil.Max(numColumnsToUse, r.Canvas.BoxColumns())
			}
		}
	}

	userInputSize := func() gowid.IRenderSize {
		var sizeForInput gowid.IRenderSize
		if !haveRows && !haveCols {
			sizeForInput = gowid.RenderFixed{}
		} else {
			sizeForInput = gowid.RenderFlowWith{C: numColumnsToUse}
		}
		return sizeForInput
	}

	var subRenderSize gowid.IRenderSize
	if haveCols {
		subRenderSize = gowid.RenderFlowWith{C: cols.Columns()}
	} else {
		subRenderSize = gowid.RenderFixed{}
	}

	forChild := false
	childSelectable := false
	curi := w.Walker().Focus()
	position := curi
	cur := w.Walker().At(position)
	if cur == nil {
		return false
	}
	startPosition := position

	dirMoved := 0
	if evm, ok := ev.(*tcell.EventMouse); ok {
		initTopMiddleBottom()
		initListOfSubRenders()
		// If the render size provided didn't specify a number of rows, then
		// add up all the rows rendered and use that. This will cause problems with
		// infinite listboxes...
		calculateScreenLines()

		_, my := evm.Position()
		curY := 0

		for i, widgetRender := range all {
			if my < curY+widgetRender.Canvas.BoxRows() && my >= curY {
				for j := 0; !widgetRender.Position.Equal(position); j++ {
					// If we use w.Walker().Next() here, we don't update the state that tracks how far
					// down the screen we are. And regardless of whether we accept the input for this
					// event or not, we are going to change focus (if we didn't click on the focus
					// list item). So change focus as we go.
					if i > midIndex {
						// Need to walk forwards
						position = w.Walker().Next(position)
						dirMoved = 1
					} else {
						// Need to walk backwards
						position = w.Walker().Previous(position)
						dirMoved = -1
					}
				}
				sizeForInput := userInputSize()
				forChild = gowid.UserInputIfSelectable(widgetRender.Widget, gowid.TranslatedMouseEvent(ev, 0, -curY), sizeForInput, gowid.Focused, app)
				childSelectable = widgetRender.Widget.Selectable()
				break
			}
			curY += widgetRender.Canvas.BoxRows()
		}
	} else {
		if position != ListPos(-1) {
			sizeForInput := userInputSize()
			forChild = gowid.UserInputIfSelectable(cur, ev, sizeForInput, focus, app)
		}
	}

	scrollDown := false
	scrollUp := false
	pgDown := false
	pgUp := false
	toHome := false
	toEnd := false

	// If the child takes the user input, and its a key, then the list will never
	// handle it
	if evk, ok := ev.(*tcell.EventKey); !forChild && ok {

		k := evk.Key()
		switch {
		case k == tcell.KeyCtrlL:
			if !w.AtBottom() {
				if w.AtTop() {
					w.GoToBottom(app)
				} else {
					if w.InMiddle() {
						w.GoToTop(app)
					} else {
						w.GoToMiddle(app)
					}
				}
			} else {
				w.GoToMiddle(app)
			}
			res = true
		case k == tcell.KeyHome:
			toHome = true
		case k == tcell.KeyEnd:
			toEnd = true
		case vim.KeyIn(evk, w.options.DownKeys):
			scrollDown = true
		case vim.KeyIn(evk, w.options.UpKeys):
			scrollUp = true
		case k == tcell.KeyPgDn:
			pgDown = true
		case k == tcell.KeyPgUp:
			pgUp = true
		default:
		}
		// But if the input is from the mouse, the list can handle it as well as any subwidget. For example,
		// if the list holds checkbox widgets, a left mouse click might check the subwidget, but it can
		// also change the focus list item.
	} else if ev2, ok := ev.(*tcell.EventMouse); ok {
		switch ev2.Buttons() {
		case tcell.WheelDown:
			if !forChild {
				scrollDown = true
			}
		case tcell.WheelUp:
			if !forChild {
				scrollUp = true
			}

		case tcell.Button1:
			app.SetClickTarget(ev2.Buttons(), w)
			res = true
		case tcell.ButtonNone:
			if childSelectable {
				// tcell will report ButtonNone for mouse events which are simply pointer movements
				// (at least in my terminal). To distinguish this from a mouse release event, we track
				// the prior input's mouse state. If the last state was a mouse click, then this event
				// is processed as a mouse button release.
				if !app.GetLastMouseState().NoButtonClicked() {
					clickit := false
					app.ClickTarget(func(k tcell.ButtonMask, v gowid.IIdentityWidget) {
						if v != nil && v.ID() == w.ID() {
							clickit = true
						}
					})
					if clickit {
						// This means the mouse button was released over widget w, after earlier having
						// been clicked on widget w
						curPosition := startPosition
						saveState := w.st

						for {
							if curPosition.Equal(position) {
								res = true
								break
							} else if dirMoved > 0 && curPosition.GreaterThan(position) {
								res = false
								w.st = saveState
								w.Walker().SetFocus(startPosition, app)
								break
							} else if dirMoved < 0 && position.GreaterThan(curPosition) {
								res = false
								w.st = saveState
								w.Walker().SetFocus(startPosition, app)
								break
							}
							if dirMoved > 0 {
								_, curPosition = w.MoveToNextFocus(subRenderSize, focus, numLinesToUse, app)
							} else if dirMoved < 0 {
								_, curPosition = w.MoveToPreviousFocus(subRenderSize, focus, numLinesToUse, app)
							} else {
								panic(BadState)
							}
						}
						//res = true
					}
				}
			}
		}
	}

	var prefCol gwutil.IntOption

	if pgDown || pgUp {

		prefCol = calcPrefPosition(w.Walker().At(w.Walker().Focus()))

		if pgDown {

			// This means we're not at rendered at the bottom, so move
			// down until we are the bottom widget (no lines at bottom)
			startedAtBottom := w.AtBottom()

			cur := w.Walker().Next(position)
			curw := w.Walker().At(cur)
			if curw != nil {

				// We scrolled at least once, therefore we accepted the input.
				res = true

				// We need to move at least one widget on page down. This is it
				candidate := cur
				oldpos := candidate

				var curLines int
				var curBoundingBox gowid.IRenderBox

				topLines := make([]int, 0, 120)
				if !startedAtBottom {
					// test this widget with focus
					curBoundingBox = gowid.RenderSize(curw, subRenderSize, gowid.NotSelected, app)
					initTopMiddleBottom()
					topLines = append(topLines, sumSubRenders(top)+curBoundingBox.BoxRows())
				}

				// TODO: factor
				if !haveRows {
					initTopMiddleBottom()
					initListOfSubRenders()
					calculateScreenLines()
				}

			Loop1:
				for {

					// test this widget with focus
					curBoundingBox = gowid.RenderSize(curw, subRenderSize, focus, app)
					curLines = curBoundingBox.BoxRows()

					// if this widget would need more lines than we have left,
					if gwutil.Sum(topLines...)+curLines > numLinesToUse {
						// we stop here, use candidate
						break
					}

					// We can move again if this widget, rendered without focus, still doesn't
					// take us over the limit. If it does, then we have to stop
					curBoundingBox = gowid.RenderSize(curw, subRenderSize, gowid.NotSelected, app)
					curLines = curBoundingBox.BoxRows()

					// if this widget would need more lines than we have left,
					if gwutil.Sum(topLines...)+curLines > numLinesToUse {
						// we stop here, use candidate
						break
					}

					candidate = cur

					for {
						cur = w.Walker().Next(cur)
						w2 := w.Walker().At(cur)
						if w2 == nil {
							break Loop1
						}
						curLines := gowid.RenderSize(w2, subRenderSize, gowid.NotSelected, app).BoxRows()
						topLines = append(topLines, curLines)
						if w2.Selectable() {
							break
						}
					}

				}
				// At end of loop, invariant holds - candidate.Pos is the position for the widget
				w.Walker().SetFocus(candidate, app)
				if !oldpos.Equal(candidate) {
					gowid.RunWidgetCallbacks(w, gowid.FocusCB{}, app, w.Walker().At(candidate))
				}

				if !startedAtBottom {
					w.GoToBottom(app)
				}
			}
		}

		if pgUp {
			// This means we're not at rendered at the top, so move
			// down until we are the bottom widget (no lines at bottom)
			startedAtTop := !w.AtBottom() && gwutil.AlmostEqual(float64(w.st.topToBottomRatio), 0.0)

			cur := w.Walker().Previous(position)
			curw := w.Walker().At(cur)
			if curw != nil {

				// We scrolled at least once, therefore we accepted the input.
				res = true

				// We need to move at least one widget on page down. This is it
				candidate := cur
				oldpos := cur

				var curLines int
				var curBoundingBox gowid.IRenderBox

				bottomLines := make([]int, 0, 120)
				if !startedAtTop {
					// test this widget with focus
					curBoundingBox = gowid.RenderSize(curw, subRenderSize, gowid.NotSelected, app)
					initTopMiddleBottom()
					bottomLines = append(bottomLines, sumSubRenders(bottom)+curBoundingBox.BoxRows())
				}

				// TODO: factor
				if !haveRows {
					initTopMiddleBottom()
					initListOfSubRenders()
					calculateScreenLines()
				}

			Loop2:
				for {

					// test this widget with focus
					curBoundingBox = gowid.RenderSize(curw, subRenderSize, focus, app)
					curLines = curBoundingBox.BoxRows()

					// if this widget would need more lines than we have left,
					if gwutil.Sum(bottomLines...)+curLines > numLinesToUse {
						// we stop here, use candidate
						break
					}

					// We can move again if this widget, rendered without focus, still doesn't
					// take us over the limit. If it does, then we have to stop
					curBoundingBox = gowid.RenderSize(curw, subRenderSize, gowid.NotSelected, app)
					curLines = curBoundingBox.BoxRows()

					// if this widget would need more lines than we have left,
					if gwutil.Sum(bottomLines...)+curLines > numLinesToUse {
						// we stop here, use candidate
						break
					}

					candidate = cur

					for {
						cur = w.Walker().Previous(cur)
						w2 := w.Walker().At(cur)
						if w2 == nil {
							break Loop2
						}
						curLines := gowid.RenderSize(w2, subRenderSize, gowid.NotSelected, app).BoxRows()
						// We are moving down one more widget, so track how far we are from the top
						bottomLines = append(bottomLines, curLines)
						if w2.Selectable() {
							break
						}
					}
				}
				// At end of loop, invariant holds - candidate.Pos is the position for the widget
				w.Walker().SetFocus(candidate, app)
				if !oldpos.Equal(candidate) {
					gowid.RunWidgetCallbacks(w, gowid.FocusCB{}, app, w.Walker().At(candidate))
				}

				if !startedAtTop {
					w.GoToTop(app)
				}

			}
		}

		if res && !prefCol.IsNone() {
			setPrefPosition(app, w.Walker().At(w.Walker().Focus()), prefCol.Val())
		}

	}

	if scrollDown || scrollUp {
		initTopMiddleBottom()

		res = true

		prefCol = calcPrefPosition(middle.Widget)

		if scrollDown {
			// This means that the middle widget could not fit entirely in the screen provided, and that
			// we have not scrolled to the bottom of the middle widget yet
			if middle.IsChopped() && (middle.Canvas.BoxRows()+w.st.linesOffTop < middle.FullCanvasLines) {
				w.st.linesOffTop += 1
			} else {
				res, _ = w.MoveToNextFocus(subRenderSize, focus, numLinesToUse, app)
			}
		}
		if scrollUp {
			// If the current widget itself is chopped, and is missing lines at the top, then reduce the number of missing lines
			if middle.IsChopped() && (w.st.linesOffTop > 0) {
				w.st.linesOffTop -= 1
			} else {
				res, _ = w.MoveToPreviousFocus(subRenderSize, focus, numLinesToUse, app)
			}
		}

		if res && !prefCol.IsNone() {
			setPrefPosition(app, w.Walker().At(w.Walker().Focus()), prefCol.Val())
		}
	}

	if toHome || toEnd {

		prefCol = calcPrefPosition(w.Walker().At(w.Walker().Focus()))
		oldpos := w.Walker().Focus()

		if toHome {
			if homer, ok := w.Walker().(IWalkerHome); ok {
				pos := homer.First()
				if pos != nil {
					w.Walker().SetFocus(pos, app)
					w.GoToTop(app)
					res = true
				}
			}
		}
		if toEnd {
			if ender, ok := w.Walker().(IWalkerEnd); ok {
				pos := ender.Last()
				if pos != nil {
					w.Walker().SetFocus(pos, app)
					w.GoToBottom(app)
					res = true
				}
			}
		}

		if res && !prefCol.IsNone() {
			setPrefPosition(app, w.Walker().At(w.Walker().Focus()), prefCol.Val())
		}

		newwandpos := w.Walker().Focus()

		if !oldpos.Equal(newwandpos) {
			gowid.RunWidgetCallbacks(w, gowid.FocusCB{}, app, w.Walker().At(newwandpos))
		}

	}

	return forChild || res
}

func (w *Widget) MoveToNextFocus(subRenderSize gowid.IRenderSize, focus gowid.Selector, screenLines int, app gowid.IApp) (bool, IWalkerPosition) {

	cur := w.Walker().Focus()
	curw := w.Walker().At(cur)
	if curw == nil {
		return false, cur
	}
	oldPos := cur
	curLinesNoFocus := gowid.RenderSize(curw, subRenderSize, gowid.NotSelected, app).BoxRows()

	// from that, get the next widget and next position. The nextw is used to run callbacks.
	var next IWalkerPosition
	var nextw gowid.IWidget
	for {
		next = w.Walker().Next(cur)
		nextw = w.Walker().At(next)
		if nextw == nil {
			return false, nil
		}
		if nextw.Selectable() {
			break
		}
		curLinesNoFocus += gowid.RenderSize(nextw, subRenderSize, gowid.NotSelected, app).BoxRows()
		cur = next
	}

	w.Walker().SetFocus(next, app)

	nextLines := gowid.RenderSize(nextw, subRenderSize, focus, app).BoxRows()

	// curWidgetLines has the number of lines used by the current focus widget when rendered. Compute how
	// many line)s should be above it, and how many below it.
	var computedLinesAbove, computedLinesBelow int
	if !w.AtBottom() {
		computedLinesAbove = gwutil.RoundFloatToInt(float32(gwutil.Max(0, screenLines)) * w.st.topToBottomRatio)
		computedLinesAbove += curLinesNoFocus
		computedLinesBelow = screenLines - (computedLinesAbove + nextLines)
		if computedLinesBelow <= 0 {
			w.GoToBottom(app)
		} else {
			w.st.topToBottomRatioValid = true
			w.st.topToBottomRatio = float32(computedLinesAbove) / float32(screenLines)
		}
	}
	w.st.linesOffTop = 0

	// Do this at the end in case the focus callback wants to save the list state too.
	if !next.Equal(oldPos) {
		gowid.RunWidgetCallbacks(w, gowid.FocusCB{}, app, nextw)
	}

	return true, next
}

func (w *Widget) MoveToPreviousFocus(subRenderSize gowid.IRenderSize, focus gowid.Selector, screenLines int, app gowid.IApp) (bool, IWalkerPosition) {

	wasAtBottom := w.AtBottom()

	cur := w.Walker().Focus()
	curw := w.Walker().At(cur)
	oldpos := cur
	curLinesFocus := gowid.RenderSize(curw, subRenderSize, focus, app).BoxRows()
	betweenNoFocus := 0

	// from that, get the next widget and next position. The nextw is used to run callbacks.
	var prev IWalkerPosition
	var prevw gowid.IWidget
	for {
		prev = w.Walker().Previous(cur)
		prevw = w.Walker().At(prev)
		if prevw == nil {
			return false, nil
		}
		if prevw.Selectable() {
			break
		}
		betweenNoFocus += gowid.RenderSize(prevw, subRenderSize, gowid.NotSelected, app).BoxRows()
		cur = prev
	}

	w.Walker().SetFocus(prev, app)

	prevLinesNoFocus := gowid.RenderSize(prevw, subRenderSize, gowid.NotSelected, app).BoxRows()

	// curWidgetLines has the number of lines used by the current focus widget when rendered. Compute how
	// many lines should be above it, and how many below it.
	var computedLinesAbove int
	if wasAtBottom {
		computedLinesAbove = gwutil.Max(0, screenLines) - (curLinesFocus + betweenNoFocus + prevLinesNoFocus)
	} else {
		computedLinesAbove = gwutil.RoundFloatToInt(float32(gwutil.Max(0, screenLines)) * w.st.topToBottomRatio)
		// Preserve lines *above* focus - it feels the most natural when scrolling. So if the
		// previous widget (below) takes 3 lines to render with focus, but only 1 without, then add just
		// one because that widget will only contribute 1 when it's no longer current.
		computedLinesAbove -= (prevLinesNoFocus + betweenNoFocus)
	}
	if computedLinesAbove <= 0 {
		prevLinesFocus := gowid.RenderSize(prevw, subRenderSize, focus, app).BoxRows()
		w.GoToTop(app)                                               // widget is logically top, but might have lines cut if too big (see next line)
		w.st.linesOffTop = gwutil.Max(0, prevLinesFocus-screenLines) // in case prev doesn't fit, start at bottom
	} else {
		w.st.topToBottomRatioValid = true
		w.st.topToBottomRatio = float32(computedLinesAbove) / float32(screenLines)
	}

	// Do this at the end in case the focus callback wants to save the list state too.
	if !prev.Equal(oldpos) {
		gowid.RunWidgetCallbacks(w, gowid.FocusCB{}, app, prevw)
	}

	return true, prev
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
