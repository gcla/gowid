// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package gowid provides widgets and tools for constructing compositional terminal user interfaces.
package gowid

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid/gwutil"
	"github.com/gdamore/tcell"
	"github.com/pkg/errors"
)

//======================================================================

type IRows interface {
	Rows() int
}

type IColumns interface {
	Columns() int
}

// IRenderSize is the type of objects that can specify how a widget is to be rendered.
// This is the empty interface, and only serves as a placeholder at the moment. In
// practise, actual rendering sizes will be determined by an IFlowDimension, IBoxDimension
// or an IFixedDimension
type IRenderSize interface{}

//======================================================================

// Widgets that are used in containers such as Pile or Columns must implement
// this interface. It specifies how each subwidget of the container should be
// rendered.
//
type IWidgetDimension interface {
	ImplementsWidgetDimension() // This exists as a marker so that IWidgetDimension is not empty, meaning satisfied by any struct.
}

type IRenderFixed interface {
	IWidgetDimension
	Fixed() // dummy
}

type IRenderFlowWith interface {
	IWidgetDimension
	FlowColumns() int
}

type IRenderFlow interface {
	IWidgetDimension
	Flow() // dummy
}

type IBox interface {
	BoxColumns() int
	BoxRows() int
}

type IRenderBox interface {
	IWidgetDimension
	IBox
}

type IRenderWithWeight interface {
	IWidgetDimension
	Weight() int
}

type IRenderRelative interface {
	IWidgetDimension
	Relative() float64
}

type IRenderWithUnits interface {
	IWidgetDimension
	Units() int
}

// Used in widgets laid out side-by-side - intended to have the effect that these widgets are
// rendered last and provided a height that corresponds to the max of the height of those
// widgets already rendered.
type IRenderMax interface {
	MaxHeight() // dummy
}

// Used in widgets laid out side-by-side - intended to limit the width of a widget
// which may otherwise be specified to be dimensioned in relation to the width available.
// This can let the layout algorithm give more space (e.g. maximized terminal) to widgets
// that can use it by constraining those that don't need it.
type IRenderMaxUnits interface {
	MaxUnits() int
}

//======================================================================

// RenderFlowWith is an object passed to a widget's Render function that specifies that
// it should be rendered with a set number of columns, but using as many rows as
// the widget itself determines it needs.
type RenderFlowWith struct {
	C int
}

func MakeRenderFlow(columns int) RenderFlowWith {
	return RenderFlowWith{C: columns}
}

func (r RenderFlowWith) FlowColumns() int {
	return r.C
}

// For IColumns
func (r RenderFlowWith) Columns() int {
	return r.FlowColumns()
}

func (r RenderFlowWith) String() string {
	return fmt.Sprintf("flowwith(c:%d)", r.C)
}

func (r RenderFlowWith) ImplementsWidgetDimension() {}

//======================================================================

// RenderBox is an object passed to a widget's Render function that specifies that
// it should be rendered with a set number of columns and rows.
type RenderBox struct {
	C int
	R int
}

func MakeRenderBox(columns, rows int) RenderBox {
	return RenderBox{C: columns, R: rows}
}

func (r RenderBox) BoxColumns() int {
	return r.C
}

func (r RenderBox) BoxRows() int {
	return r.R
}

// For IColumns
func (r RenderBox) Columns() int {
	return r.BoxColumns()
}

// For IRows
func (r RenderBox) Rows() int {
	return r.BoxRows()
}

func (r RenderBox) ImplementsWidgetDimension() {}

func (r RenderBox) String() string {
	return fmt.Sprintf("box(c:%d,r:%d)", r.C, r.R)
}

//======================================================================

// RenderFixed is an object passed to a widget's Render function that specifies that
// the widget itself will determine its own size.
type RenderFixed struct{}

func MakeRenderFixed() RenderFixed {
	return RenderFixed{}
}

func (f RenderFixed) String() string {
	return "fixed"
}

func (f RenderFixed) Fixed() {}

func (r RenderFixed) ImplementsWidgetDimension() {}

//======================================================================

type RenderWithWeight struct {
	W int
}

func (f RenderWithWeight) String() string {
	return fmt.Sprintf("weight(%d)", f.W)
}

func (f RenderWithWeight) Weight() int {
	return f.W
}

func (r RenderWithWeight) ImplementsWidgetDimension() {}

//======================================================================

// RenderFlow is used by widgets that embed an inner widget, like hpadding.Widget.
// It directs the outer widget how it should render the inner widget. If the outer
// widget is rendered in box mode, the inner widget should be rendered in flow mode,
// using the box's number of columns. If the outer widget is rendered in flow mode,
// the inner widget should be rendered in flow mode with the same number of columns.
type RenderFlow struct{}

func (s RenderFlow) Flow() {}

func (f RenderFlow) String() string {
	return "flow"
}

func (r RenderFlow) ImplementsWidgetDimension() {}

//======================================================================

// RenderMax is used in widgets laid out side-by-side - it's intended to
// have the effect that these widgets are rendered last and provided a
// height/width that corresponds to the max of the height/width of those widgets
// already rendered.
type RenderMax struct{}

func (s RenderMax) MaxHeight() {}

func (f RenderMax) String() string {
	return "maxheight"
}

//======================================================================

// RenderWithUnits is used by widgets within a container. It specifies the number
// of columns or rows to use when rendering.
type RenderWithUnits struct {
	U int
}

func (f RenderWithUnits) Units() int {
	return f.U
}

func (f RenderWithUnits) String() string {
	return fmt.Sprintf("units(%d)", f.U)
}

func (r RenderWithUnits) ImplementsWidgetDimension() {}

//======================================================================

// RenderWithRatio is used by widgets within a container
type RenderWithRatio struct {
	R float64
}

func (f RenderWithRatio) Relative() float64 {
	return f.R
}

func (f RenderWithRatio) String() string {
	return fmt.Sprintf("ratio(%f)", f.R)
}

func (r RenderWithRatio) ImplementsWidgetDimension() {}

//======================================================================

type DimensionError struct {
	Size IRenderSize
	Dim  IWidgetDimension
	Row  int
}

var _ error = DimensionError{}

func (e DimensionError) Error() string {
	if e.Row == -1 {
		return fmt.Sprintf("Dimension spec %v of type %T cannot be used with render size %v of type %T", e.Dim, e.Dim, e.Size, e.Size)
	} else {
		return fmt.Sprintf("Dimension spec %v of type %T cannot be used with render size %v of type %T and advised row %d", e.Dim, e.Dim, e.Size, e.Size, e.Row)
	}
}

