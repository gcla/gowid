// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package gowid

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

//======================================================================

// IGetScreen provides access to a tcell.Screen object e.g. for rendering
// a canvas to the terminal.
type IGetScreen interface {
	GetScreen() tcell.Screen
}

// IColorMode provides access to a ColorMode value which represents the current
// mode of the terminal e.g. 24-bit color, 256-color, monochrome.
type IColorMode interface {
	GetColorMode() ColorMode
}

// IPalette provides application "palette" information - it can look up a
// Cell styling interface by name (e.g. "main text" -> (black, white, underline))
// and it can let clients apply a function to each member of the palette (e.g.
// in order to construct a new modified palette).
type IPalette interface {
	CellStyler(name string) (ICellStyler, bool)
	RangeOverPalette(f func(key string, value ICellStyler) bool)
}

// IRenderContext proviees palette and color mode information.
type IRenderContext interface {
	IPalette
	IColorMode
}

// IApp is the interface of the application passed to every widget during Render or UserInput.
// It provides several features:
// - a function to terminate the application
// - access to the state of the mouse
// - access to the underlying tcell screen
// - access to an application-specific logger
// - functions to get and set the root widget of the widget hierarchy
// - a method to keep track of which widgets were last "clicked"
//
type IApp interface {
	IRenderContext
	IGetScreen
	ISettableComposite
	Quit()                                                     // Terminate the running gowid app + main loop soon
	Redraw()                                                   // Issue a redraw of the terminal soon
	Sync()                                                     // From tcell's screen - refresh every screen cell e.g. if screen becomes corrupted
	SetColorMode(mode ColorMode)                               // Change the terminal's color mode - 256, 16, mono, etc
	Run(f IAfterRenderEvent) error                             // Send a function to run on the widget rendering goroutine
	SetClickTarget(k tcell.ButtonMask, w IIdentityWidget) bool // When a mouse is clicked on a widget, track that widget. So...
	ClickTarget(func(tcell.ButtonMask, IIdentityWidget))       // when the button is released, we can activate the widget if we are still "over" it
	GetMouseState() MouseState                                 // Which buttons are currently clicked
	GetLastMouseState() MouseState                             // Which buttons were clicked before current event
	RegisterMenu(menu IMenuCompatible)                         // Required for an app to display an overlaying menu
	UnregisterMenu(menu IMenuCompatible) bool                  // Returns false if the menu is not found in the hierarchy
	InCopyMode(...bool) bool                                   // A getter/setter - to set the app into copy mode. Widgets might render differently as a result
	CopyModeClaimedAt(...int) int                              // the level that claims copy, 0 means deepest should claim
	CopyModeClaimedBy(...IIdentity) IIdentity                  // the level that claims copy, 0 means deepest should claim
	RefreshCopyMode()                                          // Give widgets another chance to display copy options (after the user perhaps adjusted the scope of a copy selection)
	Clips() []ICopyResult                                      // If in copy-mode, the app will descend the widget hierarchy with a special user input, gathering options for copying data
	CopyLevel(...int) int                                      // level we're at as we descend
}

// App is an implementation of IApp. The App struct conforms to IApp and
// provides services to a running gowid application, such as access to the
// palette, the screen and the state of the mouse.
type App struct {
	IPalette                                 // App holds an IPalette and provides it to each widget when rendering
	screen            tcell.Screen           // Each app has one screen
	TCellEvents       chan tcell.Event       // Events from tcell e.g. resize
	AfterRenderEvents chan IAfterRenderEvent // Functions intended to run on the widget goroutine
	closing           bool                   // If true then app is in process of closing - it may be draining AfterRenderEvents.
	closingMtx        sync.Mutex             // Make sure an AfterRenderEvent and closing don't race.
	viewPlusMenus     IWidget                // The base widget that is displayed - includes registered menus
	view              IWidget                // The base widget that is displayed under registered menus
	colorMode         ColorMode              // The current color mode of the terminal - 256, 16, mono, etc
	inCopyMode        bool                   // True if the app has been switched into "copy mode", for the user to copy a widget value
	copyClaimed       int                    // True if a widget has "claimed" copy mode during this Render pass
	copyClaimedBy     IIdentity
	copyLevel         int
	refreshCopy       bool

	lastMouse    MouseState    // So I can tell if a button was previously clicked
	MouseState                 // Track which mouse buttons are currently down
	ClickTargets               // When mouse is clicked, track potential interaction here
	log          log.StdLogger // For any application logging
}

