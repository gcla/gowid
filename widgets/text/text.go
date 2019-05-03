// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package text provides a text field widget.
package text

import (
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/mattn/go-runewidth"
)

//======================================================================

type ICursor interface {
	CursorEnabled() bool
	SetCursorDisabled()
	CursorPos() int
	SetCursorPos(pos int, app gowid.IApp)
}

type SimpleCursor struct {
	Pos int
}

func (c *SimpleCursor) CursorEnabled() bool {
	return c.Pos != -1
}

func (c *SimpleCursor) SetCursorDisabled() {
	c.Pos = -1
}

func (c *SimpleCursor) CursorPos() int {
	return c.Pos
}

func (c *SimpleCursor) SetCursorPos(pos int, app gowid.IApp) {
	c.Pos = pos
}

//======================================================================

// For callback registration
type ContentCB struct{}

//======================================================================

// IContent represents a styled range of text. Different sections of the text can
// have different styles. Behind the scenes, this is just implemented as an array of
// (rune, ICellStyler) pairs - maybe nothing more complicated would ever be needed
// in practise. See also TextContent.
type IContent interface {
	Length() int
	Width() int
	ChrAt(idx int) rune
	RangeOver(start, end int, attrs gowid.IRenderContext, proc gowid.ICellProcessor)
	AddAt(idx int, content ContentSegment)
	DeleteAt(idx, length int)
	fmt.Stringer
}

// ContentSegment represents some text each character of which is styled the same
// way.
type ContentSegment struct {
	Style gowid.ICellStyler
	Text  string
}

// StringContent makes a ContentSegment from a simple string.
func StringContent(s string) ContentSegment {
	return ContentSegment{nil, s}
}

// StyledContent makes a ContentSegment from a string and an ICellStyler.
func StyledContent(text string, style gowid.ICellStyler) ContentSegment {
	return ContentSegment{style, text}
}

// StyledRune is a styled rune.
type StyledRune struct {
	Chr  rune
	Attr gowid.ICellStyler
}

// Content is an array of AttributedRune and implements IContent.
type Content []StyledRune

// NewContent constructs Content suitable for initializing a text Widget.
func NewContent(content []ContentSegment) *Content {
	var length int
	for _, m := range content {
		length += len(m.Text) // might be underestimate
	}
	res := Content(make([]StyledRune, 0, length))
	for _, m := range content {
		res = append(res, MakeAttributedRunes(m)...)
	}
	return &res
}

// MakeAttributedRunes converts a ContentSegment into an array of AttributeRune,
// which is used to build a Content implementing IContent.
func MakeAttributedRunes(m ContentSegment) []StyledRune {
	res := make([]StyledRune, 0, len(m.Text)) // might be underestimate
	s := m.Text
	for len(s) > 0 {
		c, size := utf8.DecodeRuneInString(s)
		res = append(res, StyledRune{c, m.Style})
		s = s[size:]
	}
	return res
}

// AddAt will insert the supplied ContentSegment at index idx.
func (h *Content) AddAt(idx int, content ContentSegment) {
	piece := MakeAttributedRunes(content)
	res := Content(make([]StyledRune, 0, len(piece)+len(*h)))
	res = append(res, (*h)[0:idx]...)
	res = append(res, piece...)
	res = append(res, (*h)[idx:]...)
	*h = res
}

// DeleteAt will remove a segment of content of the provided length starting at index idx.
func (h *Content) DeleteAt(idx int, length int) {
	*h = append((*h)[0:idx], (*h)[idx+length:]...)
}

// RangeOver will call the supplied ICellProcessor for each element of the content between start and
// end, having first transformed that content element into an AttributedRune by using the
// accompanying ICellStyler and the IRenderContext. You can use this to build up an array
// of Cells, for example, in the process of converting a text widget to something that
// can be rendered in a canvas.
func (h Content) RangeOver(start, end int, attrs gowid.IRenderContext, proc gowid.ICellProcessor) {
	var curStyler gowid.ICellStyler
	var f gowid.IColor
	var g gowid.IColor
	var s gowid.StyleAttrs
	var f2 gowid.TCellColor
	var g2 gowid.TCellColor

	for idx, j := start, 0; idx < end; idx, j = idx+1, j+1 {
		if h[idx].Attr != nil {
			if h[idx].Attr != curStyler {
				f, g, s = h[idx].Attr.GetStyle(attrs)
				f2 = gowid.IColorToTCell(f, gowid.ColorNone, attrs.GetColorMode())
				g2 = gowid.IColorToTCell(g, gowid.ColorNone, attrs.GetColorMode())
				curStyler = h[idx].Attr
			}
			proc.ProcessCell(gowid.MakeCell(h[idx].Chr, f2, g2, s))
		} else {
			proc.ProcessCell(gowid.MakeCell(h[idx].Chr, gowid.ColorNone, gowid.ColorNone, gowid.StyleNone))
		}
	}
}

