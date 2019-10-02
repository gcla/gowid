// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package asciigraph provides a simple plotting widget.
package asciigraph

import (
	"strings"
	"unicode/utf8"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/guptarohit/asciigraph"
)

//======================================================================

type IAsciiGraph interface {
	GetData() []float64
	GetConf() []asciigraph.Option
}

type IWidget interface {
	gowid.IWidget
	IAsciiGraph
}

type Widget struct {
	Data []float64
	Conf []asciigraph.Option
	gowid.RejectUserInput
	gowid.NotSelectable
}

func New(series []float64, config []asciigraph.Option) *Widget {
	res := &Widget{}
	res.Data = series
	res.Conf = config
	var _ IWidget = res
	return res
}

func (w *Widget) String() string {
	return "asciigraph"
}

func (w *Widget) GetData() []float64 {
	return w.Data
}

func (w *Widget) SetData(data []float64, app gowid.IApp) {
	w.Data = data
}

func (w *Widget) GetConf() []asciigraph.Option {
	return w.Conf
}

func (w *Widget) SetConf(conf []asciigraph.Option, app gowid.IApp) {
	w.Conf = conf
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

// TODO: ILineWidget?
func RenderSize(w IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func Render(w IAsciiGraph, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {

	grender := strings.Split(asciigraph.Plot(w.GetData(), w.GetConf()...), "\n")

	rows := len(grender)
	cols := 0
	if rows > 0 {
		cols = utf8.RuneCountInString(grender[0])
	}

	switch sz := size.(type) {
	case gowid.IRenderBox:
		cols = sz.BoxColumns()
		rows = sz.BoxRows()
	case gowid.IRenderFlowWith:
		panic(gowid.WidgetSizeError{Widget: w, Size: size})
	}

	blank := fill.NewEmpty()
	res := blank.Render(gowid.RenderBox{C: cols, R: rows}, gowid.NotSelected, app)

	for y := 0; y < gwutil.Min(len(grender), res.BoxRows()); y++ {
		x := 0
		for _, r := range grender[y] {
			if x >= res.BoxColumns() {
				break
			}
			res.SetCellAt(x, y, res.CellAt(x, y).WithRune(r))
			x++
		}
	}

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
