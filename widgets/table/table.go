// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package table provides a widget that renders tabular output.
package table

import (
	"fmt"
	"strconv"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/isselected"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gdamore/tcell"
	lru "github.com/hashicorp/golang-lru"
)

//======================================================================

// RowId is used to uniquely identify a row. The idea here is that a table
// row can move in the order of rows rendered, if the table is sorted, but
// we would like to preserve the ability to cache the row's widgets (e.g.
// the selected column in the row). So a client of an ITable should first
// look up a RowId given an actual row to be rendered (1st, 2nd, etc). Then
// with the RowId, the client asks for the RowWidgets. Clients of ITable
// can then cache the rendered row using the RowId as a lookup field. Even
// if that row moves around in the order rendered, it can be found in the
// cache.
type RowId int

// IModel is implemented by any type which can provide arrays of
// widgets for a given table row, and optionally header widgets.
type IModel interface {
	Columns() int
	RowIdentifier(row int) (RowId, bool)   // return a unique ID for row
	CellWidgets(row RowId) []gowid.IWidget // nil means EOD
	HeaderWidgets() []gowid.IWidget        // nil means no headers
	VerticalSeparator() gowid.IWidget
	HorizontalSeparator() gowid.IWidget
	HeaderSeparator() gowid.IWidget
	Widths() []gowid.IWidgetDimension
}

type IMakeHeader interface {
	HeaderWidget([]gowid.IWidget, int) gowid.IWidget
}

// IBoundedTable implements ITable and can also provide the total number of
// rows in the table.
type IBoundedModel interface {
	IModel
	Rows() int
}

// ICompare is the type of the compare function used when sorting a table's
// rows.
type ICompare interface {
	Less(i, j string) bool
}

//======================================================================

// StringCompare is a unit type that satisfies ICompare, and can be used
// for lexicographically comparing strings.
type StringCompare struct{}

func (s StringCompare) Less(i, j string) bool {
	return i < j
}

var _ ICompare = StringCompare{}

// IntCompare is a unit type that satisfies ICompare, and can be used
// for numerically comparing ints.
type IntCompare struct{}

func (s IntCompare) Less(i, j string) bool {
	x, err1 := strconv.Atoi(i)
	y, err2 := strconv.Atoi(j)
	if err1 == nil && err2 == nil {
		return x < y
	} else {
		return false
	}
}

var _ ICompare = IntCompare{}

// FloatCompare is a unit type that satisfies ICompare, and can be used
// for numerically comparing float64 values.
type FloatCompare struct{}

func (s FloatCompare) Less(i, j string) bool {
	x, err1 := strconv.ParseFloat(i, 64)
	y, err2 := strconv.ParseFloat(j, 64)
	if err1 == nil && err2 == nil {
		return x < y
	} else {
		return false
	}
}

var _ ICompare = FloatCompare{}

//======================================================================

// ListWithPreferedColumn acts like a list.Widget but also satisfies gowid.IPreferedPosition.
// The idea is that if the list rows consist of columns, then moving up and down the list
// should preserve the selected column.
type ListWithPreferedColumn struct {
	list.IWidget // so we can use bounded or unbounded lists
}

var _ gowid.IPreferedPosition = (*ListWithPreferedColumn)(nil)
var _ gowid.IComposite = (*ListWithPreferedColumn)(nil)

func (l *ListWithPreferedColumn) SubWidget() gowid.IWidget {
	return l.IWidget
}

func (l *ListWithPreferedColumn) GetPreferedPosition() gwutil.IntOption {
	res := gwutil.NoneInt()
	fpos := l.IWidget.Walker().Focus()
	w := l.IWidget.Walker().At(fpos)
	if w != nil {
		res = gowid.PrefPosition(w)
	}
	return res
}