// ChrAt will return the unstyled rune at index idx.
func (h Content) ChrAt(idx int) rune {
	return h[idx].Chr
}

// Length will return the length of the content i.e. the number of runes it comprises.
func (h Content) Length() int {
	return len(h)
}

// Width returns the number of screen cells the content takes. Different from Length if >1-width runes are used.
func (h Content) Width() int {
	res := 0
	for _, r := range h {
		res += runewidth.RuneWidth(r.Chr)
	}
	return res
}

// String implements fmt.Stringer.
func (h Content) String() string {
	chars := make([]rune, h.Length())
	for i := 0; i < h.Length(); i++ {
		chars[i] = h.ChrAt(i)
	}
	return string(chars)
}

//======================================================================

// Determines how a text widget's text is wrapped - clip means anything beyond the
// specified column is clipped to the next newline

type WrapType int

const (
	WrapAny WrapType = iota
	WrapClip
)

// Widget can be used to display text on the screen, with optional styling for
// specified regions of the text.
type Widget struct {
	text         IContent
	wrap         WrapType
	align        gowid.IHAlignment
	opts         Options
	linesFromTop int
	Callbacks    *gowid.Callbacks
	gowid.RejectUserInput
	gowid.NotSelectable
}

var _ gowid.IWidget = (*Widget)(nil)
var _ io.Reader = (*Widget)(nil)
var _ fmt.Stringer = (*Widget)(nil)

type CopyableWidget struct {
	*Widget
	gowid.IIdentity
	gowid.IClipboardSelected
}

var _ gowid.IIdentityWidget = (*CopyableWidget)(nil)
var _ gowid.IClipboard = (*CopyableWidget)(nil)
var _ gowid.IClipboardSelected = (*CopyableWidget)(nil)

// ISimple is a gowid Widget that supports getting and setting plain unstyled text.
// It is used by edit.Widget, for example. This package's Widget type implements it.
type ISimple interface {
	gowid.IWidget
	Text() string
	SetText(text string, app gowid.IApp)
}

// IWidget is a gowid IWidget with the following extra APIs.
type IWidget interface {
	gowid.IWidget
	// Content returns an interface that provides access to the text and styling used.
	Content() IContent
	// Wrap determines whether the text is clipped, if too long, or flows onto the next line.
	Wrap() WrapType
	// Align can be used to keep each line of text left, right or center aligned.
	Align() gowid.IHAlignment
	// LinesFromTop is used to track how many widget lines are not in view off the top
	// given the current render. The widget tries to keep this the same when the widget
	// is re-rendered at a different size (e.g. the terminal is resized).
	LinesFromTop() int
	ClipIndicator() string
}

// Options is used to provide arguments to the various New initialization functions.
type Options struct {
	Wrap          WrapType
	ClipIndicator string
	Align         gowid.IHAlignment
}

// New initializes a text widget with a string and some extra arguments e.g. to align
// the text within each line, and to determine whether or not it's clipped.
func New(text string, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	content := &ContentSegment{nil, text}
	holder := NewContent([]ContentSegment{*content})
	return NewFromContentExt(holder, opt)
}

func NewCopyable(text string, id gowid.IIdentity, cs gowid.IClipboardSelected, opts ...Options) *CopyableWidget {
	return &CopyableWidget{
		Widget:             New(text, opts...),
		IIdentity:          id,
		IClipboardSelected: cs,
	}
}

// NewFromContent initializes a text widget with IContent, which can be built from a set
// of content segments. This is a way of making a text widget with styling.
func NewFromContent(content IContent) *Widget {
	res := &Widget{
		text:         content,
		linesFromTop: 0,
		Callbacks:    gowid.NewCallbacks(),
	}
	return res
}