//======================================================================

type WidgetSizeError struct {
	Widget   interface{}
	Size     IRenderSize
	Required string // in case I only need an interface - not sure how to capture it and not concrete type
}

var _ error = WidgetSizeError{}

func (e WidgetSizeError) Error() string {
	if e.Required == "" {
		return fmt.Sprintf("Widget %v cannot be rendered with %v of type %T", e.Widget, e.Size, e.Size)
	} else {
		return fmt.Sprintf("Widget %v cannot be rendered with %v of type %T - requires %s", e.Widget, e.Size, e.Size, e.Required)
	}
}

//======================================================================

// IContainerWidget is the type of an object that contains a widget and
// a render object that determines how it is rendered within a container of
// widgets. Note that it itself is an IWidget.
type IContainerWidget interface {
	IWidget
	IComposite
	Dimension() IWidgetDimension
	SetDimension(IWidgetDimension)
}

//======================================================================

// ContainerWidget is a simple implementation that conforms to
// IContainerWidget.  It can be used to pass widgets to containers like
// pile.Widget and columns.Widget.
type ContainerWidget struct {
	IWidget
	D IWidgetDimension
}

func (ww ContainerWidget) Dimension() IWidgetDimension {
	return ww.D
}

func (ww *ContainerWidget) SetDimension(d IWidgetDimension) {
	ww.D = d
}

func (w *ContainerWidget) SubWidget() IWidget {
	return w.IWidget
}

func (w *ContainerWidget) SetSubWidget(wi IWidget, app IApp) {
	w.IWidget = wi
}

func (w *ContainerWidget) String() string {
	return fmt.Sprintf("container[%v,%v]", w.D, w.IWidget)
}

var _ IContainerWidget = (*ContainerWidget)(nil)

//======================================================================

// Three states - false+false, false+true, true+true
type Selector struct {
	Focus    bool
	Selected bool
}

var Focused = Selector{
	Focus:    true,
	Selected: true,
}

var Selected = Selector{
	Focus:    false,
	Selected: true,
}

var NotSelected = Selector{
	Focus:    false,
	Selected: false,
}

// SelectIf returns a Selector with the Selected field set dependent on the
// supplied condition only. The Focus field is set based on the supplied condition
// AND the receiver's Focus field. Used by composite widgets with multiple children
// to allow children to change their state dependent on whether they are selected
// but independent of whether the widget is currently in focus.
func (s Selector) SelectIf(cond bool) Selector {
	return Selector{
		Focus:    s.Focus && cond,
		Selected: cond,
	}
}

// And returns a Selector with both Selected and Focus set dependent on the
// supplied condition AND the receiver. Used to propagate Selected and Focus
// state to sub widgets for input and rendering.
func (s Selector) And(cond bool) Selector {
	return Selector{
		Focus:    s.Focus && cond,
		Selected: s.Selected && cond,
	}
}

func (s Selector) String() string {
	return fmt.Sprintf("[Focus:%v,Selected:%v]", s.Focus, s.Selected)
}

// ISelectChild is implemented by any type that controls whether or not it
// will set focus.Selected on its currently "selected" child. For example, a
// columns widget will have a notion of a child widget that will take focus.
// The user may want to render that widget in a way that highlights the
// selected child, even when the columns widget itself does not have
// focus. The columns widget will set focus.Selected on Render() and
// UserInput() calls depending on the result of SelectChild() - if
// focus.Selected is set, then a styling widget can change the look of the
// widget appropriately.
//
type ISelectChild interface {
	SelectChild(Selector) bool // Whether or not this widget will set focus.Selected for its selected child
}

// IWidget is the interface of any object acting as a gowid widget.
//
// Render() is provided a size (cols, maybe rows), whether or not the widget
// is in focus, and a context (palette, etc). It must return an object conforming
// to gowid's ICanvas, which is a representation of what can be displayed in the
// terminal.
//
// RenderSize() is used by clients that need to know only how big the widget
// will be when rendered. It is expected to be cheaper to compute in some cases
// than Render(), but a fallback is to run Render() then return the size of the
// canvas.
//
// Selectable() should return true if this widget is designed for interaction e.g.
// a Button would return true, but a Text widget would return false. Note that,
// like urwid, returning false does not guarantee the widget will never have
// focus - it might be given focus if there is no other option (no other
// selectable widgets in the container, for example).
//
// UserEvent() is provided the TCell event (mouse or keyboard action),
// the size spec that would be given to Render(), whether or not the widget
// has focus, and access to the application, useful for effecting changes
// like changing colors, running a function, or quitting. The render size is
// needed because the widget might have to pass the event down to children
// widgets, and the correct one may depend on the coordinates of a mouse
// click relative to the dimensions of the widget itself.
//
type IWidget interface {
	Render(size IRenderSize, focus Selector, app IApp) ICanvas
	RenderSize(size IRenderSize, focus Selector, app IApp) IRenderBox
	UserInput(ev interface{}, size IRenderSize, focus Selector, app IApp) bool
	Selectable() bool
}

// IIdentity is used for widgets that support being a click target - so it
// is possible to link the widget that is the target of MouseReleased with
// the one that was the target of MouseLeft/Right/Middle when they might
// not necessarily be the same object (i.e. rebuilt widget hierarchy in
// between). Also used to name callbacks so they can be removed (since
// function objects can't be compared)
//
type IIdentity interface {
	ID() interface{}
}

// IComposite is an interface for anything that has a concept of a single
// "inner" widget. This applies to certain widgets themselves
// (e.g. ButtonWidget) and also to the App object which holds the top-level
// view.
//
type IComposite interface {
	SubWidget() IWidget
}

// IComposite is an interface for anything that has a concept of a single
// settable "inner" widget. This applies to certain widgets themselves
// (e.g. ButtonWidget) and also to the App object which holds the top-level
// view.
//
type ISettableComposite interface {
	IComposite
	SetSubWidget(IWidget, IApp)
}

// ISubWidgetSize returns the size argument that should be provided to
// render the inner widget based on the size argument provided to the
// containing widget.
type ISubWidgetSize interface {
	SubWidgetSize(size IRenderSize, focus Selector, app IApp) IRenderSize
}

// ICompositeWidget is an interface implemented by widgets that contain one
// subwidget. Further implented methods could make it an IButtonWidget for
// example, which then means the RenderButton() function can be exploited
// to implement Render(). If you make a new Button by embedding
// ButtonWidget, you may be able to implement Render() by simply calling
// RenderButton().
//
type ICompositeWidget interface {
	IWidget
	IComposite
	ISubWidgetSize
}

