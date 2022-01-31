// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package terminal

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	"github.com/stretchr/testify/assert"
)

//======================================================================

type FakeTerminal struct {
	modes *Modes
}

func (f *FakeTerminal) Modes() *Modes {
	return f.modes
}

func (f *FakeTerminal) Terminfo() *terminfo.Terminfo {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) Width() int {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) Height() int {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) Connected() bool {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) Write([]byte) (int, error) {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) Bell(gowid.IApp) {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) SetTitle(string, gowid.IApp) {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) GetTitle() string {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) SetLEDs(app gowid.IApp, mode LEDSState) {
	panic(errors.New("Must not call!"))
}

func (f *FakeTerminal) GetLEDs() LEDSState {
	panic(errors.New("Must not call!"))
}

func TestCanvas30(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(10, 1, 100, &f)
	_, err := io.Copy(c, strings.NewReader("hello"))
	assert.NoError(t, err)
	res := strings.Join([]string{"hello     "}, "\n")
	assert.Equal(t, res, c.String())
}

func TestCanvas31(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 2, 100, &f)
	_, err := io.Copy(c, strings.NewReader("\033#8"))
	assert.NoError(t, err)

	res := strings.Join([]string{"EEEE", "EEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[2J"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("123"))
	assert.NoError(t, err)
	res = strings.Join([]string{"123 ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[2Jab"))
	assert.NoError(t, err)
	res = strings.Join([]string{"   a", "b   "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[1;1Hxy"))
	assert.NoError(t, err)
	res = strings.Join([]string{"xy a", "b   "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// going beyond bounds
	_, err = io.Copy(c, strings.NewReader("\033[10;10Hk"))
	assert.NoError(t, err)
	res = strings.Join([]string{"xy a", "b  k"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[?3lv"))
	assert.NoError(t, err)
	res = strings.Join([]string{"v   ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// DecALN - fill canvas with E - "\033#8"
	// Move cursor to (0,0)        - "\033[1;1H"
	// Erase line                  -
	_, err = io.Copy(c, strings.NewReader("\033#8\033[1;1H\033[2K"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "EEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033#8\033[1;1H\033Da"))
	assert.NoError(t, err)
	res = strings.Join([]string{"EEEE", "aEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// cursor at 2,2 (terminal coords)
	_, err = io.Copy(c, strings.NewReader("\033[2K"))
	assert.NoError(t, err)
	res = strings.Join([]string{"EEEE", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
}

func SetupCanvas1(c *Canvas, t *testing.T) {
	_, err := io.Copy(c, strings.NewReader("\033#8"))
	assert.NoError(t, err)
	res := strings.Join([]string{"EEEE", "EEEE", "EEEE", "EEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Set coords=1,1
	_, err = io.Copy(c, strings.NewReader("\033[1;1H"))
	assert.NoError(t, err)
	res = strings.Join([]string{"EEEE", "EEEE", "EEEE", "EEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	x, y := c.TermCursor()
	assert.Equal(t, x, 0, "Failed")
	assert.Equal(t, y, 0, "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[1;1Ha\033[2;1Hb\033[3;1Hc\033[4;1Hd"))
	assert.NoError(t, err)
	res = strings.Join([]string{"aEEE", "bEEE", "cEEE", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
}

func DoScroll(c *Canvas, t *testing.T) {
	// Set scroll region
	_, err := io.Copy(c, strings.NewReader("\033[1;3r"))
	assert.NoError(t, err)
	// Constrain scrolling, coords=1,1
	_, err = io.Copy(c, strings.NewReader("\033[?6h"))
	assert.NoError(t, err)
}

func TestCanvas32(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 4, 100, &f)
	SetupCanvas1(c, t) // "aEEE", "bEEE", "cEEE", "dEEE"
	DoScroll(c, t)     // set scroll region to [0,2] (both inclusive)

	res := strings.Join([]string{"aEEE", "bEEE", "cEEE", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Move cursor to (0, 1) - "\033[2;1H"
	// Insert one line       - "\033[1L"
	_, err := io.Copy(c, strings.NewReader("\033[2;1H\033[1L"))
	assert.NoError(t, err)
	res = strings.Join([]string{"aEEE", "    ", "bEEE", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Erase line
	_, err = io.Copy(c, strings.NewReader("\033[1M"))
	assert.NoError(t, err)
	res = strings.Join([]string{"aEEE", "bEEE", "    ", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

}

func TestCanvas33(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 4, 100, &f)
	SetupCanvas1(c, t)
	DoScroll(c, t)

	// Insert line
	_, err := io.Copy(c, strings.NewReader("\033[2;1H\033[2L"))
	assert.NoError(t, err)
	res := strings.Join([]string{"aEEE", "    ", "    ", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Erase line
	_, err = io.Copy(c, strings.NewReader("\033[2M"))
	assert.NoError(t, err)
	res = strings.Join([]string{"aEEE", "    ", "    ", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

}

func TestCanvas34(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 4, 100, &f)
	SetupCanvas1(c, t)

	res := strings.Join([]string{"aEEE", "bEEE", "cEEE", "dEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Set cursor to y=1 x=0                     - "\033[2;1H"
	// Insert 1 line!! Scroll region not set     - "\033[2L"
	// http://www.inwap.com/pdp10/ansicode.txt
	//runtime.Breakpoint()
	_, err := io.Copy(c, strings.NewReader("\033[2;1H\033[2L"))
	assert.NoError(t, err)
	res = strings.Join([]string{"aEEE", "    ", "bEEE", "cEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Erase 1 line!! Scroll region not set     - "\033[2M"
	// http://www.inwap.com/pdp10/ansicode.txt
	_, err = io.Copy(c, strings.NewReader("\033[2M"))
	assert.NoError(t, err)
	res = strings.Join([]string{"aEEE", "bEEE", "cEEE", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

}

func TestCanvas35(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 2, 100, &f)

	_, err := io.Copy(c, strings.NewReader("\033[1;1HAAAA\033[1;2HB"))
	assert.NoError(t, err)
	res := strings.Join([]string{"ABAA", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[1D"))
	assert.NoError(t, err)
	x, y := c.TermCursor()
	assert.Equal(t, x, 1, "Failed")
	assert.Equal(t, y, 0, "Failed")

	// set terminal insert mode    - "\033[4h"
	// insert "**"
	// set terminal overwrite mode - "\033[4l"
	_, err = io.Copy(c, strings.NewReader("\033[4h**\033[4l"))
	assert.NoError(t, err)
	res = strings.Join([]string{"A**B", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

}

func TestCanvas36(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 2, 100, &f)

	_, err := io.Copy(c, strings.NewReader("\033[1;1HA B "))
	assert.NoError(t, err)
	res := strings.Join([]string{"A B ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[2;1HB\x08\033[2@A\x08"))
	assert.NoError(t, err)
	res = strings.Join([]string{"A B ", "A B "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
}

func TestCanvas37(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 6, 100, &f)

	// Set scroll region to top=2, bottom=6 - "\033[2;6r"
	// Constrain scrolling to this region   - "\033[?6h"
	// Move cursor to row=4, col=0          - "\033[5;1H"
	_, err := io.Copy(c, strings.NewReader("\033[2;6r\033[?6h\033[5;1H"))
	assert.NoError(t, err)
	res := strings.Join([]string{"    ", "    ", "    ", "    ", "    ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("A"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "    ", "A   "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Move cursor to row=4, col=3          - "\033[5;4H"
	// Insert a                             - "a"
	_, err = io.Copy(c, strings.NewReader("\033[5;4Ha"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "    ", "A  a"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	// Insert a CR/LF
	//runtime.Breakpoint()
	_, err = io.Copy(c, strings.NewReader("\x0d\x0a"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "A  a", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("[4;4Ha"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "A  a", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("B"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "A  a", "B   "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[5;4HB"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "A  a", "B  B"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\x08 b"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "    ", "A  a", "B  b"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\x0d\x0a"))
	assert.NoError(t, err)
	res = strings.Join([]string{"    ", "    ", "    ", "A  a", "B  b", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

}

func AssertTermPositionIs(x2, y2 int, c *Canvas, t *testing.T) {
	x, y := c.TermCursor()
	assert.Equal(t, x2, x, "Failed")
	assert.Equal(t, y2, y, "Failed")
}

func TestCanvas38(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 2, 100, &f)

	AssertTermPositionIs(0, 0, c, t)

	_, err := io.Copy(c, strings.NewReader("\033[A"))
	assert.NoError(t, err)
	AssertTermPositionIs(0, 0, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[B"))
	assert.NoError(t, err)
	AssertTermPositionIs(0, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[B"))
	assert.NoError(t, err)
	AssertTermPositionIs(0, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[C"))
	assert.NoError(t, err)
	AssertTermPositionIs(1, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[C\033[C\033[C"))
	assert.NoError(t, err)
	AssertTermPositionIs(3, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[D"))
	assert.NoError(t, err)
	AssertTermPositionIs(2, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[D\033[D\033[D\033[D\033[D"))
	assert.NoError(t, err)
	AssertTermPositionIs(0, 1, c, t)

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	AssertTermPositionIs(0, 0, c, t)

	_, err = io.Copy(c, strings.NewReader("\033#8"))
	assert.NoError(t, err)
	res := strings.Join([]string{"EEEE", "EEEE"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	c.SetTermCursor(gwutil.SomeInt(2), gwutil.SomeInt(0))
	_, err = io.Copy(c, strings.NewReader("\033[J"))
	assert.NoError(t, err)
	res = strings.Join([]string{"EE  ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	c.SetTermCursor(gwutil.SomeInt(3), gwutil.SomeInt(0))
	_, err = io.Copy(c, strings.NewReader("X"))
	assert.NoError(t, err)
	res = strings.Join([]string{"EE X", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	c.SetTermCursor(gwutil.SomeInt(1), gwutil.SomeInt(0))
	_, err = io.Copy(c, strings.NewReader("\033[K"))
	assert.NoError(t, err)
	res = strings.Join([]string{"E   ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")

	c.Clear(gwutil.SomeInt(0), gwutil.SomeInt(0))
	_, err = io.Copy(c, strings.NewReader("\033[7mx"))
	assert.NoError(t, err)
	res = strings.Join([]string{"x   ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	assert.Equal(t, tcell.AttrReverse, c.CellAt(0, 0).Style().OnOff, "Failed")

}

func TestCanvas39(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(4, 2, 100, &f)

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	// save reverse
	_, err := io.Copy(c, strings.NewReader("\033[7mx\0337\033[0my"))
	assert.NoError(t, err)

	res := strings.Join([]string{"xy  ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	assert.Equal(t, c.CellAt(0, 0).Style().OnOff, tcell.AttrReverse, "Failed")
	assert.Equal(t, c.CellAt(0, 1).Style().OnOff, tcell.AttrNone, "Failed")

	AssertTermPositionIs(2, 0, c, t)
	_, err = io.Copy(c, strings.NewReader("\033[D\033[D"))
	assert.NoError(t, err)
	AssertTermPositionIs(0, 0, c, t)
	//io.Copy(c, strings.NewReader("\0337"))
	//AssertTermPositionIs(0, 0, c, t)
	_, err = io.Copy(c, strings.NewReader("z"))
	assert.NoError(t, err)
	AssertTermPositionIs(1, 0, c, t)
	res = strings.Join([]string{"zy  ", "    "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	assert.Equal(t, c.Lines[0][0].Style().OnOff, tcell.AttrNone, "Failed")
	_, err = io.Copy(c, strings.NewReader("\0338"))
	assert.NoError(t, err)
	AssertTermPositionIs(1, 0, c, t)
	_, err = io.Copy(c, strings.NewReader("\033[D"))
	assert.NoError(t, err)
	_, err = io.Copy(c, strings.NewReader("k"))
	assert.NoError(t, err)
	AssertTermPositionIs(1, 0, c, t)
	strings.Join([]string{"ky  ", "    "}, "\n")
	assert.Equal(t, c.CellAt(0, 0).Style().OnOff, tcell.AttrReverse, "Failed")

}

func TestCanvas40(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(3, 5, 100, &f)

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	res := strings.Join([]string{"   ", "   ", "   ", "   ", "   "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 0, c, t)

	_, err := io.Copy(c, strings.NewReader("a\x0d\x0ab\x0d\x0ac\x0d\x0ad\x0d\x0ae"))
	assert.NoError(t, err)
	res = strings.Join([]string{"a  ", "b  ", "c  ", "d  ", "e  "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(1, 4, c, t)

	// scroll region
	_, err = io.Copy(c, strings.NewReader("\033[2;4r"))
	assert.NoError(t, err)
	res = strings.Join([]string{"a  ", "b  ", "c  ", "d  ", "e  "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 0, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[4;1H"))
	assert.NoError(t, err)
	res = strings.Join([]string{"a  ", "b  ", "c  ", "d  ", "e  "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 3, c, t)

	_, err = io.Copy(c, strings.NewReader("\x0a"))
	assert.NoError(t, err)
	res = strings.Join([]string{"a  ", "c  ", "d  ", "   ", "e  "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 3, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[2;1H"))
	assert.NoError(t, err)
	res = strings.Join([]string{"a  ", "c  ", "d  ", "   ", "e  "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 1, c, t)

	//
	_, err = io.Copy(c, strings.NewReader("\033M"))
	assert.NoError(t, err)
	res = strings.Join([]string{"a  ", "   ", "c  ", "d  ", "e  "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 1, c, t)

}

func TestEncoded1(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(8, 2, 100, &f)
	f.Modes().Charset = CharsetUTF8

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	res := strings.Join([]string{"        ", "        "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 0, c, t)

	_, err := io.Copy(c, strings.NewReader("\033[1;0Habc你xyz"))
	assert.NoError(t, err)
	res = strings.Join([]string{"abc你xyz", "        "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(7, 0, c, t)

	c.CarriageReturn()
	AssertTermPositionIs(0, 0, c, t)
	c.LineFeed(false)
	AssertTermPositionIs(0, 1, c, t)
	c.PushCursor('p')
	c.PushCursor('q')
	res = strings.Join([]string{"abc你xyz", "pq      "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(2, 1, c, t)
}

func TestEncoded2(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(3, 2, 100, &f)
	f.Modes().Charset = CharsetUTF8

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	res := strings.Join([]string{"   ", "   "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 0, c, t)

	_, err := io.Copy(c, strings.NewReader("\033[1;0Hab你x"))
	assert.NoError(t, err)
	res = strings.Join([]string{"ab ", "你x"}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(2, 1, c, t)
}

func TestPrivacy1(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(8, 2, 100, &f)
	f.Modes().Charset = CharsetUTF8

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	res := strings.Join([]string{"        ", "        "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 0, c, t)

	_, err := io.Copy(c, strings.NewReader("ab\033^foobar\033\\c"))
	assert.NoError(t, err)
	res = strings.Join([]string{"abc     ", "        "}, "\n")
	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(3, 0, c, t)
}

func TestCanvasVttest1(t *testing.T) {
	f := FakeTerminal{modes: &Modes{}}
	c := NewCanvasOfSize(80, 24, 100, &f)

	c.SetTermCursor(gwutil.SomeInt(0), gwutil.SomeInt(0))
	res := strings.Join([]string{
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 0, c, t)

	_, err := io.Copy(c, strings.NewReader("\033[2J\033[?3l\033[2J\033[1;1H\033[1;1HAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA\033[2;1HBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB\033[3;1HCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC\033[4;1HDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD\033[5;1HEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE\033[6;1HFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF\033[7;1HGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG\033[8;1HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH\033[9;1HIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII\033[10;1HJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJ\033[11;1HKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK\033[12;1HLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL\033[13;1HMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM\033[14;1HNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN\033[15;1HOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO\033[16;1HPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP\033[17;1HQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ\033[18;1HRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR\033[19;1HSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS\033[20;1HTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT\033[21;1HUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU\033[22;1HVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV\033[23;1HWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW\033[24;1HXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\033[4;1HScreen accordion test (Insert & Delete Line). Push <RETURN>"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		"CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
		"Screen accordion test (Insert & Delete Line). Push <RETURN>DDDDDDDDDDDDDDDDDDDDD",
		"EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE",
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		"GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG",
		"HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH",
		"IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII",
		"JJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJ",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM",
		"NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN",
		"OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO",
		"PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP",
		"QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ",
		"RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR",
		"SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS",
		"TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT",
		"UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU",
		"VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV",
		"WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW",
		"XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(59, 3, c, t)

	_, err = io.Copy(c, strings.NewReader("\x0a\033M\033[2K"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		"CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
		"                                                                                ",
		"EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE",
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		"GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG",
		"HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH",
		"IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII",
		"JJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJ",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM",
		"NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN",
		"OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO",
		"PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP",
		"QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ",
		"RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR",
		"SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS",
		"TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT",
		"UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU",
		"VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV",
		"WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW",
		"XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[2;23r\033[?6h\033[1;1H\033[1L\033[1M"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
		"CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC",
		"                                                                                ",
		"EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEE",
		"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		"GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG",
		"HHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH",
		"IIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII",
		"JJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJ",
		"KKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK",
		"LLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL",
		"MMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMM",
		"NNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN",
		"OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO",
		"PPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP",
		"QQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ",
		"RRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR",
		"SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS",
		"TTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT",
		"UUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU",
		"VVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV",
		"                                                                                ",
		"XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[2L\033[2M\033[3L\033[3M\033[4L\033[4M\033[5L\033[5M\033[6L\033[6M\033[7L\033[7M\033[8L\033[8M\033[9L\033[9M\033[10L\033[10M\033[11L\033[11M\033[12L\033[12M\033[13L\033[13M\033[14L\033[14M\033[15L\033[15M\033[16L\033[16M\033[17L\033[17M\033[18L\033[18M\033[19L\033[19M\033[20L\033[20M\033[21L\033[21M\033[22L\033[22M\033[23L\033[23M\033[24L\033[24M"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(0, 1, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[?6l\033[r\033[2;1HTop line: A's, bottom line: X's, this line, nothing more. Push <RETURN>"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"Top line: A's, bottom line: X's, this line, nothing more. Push <RETURN>         ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")

	_, err = io.Copy(c, strings.NewReader("\033[2;1H\033[0J\033[1;2HB\033[1D"))
	//_, err := io.Copy(c, strings.NewReader("\033[2;1H\033[0J\033[1;2HB\033[1D\033[4h******************************************************************************\033[4l\033[4;1HTest of 'Insert Mode'. The top line should be 'A*** ... ***B'. Push <RETURN>")
	assert.NoError(t, err)

	res = strings.Join([]string{
		"ABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(1, 0, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[4h******************************************************************************"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"A******************************************************************************B",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(79, 0, c, t)

	_, err = io.Copy(c, strings.NewReader("\033[4l\033[4;1HTest of 'Insert Mode'. The top line should be 'A*** ... ***B'. Push <RETURN>"))
	assert.NoError(t, err)

	res = strings.Join([]string{
		"A******************************************************************************B",
		"                                                                                ",
		"                                                                                ",
		"Test of 'Insert Mode'. The top line should be 'A*** ... ***B'. Push <RETURN>    ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
		"                                                                                ",
	}, "\n")

	assert.Equal(t, res, c.String(), "Failed")
	AssertTermPositionIs(76, 3, c, t)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