// NewFromContentExt initialized a text widget with IContent and some extra options such
// as wrapping, alignment, etc.
func NewFromContentExt(content IContent, opts Options) *Widget {
	if opts.Align == nil {
		opts.Align = gowid.HAlignLeft{}
	}
	res := &Widget{
		text:      content,
		wrap:      opts.Wrap,
		align:     opts.Align,
		opts:      opts,
		Callbacks: gowid.NewCallbacks(),
	}
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("text")
}

// Writer is a wrapper around a text Widget which, by including the app, can be used
// to implement io.Writer.
type Writer struct {
	*Widget
	gowid.IApp
}

// Write implements io.Writer. The app is required because the content setting API
// requires the app because callbacks might be invoked which themselves require the
// app.
func (w *Writer) Write(p []byte) (n int, err error) {
	content := &ContentSegment{nil, string(p)}
	w.SetContent(w.IApp, NewContent([]ContentSegment{*content}))
	return len(p), nil
}

// Read makes Widget implement io.Reader.
func (w *Widget) Read(p []byte) (n int, err error) {
	runes := make([]rune, 0)

	for i := 0; i < w.Content().Length(); i++ {
		runes = append(runes, w.Content().ChrAt(i))
	}

	runesString := string(runes)

	num := copy(p, runesString)
	if num < len(p) {
		return num, io.EOF
	} else {
		return num, nil
	}
}

func (w *Widget) ClipIndicator() string {
	return w.opts.ClipIndicator
}

func (w *Widget) Content() IContent {
	return w.text
}

func (w *Widget) SetContent(app gowid.IApp, content IContent) {
	w.text = content
	gowid.RunWidgetCallbacks(w.Callbacks, ContentCB{}, app, w)
}

func (w *Widget) SetText(text string, app gowid.IApp) {
	content := &ContentSegment{nil, text}
	w.SetContent(app, NewContent([]ContentSegment{*content}))
}

func (w *Widget) Wrap() WrapType {
	return w.wrap
}

func (w *Widget) SetWrap(wrap WrapType, app gowid.IApp) {
	w.wrap = wrap
}

func (w *Widget) Align() gowid.IHAlignment {
	return w.align
}

func (w *Widget) SetAlign(align gowid.IHAlignment, app gowid.IApp) {
	w.align = align
}

func (w *Widget) LinesFromTop() int {
	return w.linesFromTop
}

func (w *Widget) SetLinesFromTop(l int, app gowid.IApp) {
	w.linesFromTop = l
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) OnContentSet(cb gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ContentCB{}, cb)
}

func (w *Widget) RemoveOnContentSet(cb gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ContentCB{}, cb)
}

func IsBreakableSpace(chr rune) bool {
	return unicode.IsSpace(chr) && chr != '\u00A0'
}

// CalculateTopMiddleBottom will, for a given size, calculate three indices:
// - the index of the line of text that should be at the top of the rendered area
// - the number of lines to display
// - the number of lines occluded from the bottom
func CalculateTopMiddleBottom(w IWidget, size gowid.IRenderSize) (int, int, int) {
	cursor := false
	crow := -1
	var cursorPos int

	w2, ok := w.(ICursor)
	if ok {
		cursor = w2.CursorEnabled()
		if cursor {
			cursorPos = w2.CursorPos()
		}
	}

	var maxCol int
	var maxRow int

	box, isBox := size.(gowid.IRenderBox)
	_, isFixed := size.(gowid.IRenderFixed)
	flow, isFlow := size.(gowid.IRenderFlowWith)
	haveMaxRow := isBox || isFixed
	content := w.Content()
	if haveMaxRow {
		if isFixed {
			maxRow = 1
			maxCol = w.Content().Width()
		} else {
			maxRow = box.BoxRows()
			maxCol = box.BoxColumns()
		}
	} else {
		if !isFlow {
			// TODO - compute content twice
			maxCol = content.Width()
		} else {
			maxCol = flow.FlowColumns()
		}
	}

	layout := MakeTextLayout(content, maxCol, w.Wrap(), w.Align())

	if cursor {
		_, crow = GetCoordsFromCursorPos(cursorPos, maxCol, layout, w.Content())
	}

	if haveMaxRow && len(layout.Lines) > maxRow {
		idxAbove := w.LinesFromTop()
		idxBelow := maxRow + idxAbove
		if cursor {
			if crow >= idxBelow {
				shift := (crow + 1) - idxBelow
				idxBelow += shift
				idxAbove += shift
			}
		}
		return idxAbove, maxRow, gwutil.Max(len(layout.Lines)-idxBelow, 0)
	} else {
		// If the client is a scrollbar, if there's no maxrow (i.e. RenderFlow), the scroll should just
		// take up all the space. Or, if the rendered lines fill up all the space available.
		return 0, 1, 0
	}
}

