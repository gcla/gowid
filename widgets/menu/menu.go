// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package menu is a widget that presents a drop-down menu.
package menu

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/holder"
	"github.com/gcla/gowid/widgets/null"
	"github.com/gcla/gowid/widgets/overlay"
	"github.com/gcla/tcell"
)

//======================================================================

type IWidget interface {
	gowid.ICompositeWidget
	Overlay() overlay.IWidgetSettable
	Open(ISite, gowid.IApp)
	Close(gowid.IApp)
	IsOpen() bool
	CloseKeys() []gowid.IKey  // Keys that should close the current submenu (e.g. left arrow)
	IgnoreKeys() []gowid.IKey // Keys that shouldn't close submenu but should be passed back to main app
	AutoClose() bool
	SetAutoClose(bool, gowid.IApp)
	Width() gowid.IWidgetDimension
	SetWidth(gowid.IWidgetDimension, gowid.IApp)
	Name() string
}

type Options struct {
	CloseKeysProvided  bool
	CloseKeys          []gowid.IKey
	IgnoreKeysProvided bool
	IgnoreKeys         []gowid.IKey
	NoAutoClose        bool
	Modal              bool
}

var (
	DefaultIgnoreKeys = []gowid.IKey{
		gowid.MakeKeyExt(tcell.KeyLeft),
		gowid.MakeKeyExt(tcell.KeyRight),
		gowid.MakeKeyExt(tcell.KeyUp),
		gowid.MakeKeyExt(tcell.KeyDown),
	}

	DefaultCloseKeys = []gowid.IKey{
		gowid.MakeKeyExt(tcell.KeyLeft),
		gowid.MakeKeyExt(tcell.KeyEscape),
	}
)

// Widget overlays one widget on top of another. The bottom widget
// is rendered without the focus at full size. The bottom widget is
// rendered between a horizontal and vertical padding widget set up with
// the sizes provided.
type Widget struct {
	overlay    *overlay.Widget        // So that I can set the "top" widget in the overlay to "open" the menu
	baseHolder *holder.Widget         // Holds the actual base widget
	modal      *rejectKeyInput        // Allow/disallow keys to lower when menu is open
	top        *NavWrapperWidget      // So that I can reinstate it in the overlay to "open" the menu
	name       string                 // The name uses for the canvas anchor
	site       ISite                  // If open, provides the name of the canvas anchor at which the top widget is rendered
	width      gowid.IWidgetDimension // For rendering the top widget
	autoClose  bool                   // If true, then close the menu if it was open, and another widget takes the input
	opts       Options
	Callbacks  *gowid.Callbacks
}

type rejectKeyInput struct {
	gowid.IWidget
	on bool
}

// New takes a widget, rather than a menu model, so that I can potentially style
// the menu. TODO - consider adding styling to menu model?
//
func New(name string, menuw gowid.IWidget, width gowid.IWidgetDimension, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	res := &Widget{
		name:      name,
		width:     width,
		autoClose: !opt.NoAutoClose,
		opts:      opt,
		Callbacks: gowid.NewCallbacks(),
	}

	var _ IWidget = res
	var _ ISiteName = res

	baseHolder := holder.New(null.New())
	closerBase := &rejectKeyInput{
		IWidget: baseHolder,
	}

	// We don't have the base widget at this point
	base := &AutoCloserWidget{
		IWidget: closerBase,
		menu:    res,
	}

	// 	// Makes sure submenu doesn't take keyboard input unless it is "selected"
	top := &NavWrapperWidget{menuw, res}

	ov := overlay.New(
		nil, base,
		gowid.VAlignTop{0}, gowid.RenderFixed{}, //gowid.RenderWithRatio{1.0},
		gowid.HAlignLeft{Margin: 0}, width,
		overlay.Options{
			BottomGetsFocus: true,
		},
	)

	res.overlay = ov
	res.baseHolder = baseHolder
	res.modal = closerBase
	res.top = top

	return res
}

func (w *Widget) AutoClose() bool {
	return w.autoClose
}

func (w *Widget) SetAutoClose(autoClose bool, app gowid.IApp) {
	w.autoClose = autoClose
}

func (w *Widget) Width() gowid.IWidgetDimension {
	return w.overlay.Width()
}

func (w *Widget) SetWidth(width gowid.IWidgetDimension, app gowid.IApp) {
	w.overlay.SetWidth(width, app)
}

func (w *Widget) Name() string {
	return w.name
}

func (w *Widget) Open(site ISite, app gowid.IApp) {
	w.site = site
	site.SetNamer(w, app)
	w.overlay.SetTop(w.top, app)
	if w.opts.Modal {
		w.modal.on = true
	}
}

func (w *Widget) Close(app gowid.IApp) {
	// protect against case where it's closed already
	if w.site != nil {
		w.site.SetNamer(nil, app)
		w.site = nil
	}
	w.overlay.SetTop(nil, app)
	w.modal.on = false
}

func (w *Widget) Overlay() overlay.IWidgetSettable {
	return w.overlay
}

func (w *Widget) String() string {
	return "menu" // TODO: should iterate over submenus
}

func (w *Widget) SetSubWidget(widget gowid.IWidget, app gowid.IApp) {
	w.baseHolder.IWidget = widget
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.baseHolder.IWidget
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return w.RenderSize(size, focus, app)
}

func (w *Widget) IsOpen() bool {
	return w.overlay.Top() != nil
}