// ICompositeMultiple is an interface for widget containers that have multiple
// children and that support specifying how the children are laid out relative
// to each other.
//
type ICompositeMultiple interface {
	SubWidgets() []IWidget
}

// ISettableSubWidgetsis implemented by a type that maintains a collection of child
// widgets (like pile, columns) and that allows them to be changed.
type ISettableSubWidgets interface {
	SetSubWidgets([]IWidget, IApp)
}

// ICompositeMultipleDimensions is an interface for collections of widget dimensions,
// used in laying out some container widgets.
//
type ICompositeMultipleDimensions interface {
	ICompositeMultiple
	Dimensions() []IWidgetDimension
}

// ISettableDimensions is implemented by types that maintain a collection of
// dimensions - to be used by containers that use these dimensions to layout
// their children widgets.
//
type ISettableDimensions interface {
	SetDimensions([]IWidgetDimension, IApp)
}

// IFocus is a container widget concept that describes which widget will be
// the target of keyboard input.
//
type IFocus interface {
	Focus() int
	SetFocus(app IApp, i int)
}

// IFindNextSelectable is for any object that can iterate to its next or
// previous object
type IFindNextSelectable interface {
	FindNextSelectable(dir Direction, wrap bool) (int, bool)
}

// ICompositeMultipleWidget is a widget that implements ICompositeMultiple. The
// widget must support computing the render-time size of any of its children
// and setting focus.
//
type ICompositeMultipleWidget interface {
	IWidget
	ICompositeMultipleDimensions
	IFocus
	// SubWidgetSize should return the IRenderSize value that will be used to render
	// an inner widget given the size used to render the outer widget and an
	// IWidgetDimension (such as units, weight, etc)
	SubWidgetSize(size IRenderSize, val int, sub IWidget, dim IWidgetDimension) IRenderSize
	// RenderSubWidgets should return an array of canvases representing each inner
	// widget, rendered in the context of the containing widget with the supplied
	// size argument.
	RenderSubWidgets(size IRenderSize, focus Selector, focusIdx int, app IApp) []ICanvas
	// RenderedSubWidgetsSizes should return a bounding box for each inner widget
	// when the containing widget is rendered with the provided size. Note that this
	// is not the same as rendering each inner widget separately, because the
	// container context might result in size adjustments e.g. adjusting the
	// height of inner widgets to make sure they're aligned vertically.
	RenderedSubWidgetsSizes(size IRenderSize, focus Selector, focusIdx int, app IApp) []IRenderBox
}

// IClickable is implemented by any type that implements a Click()
// method, intended to be run in response to a user interaction with the
// type such as left mouse click or hitting enter.
//
type IClickable interface {
	Click(app IApp)
}

// IKeyPress is implemented by any type that implements a KeyPress()
// method, intended to be run in response to a user interaction with the
// type such as hitting the escape key.
//
type IKeyPress interface {
	KeyPress(key IKey, app IApp)
}

// IClickTracker is implemented by any type that can track the state of whether it
// was clicked. This is trivial, and may just be a boolean flag. It's intended for
// widgets that want to change their look when a mouse button is clicked when they are
// in focus, but before the button is released - to indicate that the widget is about
// to be activated. Of course if the user moves the cursor off the widget then
// releases the mouse button, the widget will not be activated.
type IClickTracker interface {
	SetClickPending(bool)
}

// IClickableWidget is implemented by any widget that implements a Click()
// method, intended to be run in response to a user interaction with the
// widget such as left mouse click or hitting enter. A widget implementing
// Click() and ID() may be able to run UserInputCheckedWidget() for its
// UserInput() implementation.
//
type IClickableWidget interface {
	IWidget
	IClickable
}

// IIdentityWidget is implemented by any widget that provides an ID()
// function that identifies itself and allows itself to be compared against
// other IIdentity implementers. This is intended be to used to check
// whether or not the widget that was in focus when a mouse click was
// issued is the same widget in focus when the mouse is released. If so
// then the widget was "clicked". This allows gowid to run the action on
// release rather than on click, which is more forgiving of mistaken
// clicks. The widget in focus on release may be logically the same widget
// as the one clicked, but possibly a different object, if the widget
// hierarchy was somehow rebuilt in response to the first click - so to
// receive the click event, make sure the newly built widget has the same
// ID() as the original (e.g. a serialized representation of a position in
// a ListWalker)
//
type IIdentityWidget interface {
	IWidget
	IIdentity
}

// IPreferedPosition is implemented by any widget that supports a prefered
// column or row (position in a dimension), meaning it understands what
// subwidget is at the current dimensional coordinate, and can move its focus
// widget to a new position. This is modeled on Urwid's get_pref_col()
// feature, which tries to provide a sensible switch of focus widget when
// moving the cursor vertically around the screen - instead of having it hop
// left and right depending on which widget happens to be in focus at the
// current y coordinate.
//
type IPreferedPosition interface {
	GetPreferedPosition() gwutil.IntOption
	SetPreferedPosition(col int, app IApp)
}

// IMenuCompatible is implemented by any widget that can set a subwidget.
// It's used by widgets like menus that need to inject themselves into
// the widget hierarchy close to the root (to be rendered over the main
// "view") i.e. the current root is made a child of the new menu widget,
// whuch becomes the new root.
type IMenuCompatible interface {
	IWidget
	ISettableComposite
}

//======================================================================

type ICopyResult interface {
	ClipName() string
	ClipValue() string
}

type CopyResult struct {
	Name string
	Val  string
}

var _ ICopyResult = CopyResult{}

func (c CopyResult) ClipName() string {
	return c.Name
}
func (c CopyResult) ClipValue() string {
	return c.Val
}

type IClipboard interface {
	Clips(app IApp) []ICopyResult
}

type IClipboardSelected interface {
	AlterWidget(w IWidget, app IApp) IWidget
}

//======================================================================

// AddressProvidesID is a convenience struct that can be embedded in widgets.
// It provides an ID() function by simply returning the pointer of its
// caller argument. The ID() function is for widgets that want to implement
// IIdentity, which is needed by containers that want to compare widgets.
// For example, if the user clicks on a button.Widget, the app can be used
// to save that widget. When the click is released, the button's UserInput
// function tries to determine whether the mouse was released over the
// same widget that was clicked. It can do this by comparing the widgets'
// ID() values. Note that this will not work if new button widgets are
// created each time Render/UserInput is called (because the caller
// will change).
type AddressProvidesID struct{}