// ContentToCellArray is a helper type; it can be used to construct a Cell array by passing
// it to a RangeOver() function.
type ContentToCellArray struct {
	Cells []gowid.Cell
	Cur   int
}

var _ gowid.ICellProcessor = (*ContentToCellArray)(nil)

func (m *ContentToCellArray) ProcessCell(cell gowid.Cell) gowid.Cell {
	m.Cells[m.Cur] = cell
	m.Cur += runewidth.RuneWidth(cell.Rune())
	return cell
}

// If rendered Fixed, then rows==1 and cols==len(text)
func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	cursor := false
	ccol := -1
	crow := -1
	var cursorPos int

	w2, ok := w.(ICursor)
	if ok {
		cursor = w2.CursorEnabled()
		if cursor {
			cursorPos = w2.CursorPos()
		}
	}

	var maxCol int
	var maxRow int

	box, isBox := size.(gowid.IRenderBox)
	_, isFixed := size.(gowid.IRenderFixed)
	flow, isFlow := size.(gowid.IRenderFlowWith)
	content := w.Content()
	haveMaxRow := isBox || isFixed
	if haveMaxRow {
		if isFixed {
			curcol := 0
			maxRow = 1
			var last rune
			// This is lame - find a better way
			for i := 0; i < w.Content().Length(); i++ {
				last = w.Content().ChrAt(i)
				if last == '\n' {
					maxRow++
					if curcol > maxCol {
						maxCol = curcol
					}
					curcol = 0
				} else {
					curcol += runewidth.RuneWidth(last)
				}
			}
			if curcol > maxCol {
				maxCol = curcol
			}
			// if last == '\n' {
			// 	maxRow--
			// }
		} else {
			maxRow = box.BoxRows()
			maxCol = box.BoxColumns()
		}
	} else {
		if !isFlow {
			maxCol = content.Width()
		} else {
			maxCol = flow.FlowColumns()
		}
	}

	layout := MakeTextLayout(content, maxCol, w.Wrap(), w.Align())

	lines := make([][]gowid.Cell, len(layout.Lines))

	// Construct an array of lines from the layout, to be used as data
	// to construct the canvas. Walk through each segment returned by
	// the layout object
	count := 0
	for x, segment := range layout.Lines {
		// Make enough cells to be able to render double-width runes. The second cell will be left
		// empty.
		lines[x] = make([]gowid.Cell, segment.EndWidth-segment.StartWidth)
		w.Content().RangeOver(segment.StartLength, segment.EndLength, app, &ContentToCellArray{Cells: lines[x]})
		if segment.Clipped {
			//for i := len(w.ClipIndicator())-1; i >=0; i-- {
			ind := w.ClipIndicator()
			j := len(ind) - 1
			for i := len(lines[x]) - 1; i >= 0; i-- {
				if j < 0 {
					break
				}
				lines[x][i] = lines[x][i].WithRune(rune(ind[j]))
				j -= 1
			}
		}

		if len(lines[x]) < maxCol {
			switch w.Align().(type) {
			case gowid.HAlignRight:
				length := maxCol - len(lines[x])
				lines[x] = append(gowid.CellsFromString(gwutil.StringOfLength(' ', length)), lines[x]...)
			case gowid.HAlignMiddle:
				length := (maxCol - len(lines[x])) / 2
				lines[x] = append(gowid.CellsFromString(gwutil.StringOfLength(' ', length)), lines[x]...)
			default:
			}
		}
		count++
	}

	if cursor {
		ccol, crow = GetCoordsFromCursorPos(cursorPos, maxCol, layout, w.Content())
	}

	res := gowid.NewCanvasWithLines(lines)
	res.SetCursorCoords(ccol, crow)

	if haveMaxRow {
		if res.BoxRows() > maxRow {
			idxAbove := w.LinesFromTop()
			idxBelow := maxRow + idxAbove
			if cursor {
				if crow >= idxBelow {
					shift := (crow + 1) - idxBelow
					idxBelow += shift
					idxAbove += shift
				}
			}
			// If we would cut below the bottom of the render box, then shift the
			// cut up to the bottom of the render box
			if idxBelow > res.BoxRows() {
				idxAbove -= (idxBelow - res.BoxRows())
				idxBelow = res.BoxRows()
			}
			res.Truncate(idxAbove, gwutil.Max(res.BoxRows()-idxBelow, 0))
		} else {
			hor := gowid.CellFromRune(' ')
			horArr := make([]gowid.Cell, res.BoxColumns())
			for i := 0; i < res.BoxColumns(); i++ {
				horArr[i] = hor
			}

			nl := res.BoxRows()
			for i := 0; i < maxRow-nl; i++ {
				res.AppendLine(horArr, false)
			}
		}
	}
	res.AlignRight()
	res.ExtendRight(gowid.EmptyLine(maxCol - res.BoxColumns()))

	return res
}

