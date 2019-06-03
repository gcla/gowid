// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package gowid

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/tcell"
	"github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
)

//======================================================================

// ICanvasLineReader can provide a particular line of Cells, at the specified y
// offset. The result may or may not be a copy of the actual Cells, and is determined
// by whether the user requested a copy and/or the capability of the ICanvasLineReader
// (maybe it has to provide a copy).
type ICanvasLineReader interface {
	Line(int, LineCopy) LineResult
}

// ICanvasMarkIterator will call the supplied function argument with the name and
// position of every mark set on the canvas. If the function returns true, the
// loop is terminated early.
type ICanvasMarkIterator interface {
	RangeOverMarks(f func(key string, value CanvasPos) bool)
}

// ICanvasCellReader can provide a Cell given a row and a column.
type ICanvasCellReader interface {
	CellAt(col, row int) Cell
}

type IAppendCanvas interface {
	IRenderBox
	ICanvasLineReader
	ICanvasMarkIterator
}

type IMergeCanvas interface {
	IRenderBox
	ICanvasCellReader
	ICanvasMarkIterator
}

type IDrawCanvas interface {
	IRenderBox
	ICanvasLineReader
	CursorEnabled() bool
	CursorCoords() CanvasPos
}

// ICanvas is the interface of any object which can generate a 2-dimensional
// array of Cells that are intended to be rendered on a terminal. This interface is
// pretty awful - cluttered and inconsistent and subject to cleanup... Note though
// that this interface is not here as the minimum requirement for providing arguments
// to a function or module - instead it's supposed to be an API surface for widgets
// so includes features that I am trying to guess will be needed, or that widgets
// already need.
type ICanvas interface {
	Duplicate() ICanvas
	MergeUnder(c IMergeCanvas, leftOffset, topOffset int, bottomGetsCursor bool)
	AppendBelow(c IAppendCanvas, doCursor bool, makeCopy bool)
	AppendRight(c IMergeCanvas, useCursor bool)
	SetCellAt(col, row int, c Cell)
	SetLineAt(row int, line []Cell)
	Truncate(above, below int)
	ExtendRight(cells []Cell)
	ExtendLeft(cells []Cell)
	TrimRight(cols int)
	TrimLeft(cols int)
	SetCursorCoords(col, row int)
	SetMark(name string, col, row int)
	GetMark(name string) (CanvasPos, bool)
	RemoveMark(name string)
	ICanvasMarkIterator
	ICanvasCellReader
	IDrawCanvas
	fmt.Stringer
}

// LineResult is returned by some Canvas Line-accessing APIs. If the Canvas
// can return a line without copying it, the Copied field will be false, and
// the caller is expected to make a copy if necessary (or risk modifying the
// original).
type LineResult struct {
	Line   []Cell
	Copied bool
}

// LineCopy is an argument provided to some Canvas APIs, like Line(). It tells
// the function how to allocate the backing array for a line if the line it
// returns must be a copy. Typically the API will return a type that indicates
// whether the result is a copy or not. Since the caller may receive a copy,
// it can help to indicate the allocation details like length and capacity in
// case the caller intends to extend the line returned for some other use.
type LineCopy struct {
	Len int
	Cap int
}

//======================================================================

// LineCanvas exists to make an array of Cells conform to some interfaces, specifically
// IRenderBox (it has a width of len(.) and a height of 1), IAppendCanvas, to allow
// an array of Cells to be passed to the canvas function AppendLine(), and ICanvasLineReader
// so that an array of Cells can act as a line returned from a canvas.
type LineCanvas []Cell

// BoxColumns lets LineCanvas conform to IRenderBox
func (c LineCanvas) BoxColumns() int {
	return len(c)
}

// BoxRows lets LineCanvas conform to IRenderBox
func (c LineCanvas) BoxRows() int {
	return 1
}

// BoxRows lets LineCanvas conform to IWidgetDimension
func (c LineCanvas) ImplementsWidgetDimension() {}

