// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Based heavily on vterm.py from urwid

package terminal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
)

//======================================================================

const (
	CharsetDefault = iota
	CharsetUTF8    = iota
)

const (
	EscByte byte = 27
)

type LEDSState int

const (
	LEDSClear      LEDSState = 0
	LEDSScrollLock LEDSState = 1
	LEDSNumLock    LEDSState = 2
	LEDSCapsLock   LEDSState = 3
)

const (
	DecSpecialChars    = "▮◆▒␉␌␍␊°±␤␋┘┐┌└┼⎺⎻─⎼⎽├┤┴┬│≤≥π≠£·"
	AltDecSpecialChars = "_`abcdefghijklmnopqrstuvwxyz{|}~"
)

type ScrollDir bool

const (
	ScrollDown ScrollDir = false
	ScrollUp   ScrollDir = true
)

type IMouseSupport interface {
	MouseEnabled() bool
	MouseIsSgr() bool
	MouseReportButton() bool
	MouseReportAny() bool
}

//======================================================================

// Modes is used to track the state of this terminal - which modes
// are enabled, etc. It tracks the mouse state in particular so implements
// IMouseSupport.
type Modes struct {
	DisplayCtrl        bool
	Insert             bool
	LfNl               bool
	KeysAutoWrap       bool
	ReverseVideo       bool
	ConstrainScrolling bool
	DontAutoWrap       bool
	InvisibleCursor    bool
	Charset            int
	VT200Mouse         bool // #define SET_VT200_MOUSE             1000
	ReportButton       bool // #define SET_BTN_EVENT_MOUSE         1002
	ReportAny          bool // #define SET_ANY_EVENT_MOUSE         1003
	SgrModeMouse       bool // #define SET_SGR_EXT_MODE_MOUSE      1006
}

func (t Modes) MouseEnabled() bool {
	return t.VT200Mouse
}

func (t Modes) MouseIsSgr() bool {
	return t.SgrModeMouse
}

func (t Modes) MouseReportButton() bool {
	return t.ReportButton
}

func (t Modes) MouseReportAny() bool {
	return t.ReportAny
}

//======================================================================

type CSIFunction func(canvas *Canvas, args []int, qmark bool)

type ICSICommand interface {
	MinArgs() int
	FallbackArg() int
	IsAlias() bool
	Alias() byte
	Call(canvas *Canvas, args []int, qmark bool)
}

type RegularCSICommand struct {
	minArgs     int
	fallbackArg int
	fn          CSIFunction
}

func (c RegularCSICommand) MinArgs() int     { return c.minArgs }
func (c RegularCSICommand) FallbackArg() int { return c.fallbackArg }
func (c RegularCSICommand) IsAlias() bool    { return false }
func (c RegularCSICommand) Alias() byte      { panic(errors.New("Do not call")) }
func (c RegularCSICommand) Call(canvas *Canvas, args []int, qmark bool) {
	c.fn(canvas, args, qmark)
}

type AliasCSICommand struct {
	alias byte
}

func (c AliasCSICommand) MinArgs() int     { panic(errors.New("Do not call")) }
func (c AliasCSICommand) FallbackArg() int { panic(errors.New("Do not call")) }
func (c AliasCSICommand) IsAlias() bool    { return true }
func (c AliasCSICommand) Alias() byte      { return c.alias }
func (c AliasCSICommand) Call(canvas *Canvas, args []int, qmark bool) {
	panic(errors.New("Do not call"))
}

type CSIMap map[byte]ICSICommand

// csiMap maps bytes to CSI mode changing functions. This closely follows urwid's structure.
var csiMap = CSIMap{
	'@': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.InsertChars(gwutil.NoneInt(), gwutil.NoneInt(), args[0], gwutil.NoneRune())
	}},
	'A': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(0, -int(args[0]), true, false, false)
	}},
	'B': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(0, int(args[0]), true, false, false)
	}},
	'C': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(int(args[0]), 0, true, false, false)
	}},
	'D': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(-int(args[0]), 0, true, false, false)
	}},
	'E': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(0, int(args[0]), false, false, true)
	}},
	'F': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(0, -int(args[0]), false, false, true)
	}},
	'G': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(int(args[0])-1, 0, false, false, true)
	}},
	'H': RegularCSICommand{2, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(int(args[1])-1, int(args[0])-1, false, false, false)
	}},
	'J': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSIEraseDisplay(args[0])
	}},
	'K': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSIEraseLine(args[0])
	}},
	'L': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.InsertLines(true, args[0])
	}},
	'M': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.RemoveLines(true, args[0])
	}},
	'P': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.RemoveChars(gwutil.NoneInt(), gwutil.NoneInt(), args[0])
	}},
	'X': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		myx, myy := canvas.TermCursor()
		canvas.Erase(myx, myy, myx+args[0]-1, myy)
	}},
	'a': AliasCSICommand{alias: 'C'},
	'c': RegularCSICommand{0, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSIGetDeviceAttributes(qmark)
	}},
	'd': RegularCSICommand{1, 1, func(canvas *Canvas, args []int, qmark bool) {
		canvas.MoveCursor(0, int(args[0])-1, false, true, false)
	}},
	'e': AliasCSICommand{alias: 'B'},
	'f': AliasCSICommand{alias: 'H'},
	'g': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSIClearTabstop(args[0])
	}},
	'h': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSISetModes(args, qmark, false)
	}},
	'l': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSISetModes(args, qmark, true)
	}},
	'm': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSISetAttr(args)
	}},
	'n': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSIStatusReport(args[0])
	}},
	'q': RegularCSICommand{1, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSISetKeyboardLEDs(args[0])
	}},
	'r': RegularCSICommand{2, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.CSISetScroll(args[0], args[1])
	}},
	's': RegularCSICommand{0, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.SaveCursor(false)
	}},
	'u': RegularCSICommand{0, 0, func(canvas *Canvas, args []int, qmark bool) {
		canvas.RestoreCursor(false)
	}},
	'`': AliasCSICommand{alias: 'G'},
}