// AppArgs is a helper struct, providing arguments for the initialization of App.
type AppArgs struct {
	View    IWidget
	Palette IPalette
	Log     log.StdLogger
}

// IUnhandledInput is used as a handler for application user input that is not handled by any
// widget in the widget hierarchy.
type IUnhandledInput interface {
	UnhandledInput(app IApp, ev interface{}) bool
}

// UnhandledInputFunc satisfies IUnhandledInput, allowing use of a simple function for
// handling input not claimed by any widget.
type UnhandledInputFunc func(app IApp, ev interface{}) bool

func (f UnhandledInputFunc) UnhandledInput(app IApp, ev interface{}) bool {
	return f(app, ev)
}

// IgnoreUnhandledInput is a helper function for main loops that don't need to deal
// with hanlding input that the widgets haven't claimed.
var IgnoreUnhandledInput UnhandledInputFunc = func(app IApp, ev interface{}) bool {
	return false
}

//======================================================================

// ClickTargets is used by the App to keep track of which widgets have been
// clicked. This allows the application to determine if a widget has been
// "selected" which may be best determined across two calls to UserInput - click
// and release.
type ClickTargets struct {
	click map[tcell.ButtonMask][]IIdentityWidget // When mouse is clicked, track potential interaction here
}

func MakeClickTargets() ClickTargets {
	return ClickTargets{
		click: make(map[tcell.ButtonMask][]IIdentityWidget),
	}
}

// SetClickTarget expects a Widget that provides an ID() function. Most
// widgets that can be clicked on can just use the default (&w). But if a
// widget might be recreated between the click down and release, and the
// widget under focus at the time of the release provides the same ID()
// (even if not the same object), then it can be given the click.
//
func (t ClickTargets) SetClickTarget(k tcell.ButtonMask, w IIdentityWidget) bool {
	targets, ok := t.click[k]
	if !ok {
		targets = make([]IIdentityWidget, 1)
		targets[0] = w
	} else {
		targets = append(targets, w)
	}
	t.click[k] = targets
	return !ok
}

func (t ClickTargets) ClickTarget(f func(tcell.ButtonMask, IIdentityWidget)) {
	for k, v := range t.click {
		for _, t := range v {
			f(k, t)
		}
	}
}

func (t ClickTargets) DeleteClickTargets(k tcell.ButtonMask) {
	if ws, ok := t.click[k]; ok {
		for _, w := range ws {
			if w2, ok := w.(IClickTracker); ok {
				w2.SetClickPending(false)
			}
		}
		delete(t.click, k)
	}
}

//======================================================================

type MouseState struct {
	MouseLeftClicked   bool
	MouseMiddleClicked bool
	MouseRightClicked  bool
}

func (m MouseState) String() string {
	return fmt.Sprintf("LeftClicked: %v, MiddleClicked: %v, RightClicked: %v",
		m.MouseLeftClicked,
		m.MouseMiddleClicked,
		m.MouseRightClicked,
	)
}

func (m MouseState) NoButtonClicked() bool {
	return !m.LeftIsClicked() && !m.MiddleIsClicked() && !m.RightIsClicked()
}

func (m MouseState) LeftIsClicked() bool {
	return m.MouseLeftClicked
}

func (m MouseState) MiddleIsClicked() bool {
	return m.MouseMiddleClicked
}

func (m MouseState) RightIsClicked() bool {
	return m.MouseRightClicked
}

//======================================================================

