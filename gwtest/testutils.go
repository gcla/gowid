// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package gwtest provides utilities for testing gowid widgets.
package gwtest

import (
	"errors"
	"testing"

	"github.com/gcla/gowid"
	tcell "github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var testAppData gowid.Palette

func init() {
	testAppData = make(gowid.Palette)
	testAppData["test1focus"] = gowid.MakePaletteEntry(gowid.ColorRed, gowid.ColorBlack)
	testAppData["test1notfocus"] = gowid.MakePaletteEntry(gowid.ColorGreen, gowid.ColorBlack)
}

type testApp struct {
	doQuit bool
	gowid.ClickTargets
	lastMouse gowid.MouseState
}

func NewTestApp() *testApp {
	a := &testApp{
		ClickTargets: gowid.MakeClickTargets(),
	}
	return a
}

var D *testApp = NewTestApp()

func ClearTestApp() {
	D.DeleteClickTargets(tcell.Button1)
	D.DeleteClickTargets(tcell.Button2)
	D.DeleteClickTargets(tcell.Button3)
	D.DeleteClickTargets(tcell.ButtonNone)
}

func (d testApp) CellStyler(name string) (gowid.ICellStyler, bool) {
	x, y := testAppData[name]
	return x, y
}

func (d testApp) RangeOverPalette(f func(name string, entry gowid.ICellStyler) bool) {
	for k, v := range testAppData {
		if !f(k, v) {
			break
		}
	}
}

func (d testApp) Quit() {
	d.doQuit = true
}

func (d testApp) Run(f gowid.IAfterRenderEvent) error {
	f.RunThenRenderEvent(&d)
	return nil
}

func (d testApp) GetColorMode() gowid.ColorMode {
	return gowid.Mode256Colors
}

func (d testApp) GetMouseState() gowid.MouseState {
	return gowid.MouseState{
		MouseLeftClicked:   true,
		MouseMiddleClicked: false,
		MouseRightClicked:  false,
	}
}

func (d *testApp) SetLastMouseState(m gowid.MouseState) {
	d.lastMouse = m
}

func (d testApp) GetLastMouseState() gowid.MouseState {
	return d.lastMouse
}

func (d testApp) InCopyMode(...bool) bool {
	return false
}

func (d testApp) Log(lvl log.Level, msg string, fields ...gowid.LogField) {
	panic(errors.New("Must not call!"))
}

func (d testApp) CopyModeClaimedBy(...gowid.IIdentity) gowid.IIdentity {
	panic(errors.New("Must not call!"))
}

func (d testApp) RefreshCopyMode()                            { panic(errors.New("Must not call!")) }
func (d testApp) CopyLevel(...int) int                        { panic(errors.New("Must not call!")) }
func (d testApp) Clips() []gowid.ICopyResult                  { panic(errors.New("Must not call!")) }
func (d testApp) CopyModeClaimedAt(...int) int                { panic(errors.New("Must not call!")) }
func (d testApp) RegisterMenu(m gowid.IMenuCompatible)        { panic(errors.New("Must not call!")) }
func (d testApp) UnregisterMenu(m gowid.IMenuCompatible) bool { panic(errors.New("Must not call!")) }
func (d testApp) GetLog() log.StdLogger                       { panic(errors.New("Must not call!")) }
func (d testApp) SetLog(log.StdLogger)                        { panic(errors.New("Must not call!")) }
func (d testApp) ID() interface{}                             { panic(errors.New("Must not call!")) }
func (d testApp) GetScreen() tcell.Screen                     { panic(errors.New("Must not call!")) }
func (d testApp) Redraw()                                     { panic(errors.New("Must not call!")) }
func (d testApp) Sync()                                       { panic(errors.New("Must not call!")) }
func (d testApp) SetColorMode(gowid.ColorMode)                { panic(errors.New("Must not call!")) }
func (d testApp) SetSubWidget(gowid.IWidget, gowid.IApp)      { panic(errors.New("Must not call!")) }
func (d testApp) SubWidget() gowid.IWidget                    { panic(errors.New("Must not call!")) }

//======================================================================

type CheckBoxTester struct {
	Gotit bool
}

func (f *CheckBoxTester) Changed(t gowid.IApp, w gowid.IWidget, data ...interface{}) {
	f.Gotit = true
}

func (f *CheckBoxTester) ID() interface{} { return "foo" }

//======================================================================

type ButtonTester struct {
	Gotit bool
}

func (f *ButtonTester) Changed(gowid.IApp, gowid.IWidget, ...interface{}) {
	f.Gotit = true
}

func (f *ButtonTester) ID() interface{} { return "foo" }

//======================================================================

func RenderBoxManyTimes(t *testing.T, w gowid.IWidget, minX, maxX, minY, maxY int) {
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			assert.NotPanics(t, func() {
				w.Render(gowid.RenderBox{C: x, R: y}, gowid.Focused, D)
			})
			c := w.Render(gowid.RenderBox{C: x, R: y}, gowid.Focused, D)
			if c.BoxRows() > 0 {
				assert.Equal(t, c.BoxColumns(), x, "foo boxcol=%v boxrows=%v x=%v y=%v", c.BoxColumns(), c.BoxRows(), x, y)
			}
			assert.Equal(t, c.BoxRows(), y)
		}
	}
}

func RenderFlowManyTimes(t *testing.T, w gowid.IWidget, minX, maxX int) {
	for x := minX; x <= maxX; x++ {
		assert.NotPanics(t, func() {
			w.Render(gowid.RenderFlowWith{C: x}, gowid.Focused, D)
		})
		c := w.Render(gowid.RenderFlowWith{C: x}, gowid.Focused, D)
		if c.BoxRows() > 0 {
			assert.Equal(t, c.BoxColumns(), x)
		}
	}
}

func RenderFixedDoesNotPanic(t *testing.T, w gowid.IWidget) {
	assert.NotPanics(t, func() {
		w.Render(gowid.RenderFixed{}, gowid.Focused, D)
	})
}

//======================================================================

//======================================================================

func ClickAt(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.Button1, 0)
}

func ClickUpAt(x, y int) *tcell.EventMouse {
	return tcell.NewEventMouse(x, y, tcell.ButtonNone, 0)
}

func KeyEvent(ch rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
}

func CursorDown() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
}

func CursorUp() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
}

func CursorLeft() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
}

func CursorRight() *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