//======================================================================

var charsetMapping = map[string]rune{
	"default": 0,
	"vt100":   '0',
	"ibmpc":   'U',
	"user":    0,
}

type Charset struct {
	SgrMapping bool
	Active     int
	Current    rune
	Mapping    []string
}

func NewTerminalCharset() *Charset {
	res := &Charset{}
	res.Mapping = []string{"default", "vt100"}
	res.Activate(0)
	return res
}

func (t *Charset) Activate(g int) {
	t.Active = g
	if val, ok := charsetMapping[t.Mapping[g]]; ok {
		t.Current = val
	} else {
		t.Current = 0
	}
}

func (t *Charset) Define(g int, charset string) {
	t.Mapping[g] = charset
	t.Activate(t.Active)
}

func (t *Charset) SetSgrIbmpc() {
	t.SgrMapping = true
}

func (t *Charset) ResetSgrIbmpc() {
	t.SgrMapping = false
	t.Activate(t.Active)
}

func (t *Charset) ApplyMapping(r rune) rune {
	if t.SgrMapping || t.Mapping[t.Active] == "ibmpc" {
		decPos := strings.IndexRune(DecSpecialChars, charmap.CodePage437.DecodeByte(byte(r)))
		if decPos >= 0 {
			t.Current = '0'
			return rune(AltDecSpecialChars[decPos])
		} else {
			t.Current = 'U'
			return r
		}
	} else {
		return r
	}
}

//======================================================================

// ViewPortCanvas implements ICanvas by embedding a Canvas pointer, but
// reimplementing Line and Cell access APIs relative to an Offset and
// a Height. The Height specifies the number of visible rows in the
// ViewPortCanvas; the rows that are not visible are logically "above"
// the visible rows. If Offset is reduced, the view of the underlying
// large Canvas is shifted up. This type is used by the terminal widget
// to hold the terminal's scrollback buffer.
type ViewPortCanvas struct {
	*gowid.Canvas
	Offset int
	Height int
}

func NewViewPort(c *gowid.Canvas, offset, height int) *ViewPortCanvas {
	res := &ViewPortCanvas{
		Canvas: c,
		Offset: offset,
		Height: height,
	}
	return res
}

func (v *ViewPortCanvas) BoxRows() int {
	return v.Height
}

func (v *ViewPortCanvas) Line(y int, cp gowid.LineCopy) gowid.LineResult {
	return v.Canvas.Line(y+v.Offset, cp)
}

func (v *ViewPortCanvas) SetLineAt(row int, line []gowid.Cell) {
	v.Canvas.SetLineAt(row+v.Offset, line)
}

func (v *ViewPortCanvas) CellAt(col, row int) gowid.Cell {
	return v.Canvas.CellAt(col, row+v.Offset)
}

func (v *ViewPortCanvas) SetCellAt(col, row int, c gowid.Cell) {
	v.Canvas.SetCellAt(col, row+v.Offset, c)
}

func (c *ViewPortCanvas) String() string {
	return gowid.CanvasToString(c)
}

//======================================================================

// Canvas implements gowid.ICanvas and stores the state of the terminal drawing area
// associated with a terminal (and TerminalWidget).
type Canvas struct {
	*ViewPortCanvas
	alternate                          *ViewPortCanvas
	alternateActive                    bool
	parsestate                         int
	scrollback                         int
	withinEscape                       bool
	savedx, savedy                     gwutil.IntOption
	savedstyles                        map[string]bool
	savedfg, savedbg                   gwutil.IntOption
	scrollRegionStart, scrollRegionEnd int
	terminal                           ITerminal
	charset                            *Charset
	tcx, tcy                           int
	styles                             map[string]bool
	tabstops                           []int
	isRottenCursor                     bool
	escbuf                             []byte
	fg, bg                             gwutil.IntOption
	utf8Buffer                         []byte
	gowid.ICallbacks
}

func NewCanvasOfSize(cols, rows int, scrollback int, widget ITerminal) *Canvas {
	res := &Canvas{
		ViewPortCanvas: NewViewPort(gowid.NewCanvasOfSize(cols, rows), 0, rows),
		alternate:      NewViewPort(gowid.NewCanvasOfSize(cols, rows), 0, rows),
		scrollback:     scrollback,
		terminal:       widget,
		utf8Buffer:     make([]byte, 0, 4),
		ICallbacks:     gowid.NewCallbacks(),
	}
	res.Reset()

	var _ io.Writer = res

	return res
}

// Write is an io.Writer for a terminal canvas, which processes the input as
// terminal codes, and writes with respect to the current cursor position.
func (c *Canvas) Write(p []byte) (n int, err error) {
	for _, b := range p {
		c.ProcessByte(b)
	}

	return len(p), nil
}

func (c *Canvas) Reset() {
	c.alternateActive = false
	c.escbuf = make([]byte, 0)
	c.charset = NewTerminalCharset()
	c.parsestate = 0
	c.withinEscape = false
	c.savedx = gwutil.NoneInt()
	c.savedy = gwutil.NoneInt()
	c.savedfg = gwutil.NoneInt()
	c.savedbg = gwutil.NoneInt()
	c.savedstyles = make(map[string]bool)
	c.fg = gwutil.NoneInt()
	c.bg = gwutil.NoneInt()
	c.styles = make(map[string]bool)
	*c.terminal.Modes() = Modes{}
	c.ResetScroll()
	c.InitTabstops(false)
	c.Clear(gwutil.SomeInt(0), gwutil.SomeInt(0))
}