// Line lets LineCanvas conform to ICanvasLineReader
func (c LineCanvas) Line(y int, cp LineCopy) LineResult {
	return LineResult{
		Line:   c,
		Copied: false,
	}
}

// RangeOverMarks lets LineCanvas conform to ICanvasMarkIterator
func (c LineCanvas) RangeOverMarks(f func(key string, value CanvasPos) bool) {}

var _ IAppendCanvas = (*LineCanvas)(nil)
var _ ICanvasLineReader = (*LineCanvas)(nil)
var _ ICanvasMarkIterator = (*LineCanvas)(nil)

//======================================================================

var emptyLine [4096]Cell

type EmptyLineTooLong struct {
	Requested int
}

var _ error = EmptyLineTooLong{}

func (e EmptyLineTooLong) Error() string {
	return fmt.Sprintf("Tried to make an empty line too long - tried %d, max is %d", e.Requested, len(emptyLine))
}

// EmptyLine provides a ready-allocated source of empty cells. Of course this is to be
// treated as read-only.
func EmptyLine(length int) []Cell {
	if length < 0 {
		length = 0
	}
	if length > len(emptyLine) {
		panic(errors.WithStack(EmptyLineTooLong{Requested: length}))
	}
	return emptyLine[0:length]
}

// CanvasPos is a convenience struct to represent the coordinates of a position on a canvas.
type CanvasPos struct {
	X, Y int
}

func (c CanvasPos) PlusX(n int) CanvasPos {
	return CanvasPos{X: c.X + n, Y: c.Y}
}

func (c CanvasPos) PlusY(n int) CanvasPos {
	return CanvasPos{X: c.X, Y: c.Y + n}
}

//======================================================================

type CanvasSizeWrong struct {
	Requested IRenderSize
	Actual    IRenderBox
}

var _ error = CanvasSizeWrong{}

func (e CanvasSizeWrong) Error() string {
	return fmt.Sprintf("Canvas size %v, %v does not match render size %v", e.Actual.BoxColumns(), e.Actual.BoxRows(), e.Requested)
}

// PanicIfCanvasNotRightSize is for debugging - it panics if the size of the supplied canvas does
// not conform to the size specified by the size argument. For a box argument, columns and rows are
// checked; for a flow argument, columns are checked.
func PanicIfCanvasNotRightSize(c IRenderBox, size IRenderSize) {
	switch sz := size.(type) {
	case IRenderBox:
		if (c.BoxColumns() != sz.BoxColumns() && c.BoxRows() > 0) || c.BoxRows() != sz.BoxRows() {
			panic(errors.WithStack(CanvasSizeWrong{Requested: size, Actual: c}))
		}
	case IRenderFlowWith:
		if c.BoxColumns() != sz.FlowColumns() {
			panic(errors.WithStack(CanvasSizeWrong{Requested: size, Actual: c}))
		}
	}
}

type IRightSizeCanvas interface {
	IRenderBox
	ExtendRight(cells []Cell)
	TrimRight(cols int)
	Truncate(above, below int)
	AppendBelow(c IAppendCanvas, doCursor bool, makeCopy bool)
}

func MakeCanvasRightSize(c IRightSizeCanvas, size IRenderSize) {
	switch sz := size.(type) {
	case IRenderBox:
		rightSizeCanvas(c, sz.BoxColumns(), sz.BoxRows())
	case IRenderFlowWith:
		rightSizeCanvasHorizontally(c, sz.FlowColumns())
	}
}

func rightSizeCanvas(c IRightSizeCanvas, cols int, rows int) {
	rightSizeCanvasHorizontally(c, cols)
	rightSizeCanvasVertically(c, rows)
}

func rightSizeCanvasHorizontally(c IRightSizeCanvas, cols int) {
	if c.BoxColumns() > cols {
		c.TrimRight(cols)
	} else if c.BoxColumns() < cols {
		c.ExtendRight(EmptyLine(cols - c.BoxColumns()))
	}
}