func (l *ListWithPreferedColumn) SetPreferedPosition(col int, app gowid.IApp) {
	fpos := l.IWidget.Walker().Focus()
	w := l.IWidget.Walker().At(fpos)
	if w != nil {
		gowid.SetPrefPosition(w, col, app)
	}
}

func (w *ListWithPreferedColumn) String() string {
	return fmt.Sprintf("listc")
}

//======================================================================

type Position int

var _ list.IBoundedWalkerPosition = Position(0)
var _ list.IWalkerPosition = Position(0)

func (t Position) ToInt() int {
	return int(t)
}

func (t Position) Equal(pos list.IWalkerPosition) bool {
	if t2, ok := pos.(Position); ok {
		return t == t2
	} else {
		panic(gowid.InvalidTypeToCompare{LHS: t, RHS: pos})
	}
}

func (t Position) GreaterThan(pos list.IWalkerPosition) bool {
	if t2, ok := pos.(Position); ok {
		return t > t2
	} else {
		panic(gowid.InvalidTypeToCompare{LHS: t, RHS: pos})
	}
}

//======================================================================

type RenderWithUnitsMax struct {
	gowid.RenderWithUnits
	gowid.RenderMax
}

var _ gowid.IRenderMax = widthOneHeightMax

//======================================================================

// Widget wraps a widget and aligns it vertically according to the supplied arguments. The wrapped
// widget can be aligned to the top, bottom or middle, and can be provided with a specific height in #lines.
//
type Widget struct {
	wrapper          *pile.Widget
	header           gowid.IWidget
	listw            *ListWithPreferedColumn
	model            IModel
	cur              int
	cache            *lru.Cache
	flowHorzDivider  *gowid.ContainerWidget
	flowVertDivider  *gowid.ContainerWidget
	flowTableDivider *gowid.ContainerWidget
	opt              Options
	*gowid.Callbacks
	gowid.FocusCallbacks
	gowid.IsSelectable
}

var _ gowid.IWidget = (*Widget)(nil)

type BoundedWidget struct {
	*Widget
}

var _ list.IBoundedWalker = (*BoundedWidget)(nil)
var _ list.IWalkerHome = (*BoundedWidget)(nil)
var _ list.IWalkerEnd = (*BoundedWidget)(nil)

type Options struct {
	CacheSize int
}

func New(model IModel, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Fill in Widget later once constructed.
	listw := &ListWithPreferedColumn{}

	// Construct the table first, then set the list later. That's because when the
	// pile is constructed, it tries to find the first selectable widget; but lw's
	// embedded list widget is nil at the time.

	sz := 4096
	if opt.CacheSize > 0 {
		sz = opt.CacheSize
	}

	cache, err := lru.New(sz)
	if err != nil {
		panic(err)
	}

	// res acts as a ListWalker and a widget
	res := &Widget{
		listw: listw,
		cur:   0,
		cache: cache,
	}

	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}

	switch model.(type) {
	case IBoundedModel:
		listw.IWidget = list.NewBounded(&BoundedWidget{res})
	default:
		listw.IWidget = list.New(res)
	}

	res.update(listw, 0, model, opt)

	return res
}

var _ gowid.IWidget = (*Widget)(nil)