func (c *Canvas) IsScrollRegionSet() bool {
	return !((c.scrollRegionStart == 0) && (c.scrollRegionEnd == c.BoxRows()-1))
}

func (c *Canvas) ResetScroll() {
	c.scrollRegionStart = 0
	c.scrollRegionEnd = c.BoxRows() - 1
}

func (c *Canvas) CarriageReturn() {
	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(c.tcy))
}

func (c *Canvas) Tab(tabstop int) {
	x, y := c.TermCursor()

	for x < c.BoxColumns()-1 {
		x += 1
		if c.IsTabstop(x) {
			break
		}
	}

	c.isRottenCursor = false
	c.SetTermCursor(gwutil.SomeInt(x), gwutil.SomeInt(y))
}

func (c *Canvas) InitTabstops(extend bool) {
	tablen, mod := c.BoxColumns()/8, c.BoxColumns()

	if mod > 0 {
		tablen += 1
	}

	if extend {
		for len(c.tabstops) < tablen {
			c.tabstops = append(c.tabstops, 1)
		}
	} else {
		c.tabstops = []int{}
		for i := 0; i < tablen; i++ {
			c.tabstops = append(c.tabstops, 1)
		}
	}
}

func (c *Canvas) SetTabstop(x2 gwutil.IntOption, remove bool, clear bool) {
	if clear {
		for tab := 0; tab < len(c.tabstops); tab++ {
			c.tabstops[tab] = 0
		}
	} else {

		var x int
		if x2.IsNone() {
			x, _ = c.TermCursor()
		} else {
			x = x2.Val()
		}

		div, mod := x/8, x%8
		if remove {
			c.tabstops[div] &= ^(1 << uint(mod))
		} else {
			c.tabstops[div] |= 1 << uint(mod)
		}
	}
}

func (c *Canvas) IsTabstop(x int) bool {
	div, mod := x/8, x%8

	return (c.tabstops[div] & (1 << uint(mod))) > 0
}

func (c *Canvas) TermCursor() (x, y int) {
	x, y = c.tcx, c.tcy
	return
}

func (c *Canvas) SetTermCursor(x2, y2 gwutil.IntOption) {

	tx, ty := c.TermCursor()
	var x, y int

	if x2.IsNone() {
		x = tx
	} else {
		x = x2.Val()
	}

	if y2.IsNone() {
		y = ty
	} else {
		y = y2.Val()
	}

	c.tcx, c.tcy = c.ConstrainCoords(x, y, false)

	if !c.terminal.Modes().InvisibleCursor {
		c.SetCursorCoords(c.tcx, c.tcy)
	} else {
		c.SetCursorCoords(-1, -1)
	}
}

func (c *Canvas) ConstrainCoords(x, y int, ignoreScrolling bool) (int, int) {
	if x >= c.BoxColumns() {
		x = c.BoxColumns() - 1
	} else if x < 0 {
		x = 0
	}

	if c.terminal.Modes().ConstrainScrolling && !ignoreScrolling {
		if y > c.scrollRegionEnd {
			y = c.scrollRegionEnd
		} else if y < c.scrollRegionStart {
			y = c.scrollRegionStart
		}
	} else {
		if y >= c.BoxRows() {
			y = c.BoxRows() - 1
		} else if y < 0 {
			y = 0
		}
	}

	return x, y
}

// ScrollBuffer will return the number of lines actually scrolled.
func (c *Canvas) ScrollBuffer(dir ScrollDir, reset bool, linesOpt gwutil.IntOption) int {
	prev := c.Offset
	if reset {
		c.Offset = c.Canvas.BoxRows() - c.BoxRows()
	} else {
		var lines int
		if linesOpt.IsNone() {
			lines = c.BoxRows() / 2
		} else {
			lines = linesOpt.Val()
		}
		if dir == ScrollDown {
			lines = -lines
		}
		maxScroll := c.Canvas.BoxRows() - c.BoxRows()
		c.Offset -= lines
		if c.Offset < 0 {
			c.Offset = 0
		} else if c.Offset > maxScroll {
			c.Offset = maxScroll
		}
	}
	c.SetTermCursor(gwutil.NoneInt(), gwutil.NoneInt())

	return c.Offset - prev
}

func (c *Canvas) Scroll(dir ScrollDir) {
	// reverse means scrolling up towards the top
	if dir == ScrollDown {

		// e.g. pgdown
		if c.IsScrollRegionSet() {
			start := c.scrollRegionStart + c.Offset
			end := c.scrollRegionEnd + c.Offset

			dummy := make([][]gowid.Cell, len(c.ViewPortCanvas.Canvas.Lines))
			n := 0
			n += copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[:start])
			n += copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[start+1:end+1])
			n += copy(dummy[n:], sliceWithOneEmptyLine(c.BoxColumns()))
			copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[end+1:])
			c.ViewPortCanvas.Canvas.Lines = dummy
		} else {

			chopline := false
			if c.Canvas.BoxRows() == c.BoxRows()+c.scrollback {
				chopline = true
			}

			var dummy [][]gowid.Cell
			n := 0
			if !chopline {
				dummy = make([][]gowid.Cell, c.Canvas.BoxRows()+1)
				n += copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines)
				c.Offset += 1
			} else {
				dummy = make([][]gowid.Cell, c.Canvas.BoxRows())
				n += copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[1:])
			}
			copy(dummy[n:], sliceWithOneEmptyLine(c.BoxColumns()))
			c.ViewPortCanvas.Canvas.Lines = dummy
		}

	} else {

		// e.g. pgup, cursor up
		if c.IsScrollRegionSet() {
			start := c.scrollRegionStart + c.Offset
			end := c.scrollRegionEnd + c.Offset

			dummy := make([][]gowid.Cell, len(c.ViewPortCanvas.Canvas.Lines))
			n := 0
			n += copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[:start])
			n += copy(dummy[n:], sliceWithOneEmptyLine(c.BoxColumns()))
			n += copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[start:end])
			copy(dummy[n:], c.ViewPortCanvas.Canvas.Lines[end+1:])
			c.ViewPortCanvas.Canvas.Lines = dummy
		} else {
			c.InsertLines(false, 1)
		}
	}
}