// NewAppSafe returns an initialized App struct, or an error on failure. It will
// initialize a tcell.Screen object behind the scenes, and enable mouse support
// meaning that tcell will receive mouse events if the terminal supports them.
func NewApp(args AppArgs) (rapp *App, rerr error) {
	screen, err := tcell.NewScreenExt()
	if err != nil {
		rerr = WithKVs(err, map[string]interface{}{"TERM": os.Getenv("TERM")})
		return
	}
	if err := screen.Init(); err != nil {
		rerr = WithKVs(err, map[string]interface{}{"TERM": os.Getenv("TERM")})
		return
	}

	var palette IPalette = args.Palette
	if palette == nil {
		palette = make(Palette)
	}

	tch := make(chan tcell.Event, 1000)
	wch := make(chan IAfterRenderEvent, 1000)

	clicks := MakeClickTargets()

	if args.Log == nil {
		logname := filepath.Base(os.Args[0])
		logname = fmt.Sprintf("%s.log", strings.TrimSuffix(logname, filepath.Ext(logname)))
		logfile, err := os.Create(logname)
		if err != nil {
			return nil, err
		}
		logger := log.New()
		logger.Out = logfile
		args.Log = logger
	}

	res := &App{
		IPalette:          palette,
		screen:            screen,
		TCellEvents:       tch,
		AfterRenderEvents: wch,
		closing:           false,
		view:              args.View,
		viewPlusMenus:     args.View,
		colorMode:         Mode256Colors,
		ClickTargets:      clicks,
		log:               args.Log,
	}

	defFg := ColorDefault
	defBg := ColorDefault
	defSt := StyleNone
	if paletteDefault, ok := res.IPalette.CellStyler("default"); ok {
		fgCol, bgCol, style := paletteDefault.GetStyle(res)
		defFg = IColorToTCell(fgCol, defFg, res.GetColorMode())
		defBg = IColorToTCell(bgCol, defBg, res.GetColorMode())
		defSt = defSt.MergeUnder(style)
	}
	defStyle := tcell.Style(defSt.OnOff).Background(defBg.ToTCell()).Foreground(defFg.ToTCell())
	// Ask TCell to set the screen's default style according to the palette's "default"
	// config, if one is provided. This might make every screen cell underlined, for example,
	// in the absence of overriding styling from widgets.
	screen.SetStyle(defStyle)
	screen.EnableMouse()
	screen.Clear()

	cols := screen.Colors()
	switch {
	case cols > 256:
		res.SetColorMode(Mode24BitColors)
	case cols == 256:
		res.SetColorMode(Mode256Colors)
	case cols == 88:
		res.SetColorMode(Mode88Colors)
	case cols == 16:
		res.SetColorMode(Mode16Colors)
	case cols < 0:
		res.SetColorMode(ModeMonochrome)
	default:
		res.SetColorMode(Mode8Colors)
	}

	rapp = res
	return
}

func (a *App) GetScreen() tcell.Screen {
	return a.screen
}

func (a *App) RefreshCopyMode() {
	a.refreshCopy = true
}

func (a *App) CopyLevel(lvl ...int) int {
	if len(lvl) > 0 {
		a.copyLevel = lvl[0]
	}
	return a.copyLevel
}

func (a *App) InCopyMode(on ...bool) bool {
	if len(on) > 0 {
		a.inCopyMode = on[0]
	}
	return a.inCopyMode
}

func (a *App) CopyModeClaimedAt(lvl ...int) int {
	if len(lvl) > 0 {
		a.copyClaimed = lvl[0]
	}
	return a.copyClaimed
}

func (a *App) CopyModeClaimedBy(id ...IIdentity) IIdentity {
	if len(id) > 0 {
		a.copyClaimedBy = id[0]
	}
	return a.copyClaimedBy
}

func (a *App) SetSubWidget(widget IWidget, app IApp) {
	a.view = widget
	if a.viewPlusMenus == nil {
		a.viewPlusMenus = widget
	}
}

func (a *App) SubWidget() IWidget {
	return a.view
}

func (a *App) SetPalette(palette IPalette) {
	a.IPalette = palette
}

func (a *App) GetPalette() IPalette {
	return a.IPalette
}

func (a *App) GetMouseState() MouseState {
	return a.MouseState
}

func (a *App) GetLastMouseState() MouseState {
	return a.lastMouse
}

func (a *App) SetColorMode(mode ColorMode) {
	a.colorMode = mode
}

func (a *App) GetColorMode() ColorMode {
	return a.colorMode
}

// TerminalSize returns the terminal's size.
func (a *App) TerminalSize() (x, y int) {
	x, y = a.screen.Size()
	return
}

type LogField struct {
	Name string
	Val  interface{}
}

type CopyModeEvent struct{}

func (c CopyModeEvent) When() time.Time {
	return time.Time{}
}

type ICopyModeClips interface {
	Collect([]ICopyResult)
}

