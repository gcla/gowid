// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package palettemap provides a widget that can change the color and style of an inner widget.
package palettemap

import (
	"fmt"

	"github.com/gcla/gowid"
)

//======================================================================

type IPaletteMapper interface {
	GetMappedColor(string) (string, bool)
}

type Map map[string]string

func (p Map) GetMappedColor(key string) (x string, y bool) {
	x, y = p[key]
	return
}

type IPaletteMap interface {
	FocusMap() IPaletteMapper
	NotFocusMap() IPaletteMapper
}

type IWidget interface {
	gowid.ICompositeWidget
	IPaletteMap
}

// Widget that adjusts the palette used - if the rendering context provides for a foreground
// color of red (when focused), this widget can provide a map from red -> green to change its
// display
type Widget struct {
	gowid.IWidget
	focusMap    IPaletteMapper
	notFocusMap IPaletteMapper
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
}

func New(inner gowid.IWidget, focusMap Map, notFocusMap Map) *Widget {
	res := &Widget{
		IWidget:     inner,
		focusMap:    focusMap,
		notFocusMap: notFocusMap,
	}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.IWidget = res
	var _ gowid.ICompositeWidget = res
	var _ IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("palettemap[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(inner gowid.IWidget, app gowid.IApp) {
	w.IWidget = inner
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) FocusMap() IPaletteMapper {
	return w.focusMap
}

func (w *Widget) NotFocusMap() IPaletteMapper {
	return w.notFocusMap
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return w.SubWidget().RenderSize(size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return w.SubWidget().RenderSize(size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return gowid.UserInputIfSelectable(w.IWidget, ev, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func Render(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	newAttrs := make(gowid.Palette)
	var mapToUse IPaletteMapper
	if focus.Focus {
		mapToUse = w.FocusMap()
	} else {
		mapToUse = w.NotFocusMap()
	}
	// walk through palette, making a copy with changes as dictated by either the focus map or
	// not-focus map.
	app.RangeOverPalette(func(k string, v gowid.ICellStyler) bool {
		done := false
		if newk, ok := mapToUse.GetMappedColor(k); ok {
			if newval, ok := app.CellStyler(newk); ok {
				newAttrs[k] = newval
				done = true
			}
		}
		if !done {
			newAttrs[k] = v
		}
		return true
	})
	override := NewOverride(app, &newAttrs)

	res := w.SubWidget().Render(size, focus, override)
	return res
}

//======================================================================

type PaletteOverride struct {
	gowid.IApp
	Palette gowid.IPalette
}

func NewOverride(app gowid.IApp, newattrs gowid.IPalette) *PaletteOverride {
	return &PaletteOverride{
		IApp:    app,
		Palette: newattrs,
	}
}

func (a *PaletteOverride) CellStyler(name string) (gowid.ICellStyler, bool) {
	p, ok := a.Palette.CellStyler(name)
	if ok {
		return p, ok
	}
	return a.IApp.CellStyler(name)
}

func (a *PaletteOverride) RangeOverPalette(f func(k string, v gowid.ICellStyler) bool) {
	a.Palette.RangeOverPalette(f)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