func sliceWithOneEmptyLine(n int) [][]gowid.Cell {
	return [][]gowid.Cell{emptyLine(n)}
}

func emptyLine(n int) []gowid.Cell {
	fillArr := make([]gowid.Cell, n)
	return fillArr
}

func (c *Canvas) LineFeed(reverse bool) {
	x, y := c.TermCursor()

	if reverse {
		if y <= 0 && 0 < c.scrollRegionStart {
		} else if y == c.scrollRegionStart {
			c.Scroll(ScrollUp)
		} else {
			y -= 1
		}
	} else {
		if y >= c.BoxRows()-1 && y > c.scrollRegionEnd {
		} else if y == c.scrollRegionEnd {
			c.Scroll(ScrollDown)
		} else {
			y += 1
		}
	}

	c.SetTermCursor(gwutil.SomeInt(x), gwutil.SomeInt(y))
}

func (c *Canvas) SaveCursor(withAttrs bool) {
	myx, myy := c.TermCursor()
	c.savedx = gwutil.SomeInt(myx)
	c.savedy = gwutil.SomeInt(myy)
	c.savedstyles = make(map[string]bool)
	if withAttrs {
		c.savedfg = c.fg
		c.savedbg = c.bg
		for k, v := range c.styles {
			c.savedstyles[k] = v
		}
	} else {
		c.savedfg = gwutil.NoneInt()
		c.savedbg = gwutil.NoneInt()
	}
}

func (c *Canvas) RestoreCursor(withAttrs bool) {
	if !(c.savedx == gwutil.NoneInt() || c.savedy == gwutil.NoneInt()) {
		c.SetTermCursor(c.savedx, c.savedy)
		if withAttrs {
			c.fg = c.savedfg
			c.bg = c.savedbg
			c.styles = make(map[string]bool)
			for k, v := range c.savedstyles {
				c.styles[k] = v
			}
		}
	}
}

func (c *Canvas) NewLine() {
	c.CarriageReturn()
	c.LineFeed(false)
}

func (c *Canvas) MoveCursor(x, y int, relative bool, relativeX bool, relativeY bool) {

	if relative {
		relativeX = true
		relativeY = true
	}

	ctx, cty := c.TermCursor()

	if relativeX {
		x = ctx + x
	}

	if relativeY {
		y = cty + y
	} else if c.terminal.Modes().ConstrainScrolling {
		y += c.scrollRegionStart
	}

	c.SetTermCursor(gwutil.SomeInt(x), gwutil.SomeInt(y))
	c.isRottenCursor = false
}

func (c *Canvas) Clear(newcx, newcy gwutil.IntOption) {
	for y := 0; y < c.BoxRows(); y++ {
		empty := emptyLine(c.BoxColumns())
		c.SetLineAt(y, empty)
	}
	if !newcx.IsNone() && !newcy.IsNone() {
		c.SetTermCursor(newcx, newcy)
	} else {
		c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	}
}

func (c *Canvas) DECAln() {
	for i := 0; i < c.BoxRows(); i++ {
		for j := 0; j < c.BoxColumns(); j++ {
			c.SetCellAt(j, i, gowid.MakeCell('E', gowid.MakeTCellColorExt(tcell.ColorDefault), gowid.MakeTCellColorExt(tcell.ColorDefault), gowid.StyleNone))
		}
	}
}

func (c *Canvas) UseAlternateScreen() {
	if !c.alternateActive {
		tmp := c.ViewPortCanvas
		c.ViewPortCanvas = c.alternate
		c.alternate = tmp
		c.alternateActive = true
	}
}

func (c *Canvas) UseOriginalScreen() {
	if c.alternateActive {
		tmp := c.ViewPortCanvas
		c.ViewPortCanvas = c.alternate
		c.alternate = tmp
		c.alternateActive = false
	}
}

func (c *Canvas) CSIClearTabstop(mode int) {
	switch mode {
	case 0:
		c.SetTabstop(gwutil.NoneInt(), true, false)
	case 3:
		c.SetTabstop(gwutil.NoneInt(), false, true)
	}
}

func (c *Canvas) CSISetKeyboardLEDs(mode int) {
	if mode >= 0 && mode <= 3 {
		c.RunCallbacks(LEDs{}, LEDSState(mode))
	}
}

func (c *Canvas) CSIStatusReport(mode int) {
	switch mode {
	case 5:
		d2 := "\033[0n"
		_, err := c.terminal.Write([]byte(d2))
		if err != nil {
			log.Warnf("Could not write all of %d bytes to terminal pty", len(d2))
		}
	case 6:
		x, y := c.TermCursor()
		d2 := fmt.Sprintf("\033[%d;%dR", y+1, x+1)
		_, err := c.terminal.Write([]byte(d2))
		if err != nil {
			log.Warnf("Could not write all of %d bytes to terminal pty", len(d2))
		}
	}
}

// Report as vt102, like vterm.py
func (c *Canvas) CSIGetDeviceAttributes(qmark bool) {
	if !qmark {
		d2 := "\033[?6c"
		_, err := c.terminal.Write([]byte(d2))
		if err != nil {
			log.Warnf("Could not write all of %d bytes to terminal pty", len(d2))
		}
	}
}