func rightSizeCanvasVertically(c IRightSizeCanvas, rows int) {
	if c.BoxRows() > rows {
		c.Truncate(0, c.BoxRows()-rows)
	} else if c.BoxRows() < rows {
		AppendBlankLines(c, rows-c.BoxRows())
	}
}

//======================================================================

// Canvas is a simple implementation of ICanvas, and is returned by the Render() function
// of all the current widgets. It represents the canvas by a 2-dimensional array of Cells -
// no tricks or attempts to optimize this yet! The canvas also stores a map of string
// identifiers to positions - for example, the cursor position is tracked this way, and the
// menu widget keeps track of where it should render a "dropdown" using canvas marks. Most
// Canvas APIs expect that each line has the same length.
type Canvas struct {
	Lines  [][]Cell // inner array is a line
	Marks  map[string]CanvasPos
	maxCol int
}

// NewCanvas returns an initialized Canvas struct. Its size is 0 columns and
// 0 rows.
func NewCanvas() *Canvas {
	lines := make([][]Cell, 0, 120)
	res := &Canvas{
		Lines: lines[:0],
		Marks: make(map[string]CanvasPos),
	}
	var _ io.Writer = res
	return res
}

// NewCanvasWithLines allocates a canvas struct and sets its contents to the
// 2-d array provided as an argument.
func NewCanvasWithLines(lines [][]Cell) *Canvas {
	c := &Canvas{
		Lines: lines,
		Marks: make(map[string]CanvasPos),
	}
	c.AlignRight()
	c.maxCol = c.ComputeCurrentMaxColumn()
	var _ io.Writer = c
	return c
}

// NewCanvasOfSize returns a canvas struct of size cols x rows, where
// each Cell is default-initialized (i.e. empty).
func NewCanvasOfSize(cols, rows int) *Canvas {
	return NewCanvasOfSizeExt(cols, rows, Cell{})
}

// NewCanvasOfSize returns a canvas struct of size cols x rows, where
// each Cell is initialized by copying the fill argument.
func NewCanvasOfSizeExt(cols, rows int, fill Cell) *Canvas {
	fillArr := make([]Cell, cols)
	for i := 0; i < cols; i++ {
		fillArr[i] = fill
	}

	res := NewCanvas()
	if rows > 0 {
		res.Lines = append(res.Lines, fillArr)
		for i := 0; i < rows-1; i++ {
			res.Lines = append(res.Lines, make([]Cell, 0, 120))
		}
	}
	res.AlignRightWith(fill)
	res.maxCol = res.ComputeCurrentMaxColumn()

	var _ io.Writer = res

	return res
}

// Duplicate returns a deep copy of the receiver canvas.
func (c *Canvas) Duplicate() ICanvas {
	res := NewCanvasOfSize(c.BoxColumns(), c.BoxRows())
	for i := 0; i < c.BoxRows(); i++ {
		copy(res.Lines[i], c.Lines[i])
	}
	for k, v := range c.Marks {
		res.Marks[k] = v
	}
	return res
}

type IRangeOverCanvas interface {
	IRenderBox
	ICanvasCellReader
	SetCellAt(col, row int, c Cell)
}

// RangeOverCanvas applies the supplied function to each cell,
// modifying it in place.
func RangeOverCanvas(c IRangeOverCanvas, f ICellProcessor) {
	for i := 0; i < c.BoxRows(); i++ {
		for j := 0; j < c.BoxColumns(); j++ {
			c.SetCellAt(j, i, f.ProcessCell(c.CellAt(j, i)))
		}
	}
}

// Line provides access to the lines of the canvas. LineCopy
// determines what the Line() function should allocate if it
// needs to make a copy of the Line. Return true if line was
// copied.
func (c *Canvas) Line(y int, cp LineCopy) LineResult {
	return LineResult{
		Line:   c.Lines[y],
		Copied: false,
	}
}