func (w *Widget) update(listw *ListWithPreferedColumn, row int, model IModel, opt Options) {
	// Save whether we were in the header or the packet lists
	pf := -1
	hf := -1
	// setPileFocus := false
	if w.wrapper != nil {
		pf = w.wrapper.Focus()
		if w.model != nil && len(w.model.HeaderWidgets()) > 0 {
			// 0th must be header
			hf = gowid.Focus(w.wrapper.SubWidgets()[0])
		}
	}
	// //if w.wrapper.Focus()
	// //}
	// //if bm, ok :=
	// if w.model != nil {
	// 	if bm, ok := w.model.(IBoundedTable); ok && bm.Rows() == 0 {
	// 		setPileFocus = true
	// 	}
	// } else {
	// 	setPileFocus = true
	// }

	// if setPileFocus {
	// 	setPileFocus = false
	// 	if model != nil {
	// 		if bm, ok := model.(IBoundedTable); ok && bm.Rows() > 0 {
	// 			setPileFocus = true
	// 		}
	// 	}
	// }

	pileWidgets := make([]gowid.IContainerWidget, 0)
	var flowHorzDivider *gowid.ContainerWidget
	var flowVertDivider *gowid.ContainerWidget
	var flowTableDivider *gowid.ContainerWidget

	if model.HorizontalSeparator() != nil {
		flowHorzDivider = &gowid.ContainerWidget{model.HorizontalSeparator(), gowid.RenderFlow{}}
	}
	if model.VerticalSeparator() != nil {
		flowVertDivider = &gowid.ContainerWidget{model.VerticalSeparator(), widthOneHeightMax}
	}
	if model.HeaderSeparator() != nil {
		flowTableDivider = &gowid.ContainerWidget{model.HeaderSeparator(), gowid.RenderFlow{}}
	}

	if flowTableDivider != nil {
		pileWidgets = append(pileWidgets, flowTableDivider)
	}
	hws := model.HeaderWidgets() // widgets
	var hw *columns.Widget
	if hws != nil && len(hws) > 0 {
		var hw2 gowid.IWidget
		if nm, ok := model.(IMakeHeader); ok {
			hw2 = nm.HeaderWidget(hws, hf)
		} else {
			cws := make([]gowid.IContainerWidget, 0)
			if flowVertDivider != nil {
				cws = append(cws, flowVertDivider)
			}

			for i, w := range hws {
				var dim gowid.IWidgetDimension = gowid.RenderWithWeight{1}
				if model.Widths() != nil && i < len(model.Widths()) {
					dim = model.Widths()[i]
				}
				cws = append(cws, &gowid.ContainerWidget{w, dim})
				if flowVertDivider != nil {
					cws = append(cws, flowVertDivider)
				}
			}
			hw = columns.New(cws, columns.Options{
				StartColumn: hf,
			})
			hw2 = hw
		}
		pileWidgets = append(pileWidgets, &gowid.ContainerWidget{hw2, gowid.RenderFlow{}})
		if flowTableDivider != nil {
			pileWidgets = append(pileWidgets, flowTableDivider)
		}
	}

	// Fill in Widget later once constructed.
	pileWidgets = append(pileWidgets, &gowid.ContainerWidget{listw, gowid.RenderWithWeight{1}})

	// Construct the table first, then set the list later. That's because when the
	// pile is constructed, it tries to find the first selectable widget; but lw's
	// embedded list widget is nil at the time.

	if hw != nil {
		w.header = hw
	}
	w.model = model
	w.flowHorzDivider = flowHorzDivider
	w.flowVertDivider = flowVertDivider
	w.flowTableDivider = flowTableDivider
	w.opt = opt

	if pf == -1 {
		// This is imperfect. If the table model is updated in such a way that
		// the dividers change, then re-using the previous pile index is wrong -
		// it needs to be done logically i.e. recalculated
		w.wrapper = pile.New(pileWidgets)
	} else {
		w.wrapper = pile.New(pileWidgets, pile.Options{
			StartRow: pf,
		})
	}
}

func (w *BoundedWidget) First() list.IWalkerPosition {
	if w.Length() == 0 {
		return nil
	}
	return Position(0)
}

func (w *BoundedWidget) Last() list.IWalkerPosition {
	if w.Length() == 0 {
		return nil
	}
	if w.flowHorzDivider != nil {
		return Position(w.Length()*2 - 2)
	} else {
		return Position(w.Length() - 1)
	}
}

func (w *BoundedWidget) BoundedWalker() list.IBoundedWalker {
	return w.listw.Walker().(list.IBoundedWalker)
}

func (w *BoundedWidget) Length() int {
	//return w.listw.IWidget.(*list.IndexedWidget).Walker().(list.IBoundedWalker).Length()
	return w.Model().(IBoundedModel).Rows()
}