// CSISetScroll sets the scrolling region in the current terminal. top is the line
// number of the first line, bottom the bottom line. If both are 0, the whole screen
// is used.
func (c *Canvas) CSISetScroll(top, bottom int) {
	if top == 0 {
		top = 1
	}
	if bottom == 0 {
		bottom = c.BoxRows()
	}

	if top < bottom && bottom <= c.BoxRows() {
		_, y1 := c.ConstrainCoords(0, top-1, true)
		c.scrollRegionStart = y1
		_, y2 := c.ConstrainCoords(0, bottom-1, true)
		c.scrollRegionEnd = y2
		c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	}
}

func (c *Canvas) CSISetModes(args []int, qmark bool, reset bool) {
	flag := !reset

	for _, mode := range args {
		c.SetMode(mode, flag, qmark, reset)
	}
}

func (c *Canvas) SetMode(mode int, flag bool, qmark bool, reset bool) {
	if qmark {
		switch mode {
		case 1:
			c.terminal.Modes().KeysAutoWrap = true
		case 3:
			c.Clear(gwutil.NoneInt(), gwutil.NoneInt())
		case 5:
			if c.terminal.Modes().ReverseVideo != flag {
				c.ReverseVideo(!flag)
			}
			c.terminal.Modes().ReverseVideo = flag
		case 6:
			c.terminal.Modes().ConstrainScrolling = flag
			c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
		case 7:
			c.terminal.Modes().DontAutoWrap = !flag
		case 25:
			c.terminal.Modes().InvisibleCursor = !flag
			c.SetTermCursor(gwutil.NoneInt(), gwutil.NoneInt())
		case 1000:
			c.terminal.Modes().VT200Mouse = flag
		case 1002:
			c.terminal.Modes().ReportButton = flag
			if flag {
				c.terminal.Modes().VT200Mouse = true
			}
		case 1003:
			c.terminal.Modes().ReportAny = flag
			if flag {
				c.terminal.Modes().VT200Mouse = true
			}
		case 1006:
			c.terminal.Modes().SgrModeMouse = flag
		case 1049:
			if flag {
				c.UseAlternateScreen()
			} else {
				c.UseOriginalScreen()
			}
		}
	} else {
		switch mode {
		case 3:
			c.terminal.Modes().DisplayCtrl = flag
		case 4:
			c.terminal.Modes().Insert = flag
		case 20:
			c.terminal.Modes().LfNl = flag
		}
	}
}

// TODO urwid uses undo - implement it
func (c *Canvas) ReverseVideo(undo bool) {
	for i := 0; i < c.BoxRows(); i++ {
		for j := 0; j < c.BoxColumns(); j++ {
			cell := c.CellAt(j, i)
			fg := cell.ForegroundColor()
			bg := cell.BackgroundColor()
			c.SetCellAt(j, i, cell.WithBackgroundColor(fg).WithForegroundColor(bg))
		}
	}
}

func (c *Canvas) InsertChars(startx, starty gwutil.IntOption, chars int, charo gwutil.RuneOption) {
	if startx.IsNone() || starty.IsNone() {
		myx, myy := c.TermCursor()
		startx = gwutil.SomeInt(myx)
		starty = gwutil.SomeInt(myy)
	}

	if chars == 0 {
		chars = 1
	}

	var cell gowid.Cell

	if charo.IsNone() {
		cell = gowid.Cell{}
	} else {
		cell = c.MakeCellFrom(charo.Val())
	}

	for chars > 0 {
		line := c.Line(starty.Val(), gowid.LineCopy{}).Line
		if startx.Val() >= len(line) {
			c.SetLineAt(starty.Val(), append(line, cell))
		} else {
			dummy := make([]gowid.Cell, len(c.Line(starty.Val(), gowid.LineCopy{}).Line))
			n := 0
			n += copy(dummy[n:], line[0:startx.Val()])
			n += copy(dummy[n:], []gowid.Cell{cell})
			n += copy(dummy[n:], line[startx.Val():])

			c.SetLineAt(starty.Val(), dummy)
		}
		chars--
	}
}

func (c *Canvas) RemoveChars(startx, starty gwutil.IntOption, chars int) {
	if startx.IsNone() || starty.IsNone() {
		myx, myy := c.TermCursor()
		startx = gwutil.SomeInt(myx)
		starty = gwutil.SomeInt(myy)
	}

	if chars == 0 {
		chars = 1
	}

	for chars > 0 {
		line := c.Line(starty.Val(), gowid.LineCopy{}).Line
		if startx.Val() >= len(line) {
			line = line[0:startx.Val()]
		} else {
			line = append(line[0:startx.Val()], line[startx.Val()+1:]...)
		}
		line = append(line, gowid.Cell{})
		c.SetLineAt(starty.Val(), line)
		chars--
	}
}

// InsertLines processes "CSI n L" e.g. "\033[5L". Lines are pushed down
// and blank lines inserted. Note that the 5 is only processed if a scroll
// region is defined - otherwise one line is inserted.
func (c *Canvas) InsertLines(atCursor bool, lines int) {
	var starty gwutil.IntOption
	if atCursor {
		_, myy := c.TermCursor()
		starty = gwutil.SomeInt(myy)
	} else {
		starty = gwutil.SomeInt(c.scrollRegionStart)
	}

	if !c.IsScrollRegionSet() {
		lines = 1
	} else if lines == 0 {
		lines = 1
	}

	region := c.scrollRegionEnd + 1 - starty.Val()

	if lines < region {
		for i := 0; i < region-lines; i++ {
			c.SetLineAt(c.scrollRegionEnd-i, c.Line(c.scrollRegionEnd-(i+lines), gowid.LineCopy{}).Line)
		}
	}

	for i := 0; i < gwutil.Min(lines, region); i++ {
		line := emptyLine(c.BoxColumns())
		c.SetLineAt(starty.Val()+i, line)
	}
}