type CopyModeClipsFn func([]ICopyResult)

func (f CopyModeClipsFn) Collect(clips []ICopyResult) {
	f(clips)
}

type CopyModeClipsEvent struct {
	Action ICopyModeClips
}

func (c CopyModeClipsEvent) When() time.Time {
	return time.Time{}
}

type privateId struct{}

func (n privateId) ID() interface{} {
	return n
}

func (a *App) Clips() []ICopyResult {
	res := make([]ICopyResult, 0)

	cb := CopyModeClipsFn(func(clips []ICopyResult) {
		res = append(res, clips...)
	})

	unh := UnhandledInputFunc(func(app IApp, ev interface{}) bool {
		return true
	})

	a.handleInputEvent(
		CopyModeClipsEvent{
			Action: cb,
		},
		unh,
	)

	return res
}

func (a *App) HandleTCellEvent(ev interface{}, unhandled IUnhandledInput) {
	switch ev := ev.(type) {
	case *tcell.EventBundle:
		needRedraw := false
		for _, ev2 := range ev.Events() {
			switch ev2.(type) {
			case *tcell.EventKey, *tcell.EventMouse, *tcell.EventResize:
				needRedraw = true
			}
			a.HandleTCellEvent2(ev2, unhandled, false)
		}
		if needRedraw {
			a.RedrawTerminal()
		}
	default:
		a.handleTCellEventV1(ev, unhandled, true)
	}
}

// HandleTCellEvent handles an event from the underlying TCell library,
// based on its type (key-press, error, etc.) User input events are sent
// to onInputEvent, which will check the widget hierarchy to see if the
// input can be processed; other events might result in gowid updating its
// internal state, like the size of the underlying terminal.
func (a *App) handleTCellEventV1(ev interface{}, unhandled IUnhandledInput, redraw bool) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		// This makes for a better experience on limited hardware like raspberry pi
		debug.SetGCPercent(-1)
		defer debug.SetGCPercent(100)
		cm := a.InCopyMode()
		a.handleInputEvent(ev, unhandled)
		newCopyMode := (!cm && a.InCopyMode())
		if newCopyMode || a.refreshCopy {
			// Now need to work out which widget claims the copy - choose deepest
			a.copyLevel = 0  // current level as we traverse - start at highest
			if newCopyMode { // newly entered
				a.copyClaimed = 100000 // won't ever nest this deep - widget claims beyond this point or at leaf
				a.copyClaimedBy = privateId{}
			}
			a.handleInputEvent(CopyModeEvent{}, unhandled)
			a.refreshCopy = false
		}
		if redraw {
			a.RedrawTerminal()
		}
	case *tcell.EventMouse:
		switch ev.Buttons() {
		case tcell.Button1:
			a.MouseLeftClicked = true
		case tcell.Button2:
			a.MouseMiddleClicked = true
		case tcell.Button3:
			a.MouseRightClicked = true
		default:
		}
		debug.SetGCPercent(-1)
		defer debug.SetGCPercent(100)
		a.handleInputEvent(ev, unhandled)
		// Make sure we don't hold on to references longer than we need to
		if ev.Buttons() == tcell.ButtonNone {
			a.ClickTargets.DeleteClickTargets(tcell.Button1)
			a.ClickTargets.DeleteClickTargets(tcell.Button2)
			a.ClickTargets.DeleteClickTargets(tcell.Button3)
		}
		a.lastMouse = a.MouseState
		a.MouseState = MouseState{}
		if redraw {
			a.RedrawTerminal()
		}
	case *tcell.EventResize:
		if flog, ok := a.log.(log.FieldLogger); ok {
			flog.WithField("event", ev).Infof("Terminal was resized")
		} else {
			a.log.Printf("Terminal was resized\n")
		}
		if redraw {
			a.RedrawTerminal()
		}
	case *tcell.EventInterrupt:
		if flog, ok := a.log.(log.FieldLogger); ok {
			flog.WithField("event", ev).Infof("Interrupt event from tcell")
		} else {
			a.log.Printf("Interrupt event from tcell: %v\n", ev)
		}
	case *tcell.EventError:
		if flog, ok := a.log.(log.FieldLogger); ok {
			flog.WithField("event", ev).WithField("error", ev.Error()).Errorf("Error event from tcell")
		} else {
			a.log.Printf("Error event from tcell: %v, %v\n", ev, ev.Error())
		}
	default:
		if flog, ok := a.log.(log.FieldLogger); ok {
			flog.WithField("event", ev).Infof("Unanticipated event from tcell")
		} else {
			a.log.Printf("Unanticipated event from tcell: %v\n", ev)
		}
	}
}