type IChrAt interface {
	ChrAt(i int) rune
}

// zero-based
func GetCoordsFromCursorPos(cursorPos int, maxCol int, layout *TextLayout, at IChrAt) (x int, y int) {
	var crow, ccol int
	for lineNumber, segment := range layout.Lines {
		if segment.StartLength <= cursorPos && cursorPos <= segment.EndLength {
			crow = lineNumber

			ccol = 0
			for i := segment.StartLength; i < gwutil.Min(segment.EndLength, cursorPos); i++ {
				ccol += runewidth.RuneWidth(at.ChrAt(i))
			}
		}
	}
	return ccol, crow
}

// GetCursorPosFromCoords translates a (col, row) coord to a cursor position. It looks up the layout structure
// at the right line, then adds the col to the segment start offset.
func GetCursorPosFromCoords(ccol int, crow int, layout *TextLayout, at IChrAt) int {
	if len(layout.Lines) == 0 {
		return 0
	} else if crow >= len(layout.Lines) {
		return layout.Lines[len(layout.Lines)-1].EndWidth
	} else {

		start := layout.Lines[crow].StartLength

		startw := layout.Lines[crow].StartWidth
		endw := layout.Lines[crow].EndWidth

		col := 0
		for i := 0; i < gwutil.Min(endw-startw, ccol); {
			i += runewidth.RuneWidth(at.ChrAt(col + start))
			col += 1
		}
		return start + col
	}
}

//======================================================================

type LineLayout struct {
	StartLength int
	StartWidth  int
	EndWidth    int
	EndLength   int
	Clipped     bool
}

type TextLayout struct {
	Lines []LineLayout
}

