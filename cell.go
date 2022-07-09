// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package gowid

//======================================================================

// Cell represents a single element of terminal output. The empty value
// is a blank cell with default colors, style, and a 'blank' rune. It is
// closely tied to TCell's underlying cell representation - colors are
// TCell-specific, so are translated from anything more general before a
// Cell is instantiated.
type Cell struct {
	codePoint rune
	fg        TCellColor
	bg        TCellColor
	style     StyleAttrs
}

// MakeCell returns a Cell initialized with the supplied run (char to display),
// foreground color, background color and style attributes. Each color can specify
// "default" meaning whatever the terminal default foreground/background is, or
// "none" meaning no preference, allowing it to be overridden when laid on top
// of another Cell during the render process.
func MakeCell(codePoint rune, fg TCellColor, bg TCellColor, Attr StyleAttrs) Cell {
	return Cell{
		codePoint: codePoint,
		fg:        fg,
		bg:        bg,
		style:     Attr,
	}
}

// MergeUnder returns a Cell representing the receiver merged "underneath" the
// Cell argument provided. This means the argument's rune value will be used
// unless it is "empty", and the cell's color and styling come from the
// argument's value in a similar fashion.
func (c Cell) MergeUnder(upper Cell) Cell {
	res := c
	if upper.codePoint != 0 {
		res.codePoint = upper.codePoint
	}
	return res.MergeDisplayAttrsUnder(upper)
}

// MergeDisplayAttrsUnder returns a Cell representing the receiver Cell with the
// argument Cell's color and styling applied, if they are explicitly set.
func (c Cell) MergeDisplayAttrsUnder(upper Cell) Cell {
	res := c
	ufg, ubg, ust := upper.GetDisplayAttrs()
	if ubg != ColorNone {
		res = res.WithBackgroundColor(ubg)
	}
	if ufg != ColorNone {
		res = res.WithForegroundColor(ufg)
	}
	res.style = res.style.MergeUnder(ust)
	return res
}

// GetDisplayAttrs returns the receiver Cell's foreground and background color
// and styling.
func (c Cell) GetDisplayAttrs() (x TCellColor, y TCellColor, z StyleAttrs) {
	x = c.ForegroundColor()
	y = c.BackgroundColor()
	z = c.Style()
	return
}

// HasRune returns true if the Cell actively specifies a rune to display; otherwise
// false, meaning there it is "empty", and a Cell layered underneath it will have its
// rune displayed.
func (c Cell) HasRune() bool {
	return c.codePoint != 0
}

// Rune will return a rune that can be displayed, if this Cell is being rendered in some
// fashion. If the Cell is empty, then a space rune is returned.
func (c Cell) Rune() rune {
	if !c.HasRune() {
		return ' '
	} else {
		return c.codePoint
	}
}

// WithRune returns a Cell equal to the receiver Cell but that will render the supplied
// rune instead.
func (c Cell) WithRune(r rune) Cell {
	c.codePoint = r
	return c
}

// BackgroundColor returns the background color of the receiver Cell.
func (c Cell) BackgroundColor() TCellColor {
	return c.bg
}

// ForegroundColor returns the foreground color of the receiver Cell.
func (c Cell) ForegroundColor() TCellColor {
	return c.fg
}

// Style returns the style of the receiver Cell.
func (c Cell) Style() StyleAttrs {
	return c.style
}

// WithRune returns a Cell equal to the receiver Cell but that will render no
// rune instead i.e. it is "empty".
func (c Cell) WithNoRune() Cell {
	c.codePoint = 0
	return c
}

// WithBackgroundColor returns a Cell equal to the receiver Cell but that
// will render with the supplied background color instead. Note that this color
// can be set to "none" by passing the value gowid.ColorNone, meaning allow
// Cells layered underneath to determine the background color.
func (c Cell) WithBackgroundColor(a TCellColor) Cell {
	c.bg = a
	return c
}

// WithForegroundColor returns a Cell equal to the receiver Cell but that
// will render with the supplied foreground color instead. Note that this color
// can be set to "none" by passing the value gowid.ColorNone, meaning allow
// Cells layered underneath to determine the background color.
func (c Cell) WithForegroundColor(a TCellColor) Cell {
	c.fg = a
	return c
}

// WithStyle returns a Cell equal to the receiver Cell but that will render
// with the supplied style (e.g. underline) instead. Note that this style
// can be set to "none" by passing the value gowid.AttrNone, meaning allow
// Cells layered underneath to determine the style.
func (c Cell) WithStyle(attr StyleAttrs) Cell {
	c.style = attr
	return c
}

//======================================================================

// CellFromRune returns a Cell with the supplied rune and with default
// coloring and styling.
func CellFromRune(r rune) Cell {
	return MakeCell(r, ColorNone, ColorNone, StyleNone)
}

// CellsFromString is a utility function to turn a string into an array
// of Cells. Note that each Cell has no color or style set.
func CellsFromString(s string) []Cell {
	res := make([]Cell, 0, len(s)) // overcommits, counts chars and not runes, but minimizes reallocations.
	for _, r := range s {
		if r != ' ' {
			res = append(res, CellFromRune(r))
		} else {
			res = append(res, Cell{})
		}
	}
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