// Close should be called by a gowid application after the user terminates the application.
// It will cleanup tcell's screen object.
func (a *App) Close() {
	a.screen.Fini()
}

// StartTCellEvents starts a goroutine that listens for events from TCell. The
// PollEvent function will block until TCell has something to report - when
// something arrives, it is written to the tcellEvents channel. The function
// is provided with a quit channel which is consulted for an event that will
// terminate this goroutine.
func (a *App) StartTCellEvents(quit <-chan Unit, wg *sync.WaitGroup) {
	wg.Add(1)
	go func(quit <-chan Unit) {
		defer wg.Done()
	Loop:
		for {
			a.TCellEvents <- a.screen.PollEvent()
			select {
			case <-quit:
				break Loop
			default:
			}
		}
	}(quit)
}

// StopTCellEvents will cause TCell to generate an interrupt event; an event is posted
// to the quit channel first to stop the TCell event goroutine.
func (a *App) StopTCellEvents(quit chan<- Unit, wg *sync.WaitGroup) {
	quit <- Unit{}
	a.screen.PostEventWait(tcell.NewEventInterrupt(nil))
	wg.Wait()
}

// SimpleMainLoop will run your application using a default unhandled input function
// that will terminate your application on q/Q, ctrl-c and escape.
func (a *App) SimpleMainLoop() {
	a.MainLoop(UnhandledInputFunc(HandleQuitKeys))
}

// HandleQuitKeys is provided as a simple way to terminate your application using typical
// "quit" keys - q/Q, ctrl-c, escape.
func HandleQuitKeys(app IApp, event interface{}) bool {
	handled := false
	if ev, ok := event.(*tcell.EventKey); ok {
		if ev.Key() == tcell.KeyCtrlC || ev.Key() == tcell.KeyEsc || ev.Rune() == 'q' || ev.Rune() == 'Q' {
			app.Quit()
			handled = true
		}
	}
	return handled
}

type AppRunner struct {
	app     *App
	wg      sync.WaitGroup
	started bool
	quitCh  chan Unit
}

func (a *App) Runner() *AppRunner {
	res := &AppRunner{
		app:    a,
		quitCh: make(chan Unit, 100),
	}
	return res
}

func (st *AppRunner) Start() {
	st.app.StartTCellEvents(st.quitCh, &st.wg)
	st.started = true
}

func (st *AppRunner) Stop() {
	if st.started {
		st.app.StopTCellEvents(st.quitCh, &st.wg)
		st.started = false
	}
}

// MainLoop is the intended gowid entry point for typical applications. After the App
// is instantiated and the widget hierarchy set up, the application should call MainLoop
// with a handler for processing input that is not consumed by any widget.
func (a *App) MainLoop(unhandled IUnhandledInput) {
	defer a.Close()
	st := a.Runner()
	st.Start()
	defer st.Stop()
	a.handleEvents(unhandled)
}

// RunThenRenderEvent dispatches the event by calling it with the
// app as an argument - then it will force the application to re-render
// itself.
func (a *App) RunThenRenderEvent(ev IAfterRenderEvent) {
	ev.RunThenRenderEvent(a)
	a.RedrawTerminal()
}

// handleEvents processes all gowid events. These can be either app-generated events
// like a function which must be executed on the render goroutine, or events from
// the underlying TCell library like user input or terminal resize.
func (a *App) handleEvents(unhandled IUnhandledInput) {
Loop:
	for {
		select {
		case ev := <-a.TCellEvents:
			a.HandleTCellEvent(ev, unhandled)
		case ev := <-a.AfterRenderEvents:
			if ev == nil {
				break Loop
			}
			a.RunThenRenderEvent(ev)
		}
	}
}