func (a *AddressProvidesID) ID() interface{} {
	return a
}

//======================================================================

// RejectUserInput is a convenience struct that can be embedded in widgets
// that don't accept any user input.
//
type RejectUserInput struct{}

func (r RejectUserInput) UserInput(ev interface{}, size IRenderSize, focus Selector, app IApp) bool {
	return false
}

//======================================================================

// NotSelectable is a convenience struct that can be embedded in widgets. It provides
// a function that simply return false to the call to Selectable()
//
type NotSelectable struct{}

func (r *NotSelectable) Selectable() bool {
	return false
}

//======================================================================

// IsSelectable is a convenience struct that can be embedded in widgets. It provides
// a function that simply return true to the call to Selectable()
//
type IsSelectable struct{}

func (r *IsSelectable) Selectable() bool {
	return true
}

//======================================================================

// SelectableIfAnySubWidgetsAre is useful for various container widgets.
//
func SelectableIfAnySubWidgetsAre(w ICompositeMultipleDimensions) bool {
	for _, widget := range w.SubWidgets() {
		if widget.Selectable() {
			return true
		}
	}
	return false
}

//======================================================================

type IHAlignment interface {
	ImplementsHAlignment()
}

type HAlignRight struct{}
type HAlignMiddle struct{}
type HAlignLeft struct {
	Margin int
}

func (h HAlignRight) ImplementsHAlignment()  {}
func (h HAlignMiddle) ImplementsHAlignment() {}
func (h HAlignLeft) ImplementsHAlignment()   {}

type IVAlignment interface {
	ImplementsVAlignment()
}
type VAlignBottom struct{}
type VAlignMiddle struct{}
type VAlignTop struct {
	Margin int
}

func (v VAlignBottom) ImplementsVAlignment() {}
func (v VAlignMiddle) ImplementsVAlignment() {}
func (v VAlignTop) ImplementsVAlignment()    {}

//======================================================================

// CalculateRenderSizeFallback can be used by widgets that cannot easily compute a value
// for RenderSize without actually rendering the widget and measuring the bounding box.
// It assumes that if IRenderBox size is provided, then the widget's canvas when rendered
// will be that large, and simply returns the box. If an IRenderFlow is provided, then
// the widget is rendered, and the bounding box is returned.
func CalculateRenderSizeFallback(w IWidget, size IRenderSize, focus Selector, app IApp) RenderBox {
	var res RenderBox
	switch sz := size.(type) {
	case IRenderBox:
		res.R = sz.BoxRows()
		res.C = sz.BoxColumns()
	default:
		c := Render(w, size, focus, app)
		res.R = c.BoxRows()
		res.C = c.BoxColumns()
	}
	return res
}

// Render currently passes control through to the widget's Render method. Having
// this function allows for easier instrumentation of the Render path. The function
// returns a canvas representing the rendered widget.
func Render(w IWidget, size IRenderSize, focus Selector, app IApp) ICanvas {
	res := w.Render(size, focus, app)

	// Enable when debugging
	// PanicIfCanvasNotRightSize(res, size)

	return res
}

// UserInputIfSelectable will return false if the widget is not selectable; otherwise it will
// try the widget's UserInput function.
func UserInputIfSelectable(w IWidget, ev interface{}, size IRenderSize, focus Selector, app IApp) bool {
	res := false
	if w.Selectable() {
		res = UserInput(w, ev, size, focus, app)
	}

	return res
}

// UserInput currently passes control through to the widget's UserInput method. Having
// this function allows for easier instrumentation of the UserInput path. UserInput
// should return true if this widget "handles" the provided input, and false
// otherwise. This return value can guide parent widgets and help them determine
// whether or not they should then consume the input event. Note that returning
// true does not guarantee no other widget will also handle the event - for example
// the ListBox widget may handle a mouse click by changing the focus widget, but
// also allow the child widget at focus to receive the click too.
func UserInput(w IWidget, ev interface{}, size IRenderSize, focus Selector, app IApp) bool {
	return w.UserInput(ev, size, focus, app)
}

// RenderSize currently passes control through to the widget's RenderSize
// method. Having this function allows for easier instrumentation of the
// RenderSize path. RenderSize is intended to compute the size of the canvas
// that will be generated when the widget is rendered. Some parent widgets
// need this value from their children, and it might be possible to compute
// it much more cheaply than rendering the widget in order to determine
// the canvas size only.
func RenderSize(w IWidget, size IRenderSize, focus Selector, app IApp) IRenderBox {
	return w.RenderSize(size, focus, app)
}

// SubWidgetSize currently passes control through to the widget's SubWidgetSize
// method. Having this function allows for easier instrumentation of the
// SubWidgetSize path. The function should compute the size that it will itself
// use to render its child widget; for example, a framing widget rendered
// with IRenderBox might return a RenderBox value that is 2 units smaller
// in both height and width.
func SubWidgetSize(w ICompositeWidget, size IRenderSize, focus Selector, app IApp) IRenderSize {
	return w.SubWidgetSize(size, focus, app)
}

// RenderRoot is called from the App application object when beginning the
// widget rendering process. It starts at the root of the widget hierarchy
// with an IRenderBox size argument equal to the size of the current terminal.
func RenderRoot(w IWidget, t *App) {
	maxX, maxY := t.TerminalSize()
	canvas := Render(w, RenderBox{C: maxX, R: maxY}, Focused, t)

	// tcell will apply its default style to empty cells. But because gowid's model
	// is to layer styles, here we explicitly merge each canvas cell on top of a cell
	// constructed with the tcell default style. Therefore if the tcell default applies
	// an underline, for example, then each canvas cell will be merged on top of a cell
	// with an underline. If the upper cell masks out underline, then it won't show. But
	// if the upper cell doesn't mask out the underline, it will show.
	if paletteDefault, ok := t.CellStyler("default"); ok {
		defFg := ColorDefault
		defBg := ColorDefault
		fgCol, bgCol, style := paletteDefault.GetStyle(t)
		mode := t.GetColorMode()
		defFg = IColorToTCell(fgCol, defFg, mode)
		defBg = IColorToTCell(bgCol, defBg, mode)
		RangeOverCanvas(canvas, CellRangeFunc(func(c Cell) Cell {
			return MakeCell(c.codePoint, defFg, defBg, style).MergeDisplayAttrsUnder(c)
		}))
	}

	Draw(canvas, t, t.GetScreen())
}

