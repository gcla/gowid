// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package divider provides a widget that draws a dividing line between other widgets.
package divider

import (
	"fmt"
	"runtime"

	"github.com/gcla/gowid"
)

//======================================================================

var (
	HorizontalLine    = '━'
	AltHorizontalLine = '▀'
)

func init() {
	if runtime.GOOS == "windows" {
		HorizontalLine = '―'
		AltHorizontalLine = HorizontalLine
	}
}

type IDivider interface {
	Opts() Options
}

type IWidget interface {
	gowid.IWidget
	IDivider
}

type Widget struct {
	opts Options
	gowid.RejectUserInput
	gowid.NotSelectable
}

type Options struct {
	Chr          rune
	Above, Below int
}

func New(opts ...Options) *Widget {
	var opt Options
	if len(opts) == 0 {
		opt = Options{
			Chr: '-',
		}
	} else {
		opt = opts[0]
	}
	res := &Widget{
		opts: opt,
	}
	var _ IWidget = res
	return res
}

func NewAscii() *Widget {
	return New(Options{
		Chr: '-',
	})
}

func NewBlank() *Widget {
	return New(Options{
		Chr: ' ',
	})
}

func NewUnicode() *Widget {
	return New(Options{
		Chr: HorizontalLine,
	})
}

func NewUnicodeAlt() *Widget {
	return New(Options{
		Chr: AltHorizontalLine,
	})
}

func (w *Widget) String() string {
	return fmt.Sprintf("div[%c]", w.Opts().Chr)
}

func (w *Widget) Opts() Options {
	return w.opts
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func RenderSize(w IDivider, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	if flow, ok := size.(gowid.IRenderFlowWith); !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFlowWith"})
	} else {
		return gowid.RenderBox{C: flow.FlowColumns(), R: w.Opts().Above + w.Opts().Below + 1}
	}
}

func Render(w IDivider, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	flow, ok := size.(gowid.IRenderFlowWith)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFlowWith"})
	}

	div := gowid.CellFromRune(w.Opts().Chr)
	divArr := make([]gowid.Cell, flow.FlowColumns())
	for i := 0; i < flow.FlowColumns(); i++ {
		divArr[i] = div
	}

	res := gowid.NewCanvas()
	for i := 0; i < w.Opts().Above; i++ {
		res.AppendLine([]gowid.Cell{}, false)
	}
	res.AppendLine(divArr, false)
	for i := 0; i < w.Opts().Below; i++ {
		res.AppendLine([]gowid.Cell{}, false)
	}
	res.AlignRight()

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