func (c *Canvas) RemoveLines(atCursor bool, lines int) {
	var starty gwutil.IntOption
	if atCursor {
		_, myy := c.TermCursor()
		starty = gwutil.SomeInt(myy)
	} else {
		starty = gwutil.SomeInt(c.scrollRegionStart)
	}

	if !c.IsScrollRegionSet() {
		lines = 1
	} else if lines == 0 {
		lines = 1
	}

	region := c.scrollRegionEnd + 1 - starty.Val()

	if lines < region {
		for i := 0; i < region-lines; i++ {
			c.SetLineAt(starty.Val()+i, c.Line(starty.Val()+i+lines, gowid.LineCopy{}).Line)
		}
	}

	for i := 0; i < gwutil.Min(lines, region); i++ {
		line := emptyLine(c.BoxColumns())
		c.SetLineAt(c.scrollRegionEnd-i, line)
	}
}

func (c *Canvas) Erase(startx, starty, endx, endy int) {
	sx, sy := c.ConstrainCoords(startx, starty, false)
	ex, ey := c.ConstrainCoords(endx, endy, false)

	if sy == ey {
		for i := sx; i < ex+1; i++ {
			c.SetCellAt(i, sy, gowid.Cell{})
		}
	} else {
		y := sy
		for y <= ey {
			if y == sy {
				for i := sx; i < c.BoxColumns(); i++ {
					c.SetCellAt(i, y, gowid.Cell{})
				}
			} else if y == ey {
				for i := 0; i < ex+1; i++ {
					c.SetCellAt(i, y, gowid.Cell{})
				}
			} else {
				for i := 0; i < c.BoxColumns(); i++ {
					c.SetCellAt(i, y, gowid.Cell{})
				}
			}
			y++
		}
	}
}

func (c *Canvas) CSIEraseLine(mode int) {
	myx, myy := c.TermCursor()
	switch mode {

	case 0:
		c.Erase(myx, myy, c.BoxColumns()-1, myy)
	case 1:
		c.Erase(0, myy, myx, myy)
	case 2:
		for i := 0; i < c.BoxColumns(); i++ {
			c.SetCellAt(i, myy, gowid.Cell{})
		}
	}
}

func (c *Canvas) CSIEraseDisplay(mode int) {
	myx, myy := c.TermCursor()
	switch mode {
	case 0:
		c.Erase(myx, myy, c.BoxColumns()-1, c.BoxRows()-1)
	case 1:
		c.Erase(0, 0, myx, myy)
	case 2:
		c.Clear(gwutil.SomeInt(myx), gwutil.SomeInt(myy))
	}
}

func (c *Canvas) CSISetAttr(args []int) {
	if args[len(args)-1] == 0 {
		c.fg = gwutil.NoneInt()
		c.bg = gwutil.NoneInt()
		c.styles = make(map[string]bool)
	}

	c.fg, c.bg, c.styles = c.SGIToAttribs(args, c.fg, c.bg, c.styles)
}

func (c *Canvas) SGIToAttribs(args []int, fg, bg gwutil.IntOption, styles map[string]bool) (gwutil.IntOption, gwutil.IntOption, map[string]bool) {
	for i := 0; i < len(args); i++ {
		attr := args[i]
		switch {
		case 30 <= attr && attr <= 37:
			fg = gwutil.SomeInt(attr + 1 - 30)
		case 90 <= attr && attr <= 97:
			fg = gwutil.SomeInt(attr - 91 + 8) // 8 basic colors
		case 40 <= attr && attr <= 47:
			bg = gwutil.SomeInt(attr + 1 - 40)
		case 100 <= attr && attr <= 107:
			bg = gwutil.SomeInt(attr - 101 + 8) // 8 basic colors
		case attr == 23:
			// TODO vim sends this
		case attr == 38:
			if i+2 < len(args) && args[i+1] == 5 && args[i+2] >= 0 && args[i+2] <= 255 {
				fg = gwutil.SomeInt(args[i+2] + 1)
				i += 2
			} else if i+4 < len(args) && args[i+1] == 2 && args[i+2] >= 0 && args[i+2] <= 255 && args[i+3] >= 0 && args[i+3] <= 255 && args[i+4] >= 0 && args[i+4] <= 255 {
				fg = gwutil.SomeInt(gowid.CubeStart + (((args[i+2] * gowid.CubeSize256) + args[i+3]) * gowid.CubeSize256) + args[i+4] + 1)
				i += 4
			}
		case attr == 39:
			delete(styles, "underline")
			fg = gwutil.NoneInt()
		case attr == 48:
			if i+2 < len(args) && args[i+1] == 5 && args[i+2] >= 0 && args[i+2] <= 255 {
				bg = gwutil.SomeInt(args[i+2] + 1)
				i += 2
			} else if i+4 < len(args) && args[i+1] == 2 && args[i+2] >= 0 && args[i+2] <= 255 && args[i+3] >= 0 && args[i+3] <= 255 && args[i+4] >= 0 && args[i+4] <= 255 {
				bg = gwutil.SomeInt(gowid.CubeStart + (((args[i+2] * gowid.CubeSize256) + args[i+3]) * gowid.CubeSize256) + args[i+4] + 1)
				i += 4
			}
		case attr == 49:
			bg = gwutil.NoneInt()
		case attr == 10:
			c.charset.ResetSgrIbmpc()
			c.terminal.Modes().DisplayCtrl = false
		case attr == 11 || attr == 12:
			c.charset.SetSgrIbmpc()
			c.terminal.Modes().DisplayCtrl = true
		case attr == 1:
			styles["bold"] = true
		case attr == 4:
			styles["underline"] = true
		case attr == 7:
			styles["reverse"] = true
		case attr == 5:
			styles["blink"] = true
		case attr == 22:
			delete(styles, "bold")
		case attr == 24:
			delete(styles, "underline")
		case attr == 25:
			delete(styles, "blink")
		case attr == 27:
			delete(styles, "reverse")
		case attr == 0:
			fg = gwutil.NoneInt()
			bg = gwutil.NoneInt()
			styles = make(map[string]bool)
		case attr == 3:
		case attr == 6:
		}
	}

	return fg, bg, styles
}