// BoxColumns helps Canvas conform to IRenderBox.
func (c *Canvas) BoxColumns() int {
	return c.maxCol
}

// BoxRows helps Canvas conform to IRenderBox.
func (c *Canvas) BoxRows() int {
	return len(c.Lines)
}

// BoxRows helps Canvas conform to IWidgetDimension.
func (c *Canvas) ImplementsWidgetDimension() {}

// ComputeCurrentMaxColumn walks the 2-d array of Cells to determine
// the length of the longest line. This is used by certain APIs that
// manipulate the canvas.
func (c *Canvas) ComputeCurrentMaxColumn() int {
	res := 0
	for _, line := range c.Lines {
		res = gwutil.Max(res, len(line))
	}
	return res
}

// Write lets Canvas conform to io.Writer. Since each Canvas Cell holds a
// rune, the byte array argument is interpreted as the UTF-8 encoding of
// a sequence of runes.
func (c *Canvas) Write(p []byte) (n int, err error) {
	return WriteToCanvas(c, p)
}

// WriteToCanvas extracts the logic of implementing io.Writer into a free
// function that can be used by any canvas implementing ICanvas.
func WriteToCanvas(c IRangeOverCanvas, p []byte) (n int, err error) {
	done := 0
	maxcol := c.BoxColumns()
	line := 0
	col := 0
	for i, chr := range string(p) {
		if c.BoxRows() > line {
			switch chr {
			case '\n':
				for col < maxcol {
					c.SetCellAt(col, line, Cell{})
					col++
				}
				line++
				col = 0
			default:
				wid := runewidth.RuneWidth(chr)
				if col+wid > maxcol {
					col = 0
					line++
				}
				c.SetCellAt(col, line, c.CellAt(col, line).WithRune(chr))
				col += wid
			}
			done = i + utf8.RuneLen(chr)
		} else {
			break
		}
	}
	return done, nil
}

// CursorEnabled returns true if the cursor is enabled in this canvas, false otherwise.
func (c *Canvas) CursorEnabled() bool {
	_, ok := c.Marks["cursor"]
	return ok
}

// CursorCoords returns a pair of ints representing the current cursor coordinates. Note
// that the caller must be sure the Canvas's cursor is enabled.
func (c *Canvas) CursorCoords() CanvasPos {
	if pos, ok := c.Marks["cursor"]; !ok {
		// Caller must check first
		panic(errors.New("Cursor is off!"))
	} else {
		return pos
	}
}

// SetCursorCoords will set the Canvas's cursor coordinates. The special input of (-1,-1)
// will disable the cursor.
func (c *Canvas) SetCursorCoords(x, y int) {
	if x == -1 && y == -1 {
		delete(c.Marks, "cursor")
	} else {
		c.SetMark("cursor", x, y)
	}
}

// SetMark allows the caller to store a string identifier at a particular position in the
// Canvas. The menu widget uses this feature to keep track of where it should "open", acting
// as an overlay over the widgets below.
func (c *Canvas) SetMark(name string, x, y int) {
	c.Marks[name] = CanvasPos{X: x, Y: y}
}

// GetMark returns the position and presence/absence of the specified string identifier
// in the Canvas.
func (c *Canvas) GetMark(name string) (CanvasPos, bool) {
	i, ok := c.Marks[name]
	return i, ok
}

// RemoveMark removes a mark from the Canvas.
func (c *Canvas) RemoveMark(name string) {
	delete(c.Marks, name)
}

// RangeOverMarks applies the supplied function to each mark and position in the
// received Canvas. If the function returns false, the loop is terminated.
func (c *Canvas) RangeOverMarks(f func(key string, value CanvasPos) bool) {
	for k, v := range c.Marks {
		if !f(k, v) {
			break
		}
	}
}