func (w *Widget) CalculateOnScreen(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) (int, int, int, error) {
	return list.CalculateOnScreen(w.listw, size, focus, app)
}

func (w *Widget) SetModel(model IModel, app gowid.IApp) {
	oldpos, olderr := w.FocusXY()
	w.cache.Purge() // gcla later todo
	w.update(w.listw, w.cur, model, w.opt)
	if olderr == nil {
		w.SetFocusXY(app, oldpos) // mght not be able to set old focus, if model shape has changed
	} else {

		// If we previously had no data and now we do, change focus to the data element
		// in the pile
		// if bm, bmok := model.(IBoundedTable); bmok && bm.Rows() > 0 && model.Columns() > 0 {
		// 	w.wrapper.SetFocus(app, 1)
		// }
		// No focus in old model, so try to set a default one
		//w.SetFocusXY(app, Coords{0, 0})
		// if bm, ok := model.(IBoundedTable); ok && bm.Rows() > 0 && model.Columns() > 0 {
		// 	if len(model.HeaderWidgets()) > 0 {
		// 		// 	//w.SetFocusXY(app, Coords{0, 1})
		// 		// 	w.SetFocusXY(app, Coords{0, 0})
		// 		// } else {
		// 		// 	w.SetFocusXY(app, Coords{0, 0})
		// 	}
		// }
	}
	newpos, newerr := w.FocusXY()
	if olderr != newerr || oldpos != newpos {
		gowid.RunWidgetCallbacks(w.Callbacks, gowid.FocusCB{}, app, w)
	}
}

func (w *Widget) Lower() *ListWithPreferedColumn {
	return w.listw
}

func (w *Widget) SetLower(l *ListWithPreferedColumn) {
	w.listw = l
}

func (w *Widget) Cache() *lru.Cache {
	return w.cache
}

func (w *Widget) Model() IModel {
	return w.model
}

func (w *Widget) CurrentRow() int {
	return w.cur
}

func (w *Widget) SetCurrentRow(p Position) {
	w.cur = int(p)
}

func (w *Widget) VertDivider() gowid.IContainerWidget {
	// Do it this way to avoid having a nil value in an interface type that isn't nil
	if w.flowVertDivider == nil {
		return nil
	} else {
		return w.flowVertDivider
	}
}

func (w *Widget) TableDivider() gowid.IContainerWidget {
	// Do it this way to avoid having a nil value in an interface type that isn't nil
	if w.flowTableDivider == nil {
		return nil
	} else {
		return w.flowTableDivider
	}
}

func (w *Widget) HorzDivider() gowid.IWidget {
	// Do it this way to avoid having a nil value in an interface type that isn't nil
	if w.flowHorzDivider == nil {
		return nil
	} else {
		return w.flowHorzDivider
	}
}

func (w *Widget) String() string {
	return fmt.Sprintf("table")
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	oldpos, olderr := w.FocusXY()
	res := w.wrapper.UserInput(ev, size, focus, app)
	newpos, newerr := w.FocusXY()
	if olderr != newerr || oldpos != newpos {
		gowid.RunWidgetCallbacks(w.Callbacks, gowid.FocusCB{}, app, w)
	}
	return res
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return w.wrapper.Render(size, focus, app)
}

func (w *Widget) Up(lines int, size gowid.IRenderSize, app gowid.IApp) {
	for i := 0; i < lines; i++ {
		w.wrapper.UserInput(tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone), size, gowid.Focused, app)
	}
}

func (w *Widget) Down(lines int, size gowid.IRenderSize, app gowid.IApp) {
	for i := 0; i < lines; i++ {
		w.wrapper.UserInput(tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone), size, gowid.Focused, app)
	}
}