func FindNextSelectableFrom(w ICompositeMultipleDimensions, start int, dir Direction, wrap bool) (int, bool) {
	dup := CopyWidgets(w.SubWidgets())
	return FindNextSelectableWidget(dup, start, dir, wrap)
}

func FindNextSelectableWidget(w []IWidget, pos int, dir Direction, wrap bool) (int, bool) {
	if len(w) == 0 {
		return -1, false
	}
	if pos == -1 {
		if dir <= 0 {
			pos = len(w)
		}
	}
	start := pos
	for {
		pos = pos + int(dir)
		if pos == -1 {
			if wrap {
				pos = len(w) - 1
			} else {
				return -1, false
			}
		} else if pos == len(w) {
			if wrap {
				pos = 0
			} else {
				return -1, false
			}
		}
		if w[pos] != nil && w[pos].Selectable() {
			return pos, true
		} else if pos == start {
			return -1, false
		}
	}
}

func FixCanvasHeight(c ICanvas, size IRenderSize) {
	// Make sure that if we're rendered as a box, we have enough rows.
	if box, ok := size.(IRenderBox); ok {
		if c.BoxRows() < box.BoxRows() {
			AppendBlankLines(c, box.BoxRows()-c.BoxRows())
		} else if c.BoxRows() > box.BoxRows() {
			c.Truncate(0, c.BoxRows()-box.BoxRows())
		}
	}
}

type IAppendBlankLines interface {
	BoxColumns() int
	AppendBelow(c IAppendCanvas, doCursor bool, makeCopy bool)
}

func AppendBlankLines(c IAppendBlankLines, iters int) {
	for i := 0; i < iters; i++ {
		line := make([]Cell, c.BoxColumns())
		c.AppendBelow(LineCanvas(line), false, false)
	}
}

//======================================================================

type ICallbackRunner interface {
	RunWidgetCallbacks(name interface{}, app IApp, w IWidget)
}

// IWidgetChangedCallback defines the types that can be used as callbacks
// that are issued when widget properties change. It expects a function
// Changed() that is called with the current app and the widget that is
// issuing the callback. It also expects to conform to IIdentity, so that
// one callback instance can be compared to another - this is to allow
// callbacks to be removed correctly, if that is required.
type IWidgetChangedCallback interface {
	IIdentity
	Changed(app IApp, widget IWidget, data ...interface{})
}

// WidgetChangedFunction meets the IWidgetChangedCallback interface, for simpler
// usage.
type WidgetChangedFunction func(app IApp, widget IWidget)

func (f WidgetChangedFunction) Changed(app IApp, widget IWidget, data ...interface{}) {
	f(app, widget)
}

// WidgetCallback is a simple struct with a name field for IIdentity and
// that embeds a WidgetChangedFunction to be issued as a callback when a widget
// property changes.
type WidgetCallback struct {
	Name interface{}
	WidgetChangedFunction
}

func MakeWidgetCallback(name interface{}, fn WidgetChangedFunction) WidgetCallback {
	return WidgetCallback{
		Name:                  name,
		WidgetChangedFunction: fn,
	}
}

func (f WidgetCallback) ID() interface{} {
	return f.Name
}

func RunWidgetCallbacks(c ICallbacks, name interface{}, app IApp, data ...interface{}) {
	data2 := append([]interface{}{app}, data...)
	c.RunCallbacks(name, data2...)
}

type widgetChangedCallbackProxy struct {
	IWidgetChangedCallback
}

func (p widgetChangedCallbackProxy) Call(args ...interface{}) {
	t := args[0].(IApp)
	w := args[1].(IWidget)
	p.IWidgetChangedCallback.Changed(t, w, args[2:]...)
}

func AddWidgetCallback(c ICallbacks, name interface{}, cb IWidgetChangedCallback) {
	c.AddCallback(name, widgetChangedCallbackProxy{cb})
}