func (c *Canvas) Resize(width, height int) {
	x, y := c.TermCursor()

	if width > c.BoxColumns() {
		c.ExtendRight(gowid.EmptyLine(width - c.BoxColumns()))
	} else if width < c.BoxColumns() {
		c.TrimRight(width)
	}

	// Move upwards - so reduce the offset from the top by the amount the new height
	// is greater than the old height.
	c.Offset -= height - c.Height
	c.Height = height
	if c.Height > c.Canvas.BoxRows() {
		c.Height = c.Canvas.BoxRows()
	} else if c.Height < 1 {
		c.Height = 1
	}
	if c.Offset < 0 {
		c.Offset = 0
	} else if c.Offset > (c.Canvas.BoxRows() - c.Height) {
		c.Offset = c.Canvas.BoxRows() - c.Height
	}

	c.ResetScroll()

	x, y = c.ConstrainCoords(x, y, false)
	c.SetTermCursor(gwutil.SomeInt(x), gwutil.SomeInt(y))

	c.InitTabstops(true)
}

func (c *Canvas) PushCursor(r rune) {
	x, y := c.TermCursor()
	wid := runewidth.RuneWidth(r)

	if !c.terminal.Modes().DontAutoWrap {
		if x+wid == c.BoxColumns() && !c.isRottenCursor {
			c.isRottenCursor = true
			c.PushRune(r, x, y)
		} else {
			x += wid
			if x >= c.BoxColumns() {
				if y >= c.scrollRegionEnd {
					c.Scroll(false)
				} else {
					y += 1
				}
				x = wid
				c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(y))
			}
			c.PushRune(r, x, y)
			c.isRottenCursor = false
		}
	} else {
		if x+wid < c.BoxColumns() {
			x += wid
		}
		c.isRottenCursor = false
		c.PushRune(r, x, y)
	}
}

func (c *Canvas) PushRune(r rune, x, y int) {
	r2 := c.charset.ApplyMapping(r)

	if c.terminal.Modes().Insert {
		c.InsertChars(gwutil.NoneInt(), gwutil.NoneInt(), 1, gwutil.SomeRune(r2))
	} else {
		c.SetRune(r2)
	}

	c.SetTermCursor(gwutil.SomeInt(x), gwutil.SomeInt(y))
}

func (c *Canvas) SetRune(r rune) {
	x, y := c.ConstrainCoords(c.tcx, c.tcy, false)
	c.SetRuneAt(x, y, r)
}

func (c *Canvas) MakeCellFrom(r rune) gowid.Cell {
	var cell gowid.Cell = gowid.MakeCell(r, gowid.MakeTCellColorExt(tcell.ColorDefault), gowid.MakeTCellColorExt(tcell.ColorDefault), gowid.StyleNone)
	if !c.fg.IsNone() {
		cell = cell.WithForegroundColor(gowid.MakeTCellColorExt(tcell.Color(c.fg.Val() - 1)))
	}
	if !c.bg.IsNone() {
		cell = cell.WithBackgroundColor(gowid.MakeTCellColorExt(tcell.Color(c.bg.Val() - 1)))
	}
	if len(c.styles) > 0 {
		for k, _ := range c.styles {
			switch k {
			case "underline":
				cell = cell.WithStyle(gowid.StyleUnderline)
			case "bold":
				cell = cell.WithStyle(gowid.StyleBold)
			case "reverse":
				cell = cell.WithStyle(gowid.StyleReverse)
			case "blink":
				cell = cell.WithStyle(gowid.StyleBlink)
			}
		}
	}
	return cell
}

func (c *Canvas) SetRuneAt(x, y int, r rune) {
	c.SetCellAt(x, y, c.MakeCellFrom(r))
}

func (c *Canvas) LeaveEscape() {
	c.withinEscape = false
	c.parsestate = 0
	c.escbuf = make([]byte, 0)
}

// TODO am I always guaranteed to have something in escbuf?
func (c *Canvas) ParseEscape(r byte) {
	leaveEscape := true
	switch {
	case c.parsestate == 1:
		if _, ok := csiMap[r]; ok {
			c.ParseCSI(r)
			c.parsestate = 0
		} else if ((r == '-') || (r == '0') || (r == '1') || (r == '2') || (r == '3') || (r == '4') || (r == '5') || (r == '6') || (r == '7') || (r == '8') || (r == '9') || (r == ';')) || (len(c.escbuf) == 0 && r == '?') {
			c.escbuf = append(c.escbuf, r)
			leaveEscape = false
		}
	case c.parsestate == 0 && r == ']':
		c.escbuf = make([]byte, 0)
		c.parsestate = 2
		leaveEscape = false
	case c.parsestate == 2 && r == '\x07':
		c.ParseOSC(gwutil.LStripByte(c.escbuf, '0'))
	case c.parsestate == 2 && len(c.escbuf) > 0 && c.escbuf[len(c.escbuf)-1] == EscByte && r == '\\':
		c.ParseOSC(gwutil.LStripByte(c.escbuf[0:len(c.escbuf)-1], '0'))
	case c.parsestate == 2 && len(c.escbuf) > 0 && c.escbuf[0] == 'P' && len(c.escbuf) == 8:
		// TODO Palette (ESC]Pnrrggbb)
	case c.parsestate == 2 && len(c.escbuf) == 0 && r == 'R':
		// TODO Reset Palette
	case c.parsestate == 2:
		c.escbuf = append(c.escbuf, r)
		leaveEscape = false
	case c.parsestate == 0 && r == '[':
		c.escbuf = make([]byte, 0)
		c.parsestate = 1
		leaveEscape = false
	case c.parsestate == 0 && ((r == '%') || (r == '#') || (r == '(') || (r == ')')):
		c.escbuf = make([]byte, 1)
		c.escbuf[0] = r
		c.parsestate = 3
		leaveEscape = false
	case c.parsestate == 3:
		c.ParseNonCSI(r, c.escbuf[0])
	case ((r == 'c') || (r == 'D') || (r == 'E') || (r == 'H') || (r == 'M') || (r == 'Z') || (r == '7') || (r == '8') || (r == '>') || (r == '=')):
		c.ParseNonCSI(r, 0)
	}

	if leaveEscape {
		c.LeaveEscape()
	}
}