func (w *Widget) UpPage(num int, size gowid.IRenderSize, app gowid.IApp) {
	for i := 0; i < num; i++ {
		w.wrapper.UserInput(tcell.NewEventKey(tcell.KeyPgUp, ' ', tcell.ModNone), size, gowid.Focused, app)
	}
}

func (w *Widget) DownPage(num int, size gowid.IRenderSize, app gowid.IApp) {
	for i := 0; i < num; i++ {
		w.wrapper.UserInput(tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone), size, gowid.Focused, app)
	}
}

type IRowToWidget interface {
	VertDivider() gowid.IContainerWidget
	Model() IModel
	Cache() *lru.Cache
}

// propagatePrefPosition will ensure prefered position is set for all candidates in an isselected
// widget, so that - if appropriate - a prefered position is rendered whether or not
// the widget has focus (or indeed is selected).
type propagatePrefPosition struct {
	*isselected.WidgetExt
}

var _ gowid.IWidget = (*propagatePrefPosition)(nil)
var _ gowid.IComposite = (*propagatePrefPosition)(nil)
var _ gowid.IPreferedPosition = (*propagatePrefPosition)(nil)

func (w *propagatePrefPosition) SubWidget() gowid.IWidget {
	return w.WidgetExt
}

func (w *propagatePrefPosition) GetPreferedPosition() gwutil.IntOption {
	// Get the Focused one because when this is called, the child widget (which will be columns)
	// will be in focus.
	//
	// This list above this will get the pref position of the current container widget, prior to
	// effecting user input. e.g. column 2. Then it might move focus to the next row, so a new
	// column widget. It will then set the prefered column of that new focus widget to 2. so when
	// querying the current pref col, get the one in focus. A big hack.
	res := gowid.PrefPosition(w.Focused)
	return res
}

func (w *propagatePrefPosition) SetPreferedPosition(col int, app gowid.IApp) {
	gowid.SetPrefPosition(w.Not, col, app)
	gowid.SetPrefPosition(w.Selected, col, app)
	gowid.SetPrefPosition(w.Focused, col, app)
}

func (t *Widget) RowToWidget(ws []gowid.IWidget) gowid.IWidget {
	return RowToWidget(t, ws)
}

func RowToWidget(t IRowToWidget, ws []gowid.IWidget) gowid.IWidget {
	var res gowid.IWidget
	if ws != nil {
		cws := make([]gowid.IContainerWidget, 0)
		if t.VertDivider() != nil {
			cws = append(cws, t.VertDivider())
		}
		for i, w := range ws {
			var dim gowid.IWidgetDimension = gowid.RenderWithWeight{1}
			if t.Model().Widths() != nil && i < len(t.Model().Widths()) {
				dim = t.Model().Widths()[i]
			}
			cws = append(cws, &gowid.ContainerWidget{w, dim})
			if t.VertDivider() != nil {
				cws = append(cws, t.VertDivider())
			}
		}
		colsWhenFocusOrSelected := columns.New(cws)
		colsWhenNotSelected := columns.New(cws, columns.Options{
			// Don't let columns set focus.select for selected child
			DoNotSetSelected: true,
		})

		// This has the following effect:
		//
		// - if a row of the table is either selected or in focus, the conditional
		//   widget will render a regular columns. In that case, the columns will
		//   set focus.Selected for the "cell" that is currently selected. Styling
		//   can be applied as appropriate.
		//
		// - if a row of the table is not selected (so also not focused), then
		//   the columns widget will not set focus.Selected. That means rows
		//   of the table that are not the current row cannot make decisions
		//   based on which cell is selected. This is likely what is needed -
		//   applying cell-level styling to the selected table row, and not
		//   other table rows (e.g. highlighting the selected column, but not
		//   in rows not in focus)
		//
		res = &propagatePrefPosition{
			WidgetExt: isselected.NewExt(
				colsWhenNotSelected,
				colsWhenFocusOrSelected,
				colsWhenFocusOrSelected,
			),
		}
	}
	return res
}

type IWidgetAt interface {
	HorzDivider() gowid.IWidget
	RowToWidget(ws []gowid.IWidget) gowid.IWidget
	IRowToWidget
}

// WidgetAt is used by the type that satisfies list.IWalker - therefore it provides
// a row for each line of the list. That means that if the table has dividers, it
// provides them too.
func (t *Widget) AtRow(pos int) gowid.IWidget {
	return WidgetAt(t, pos)
}

// Provide the pos'th "row" widget
func WidgetAt(t IWidgetAt, pos int) gowid.IWidget {
	if pos < 0 {
		return nil
	}
	if t.HorzDivider() != nil {
		if pos%2 == 1 {
			return t.HorzDivider()
		} else {
			pos = pos / 2
		}
	}

	var res gowid.IWidget
	if rid, ok := t.Model().RowIdentifier(pos); ok {
		if cw, ok := t.Cache().Get(rid); ok {
			res = cw.(gowid.IWidget)
		} else {
			ws := t.Model().CellWidgets(rid)
			res = t.RowToWidget(ws)
			if res != nil {
				t.Cache().Add(rid, res)
			}
			return res
		}
	}
	return res
}

type IFocus interface {
	IWidgetAt
	AtRow(int) gowid.IWidget
	CurrentRow() int
}

// list.IWalker
func (t *Widget) Focus() list.IWalkerPosition {
	return Focus(t)
}

// list.IWalker
func (t *Widget) At(pos list.IWalkerPosition) gowid.IWidget {
	return t.AtRow(int(pos.(Position)))
}

// list.IWalker
func Focus(t IFocus) list.IWalkerPosition {
	return Position(t.CurrentRow())
}

type ISetFocus interface {
	SetCurrentRow(pos Position)
}

// In order to implement list.IWalker
// list.IWalker
func (t *Widget) SetFocus(pos list.IWalkerPosition, app gowid.IApp) {
	SetFocus(t, pos)
}

// In order to implement list.IWalker
func SetFocus(t ISetFocus, pos list.IWalkerPosition) {
	if t2, ok := pos.(Position); !ok {
		panic(fmt.Errorf("Invalid position %v passed to SetFocus", pos))
	} else {
		t.SetCurrentRow(t2)
	}
}

// In order to implement list.IWalker
// list.IWalker
func (t *Widget) Next(ipos list.IWalkerPosition) list.IWalkerPosition {
	if pos, ok := ipos.(Position); !ok {
		panic(fmt.Errorf("Invalid position %v passed to Next", ipos))
	} else {
		npos := int(pos) + 1
		return Position(npos)
	}
}

// In order to implement list.IWalker
// list.IWalker
func (t *Widget) Previous(ipos list.IWalkerPosition) list.IWalkerPosition {
	if pos, ok := ipos.(Position); !ok {
		panic(fmt.Errorf("Invalid position %v passed to Prev", ipos))
	} else {
		ppos := int(pos) - 1
		return Position(ppos)
	}
}

type IGoToBottom interface {
	GoToBottom(app gowid.IApp)
}

func (t *Widget) GoToBottom(app gowid.IApp) bool {
	if b, ok := t.listw.IWidget.(IGoToBottom); ok {
		b.GoToBottom(app)
		return true
	}
	return false
}

type IGoToMiddle interface {
	GoToMiddle(app gowid.IApp)
}

func (t *Widget) GoToMiddle(app gowid.IApp) bool {
	if b, ok := t.listw.IWidget.(IGoToMiddle); ok {
		b.GoToMiddle(app)
		return true
	}
	return false
}

type Coords struct {
	Column int
	Row    int
}

func (c Coords) String() string {
	return fmt.Sprintf("(%d,%d)", c.Column, c.Row)
}

type NoFocus struct{}

func (n NoFocus) Error() string {
	return "No table data"
}

func findFocus(w gowid.IWidget) gowid.IWidget {
	w = gowid.FindInHierarchy(w, true, gowid.WidgetPredicate(func(w gowid.IWidget) bool {
		var res bool
		if _, ok := w.(gowid.IFocus); ok {
			res = true
		}
		return res
	}))
	return w
}

// FocusXY returns the coordinates of the focus widget in the table, potentially including
// the header if one is configured. This is grungy and needs to account for the cell separators
// in its arithmetic.
func (t *Widget) FocusXY() (Coords, error) {
	var col, row int
	addOne := false
	if t.header != nil {
		focusedOnHeader := false
		if t.TableDivider() != nil && t.wrapper.Focus() == 1 {
			focusedOnHeader = true
		} else if t.TableDivider() == nil && t.wrapper.Focus() == 0 {
			focusedOnHeader = true
		}
		if focusedOnHeader {
			row = 0
			fw := findFocus(t.header)
			if fw == nil {
				return Coords{}, NoFocus{}
			}
			col = fw.(gowid.IFocus).Focus()
			if t.VertDivider() != nil {
				col = (col - 1) / 2
			}
			return Coords{Column: col, Row: row}, nil
		} else {
			addOne = true
		}
	}
	rw := t.listw.Walker().Focus()
	rww := t.listw.Walker().At(rw)
	colwi := gowid.FindInHierarchy(rww, true, gowid.WidgetPredicate(func(w gowid.IWidget) bool {
		_, ok := w.(*columns.Widget)
		return ok
	}))
	if colwi == nil {
		//panic(fmt.Errorf("Could not find columns widget within table structure"))
		return Coords{}, NoFocus{}
	}
	colw := colwi.(*columns.Widget)
	col = colw.Focus()
	if t.VertDivider() != nil {
		col = (col - 1) / 2
	}
	row = int(rw.(Position))
	if t.HorzDivider() != nil {
		row = row / 2
	}
	if addOne {
		row++
	}
	return Coords{Column: col, Row: row}, nil
}

func (t *Widget) SetFocusXY(app gowid.IApp, xy Coords) {
	oldpos, olderr := t.FocusXY()
	defer func() {
		newpos, newerr := t.FocusXY()
		if olderr != newerr || oldpos != newpos {
			gowid.RunWidgetCallbacks(t.Callbacks, gowid.FocusCB{}, app, t)
		}
	}()

	if t.header != nil {
		if xy.Row == 0 {
			if t.TableDivider() != nil {
				t.wrapper.SetFocus(app, 1)
			} else {
				t.wrapper.SetFocus(app, 0)
			}

			fw := findFocus(t.header)
			if fw == nil {
				return
			}
			fw2 := fw.(gowid.IFocus)

			if t.VertDivider() != nil {
				fw2.SetFocus(app, xy.Column*2+1)
			} else {
				fw2.SetFocus(app, xy.Column)
			}
			return
		} else {
			if t.TableDivider() != nil {
				t.wrapper.SetFocus(app, 3)
			} else {
				t.wrapper.SetFocus(app, 1)
			}
			xy.Row--
		}
	}
	// Have to set list
	walker := t.listw.Walker()
	if t.HorzDivider() != nil {
		walker.SetFocus(Position(xy.Row*2), app)
	} else {
		walker.SetFocus(Position(xy.Row), app)
	}

	colwi := gowid.FindInHierarchy(walker.At(walker.Focus()), true, gowid.WidgetPredicate(func(w gowid.IWidget) bool {
		_, ok := w.(*columns.Widget)
		return ok
	}))
	if colwi != nil {
		cols := colwi.(*columns.Widget)
		//cols := walker.Focus().Widget.(*propagatePrefPosition).Widget.(*isselected.WidgetExt).Not.(*columns.Widget)
		if t.VertDivider() != nil {
			cols.SetFocus(app, xy.Column*2+1)
		} else {
			cols.SetFocus(app, xy.Column)
		}
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