func RemoveWidgetCallback(c ICallbacks, name interface{}, id IIdentity) {
	c.RemoveCallback(name, id)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''
// Common callbacks

// QuitFn can be used to construct a widget callback that terminates your
// application. It can be used as the second argument of the
// WidgetChangedCallback struct which implements IWidgetChangedCallback.
func QuitFn(app IApp, widget IWidget) {
	app.Quit()
}

// SubWidgetCallbacks is a convenience struct for embedding in a widget, providing methods
// to add and remove callbacks that are executed when the widget's child is modified.
type SubWidgetCallbacks struct {
	ICallbacks
}

func (w *SubWidgetCallbacks) OnSetSubWidget(f IWidgetChangedCallback) {
	AddWidgetCallback(w, SubWidgetCB{}, f)
}

func (w *SubWidgetCallbacks) RemoveOnSetSubWidget(f IIdentity) {
	RemoveWidgetCallback(w, SubWidgetCB{}, f)
}

//======================================================================

// SubWidgetsCallbacks is a convenience struct for embedding in a widget, providing methods
// to add and remove callbacks that are executed when the widget's children are modified.
type SubWidgetsCallbacks struct {
	ICallbacks
}

func (w *SubWidgetsCallbacks) OnSetSubWidgets(f IWidgetChangedCallback) {
	AddWidgetCallback(w, SubWidgetsCB{}, f)
}

func (w *SubWidgetsCallbacks) RemoveOnSetSubWidgets(f IIdentity) {
	RemoveWidgetCallback(w, SubWidgetsCB{}, f)
}

//======================================================================

// ClickCallbacks is a convenience struct for embedding in a widget, providing methods
// to add and remove callbacks that are executed when the widget is "clicked".
type ClickCallbacks struct {
	ICallbacks
}

func (w *ClickCallbacks) OnClick(f IWidgetChangedCallback) {
	AddWidgetCallback(w, ClickCB{}, f)
}

func (w *ClickCallbacks) RemoveOnClick(f IIdentity) {
	RemoveWidgetCallback(w, ClickCB{}, f)
}

//======================================================================

// KeyPressCallbacks is a convenience struct for embedding in a widget, providing methods
// to add and remove callbacks that are executed when the widget is "clicked".
type KeyPressCallbacks struct {
	ICallbacks
}

func (w *KeyPressCallbacks) OnKeyPress(f IWidgetChangedCallback) {
	AddWidgetCallback(w, KeyPressCB{}, f)
}

func (w *KeyPressCallbacks) RemoveOnKeyPress(f IIdentity) {
	RemoveWidgetCallback(w, KeyPressCB{}, f)
}

//======================================================================

// FocusCallbacks is a convenience struct for embedding in a widget, providing methods
// to add and remove callbacks that are executed when the widget's focus widget changes.
type FocusCallbacks struct {
	ICallbacks
}

func (w *FocusCallbacks) OnFocusChanged(f IWidgetChangedCallback) {
	AddWidgetCallback(w, FocusCB{}, f)
}

func (w *FocusCallbacks) RemoveOnFocusChanged(f IIdentity) {
	RemoveWidgetCallback(w, FocusCB{}, f)
}

//======================================================================

// ICellProcessor is a general interface used by several gowid types for
// processing a range of Cell types. For example, a canvas provides a function
// to range over its contents, each cell being handed to an ICellProcessor.
type ICellProcessor interface {
	ProcessCell(cell Cell) Cell
}

// CellRangeFunc is an adaptor for a simple function to implement ICellProcessor.
type CellRangeFunc func(cell Cell) Cell

// ProcessCell hands over processing to the adapted function.
func (f CellRangeFunc) ProcessCell(cell Cell) Cell {
	return f(cell)
}

//======================================================================

// IKey represents a keypress. It's a subset of tcell.EventKey because it doesn't
// capture the time of the keypress. It can be used by widgets to customize what
// keypresses they respond to.
type IKey interface {
	Rune() rune
	Key() tcell.Key
	Modifiers() tcell.ModMask
}

func KeysEqual(k1, k2 IKey) bool {
	res := true
	res = res && (k1.Key() == k2.Key())
	if k1.Key() == tcell.KeyRune && k1.Key() == tcell.KeyRune {
		res = res && (k1.Modifiers() == k2.Modifiers())
		res = res && (k1.Rune() == k2.Rune())
	}
	return res
}

// Key is a trivial representation of a keypress, a subset of tcell.Key. Key
// implements IKey. This exists as a convenience to widgets looking to
// customize keypress responses.
type Key struct {
	mod tcell.ModMask
	key tcell.Key
	ch  rune
}

func MakeKey(ch rune) Key {
	return Key{ch: ch, key: tcell.KeyRune}
}

func MakeKeyExt(key tcell.Key) Key {
	return Key{key: key}
}

func MakeKeyExt2(mod tcell.ModMask, key tcell.Key, ch rune) Key {
	return Key{
		mod: mod,
		key: key,
		ch:  ch,
	}
}

func (k Key) Rune() rune {
	return k.ch
}

func (k Key) Key() tcell.Key {
	return k.key
}

func (k Key) Modifiers() tcell.ModMask {
	return k.mod
}

// Stolen from tcell, but omit the Rune[...]
func (k Key) String() string {
	s := ""
	m := []string{}
	if k.mod&tcell.ModShift != 0 {
		m = append(m, "Shift")
	}
	if k.mod&tcell.ModAlt != 0 {
		m = append(m, "Alt")
	}
	if k.mod&tcell.ModMeta != 0 {
		m = append(m, "Meta")
	}
	if k.mod&tcell.ModCtrl != 0 {
		m = append(m, "Ctrl")
	}

	ok := false
	if s, ok = tcell.KeyNames[k.key]; !ok {
		if k.key == tcell.KeyRune {
			s = fmt.Sprintf("%c", k.ch)
		} else {
			s = fmt.Sprintf("Key[%d,%d]", k.key, int(k.ch))
		}
	}
	if len(m) != 0 {
		if k.mod&tcell.ModCtrl != 0 && strings.HasPrefix(s, "Ctrl-") {
			s = s[5:]
		}
		return fmt.Sprintf("%s+%s", strings.Join(m, "+"), s)
	}
	return s
}

var _ IKey = Key{}
var _ fmt.Stringer = Key{}

//======================================================================

// ComputeVerticalSubSizeUnsafe calls ComputeVerticalSubSize but returns only
// a single value - the IRenderSize. If there is an error the function will
// panic.
func ComputeVerticalSubSizeUnsafe(size IRenderSize, d IWidgetDimension, maxCol int, advRow int) IRenderSize {
	subSize, err := ComputeVerticalSubSize(size, d, maxCol, advRow)
	if err != nil {
		panic(err)
	}
	return subSize
}

// ComputeVerticalSubSize is used to determine the size with which a child
// widget should be rendered given the parent's render size, and an
// IWidgetDimension. The function will make adjustments to the size's
// number of rows i.e. in the vertical dimension, and as such is used by
// vpadding and pile. For example, if the parent render size is
// RenderBox{C: 20, R: 5} and the IWidgetDimension argument is
// RenderFlow{}, the function will return RenderFlowWith{C: 20}, i.e. it
// will transform a RenderBox to a RenderFlow of the same width. Another
// example is to transform a RenderBox to a shorter RenderBox if the
// IWidgetDimension specifies a RenderWithUnits{} - so it allows widgets
// like pile and vpadding to force widgets to be of a certain height, or
// to have their height be in a certain ratio to other widgets.
func ComputeVerticalSubSize(size IRenderSize, d IWidgetDimension, maxCol int, advRow int) (IRenderSize, error) {
	var subSize IRenderSize
	switch sz := size.(type) {
	case IRenderFixed:
		switch d2 := d.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: d2.BoxColumns(), R: d2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: d2.FlowColumns()}
		case IRenderFlow:
			if maxCol >= 0 {
				subSize = RenderFlowWith{C: maxCol}
			} else {
				return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
			}
		case IRenderWithUnits:
			subSize = RenderFixed{} // assumes the outer widget will respect the units - in general when we call this
			// we don't have a way to convert this to something else with a set number of rows. If we wanted to convert
			// to flow, we should use RenderFlowWith{}.
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
		}
	case IRenderBox:
		switch d2 := d.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: d2.FlowColumns()}
		case IRenderFlow:
			subSize = RenderFlowWith{C: sz.BoxColumns()}
		case IRenderWithUnits:
			subSize = RenderBox{C: sz.BoxColumns(), R: d2.Units()}
		case IRenderRelative:
			subSize = RenderBox{C: sz.BoxColumns(), R: int((d2.Relative() * float64(sz.BoxRows())) + 0.5)}
		case IRenderWithWeight:
			if advRow >= 0 {
				subSize = RenderBox{C: sz.BoxColumns(), R: advRow}
			} else {
				return nil, errors.WithStack(DimensionError{Size: size, Dim: d, Row: advRow})
			}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
		}
	case IRenderFlowWith:
		switch d2 := d.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderFlow:
			subSize = RenderFlowWith{C: sz.FlowColumns()}
		case IRenderWithUnits:
			subSize = RenderBox{C: sz.FlowColumns(), R: d2.Units()}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
		}
	default:
		return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
	}

	return subSize, nil
}