// handleInputEvent manages key-press events. A keybinding handler is called when
// a key-press or mouse event satisfies a configured keybinding. Furthermore,
// currentView's internal buffer is modified if currentView.Editable is true.
func (a *App) handleInputEvent(ev interface{}, unhandled IUnhandledInput) {
	switch ev.(type) {
	case *tcell.EventKey, *tcell.EventMouse:
		x, y := a.TerminalSize()
		handled := UserInputIfSelectable(a.viewPlusMenus, ev, RenderBox{C: x, R: y}, Focused, a)
		if !handled {
			handled = unhandled.UnhandledInput(a, ev)
			if !handled {
				if flog, ok := a.log.(log.FieldLogger); ok {
					flog.WithField("event", ev).Debugf("Input was not handled")
				} else {
					a.log.Printf("Input was not handled: %v\n", ev)
				}
			}
		}
	default:
		x, y := a.TerminalSize()
		UserInputIfSelectable(a.viewPlusMenus, ev, RenderBox{C: x, R: y}, Focused, a)
	}
}

// Sync defers immediately to tcell's Screen's Sync() function - it is for updating
// every screen cell in the event something corrupts the screen (e.g. ssh -v logging)
func (a *App) Sync() {
	a.screen.Sync()
}

// RedrawTerminal updates the gui, re-drawing frames and buffers. Call this from
// the widget-handling goroutine only. Intended for use by apps that construct their
// own main loops and handle gowid events themselves.
func (a *App) RedrawTerminal() {
	RenderRoot(a.viewPlusMenus, a)
	a.screen.Show()
}

// RegisterMenu should be called by any widget that wants to display a
// menu. The call could be made after initializing the App object. This call
// adds the menu above the current root of the widget hierarchy - when the App
// renders from the root down, any open menus will be rendered on top of the
// original root (using the overlay widget).
func (a *App) RegisterMenu(menu IMenuCompatible) {
	menu.SetSubWidget(a.viewPlusMenus, a)
	a.viewPlusMenus = menu
}

type menuView struct {
	*App
}

// SetSubWidget will set the real root of the widget hierarchy rather than
// the one visible to users of the App. i.e. it allows for a menu to be injected
// into the hierarchy.
func (a *menuView) SetSubWidget(widget IWidget, app IApp) {
	a.viewPlusMenus = widget
}

func (a *App) unregisterMenu(cur ISettableComposite, removeMe IMenuCompatible) bool {
	res := true
	for {
		if sm, ok := cur.SubWidget().(IMenuCompatible); ok {
			if sm == removeMe {
				cur.SetSubWidget(sm.SubWidget(), a)
				break
			} else {
				cur = sm
			}
		} else {
			res = false
			break
		}
	}
	return res
}

// UnregisterMenu will remove a menu from the widget hierarchy. If it's not found,
// false is returned.
func (a *App) UnregisterMenu(menu IMenuCompatible) bool {
	return a.unregisterMenu(&menuView{a}, menu)
}

//======================================================================

type RunFunction func(IApp)

// IAfterRenderEvent is implemented by clients that wish to run a function on the
// gowid rendering goroutine, directly after the widget hierarchy is rendered. This
// allows the client to be sure that there is no race condition with the
// widget rendering code.
type IAfterRenderEvent interface {
	RunThenRenderEvent(IApp)
}

// RunThenRenderEvent lets the receiver RunOnRenderFunction implement IOnRenderEvent. This
// lets a regular function be executed on the same goroutine as the rendering code.
func (f RunFunction) RunThenRenderEvent(app IApp) {
	f(app)
}

var AppClosingErr = fmt.Errorf("App is closing - no more events accepted.")

// Run executes this function on the goroutine that renders
// widgets and processes their callbacks. Any function that manipulates
// widget state outside of the Render/UserInput chain should be run this
// way for thread-safety e.g. a function that changes the UI from a timer
// event.
func (a *App) Run(f IAfterRenderEvent) error {
	a.closingMtx.Lock()
	defer a.closingMtx.Unlock()

	if !a.closing {
		a.AfterRenderEvents <- f
		return nil
	}
	return AppClosingErr
}

// Redraw will re-render the widget hierarchy.
func (a *App) Redraw() {
	a.Run(RunFunction(func(IApp) {}))
}

// Quit will terminate the gowid main loop.
func (a *App) Quit() {
	a.closingMtx.Lock()
	defer a.closingMtx.Unlock()

	a.closing = true
	close(a.AfterRenderEvents)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