func (w *Widget) CloseKeys() []gowid.IKey {
	closeKeys := w.opts.CloseKeys
	if !w.opts.CloseKeysProvided && len(w.opts.CloseKeys) == 0 {
		closeKeys = DefaultCloseKeys
	}
	return closeKeys
}

func (w *Widget) IgnoreKeys() []gowid.IKey {
	ignoreKeys := w.opts.IgnoreKeys
	if !w.opts.IgnoreKeysProvided && len(w.opts.IgnoreKeys) == 0 {
		ignoreKeys = DefaultIgnoreKeys
	}
	return ignoreKeys
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func (w *Widget) Selectable() bool {
	return true
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return UserInput(w, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *rejectKeyInput) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	if _, ok := ev.(*tcell.EventKey); ok && w.on {
		return false
	}
	return gowid.UserInput(w.IWidget, ev, size, focus, app)
}

//======================================================================

type CachedWidget struct {
	gowid.IWidget
	c gowid.ICanvas
}

func (w *CachedWidget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return w.c
}

type CachedOverlay struct {
	overlay.IWidget
	c gowid.ICanvas
}

func (w *CachedOverlay) Bottom() gowid.IWidget {
	return &CachedWidget{w.IWidget.Bottom(), w.c}
}

func (w *CachedOverlay) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return overlay.Render(w, size, focus, app)
}

//======================================================================

func UserInput(w IWidget, ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return gowid.UserInputIfSelectable(w.Overlay(), ev, size, focus, app)
}

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	bfocus := focus.And(w.Overlay().BottomGetsFocus())

	bottomC := gowid.Render(w.Overlay().Bottom(), size, bfocus, app)

	off, ok := bottomC.GetMark(w.Name())
	if !ok {
		// Means menu is closed
		return bottomC
	}

	w.Overlay().SetVAlign(gowid.VAlignTop{off.Y}, app)
	w.Overlay().SetHAlign(gowid.HAlignLeft{off.X}, app)

	// So we don't need to render the bottom canvas twice
	fakeOverlay := &CachedOverlay{w.Overlay(), bottomC}

	return gowid.Render(fakeOverlay, size, focus, app)
}

//======================================================================

type ISiteName interface {
	Name() string
}

type ISite interface {
	Namer() ISiteName
	SetNamer(ISiteName, gowid.IApp)
}

// SiteWidget is a zero-width widget which acts as the coordinates at which a submenu will open
type SiteWidget struct {
	gowid.IWidget
	Options SiteOptions
}

type SiteOptions struct {
	Namer   ISiteName
	XOffset int
	YOffset int
}

func NewSite(opts ...SiteOptions) *SiteWidget {
	var opt SiteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	res := &SiteWidget{
		IWidget: null.New(),
		Options: opt,
	}
	var _ gowid.IWidget = res
	var _ ISite = res

	return res
}

func (w *SiteWidget) Selectable() bool {
	return false
}

func (w *SiteWidget) Namer() ISiteName {
	return w.Options.Namer
}

func (w *SiteWidget) SetNamer(m ISiteName, app gowid.IApp) {
	w.Options.Namer = m
}

func (w *SiteWidget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	res := gowid.Render(w.IWidget, size, focus, app)
	if w.Options.Namer != nil {
		res.SetMark(w.Options.Namer.Name(), w.Options.XOffset, w.Options.YOffset)
	}
	return res
}

func (w *SiteWidget) String() string {
	return fmt.Sprintf("menusite[->%v]", w.Options.Namer)
}

//======================================================================

// AutoCloserWidget is used to detect if a given menu is open when a widget responds to user
// input. Then some action can be taken after that user input (e.g. closing the menu)
type AutoCloserWidget struct {
	gowid.IWidget
	menu IWidget
}

func (w *AutoCloserWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	wasOpen := w.menu.IsOpen()
	res := gowid.UserInput(w.IWidget, ev, size, focus, app)

	// Close the menu if it was open prior to this input operation (i.e. not just opened) and
	// if the non-menu part of the UI took the current input - but only if the input was mouse.
	// This makes it harder to accidentally close the menu by hitting e.g. a right cursor key and
	// it not being accepted by the menu, instead being accepted by the base widget
	if w.menu.AutoClose() && wasOpen && res {
		if _, ok := ev.(*tcell.EventMouse); ok {
			w.menu.Close(app)
		}
	}

	return res
}

//======================================================================

// NavWrapperWidget is used to detect if a given menu is open when a widget responds to user
// input. Then some action can be taken after that user input (e.g. closing the menu)
type NavWrapperWidget struct {
	gowid.IWidget
	menu IWidget
	//index int
}

func (w *NavWrapperWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := false

	// if _, ok := ev.(*tcell.EventKey); ok {
	// 	if w.index != w.menu.Active() {
	// 		return res
	// 	}
	// }

	// Test the subwidget first. It might want to capture certain keys
	res = gowid.UserInput(w.IWidget, ev, size, focus, app)

	// If the submenu itself didn't claim the input, check the close keys
	if !res {
		if evk, ok := ev.(*tcell.EventKey); ok {
			for _, k := range w.menu.CloseKeys() {
				if gowid.KeysEqual(k, evk) {
					w.menu.Close(app)
					res = true
					break
				}
			}
			if !res {
				for _, k := range w.menu.IgnoreKeys() {
					if gowid.KeysEqual(k, evk) {
						res = true
						break
					}
				}
			}
		}
	}
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