// CellAt returns the Cell at the Canvas position provided. Note that the
// function assumes the caller has ensured the position is not out of
// bounds.
func (c *Canvas) CellAt(col, row int) Cell {
	return c.Lines[row][col]
}

// SetCellAt sets the Canvas Cell at the position provided. Note that the
// function assumes the caller has ensured the position is not out of
// bounds.
func (c *Canvas) SetCellAt(col, row int, cell Cell) {
	c.Lines[row][col] = cell
}

// SetLineAt sets a line of the Canvas at the given y position. The function
// assumes a line of the correct width has been provided.
func (c *Canvas) SetLineAt(row int, line []Cell) {
	c.Lines[row] = line
}

// AppendLine will append the array of Cells provided to the bottom of
// the receiver Canvas. If the makeCopy argument is true, a copy is made
// of the provided Cell array; otherwise, a slice is taken and used
// directly, meaning the Canvas will hold a reference to the underlying
// array.
func (c *Canvas) AppendLine(line []Cell, makeCopy bool) {
	newwidth := gwutil.Max(c.BoxColumns(), len(line))
	var newline []Cell
	if cap(line) < newwidth {
		makeCopy = true
	}
	if makeCopy {
		newline = make([]Cell, newwidth)
		copy(newline, line)
	} else if len(line) < newwidth {
		// extend slice
		newline = line[0:newwidth]
	} else {
		newline = line
	}
	c.Lines = append(c.Lines, newline)
	c.AlignRight()
}

// String lets Canvas conform to fmt.Stringer.
func (c *Canvas) String() string {
	return CanvasToString(c)
}

func CanvasToString(c ICanvas) string {
	lineStrings := make([]string, c.BoxRows())
	for i := 0; i < c.BoxRows(); i++ {
		line := c.Line(i, LineCopy{}).Line
		curLine := make([]rune, 0)
		for x := 0; x < len(line); {
			r := line[x].Rune()
			curLine = append(curLine, r)
			x += runewidth.RuneWidth(r)
		}
		lineStrings[i] = string(curLine)
	}
	return strings.Join(lineStrings, "\n")
}

// ExtendRight appends to each line of the receiver Canvas the array of
// Cells provided as an argument.
func (c *Canvas) ExtendRight(cells []Cell) {
	if len(cells) > 0 {
		for i := 0; i < len(c.Lines); i++ {
			if len(c.Lines[i])+len(cells) > cap(c.Lines[i]) {
				widerLine := make([]Cell, len(c.Lines[i]), len(c.Lines[i])+len(cells))
				copy(widerLine, c.Lines[i])
				c.Lines[i] = widerLine
			}
			c.Lines[i] = append(c.Lines[i], cells...)
		}
		c.maxCol += len(cells)
	}
}

// ExtendLeft prepends to each line of the receiver Canvas the array of
// Cells provided as an argument.
func (c *Canvas) ExtendLeft(cells []Cell) {
	if len(cells) > 0 {
		for i := 0; i < len(c.Lines); i++ {
			cellsCopy := make([]Cell, len(cells)+len(c.Lines[i]))
			copy(cellsCopy, cells)
			copy(cellsCopy[len(cells):], c.Lines[i])
			c.Lines[i] = cellsCopy
		}
		for k, pos := range c.Marks {
			c.Marks[k] = pos.PlusX(len(cells))
		}
		c.maxCol += len(cells)
	}
}