//======================================================================

// ComputeHorizontalSubSizeUnsafe calls ComputeHorizontalSubSize but
// returns only a single value - the IRenderSize. If there is an error the
// function will panic.
func ComputeHorizontalSubSizeUnsafe(size IRenderSize, d IWidgetDimension) IRenderSize {
	subSize, err := ComputeHorizontalSubSize(size, d)
	if err != nil {
		panic(err)
	}
	return subSize
}

// ComputeHorizontalSubSize is used to determine the size with which a
// child widget should be rendered given the parent's render size, and an
// IWidgetDimension. The function will make adjustments to the size's
// number of columns i.e. in the horizontal dimension, and as such is used
// by hpadding and columns. For example the function can transform a
// RenderBox to a narrower RenderBox if the IWidgetDimension specifies a
// RenderWithUnits{} - so it allows widgets like columns and hpadding to
// force widgets to be of a certain width, or to have their width be in a
// certain ratio to other widgets.
func ComputeHorizontalSubSize(size IRenderSize, d IWidgetDimension) (IRenderSize, error) {
	var subSize IRenderSize

	switch sz := size.(type) {
	case IRenderFixed:
		switch w2 := d.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: w2.BoxColumns(), R: w2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: w2.FlowColumns()}
		case IRenderWithUnits:
			subSize = RenderFlowWith{C: w2.Units()}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
		}
	case IRenderBox:
		switch w2 := d.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: w2.BoxColumns(), R: w2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: w2.FlowColumns()}
		case IRenderFlow:
			subSize = RenderFlowWith{C: sz.BoxColumns()}
		case IRenderRelative:
			subSize = RenderBox{C: int((w2.Relative() * float64(sz.BoxColumns())) + 0.5), R: sz.BoxRows()}
		case IRenderWithUnits:
			subSize = RenderBox{C: w2.Units(), R: sz.BoxRows()}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
		}
	case IRenderFlowWith:
		switch w2 := d.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: w2.BoxColumns(), R: w2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: w2.FlowColumns()}
		case IRenderFlow:
			subSize = RenderFlowWith{C: sz.FlowColumns()}
		case IRenderRelative:
			subSize = RenderFlowWith{C: int((w2.Relative() * float64(sz.FlowColumns())) + 0.5)}
		case IRenderWithUnits:
			subSize = RenderFlowWith{C: w2.Units()}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
		}
	default:
		return nil, errors.WithStack(DimensionError{Size: size, Dim: d})
	}

	return subSize, nil
}

func ComputeSubSizeUnsafe(size IRenderSize, w IWidgetDimension, h IWidgetDimension) IRenderSize {
	subSize, err := ComputeSubSize(size, w, h)
	if err != nil {
		panic(err)
	}
	return subSize
}

// TODO - doc
func ComputeSubSize(size IRenderSize, w IWidgetDimension, h IWidgetDimension) (IRenderSize, error) {
	var subSize IRenderSize

	switch sz := size.(type) {
	case IRenderFixed:
		switch w2 := w.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: w2.BoxColumns(), R: w2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: w2.FlowColumns()}
		case IRenderWithUnits:
			switch h2 := h.(type) {
			case IRenderWithUnits:
				subSize = RenderBox{C: w2.Units(), R: h2.Units()}
			default:
				subSize = RenderFixed{}
			}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: w})
		}
	case IRenderBox:
		switch w2 := w.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: w2.BoxColumns(), R: w2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: w2.FlowColumns()}
		case IRenderFlow:
			subSize = RenderFlowWith{C: sz.BoxColumns()}
		case IRenderRelative:
			cols := int((w2.Relative() * float64(sz.BoxColumns())) + 0.5)
			switch h2 := h.(type) {
			case IRenderRelative:
				rows := int((h2.Relative() * float64(sz.BoxRows())) + 0.5)
				subSize = RenderBox{C: cols, R: rows}
			case IRenderWithUnits:
				rows := h2.Units()
				subSize = RenderBox{C: cols, R: rows}
			case IRenderFlow:
				subSize = RenderFlowWith{C: cols}
			case IRenderFixed:
				subSize = RenderFlowWith{C: cols}
			default:
				return nil, errors.WithStack(DimensionError{Size: size, Dim: w})
			}
		case IRenderWithUnits:
			cols := w2.Units()
			switch h2 := h.(type) {
			case IRenderRelative:
				rows := int((h2.Relative() * float64(sz.BoxRows())) + 0.5)
				subSize = RenderBox{C: cols, R: rows}
			case IRenderWithUnits:
				rows := h2.Units()
				subSize = RenderBox{C: cols, R: rows}
			case IRenderFlow:
				subSize = RenderFlowWith{C: cols}
			case IRenderFixed:
				subSize = RenderFlowWith{C: cols}
			default:
				return nil, errors.WithStack(DimensionError{Size: size, Dim: w})
			}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: w})
		}
	case IRenderFlowWith:
		switch w2 := w.(type) {
		case IRenderFixed:
			subSize = RenderFixed{}
		case IRenderBox:
			subSize = RenderBox{C: w2.BoxColumns(), R: w2.BoxRows()}
		case IRenderFlowWith:
			subSize = RenderFlowWith{C: w2.FlowColumns()}
		case IRenderFlow:
			subSize = RenderFlowWith{C: sz.FlowColumns()}
		case IRenderRelative:
			subSize = RenderFlowWith{C: int((w2.Relative() * float64(sz.FlowColumns())) + 0.5)}
		case IRenderWithUnits:
			subSize = RenderFlowWith{C: w2.Units()}
		default:
			return nil, errors.WithStack(DimensionError{Size: size, Dim: w})
		}
	default:
		return nil, errors.WithStack(DimensionError{Size: size, Dim: w})
	}

	return subSize, nil
}

//======================================================================

// PrefPosition repeatedly unpacks composite widgets until it has to stop. It
// looks for a type exports a prefered position API. The widget might be
// ContainerWidget/StyledWidget/...
func PrefPosition(curw interface{}) gwutil.IntOption {
	var res gwutil.IntOption
	for {
		if ipos, ok := curw.(IPreferedPosition); ok {
			res = ipos.GetPreferedPosition()
			break
		}
		if curw2, ok2 := curw.(IComposite); ok2 {
			curw = curw2.SubWidget()
		} else {
			break
		}
	}
	return res
}