// MakeTextLayout builds an array of line layouts from an IContent object. It applies the provided
// text wrapping and alignment options. The line layouts can then be used to index the IContent
// in order to build a canvas for rendering.
func MakeTextLayout(content IContent, width int, wrap WrapType, align gowid.IHAlignment) *TextLayout {
	lines := make([]LineLayout, 0)
	if width > 0 {
		switch wrap {
		case WrapClip:
			indexInLineWidth := 0        // current line index based on screen cells
			indexInLineLength := 0       // current line index based on runes
			skippingToEndOfLine := false // true if we had to cut off the text and are looking for a newline
			startOfCurrentLineLength := 0
			startOfCurrentLineWidth := 0
			for startOfCurrentLineLength+indexInLineLength < content.Length() {
				c := content.ChrAt(startOfCurrentLineLength + indexInLineLength)
				wid := runewidth.RuneWidth(c)
				if !skippingToEndOfLine && indexInLineWidth+wid > width { // end of space and no newline found
					lines = append(lines, LineLayout{
						StartLength: startOfCurrentLineLength,
						StartWidth:  startOfCurrentLineWidth,
						EndLength:   startOfCurrentLineLength + indexInLineLength,
						EndWidth:    startOfCurrentLineWidth + indexInLineWidth,
						Clipped:     true,
					})
					skippingToEndOfLine = true
					indexInLineWidth += wid
					indexInLineLength++
				} else if c == '\n' {
					if !skippingToEndOfLine {
						lines = append(lines, LineLayout{
							StartLength: startOfCurrentLineLength,
							StartWidth:  startOfCurrentLineLength,
							EndLength:   startOfCurrentLineLength + indexInLineLength,
							EndWidth:    startOfCurrentLineWidth + indexInLineWidth,
							Clipped:     false,
						})
					}
					skippingToEndOfLine = false
					startOfCurrentLineLength += (indexInLineLength + 1)
					startOfCurrentLineWidth += (indexInLineWidth + 1)
					indexInLineLength = 0
					indexInLineWidth = 0
				} else {
					indexInLineWidth += wid
					indexInLineLength += 1
				}
			}
			if !skippingToEndOfLine {
				lines = append(lines, LineLayout{
					StartLength: startOfCurrentLineLength,
					StartWidth:  startOfCurrentLineWidth,
					EndLength:   startOfCurrentLineLength + indexInLineLength,
					EndWidth:    startOfCurrentLineWidth + indexInLineWidth,
					Clipped:     false,
				})
			}

		case WrapAny:
			indexInSegmentLength := 0 // current line index
			indexInSegmentWidth := 0  // current line index
			startOfCurrentSegmentLength := 0
			startOfCurrentSegmentWidth := 0
			for startOfCurrentSegmentLength+indexInSegmentLength < content.Length() {
				c := content.ChrAt(startOfCurrentSegmentLength + indexInSegmentLength)
				if indexInSegmentWidth+runewidth.RuneWidth(c) > width { // end of space and no newline found
					lines = append(lines, LineLayout{
						StartLength: startOfCurrentSegmentLength,
						StartWidth:  startOfCurrentSegmentWidth,
						EndLength:   startOfCurrentSegmentLength + indexInSegmentLength,
						EndWidth:    startOfCurrentSegmentWidth + indexInSegmentWidth,
						Clipped:     false,
					})
					startOfCurrentSegmentLength += indexInSegmentLength
					startOfCurrentSegmentWidth += indexInSegmentWidth
					indexInSegmentLength = 0
					indexInSegmentWidth = 0
				} else if c == '\n' {
					lines = append(lines, LineLayout{
						StartLength: startOfCurrentSegmentLength,
						StartWidth:  startOfCurrentSegmentWidth,
						EndLength:   startOfCurrentSegmentLength + indexInSegmentLength,
						EndWidth:    startOfCurrentSegmentWidth + indexInSegmentWidth,
						Clipped:     false,
					})
					startOfCurrentSegmentLength += (indexInSegmentLength + 1)
					startOfCurrentSegmentWidth += (indexInSegmentWidth + 1)
					indexInSegmentLength = 0
					indexInSegmentWidth = 0
				} else {
					indexInSegmentWidth += runewidth.RuneWidth(c)
					indexInSegmentLength += 1
				}
			}
			lines = append(lines, LineLayout{
				StartLength: startOfCurrentSegmentLength,
				StartWidth:  startOfCurrentSegmentWidth,
				EndLength:   startOfCurrentSegmentLength + indexInSegmentLength,
				EndWidth:    startOfCurrentSegmentWidth + indexInSegmentWidth,
				Clipped:     false,
			})

		default:
			panic(fmt.Errorf("Wrap %v not supported yet", wrap))
		}
	}
	return &TextLayout{lines}
}

//======================================================================

// This meets both IText and ICursor, and allows me to make a canvas from a text widget
// and a separately specified cursor position
type WidgetWithCursor struct {
	*Widget
	*SimpleCursor
}

func (w *WidgetWithCursor) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *WidgetWithCursor) CalculateTopMiddleBottom(size gowid.IRenderSize) (int, int, int) {
	return CalculateTopMiddleBottom(w, size)
}

//======================================================================

func (w *CopyableWidget) Clips(app gowid.IApp) []gowid.ICopyResult {
	return []gowid.ICopyResult{
		gowid.CopyResult{
			Name: "Displayed text",
			Val:  w.Content().String(),
		},
	}
}

func (w *CopyableWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	claimed := false
	if _, ok := ev.(gowid.CopyModeEvent); ok {
		// zero means deepest should try to claim - leaf knows it's a leaf.
		if app.InCopyMode() && app.CopyLevel() <= app.CopyModeClaimedAt() {
			app.CopyModeClaimedAt(app.CopyLevel())
			app.CopyModeClaimedBy(w)
			claimed = true
		}
	} else if evc, ok := ev.(gowid.CopyModeClipsEvent); ok {
		evc.Action.Collect(w.Clips(app))
		claimed = true
	}
	return claimed
}

func (w *CopyableWidget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	if app.InCopyMode() && app.CopyModeClaimedBy().ID() == w.ID() && focus.Focus {
		w2 := w.AlterWidget(w.Widget, app)
		return w2.Render(size, focus, app)
	} else {
		return gowid.Render(w.Widget, size, focus, app)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