// AppendBelow appends the supplied Canvas to the "bottom" of the receiver Canvas. If
// doCursor is true and the supplied Canvas has an enabled cursor, it is applied to
// the received Canvas, with a suitable Y offset. If makeCopy is true then the supplied
// Canvas is copied; if false, and the supplied Canvas is capable of giving up
// ownership of its data structures, then they are moved to the receiver Canvas.
func (c *Canvas) AppendBelow(c2 IAppendCanvas, doCursor bool, makeCopy bool) {
	cw := c.BoxColumns()
	lenc := len(c.Lines)
	for i := 0; i < c2.BoxRows(); i++ {
		lr := c2.Line(i, LineCopy{
			Len: cw,
			Cap: cw,
		})
		if makeCopy && !lr.Copied {
			line := make([]Cell, cw)
			copy(line, lr.Line)
			c.Lines = append(c.Lines, line)
		} else {
			c.Lines = append(c.Lines, lr.Line)
		}
	}
	c.AlignRight()
	c2.RangeOverMarks(func(k string, pos CanvasPos) bool {
		if doCursor || (k != "cursor") {
			c.Marks[k] = pos.PlusY(lenc)
		}
		return true
	})
}

// Truncate removes "above" lines from above the receiver Canvas, and
// "below" lines from below.
func (c *Canvas) Truncate(above, below int) {
	if above < 0 {
		panic(errors.New("Lines to cut above must be >= 0"))
	}
	if below < 0 {
		panic(errors.New("Lines to cut below must be >= 0"))
	}
	cutAbove := gwutil.Min(len(c.Lines), above)
	c.Lines = c.Lines[cutAbove:]
	cutBelow := len(c.Lines) - gwutil.Min(len(c.Lines), below)
	c.Lines = c.Lines[:cutBelow]
	for k, pos := range c.Marks {
		c.Marks[k] = pos.PlusY(-cutAbove)
	}
}

type CellMergeFunc func(lower, upper Cell) Cell

// MergeWithFunc merges the supplied Canvas with the receiver canvas, where the receiver canvas
// is considered to start at column leftOffset and at row topOffset, therefore translated some
// distance from the top-left, and the receiver Canvas is the one modified. A function argument
// is supplied which specifies how Cells are merged, one by one e.g. which style takes effect,
// which rune, and so on.
func (c *Canvas) MergeWithFunc(c2 IMergeCanvas, leftOffset, topOffset int, fn CellMergeFunc, bottomGetsCursor bool) {
	c2w := c2.BoxColumns()
	for i := 0; i < c2.BoxRows(); i++ {
		if i+topOffset < len(c.Lines) {
			cl := len(c.Lines[i+topOffset])
			for j := 0; j < c2w; j++ {
				if j+leftOffset < cl {
					c2ij := c2.CellAt(j, i)
					c.Lines[i+topOffset][j+leftOffset] = fn(c.Lines[i+topOffset][j+leftOffset], c2ij)
				} else {
					break
				}
			}
		}
	}
	c2.RangeOverMarks(func(k string, v CanvasPos) bool {
		// Special treatment for the cursor mark - to allow widgets to display the cursor via
		// a "lower" widget. The terminal will typically support displaying one cursor only.
		if k != "cursor" || !bottomGetsCursor {
			c.Marks[k] = v.PlusX(leftOffset).PlusY(topOffset)
		}
		return true
	})
}

// MergeUnder merges the supplied Canvas "under" the receiver Canvas, meaning the
// receiver Canvas's Cells' settings are given priority.
func (c *Canvas) MergeUnder(c2 IMergeCanvas, leftOffset, topOffset int, bottomGetsCursor bool) {
	c.MergeWithFunc(c2, leftOffset, topOffset, Cell.MergeUnder, bottomGetsCursor)
}

// AppendRight appends the supplied Canvas to the right of the receiver Canvas. It
// assumes both Canvases have the same number of rows. If useCursor is true and the
// supplied Canvas has an enabled cursor, then it is applied with a suitable X
// offset applied.
func (c *Canvas) AppendRight(c2 IMergeCanvas, useCursor bool) {
	m := c.BoxColumns()
	c2w := c2.BoxColumns()
	for y := 0; y < c2.BoxRows(); y++ {
		if cap(c.Lines[y]) < len(c.Lines[y])+c2w {
			widerLine := make([]Cell, len(c.Lines[y])+c2w)
			copy(widerLine, c.Lines[y])
			c.Lines[y] = widerLine
		} else {
			c.Lines[y] = c.Lines[y][0 : len(c.Lines[y])+c2w]
		}
		for x := 0; x < c2w; x++ {
			c.Lines[y][x+m] = c2.CellAt(x, y)
		}
	}

	c2.RangeOverMarks(func(k string, v CanvasPos) bool {
		if (k != "cursor") || useCursor {
			c.Marks[k] = v.PlusX(m)
		}
		return true
	})
	c.maxCol = m + c2w
}