func SetPrefPosition(curw interface{}, prefPos int, app IApp) bool {
	var res bool
	for {
		if ipos, ok := curw.(IPreferedPosition); ok {
			ipos.SetPreferedPosition(prefPos, app)
			res = true
			break
		}
		if curw2, ok2 := curw.(IComposite); ok2 {
			curw = curw2.SubWidget()
		} else {
			break
		}
	}
	return res
}

//======================================================================

type WidgetPredicate func(w IWidget) bool

// FindInHierarchy starts at w, and applies the supplied predicate function; if it
// returns true, w is returned. If not, then the hierarchy is descended. If w has
// a child widget, then the predicate is applied to that child. If w has a set of
// children with a concept of one with focus, the predicate is applied to the child
// in focus. This repeats until a suitable widget is found, or the hierarchy terminates.
func FindInHierarchy(w IWidget, includeMe bool, pred WidgetPredicate) IWidget {
	var res IWidget
	for {
		if includeMe && pred(w) {
			res = w
			break
		}
		includeMe = true
		if cw, ok := w.(IComposite); ok {
			w = cw.SubWidget()
		} else if cw, ok := w.(ICompositeMultipleFocus); ok {
			f := cw.Focus()
			if f < 0 {
				break
			}
			w = cw.SubWidgets()[cw.Focus()]
		} else {
			break
		}
	}
	return res
}

type IFocusSelectable interface {
	IFocus
	IFindNextSelectable
}

type ICompositeMultipleFocus interface {
	IFocus
	ICompositeMultiple
}

type IChangeFocus interface {
	ChangeFocus(dir Direction, wrap bool, app IApp) bool
}

// ChangeFocus is a general algorithm for applying a change of focus to a type. If the type
// supports IChangeFocus, then that method is called directly. If the type supports IFocusSelectable,
// then the next widget is found, and set. Otherwise, if the widget has a child or children, the
// call is passed to them.
func ChangeFocus(w IWidget, dir Direction, wrap bool, app IApp) bool {
	w = FindInHierarchy(w, true, WidgetPredicate(func(w IWidget) bool {
		var res bool
		if _, ok := w.(IChangeFocus); ok {
			res = true
		} else if _, ok := w.(IFocusSelectable); ok {
			res = true
		}
		return res
	}))

	var res bool
	if w != nil {
		if sw, ok := w.(IChangeFocus); ok {
			res = sw.ChangeFocus(dir, wrap, app)
		} else if sw, ok := w.(IFocusSelectable); ok {
			next, ok := sw.FindNextSelectable(dir, wrap)
			if ok {
				sw.SetFocus(app, next)
				res = true
			}
		}
	}
	return res
}

type IGetFocus interface {
	Focus() int
}

func Focus(w IWidget) int {
	w = FindInHierarchy(w, true, WidgetPredicate(func(w IWidget) bool {
		var res bool
		if _, ok := w.(IGetFocus); ok {
			res = true
		}
		return res
	}))

	res := -1
	if w != nil {
		res = w.(IGetFocus).Focus()
	}

	return res
}

//======================================================================

// FocusPath returns a list of positions, each representing the focus
// position at that level in the widget hierarchy. The returned list may
// be shorter than the focus path through the hierarchy - only widgets
// that have more than one option for the focus will contribute.
func FocusPath(w IWidget) []interface{} {
	res := make([]interface{}, 0)
	includeMe := true
	for {
		w = FindInHierarchy(w, includeMe, WidgetPredicate(func(w IWidget) bool {
			_, ok := w.(IFocus)
			return ok
		}))
		if w == nil {
			break
		}
		includeMe = false
		wf, _ := w.(IFocus)
		res = append(res, wf.Focus())
	}

	return res
}

type FocusPathResult struct {
	Succeeded   bool
	FailedLevel int
}

func (f FocusPathResult) Error() string {
	return fmt.Sprintf("Focus at level %d could not be applied.", f.FailedLevel)
}

// SetFocusPath takes an array of focus positions, and applies them down the
// widget hierarchy starting at the supplied widget, w. If not all positions
// can be applied, the result's Succeeded field is set to false, and the
// FailedLevel field provides the index in the array of paths that could not
// be applied.
func SetFocusPath(w IWidget, path []interface{}, app IApp) FocusPathResult {
	res := FocusPathResult{
		Succeeded: true,
	}
	includeMe := true
	for i, v := range path {
		w = FindInHierarchy(w, includeMe, WidgetPredicate(func(w IWidget) bool {
			_, ok := w.(IFocus)
			return ok
		}))
		if w == nil {
			res.Succeeded = false
			res.FailedLevel = i
			break
		}
		includeMe = false
		wf, _ := w.(IFocus)
		wf.SetFocus(app, v.(int))
	}
	return res
}

//======================================================================

type ICopyModeWidget interface {
	IComposite
	IIdentity
	IClipboard
	CopyModeLevels() int
}

// CopyModeUserInput processes copy mode events in a typical fashion - a widget that wraps one
// with potentially copyable information could defer to this implementation of UserInput.
func CopyModeUserInput(w ICopyModeWidget, ev interface{}, size IRenderSize, focus Selector, app IApp) bool {
	res := false

	lvls := w.CopyModeLevels()

	if _, ok := ev.(CopyModeEvent); ok {
		if app.CopyModeClaimedAt() >= app.CopyLevel() && app.CopyModeClaimedAt() < app.CopyLevel()+lvls {
			app.CopyModeClaimedBy(w)
			res = true
		} else {
			cl := app.CopyLevel()
			app.CopyLevel(cl + lvls) // this is how many levels hexdumper will support
			res = UserInput(w.SubWidget(), ev, size, focus, app)
			app.CopyLevel(cl)

			if !res {
				app.CopyModeClaimedAt(app.CopyLevel() + lvls)
				app.CopyModeClaimedBy(w)
			}
		}
	} else if evc, ok := ev.(CopyModeClipsEvent); ok && (app.CopyModeClaimedAt() >= app.CopyLevel() && app.CopyModeClaimedAt() < app.CopyLevel()+lvls+1) {
		evc.Action.Collect(w.Clips(app))
		res = true
	} else {
		res = UserInput(w.SubWidget(), ev, size, focus, app)
	}
	return res
}

//======================================================================

// CopyWidgets is a trivial utility to return a copy of the array of widgets supplied.
// Note that this is not a deep copy! The array is different, but the IWidgets are not.
func CopyWidgets(w []IWidget) []IWidget {
	res := make([]IWidget, len(w))
	copy(res, w)
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 78
// End:
