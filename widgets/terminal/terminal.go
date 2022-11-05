// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package terminal provides a widget that functions as a unix terminal. Like urwid, it emulates
// a vt220 (roughly). Mouse support is provided. See the terminal demo for more.
package terminal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/holder"
	"github.com/gcla/gowid/widgets/null"
	"github.com/gcla/gowid/widgets/vscroll"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	"github.com/gdamore/tcell/v2/terminfo/dynamic"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

//======================================================================

// ITerminal is the interface required by terminal.Canvas. For example, when
// the pty sends a byte sequence, the canvas needs to pass it on to the terminal
// implementation - hence io.Writer.
type ITerminal interface {
	io.Writer
	Width() int
	Height() int
	Modes() *Modes
	Terminfo() *terminfo.Terminfo
}

// IWidget encapsulates the requirements of a gowid widget that can represent
// and interact with a terminal.
type IWidget interface {
	// All the usual widget requirements
	gowid.IWidget
	// Support terminal interfaces needed by terminal.Canvas
	ITerminal
	// IHotKeyProvider specifies the keypress that will "unfocus" the widget, that is that will
	// for a period of time ensure that the widget does not accept keypresses. This allows
	// the containing gowid application to change focus to another widget e.g. by hitting
	// the cursor key inside a pile or column widget.
	IHotKeyProvider
	// IHotKeyPersistence determines how long a press of the hotkey will be in effect before
	// keyboard user input is sent back to the underlying terminal.
	IHotKeyPersistence
	// IPaste tracks whether the paste start sequence has been seen wthout a matching
	// paste end sequence
	IPaste
	// HotKeyActive returns true if the hotkey is currently in effect.
	HotKeyActive() bool
	// SetHotKeyActive sets the state of HotKeyActive.
	SetHotKeyActive(app gowid.IApp, down bool)
	// HotKeyDownTime returns the time at which the hotkey was pressed.
	HotKeyDownTime() time.Time
	// Scroll the terminal's buffer.
	Scroll(dir ScrollDir, page bool, lines int)
	// Reset the the terminal's buffer scroll; display what was current.
	ResetScroll()
	// Currently scrolled away from normal view
	Scrolling() bool
}

type IHotKeyFunctions interface {
	// Customized handling of hotkey sequences
	HotKeyFunctions() []IHotKeyInputFn
}

type IScrollbar interface {
	ScrollbarEnabled() bool
	EnableScrollbar(app gowid.IApp)
	DisableScrollbar(app gowid.IApp)
}

type IHotKeyPersistence interface {
	HotKeyDuration() time.Duration
}

type IHotKeyProvider interface {
	HotKey() tcell.Key
}

type IPaste interface {
	PasteState(...bool) bool
}

type HotKeyInputFnExt func(ev *tcell.EventKey, size gowid.IRenderSize, w IWidget, app gowid.IApp) bool
type HotKeyInputFn func(ev *tcell.EventKey, w IWidget, app gowid.IApp) bool

type HotKeyDuration struct {
	D time.Duration
}

func (t HotKeyDuration) HotKeyDuration() time.Duration {
	return t.D
}

type HotKey struct {
	K tcell.Key
}

func (t HotKey) HotKey() tcell.Key {
	return t.K
}

// For callback registration
type Bell struct{}
type LEDs struct{}
type Title struct{}
type ProcessExited struct{}
type HotKeyCB struct{}

type bell struct{}
type leds struct{}
type title struct{}
type hotkey struct{}

type IHotKeyInputFn interface {
	ProcessHotKey(ev *tcell.EventKey, w IWidget, app gowid.IApp) bool
}

type IHotKeyInputFnExt interface {
	ProcessHotKeyExt(ev *tcell.EventKey, size gowid.IRenderSize, w IWidget, app gowid.IApp) bool
}

func (f HotKeyInputFnExt) ProcessHotKeyExt(ev *tcell.EventKey, size gowid.IRenderSize, w IWidget, app gowid.IApp) bool {
	return f(ev, size, w, app)
}

func (f HotKeyInputFnExt) ProcessHotKey(ev *tcell.EventKey, w IWidget, app gowid.IApp) bool {
	panic(nil)
}

func (f HotKeyInputFn) ProcessHotKey(ev *tcell.EventKey, w IWidget, app gowid.IApp) bool {
	return f(ev, w, app)
}

type Options struct {
	Command                 []string
	Env                     []string
	HotKey                  IHotKeyProvider
	HotKeyPersistence       IHotKeyPersistence // the period of time a hotKey sticks after the first post-hotKey keypress
	Scrollback              int
	Scrollbar               bool             // disabled regardless of setting if there is no scrollback
	HotKeyFns               []IHotKeyInputFn // allow custom behavior after pressing the hotkey
	EnableBracketedPaste    bool
	ManualResize            bool // if true, don't resize terminal when gowid screen size changes
	KeyPressToEndScrollMode bool // set to true to enable legacy behavior - when the user has scrolled
	// back to the prompt, still require a keypress (q or Q) to end scroll-mode.
}

// Widget is a widget that hosts a terminal-based application. The user provides the
// command to run, an optional environment in which to run it, and an optional hotKey. The hotKey is
// used to "escape" from the terminal (if using only the keyboard), and serves a similar role to the
// default ctrl-b in tmux. For example, to move focus to a widget to the right, the user could hit
// ctrl-b <right>. See examples/gowid-editor for a demo.
type Widget struct {
	IHotKeyProvider
	IHotKeyPersistence
	params              Options
	Cmd                 *exec.Cmd
	master              *os.File
	canvas              *Canvas
	modes               Modes
	curWidth, curHeight int
	terminfo            *terminfo.Terminfo
	title               string
	leds                LEDSState
	hotKeyDown          bool
	hotKeyDownTime      time.Time
	hotKeyTimer         *time.Timer
	isScrolling         bool
	paste               bool
	hold                *holder.Widget  // used if scrollbar is enabled
	cols                *columns.Widget // used if scrollbar is enabled
	sbar                *vscroll.Widget // used if scrollbar is enabled
	scrollbarTmpOff     bool            // a simple hack to help with UserInput and Render
	Callbacks           *gowid.Callbacks
	gowid.IsSelectable
}

func New(command []string) (*Widget, error) {
	return NewExt(Options{
		Command: command,
		Env:     os.Environ(),
	})
}

func NewExt(opts Options) (*Widget, error) {
	var err error
	var ti *terminfo.Terminfo

	var term string
	for _, s := range opts.Env {
		if strings.HasPrefix(s, "TERM=") {
			term = s[len("TERM="):]
			break
		}
	}

	useDefault := true

	if term != "" {
		ti, err = findTerminfo(term)
		if err == nil {
			useDefault = false
		}
	}

	if useDefault {
		ti, err = findTerminfo("xterm")
	}

	if err != nil {
		return nil, err
	}

	if opts.HotKey == nil {
		opts.HotKey = HotKey{tcell.KeyCtrlB}
	}

	if opts.Scrollback <= 0 {
		opts.Scrollbar = false
	}

	var persistence IHotKeyPersistence
	if opts.HotKeyPersistence != nil {
		persistence = opts.HotKeyPersistence
	} else {
		persistence = &HotKeyDuration{
			D: 2 * time.Second,
		}
	}

	// Always allocate so the scrollbar can be turned on later
	sbar := vscroll.NewExt(vscroll.VerticalScrollbarUnicodeRunes)

	hold := holder.New(null.New())

	cols := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{hold, gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{sbar, gowid.RenderWithUnits{U: 1}},
	})

	res := &Widget{
		params:             opts,
		IHotKeyProvider:    opts.HotKey,
		IHotKeyPersistence: persistence,
		terminfo:           ti,
		sbar:               sbar,
		cols:               cols,
		hold:               hold,
		Callbacks:          gowid.NewCallbacks(),
	}

	res.hold.SetSubWidget(res, nil)
	res.cols.SetFocus(nil, 0)

	sbar.OnClickAbove(gowid.WidgetCallback{"cb", res.clickUp})
	sbar.OnClickBelow(gowid.WidgetCallback{"cb", res.clickDown})
	sbar.OnClickUpArrow(gowid.WidgetCallback{"cb", res.clickUpArrow})
	sbar.OnClickDownArrow(gowid.WidgetCallback{"cb", res.clickDownArrow})

	var _ gowid.IWidget = res
	var _ ITerminal = res
	var _ IWidget = res
	var _ IHotKeyFunctions = res
	var _ IScrollbar = res
	var _ io.Writer = res

	return res, nil
}

func (w *Widget) String() string {
	return fmt.Sprintf("terminal")
}

func (w *Widget) Scrolling() bool {
	return w.isScrolling
}

func (w *Widget) Modes() *Modes {
	return &w.modes
}

func (w *Widget) Terminfo() *terminfo.Terminfo {
	return w.terminfo
}

func (w *Widget) ScrollbarEnabled() bool {
	return w.params.Scrollbar
}

func (w *Widget) EnableScrollbar(app gowid.IApp) {
	w.params.Scrollbar = true
}

func (w *Widget) DisableScrollbar(app gowid.IApp) {
	w.params.Scrollbar = false
}

func (w *Widget) HotKeyFunctions() []IHotKeyInputFn {
	return w.params.HotKeyFns
}

func (w *Widget) Bell(app gowid.IApp) {
	gowid.RunWidgetCallbacks(w.Callbacks, Bell{}, app, w)
}

func (w *Widget) SetLEDs(app gowid.IApp, mode LEDSState) {
	w.leds = mode
	gowid.RunWidgetCallbacks(w.Callbacks, LEDs{}, app, w)
}

func (w *Widget) GetLEDs() LEDSState {
	return w.leds
}

func (w *Widget) SetTitle(title string, app gowid.IApp) {
	w.title = title
	gowid.RunWidgetCallbacks(w.Callbacks, Title{}, app, w)
}

func (w *Widget) GetTitle() string {
	return w.title
}

func (w *Widget) OnProcessExited(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, ProcessExited{}, f)
}

func (w *Widget) RemoveOnProcessExited(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, ProcessExited{}, f)
}

func (w *Widget) OnSetTitle(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, Title{}, f)
}

func (w *Widget) RemoveOnSetTitle(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, Title{}, f)
}

func (w *Widget) OnBell(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, Bell{}, f)
}

func (w *Widget) RemoveOnBell(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, Bell{}, f)
}

func (w *Widget) OnHotKey(f gowid.IWidgetChangedCallback) {
	gowid.AddWidgetCallback(w.Callbacks, HotKeyCB{}, f)
}

func (w *Widget) RemoveOnHotKey(f gowid.IIdentity) {
	gowid.RemoveWidgetCallback(w.Callbacks, HotKeyCB{}, f)
}

func (w *Widget) PasteState(b ...bool) bool {
	if len(b) > 0 {
		w.paste = b[0]
	}
	return w.paste
}

func (w *Widget) HotKeyActive() bool {
	return w.hotKeyDown
}

func (w *Widget) SetHotKeyActive(app gowid.IApp, down bool) {
	w.hotKeyDown = down

	if w.hotKeyTimer != nil {
		w.hotKeyTimer.Stop()
	}

	gowid.RunWidgetCallbacks(w.Callbacks, HotKeyCB{}, app, w)

	if down {
		w.hotKeyDownTime = time.Now()
		w.hotKeyTimer = time.AfterFunc(w.HotKeyDuration(), func() {
			app.Run(&appRunExt{
				fn: func(app gowid.IApp) bool {
					w.SetHotKeyActive(app, false)
					gowid.RunWidgetCallbacks(w.Callbacks, HotKeyCB{}, app, w)
					return false
				},
			})
		})
	}
}

func (w *Widget) HotKeyDownTime() time.Time {
	return w.hotKeyDownTime
}

func (w *Widget) Scroll(dir ScrollDir, page bool, lines int) {
	if page {
		lines = w.canvas.ScrollBuffer(dir, false, gwutil.NoneInt())
	} else {
		lines = w.canvas.ScrollBuffer(dir, false, gwutil.SomeInt(lines))
	}

	wasScrolling := w.isScrolling
	if lines != 0 {
		w.isScrolling = true
	} else if !w.params.KeyPressToEndScrollMode && dir == ScrollDown {
		// Disable scroll if we are at the bottom and we tried to scroll down
		// Thanks @Peter2121 !
		w.isScrolling = false
	}
	if wasScrolling && !w.isScrolling {
		w.ResetScroll()
	}
}

func (w *Widget) ResetScroll() {
	w.isScrolling = false
	w.canvas.ScrollBuffer(false, true, gwutil.NoneInt())
}

func (w *Widget) Width() int {
	return w.curWidth
}

func (w *Widget) Height() int {
	return w.curHeight
}

func (w *Widget) Connected() bool {
	return w.master != nil
}

func (w *Widget) Canvas() *Canvas {
	return w.canvas
}

func (w *Widget) SetCanvas(app gowid.IApp, c *Canvas) {
	w.canvas = c
	if app.GetScreen().CharacterSet() == "UTF-8" {
		w.canvas.terminal.Modes().Charset = CharsetUTF8
	}
}

func (w *Widget) Write(p []byte) (n int, err error) {
	n, err = w.master.Write(p)
	return
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if !w.scrollbarTmpOff && w.params.Scrollbar {
		w.scrollbarTmpOff = true
		res := w.cols.UserInput(ev, size, focus, app)
		w.scrollbarTmpOff = false
		w.cols.SetFocus(app, 0)
		return res
	}
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	box, ok := size.(gowid.IRenderBox)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderBox"})
	}

	logrus.Infof("GCLA: TERM: rerendering widget %p size is %v %T", w, size, size)

	if !w.scrollbarTmpOff && w.params.Scrollbar {
		w.scrollbarTmpOff = true
		c := w.cols.Render(size, focus, app)
		w.scrollbarTmpOff = false
		return c
	}

	w.TouchTerminal(box.BoxColumns(), box.BoxRows(), app)

	w.sbar.Top = w.canvas.Offset
	w.sbar.Middle = w.canvas.scrollRegionEnd
	w.sbar.Bottom = gwutil.Max(0, w.canvas.ViewPortCanvas.Canvas.BoxRows()-(box.BoxRows()+w.canvas.Offset))

	return w.canvas
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	box, ok := size.(gowid.IRenderBox)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderBox"})
	}

	return gowid.RenderBox{C: box.BoxColumns(), R: box.BoxRows()}
}

type terminalSizeSpec struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func (w *Widget) SetUnderlyingTerminalSize(width, height int) error {
	spec := &terminalSizeSpec{
		Row: uint16(height),
		Col: uint16(width),
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		w.master.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(spec)),
	)

	var err error
	if errno != 0 {
		err = errno
	}

	// logrus.Infof("GCLA: TERM: set term size to %d %d err is %v", width, height, err)

	// _, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
	// 	w.master.Fd(),
	// 	syscall.TIOCGWINSZ,
	// 	uintptr(unsafe.Pointer(spec)),
	// )

	// logrus.Infof("GCLA: TERM: fetched term size to %d %d err is %v", spec.Col, spec.Row, errno)

	return err
}

func (w *Widget) SetTerminalSize(width, height int) error {
	err := w.SetUnderlyingTerminalSize(width, height)
	if err == nil {
		w.Canvas().Resize(width, height)
		w.curWidth = width
		w.curHeight = height
	}

	return err
}

type StartCommandError struct {
	Command []string
	Err     error
}

var _ error = StartCommandError{}

func (e StartCommandError) Error() string {
	return fmt.Sprintf("Error running command %v: %v", e.Command, e.Err)
}

func (e StartCommandError) Cause() error {
	return e.Err
}

func (e StartCommandError) Unwrap() error {
	return e.Err
}

//	TouchTerminal2(w, width, height, app)
//}
//
//func TouchTerminal2(w IWidget, width int, height int, app gowid.IApp) {

func (w *Widget) TouchTerminal(width, height int, app gowid.IApp) {
	setTermSize := false

	if w.Canvas() == nil {
		w.SetCanvas(app, NewCanvasOfSize(width, height, w.params.Scrollback, w))
	}
	if !w.Connected() {
		logrus.Infof("GCLA: widget %v %T not connected", w, w)
		err := w.StartCommand(app, width, height) // TODO check for errors
		if err != nil {
			panic(StartCommandError{Command: w.params.Command, Err: err})
		}
		setTermSize = true
	}

	if !w.params.ManualResize && !(w.Width() == width && w.Height() == height) {
		//logrus.Infof("GCLA: TERM: set size width is %d, w is %d, height is %d, h is %d", w.Width(), width, w.Height(), height)
		if !setTermSize {
			err := w.SetTerminalSize(width, height)
			if err != nil {
				log.WithFields(log.Fields{
					"width":  width,
					"height": height,
					"error":  err,
				}).Warn("Could not set terminal size")
			}
		}

		w.Canvas().Resize(width, height)

		w.curWidth = width
		w.curHeight = height
	}

}

func (w *Widget) Signal(sig syscall.Signal) error {
	var err error
	if w.Cmd != nil {
		err = w.Cmd.Process.Signal(sig)
	}
	return err
}

func (w *Widget) RequestTerminate() error {
	return w.Signal(syscall.SIGTERM)
}

func (w *Widget) StartCommand(app gowid.IApp, width, height int) error {
	w.Cmd = exec.Command(w.params.Command[0], w.params.Command[1:]...)
	var err error
	var tty *os.File
	w.master, tty, err = PtyStart1(w.Cmd)
	if err != nil {
		return err
	}
	defer tty.Close()

	err = w.SetTerminalSize(width, height)
	if err != nil {
		log.WithFields(log.Fields{
			"width":  width,
			"height": height,
			"error":  err,
		}).Warn("Could not set terminal size")
	}

	err = w.Cmd.Start()
	logrus.Infof("GCLA: TERM: started w.Cmd %v", w.Cmd)
	if err != nil {
		w.master.Close()
		return err
	}

	master := w.master
	canvas := w.canvas

	canvas.AddCallback(Title{}, gowid.Callback{title{}, func(args ...interface{}) {
		title := args[0].(string)
		app.Run(&appRunExt{
			fn: func(app gowid.IApp) bool {
				w.SetTitle(title, app)
				return false
			},
		})
	}})

	canvas.AddCallback(Bell{}, gowid.Callback{bell{}, func(args ...interface{}) {
		app.Run(&appRunExt{
			fn: func(app gowid.IApp) bool {
				w.Bell(app)
				return false
			},
		})
	}})

	canvas.AddCallback(LEDs{}, gowid.Callback{leds{}, func(args ...interface{}) {
		mode := args[0].(LEDSState)
		app.Run(&appRunExt{
			fn: func(app gowid.IApp) bool {
				w.SetLEDs(app, mode)
				return false
			},
		})
	}})

	if w.params.EnableBracketedPaste {
		app.Run(&appRunExt{
			fn: func(app gowid.IApp) bool {
				redraw := false
				for _, b := range enablePaste(w.terminfo) {
					if canvas.ProcessByteExt(b) {
						redraw = true
					}
				}
				return redraw
			},
		})
	}

	go func() {
		logrus.Infof("GCLA: TERM: goroutine called")

		var wg sync.WaitGroup

		data := make([]byte, 4096)

		for {
			//data := make([]byte, 4096)
			wg.Wait()

			n, err := master.Read(data)
			// s := string(data[0:n])
			// clean := strings.Map(func(r rune) rune {
			// 	if unicode.IsGraphic(r) {
			// 		return r
			// 	}
			// 	return -1
			// }, s)

			// logrus.Infof("GCLA: TERM: read %d from master %p", n, master)
			// logrus.Infof("GCLA: TERM: read %v err is %v", clean, err)

			// n, err = master.Read(data)
			// s = string(data[0:n])
			// clean = strings.Map(func(r rune) rune {
			// 	if unicode.IsGraphic(r) {
			// 		return r
			// 	}
			// 	return -1
			// }, s)

			// logrus.Infof("GCLA: TERM: read2 %d from master %p", n, master)
			// logrus.Infof("GCLA: TERM: read2 %v err is %v", clean, err)

			if n == 0 && err == io.EOF {
				w.Cmd.Wait()
				app.Run(&appRunExt{
					fn: func(app gowid.IApp) bool {
						gowid.RunWidgetCallbacks(w.Callbacks, ProcessExited{}, app, w)
						return false
					},
				})

				break
			} else if err != nil {
				w.Cmd.Wait()
				app.Run(&appRunExt{
					fn: func(app gowid.IApp) bool {
						gowid.RunWidgetCallbacks(w.Callbacks, ProcessExited{}, app, w)
						return false
					},
				})
				break
			}

			wg.Add(1)
			app.Run(&appRunExt{
				fn: func(app gowid.IApp) bool {
					//render := false
					for _, b := range data[0:n] {
						if canvas.ProcessByteExt(b) {
							//render = true
						}
					}
					defer wg.Done()
					defer logrus.Infof("GCLA: TERM: done writing bytes, about to call done")
					return true
				},
			})
		}
	}()

	return nil
}

type runFunctionExt func(gowid.IApp) bool

// appRunExt implements a type that can be passed to app.Run. From the fn, return bool if
// the terminal must be redrawn.
type appRunExt struct {
	fn runFunctionExt
}

var _ gowid.IAppRun = (*appRunExt)(nil)

// I need to implement this to pass to app.Run(). But if I also implement
// RunThenOptionallyRenderEvent, I can better control whether a redraw is issued.
func (t *appRunExt) RunThenRenderEvent(app gowid.IApp) {
	t.fn(app) // old interface throws away result
}

func (t *appRunExt) RunThenOptionallyRenderEvent(app gowid.IApp) bool {
	return t.fn(app)
}

func (w *Widget) StopCommand() {
	if w.master != nil {
		w.master.Close()
	}
}

func (w *Widget) clickUp(app gowid.IApp, w2 gowid.IWidget) {
	w.Scroll(ScrollUp, true, 1)
}

func (w *Widget) clickDown(app gowid.IApp, w2 gowid.IWidget) {
	w.Scroll(ScrollDown, true, 1)
}

func (w *Widget) clickUpArrow(app gowid.IApp, w2 gowid.IWidget) {
	w.Scroll(ScrollUp, false, 1)
}

func (w *Widget) clickDownArrow(app gowid.IApp, w2 gowid.IWidget) {
	w.Scroll(ScrollDown, false, 1)
}

// func (w *Widget) SetWidgetSize(app gowid.IApp, wi int, he int) {
// 	w.Canvas().Resize(wi, he)
// 	w.curWidth = wi
// 	w.curHeight = he
// }

type iTerminalResize interface {
	SetTerminalSize(w, h int) error
	//Canvas() *Canvas
	//SetWidgetSize(app gowid.IApp, w int, h int)
}

//func ResizeTerminalHotKeyFn(ev *tcell.EventKey, size gowid.IRenderSize, w IWidget, app gowid.IApp) bool {

var ResizeTerminalHotKeyFn HotKeyInputFnExt = func(ev *tcell.EventKey, size gowid.IRenderSize, w IWidget, app gowid.IApp) bool {
	logrus.Infof("GCLA: resize11")
	//if ev.Key() == tcell.KeyRune && ev.Rune() == 'l' {
	if ev.Key() == tcell.KeyCtrlL {
		//switch {
		//case k == tcell.KeyCtrlL:

		logrus.Infof("GCLA: resize")

		if size, ok := size.(gowid.IRenderBox); ok {
			if w, ok := w.(iTerminalResize); ok {
				err := w.SetTerminalSize(size.BoxColumns(), size.BoxRows())
				if err != nil {
					log.WithFields(log.Fields{
						"width":  size.BoxColumns(),
						"height": size.BoxRows(),
						"error":  err,
					}).Warn("Could not set terminal size")
					// } else {
					// 	w.SetWidgetSize(app, size.BoxColumns(), size.BoxRows())
					//w.Canvas().Resize(size.BoxColumns(), size.BoxRows())

					//w.curWidth = size.BoxColumns()
					//w.curHeight = size.BoxRows()
				}
			}
			return true
		}
	}
	return false
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	// Set true if this function has claimed the input
	res := false
	// True if input should be sent to tty
	passToTerminal := true

	if evk, ok := ev.(*tcell.EventKey); ok {
		if w.Scrolling() {
			// If we're currently scrolling, then this user input should
			// never be sent to the terminal. It's for controlling or exiting
			// scrolling.
			passToTerminal = false
			res = true
			switch evk.Key() {
			case tcell.KeyPgUp:
				w.Scroll(ScrollUp, true, 0)
			case tcell.KeyPgDn:
				w.Scroll(ScrollDown, true, 0)
			case tcell.KeyUp:
				w.Scroll(ScrollUp, false, 1)
			case tcell.KeyDown:
				w.Scroll(ScrollDown, false, 1)
			case tcell.KeyRune:
				switch evk.Rune() {
				case 'q', 'Q':
					w.ResetScroll()
				}
			default:
				res = false
			}
		} else if w.HotKeyActive() {
			// If we're not scrolling but the hotkey is still active (recently
			// pressed) then the input will not go to the terminal - it's hotkey
			// function processing.
			passToTerminal = false
			res = false
			deactivate := false
			if whk, ok := w.(IHotKeyFunctions); ok {
				for _, fn := range whk.HotKeyFunctions() {
					if fne, ok := fn.(IHotKeyInputFnExt); ok {
						res = fne.ProcessHotKeyExt(evk, size, w, app)
					} else {
						res = fn.ProcessHotKey(evk, w, app)
					}
					if res {
						deactivate = true
						break
					}
				}
			}
			if !res {
				res = true
				switch evk.Key() {
				case w.HotKey():
					deactivate = true
				case tcell.KeyPgUp:
					w.Scroll(ScrollUp, true, 0)
					deactivate = true
				case tcell.KeyPgDn:
					w.Scroll(ScrollDown, true, 0)
					deactivate = true
				case tcell.KeyUp:
					w.Scroll(ScrollUp, false, 1)
					deactivate = true
				case tcell.KeyDown:
					w.Scroll(ScrollDown, false, 1)
					deactivate = true
				default:
					res = false
				}
			}
			if deactivate {
				w.SetHotKeyActive(app, false)
			}
		} else if evk.Key() == w.HotKey() {
			passToTerminal = false
			w.SetHotKeyActive(app, true)
			res = true // handled
		}
	}
	// If nothing has claimed the user input yet, then if the input is
	// mouse input, disqualify it if it's out of bounds of the terminal.
	if !res {
		if ev2, ok := ev.(*tcell.EventMouse); ok {
			mx, my := ev2.Position()
			if !((mx < w.Width()) && (my < w.Height())) {
				passToTerminal = false
			}
		}
	}
	if passToTerminal {
		seq, parsed := TCellEventToBytes(ev, w.Modes(), app.GetLastMouseState(), w, w.Terminfo())

		if parsed {
			logrus.Infof("GCLA: TERM: writing parsed")
			_, err := w.Write(seq)
			if err != nil {
				log.WithField("error", err).Warn("Could not send all input to terminal")
			}
			res = true
		}
	}

	return res
}

// PtyStart1 connects the supplied Cmd's stdin/stdout/stderr to a new tty
// object. The function returns the pty and tty, and also an error which is
// nil if the operation was successful.
func PtyStart1(c *exec.Cmd) (pty2, tty *os.File, err error) {
	pty2, tty, err = pty.Open()
	if err != nil {
		return nil, nil, err
	}
	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setctty = true
	c.SysProcAttr.Setsid = true
	return pty2, tty, err
}

//======================================================================

var cachedTerminfo map[string]*terminfo.Terminfo
var cachedTerminfoMutex sync.Mutex

func init() {
	cachedTerminfo = make(map[string]*terminfo.Terminfo)
}

// findTerminfo returns a terminfo struct via tcell's dynamic method first,
// then using the built-in databases. The aim is to use the terminfo database
// most likely to be correct. Maybe even better would be parsing the terminfo
// file directly using something like https://github.com/beevik/terminfo/, to
// avoid the extra process.
func findTerminfo(name string) (*terminfo.Terminfo, error) {
	cachedTerminfoMutex.Lock()
	if ti, ok := cachedTerminfo[name]; ok {
		cachedTerminfoMutex.Unlock()
		return ti, nil
	}
	ti, _, e := dynamic.LoadTerminfo(name)
	if e == nil {
		cachedTerminfo[name] = ti
		cachedTerminfoMutex.Unlock()
		return ti, nil
	}
	ti, e = terminfo.LookupTerminfo(name)
	cachedTerminfoMutex.Unlock()
	return ti, e
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