func (c *Canvas) ParseOSC(osc []byte) {
	switch {
	case len(osc) > 0 && osc[0] == ';':
		c.RunCallbacks(Title{}, string(osc[1:]))
	case len(osc) > 1 && osc[0] == '3' && osc[1] == ';':
		c.RunCallbacks(Title{}, string(osc[2:]))
	}
}

func (c *Canvas) SetG01(r byte, mod byte) {
	if c.terminal.Modes().Charset == CharsetDefault {
		g := 1
		if mod == '(' {
			g = 0
		}

		var cset string
		switch r {
		case '0':
			cset = "vt100"
		case 'U':
			cset = "ibmpc"
		case 'K':
			cset = "user"
		default:
			cset = "default"
		}

		c.charset.Define(g, cset)
	}
}

func (c *Canvas) ParseNonCSI(r byte, mod byte) {
	switch {
	case r == '8' && mod == '#':
		c.DECAln()
	case mod == '%':
		if r == '@' {
			c.terminal.Modes().Charset = CharsetDefault
		} else if r == '8' || r == 'G' {
			c.terminal.Modes().Charset = CharsetUTF8
		}
	case mod == '(' || mod == ')':
		c.SetG01(r, mod)
	case r == 'M':
		c.LineFeed(true)
	case r == 'D':
		c.LineFeed(false)
	case r == 'c':
		c.Reset()
	case r == 'E':
		c.NewLine()
	case r == 'H':
		c.SetTabstop(gwutil.NoneInt(), false, false)
	case r == 'Z':
		c.CSIGetDeviceAttributes(true)
	case r == '7':
		c.SaveCursor(true)
	case r == '8':
		c.RestoreCursor(true)
	}

}

func (c *Canvas) ParseCSI(r byte) {
	numbuf := make([]int, 0)
	qmark := false

	for i, u := range bytes.Split(c.escbuf, []byte{';'}) {
		if (i == 0) && (len(u) > 0) && (u[0] == '?') {
			qmark = true
			u = u[1:]
		}

		num, err := strconv.Atoi(string(u))
		if err == nil {
			numbuf = append(numbuf, num)
		}
	}

	if cmd, ok := csiMap[r]; ok {
		for cmd.IsAlias() {
			cmd = csiMap[cmd.Alias()]
		}
		for len(numbuf) < cmd.MinArgs() {
			numbuf = append(numbuf, cmd.FallbackArg())
		}
		for i, _ := range numbuf {
			if numbuf[i] == 0 {
				// TODO fishy...
				numbuf[i] = cmd.FallbackArg()
			}
		}
		cmd.Call(c, numbuf, qmark)
	}
}

func (c *Canvas) ProcessByte(b byte) {
	var r rune
	if c.terminal.Modes().Charset == CharsetUTF8 {
		c.utf8Buffer = append(c.utf8Buffer, b)
		r, _ = utf8.DecodeRune(c.utf8Buffer)
		if r == utf8.RuneError {
			return
		}
		c.utf8Buffer = c.utf8Buffer[:0]
	} else {
		r = rune(b)
	}

	c.ProcessByteOrCommand(r)
}

func (c *Canvas) ProcessByteOrCommand(r rune) {
	x, y := c.TermCursor()
	dc := c.terminal.Modes().DisplayCtrl

	switch {
	case r == '\x1b' && c.parsestate != 2:
		c.withinEscape = true
	case r == '\x0d' && !dc:
		c.CarriageReturn()
	case r == '\x0f' && !dc:
		c.charset.Activate(0)
	case r == '\x0e' && !dc:
		c.charset.Activate(1)
	case ((r == '\x0a') || (r == '\x0b') || (r == '\x0c')) && !dc:
		c.LineFeed(false)
		if c.terminal.Modes().LfNl {
			c.CarriageReturn()
		}
	case r == '\x09' && !dc:
		c.Tab(8)
	case r == '\x08' && !dc:
		if x > 0 {
			c.SetTermCursor(gwutil.SomeInt(x-1), gwutil.SomeInt(y))
		}
	case r == '\x07' && c.parsestate != 2 && !dc:
		c.RunCallbacks(Bell{})
	case ((r == '\x18') || (r == '\x1a')) && !dc:
		c.LeaveEscape()
	case ((r == '\x00') || (r == '\x7f')) && !dc:
		// Ignored
	case c.withinEscape:
		c.ParseEscape(byte(r))
	case r == '\x9b' && !dc:
		c.withinEscape = true
		c.escbuf = make([]byte, 0)
		c.parsestate = 1
	default:
		c.PushCursor(r)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