// TrimRight removes columns from the right of the receiver Canvas until there
// is the specified number left.
func (c *Canvas) TrimRight(colsToHave int) {
	for i := 0; i < len(c.Lines); i++ {
		if len(c.Lines[i]) > colsToHave {
			c.Lines[i] = c.Lines[i][0:colsToHave]
		}
	}
	c.maxCol = colsToHave
}

// TrimLeft removes columns from the left of the receiver Canvas until there
// is the specified number left.
func (c *Canvas) TrimLeft(colsToHave int) {
	colsToTrim := 0
	for i := 0; i < len(c.Lines); i++ {
		colsToTrim = gwutil.Max(colsToTrim, len(c.Lines[i])-colsToHave)
	}
	for i := 0; i < len(c.Lines); i++ {
		if len(c.Lines[i]) >= colsToTrim {
			c.Lines[i] = c.Lines[i][colsToTrim:]
		}
	}
	for k, v := range c.Marks {
		c.Marks[k] = v.PlusX(-colsToTrim)
	}
}

func appendCell(slice []Cell, data Cell, num int) []Cell {
	m := len(slice)
	n := m + num
	if n > cap(slice) { // if necessary, reallocate
		newSlice := make([]Cell, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	for i := 0; i < num; i++ {
		slice[m+i] = data
	}
	return slice
}

// AlignRightWith will extend each row of Cells in the receiver Canvas with
// the supplied Cell in order to ensure all rows are the same length. Note
// that the Canvas will not increase in width as a result.
func (c *Canvas) AlignRightWith(cell Cell) {
	m := c.ComputeCurrentMaxColumn()
	for j, line := range c.Lines {
		lineLen := len(line)
		cols := m - lineLen
		if len(c.Lines[j])+cols > cap(c.Lines[j]) {
			tmp := make([]Cell, len(c.Lines[j]), len(c.Lines[j])+cols+32)
			copy(tmp, c.Lines[j])
			c.Lines[j] = tmp[0:len(c.Lines[j])]
		}
		c.Lines[j] = appendCell(c.Lines[j], cell, m-lineLen)
	}
	c.maxCol = m
}

// AlignRight will extend each row of Cells in the receiver Canvas with an
// empty Cell in order to ensure all rows are the same length. Note that
// the Canvas will not increase in width as a result.
func (c *Canvas) AlignRight() {
	c.AlignRightWith(Cell{})
}

// Draw will render a Canvas to a tcell Screen.
func Draw(canvas IDrawCanvas, mode IColorMode, screen tcell.Screen) {
	cpos := CanvasPos{X: -1, Y: -1}
	if canvas.CursorEnabled() {
		cpos = canvas.CursorCoords()
	}

	screen.ShowCursor(-1, -1)

	for y := 0; y < canvas.BoxRows(); y++ {
		line := canvas.Line(y, LineCopy{})
		vline := line.Line
		for x := 0; x < len(vline); {
			c := vline[x]
			f, b, s := c.ForegroundColor(), c.BackgroundColor(), c.Style()
			st := MakeCellStyle(f, b, s)
			screen.SetContent(x, y, c.Rune(), nil, st)
			x += runewidth.RuneWidth(c.Rune())

			if x == cpos.X && y == cpos.Y {
				screen.ShowCursor(x, y)
			}
		}
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
