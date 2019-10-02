// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package bargraph provides a simple plotting widget.
package bargraph

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/overlay"
	"github.com/gcla/gowid/widgets/pile"
)

//======================================================================

type IBarGraph interface {
	GetData() [][]int
	GetAttrs() []gowid.IColor
	GetMax() int
}

type IWidget interface {
	gowid.IWidget
	IBarGraph
}

type Widget struct {
	Data  [][]int
	Max   int
	Attrs []gowid.IColor
	gowid.RejectUserInput
	gowid.NotSelectable
}

func New(atts []gowid.IColor) *Widget {
	res := &Widget{}
	res.Data = make([][]int, 0)
	res.Max = 0
	res.Attrs = atts
	var _ gowid.IWidget = res
	return res
}

func (w *Widget) String() string {
	return "bargraph"
}

func (w *Widget) GetData() [][]int {
	return w.Data
}

func (w *Widget) SetData(l [][]int, max int, app gowid.IApp) {
	w.Data = l
	w.Max = max
}

func (w *Widget) GetAttrs() []gowid.IColor {
	return w.Attrs
}

func (w *Widget) GetMax() int {
	return w.Max
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return RenderSize(w, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	return Render(w, size, focus, app)
}

//''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''''

func RenderSize(w gowid.IWidget, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return gowid.CalculateRenderSizeFallback(w, size, focus, app)
}

func Render(w IBarGraph, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	_, ok := size.(gowid.IRenderBox)
	if !ok {
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderBox"})
	}

	weight1 := gowid.RenderWithWeight{1}
	bgTCellColor := gowid.IColorToTCell(w.GetAttrs()[0], gowid.ColorDefault, app.GetColorMode())

	// TODO - check case when data is empty
	dataIdxLimit := 0
	if len(w.GetData()) > 0 {
		dataIdxLimit = len(w.GetData()[0])
	}

	dataWidgets := make([]*columns.Widget, dataIdxLimit)
	for dataIdx := 0; dataIdx < dataIdxLimit; dataIdx++ {
		cols := make([]gowid.IContainerWidget, len(w.GetData()))
		for i, d := range w.GetData() {
			datum := d[dataIdx]
			dataTCellColor := gowid.IColorToTCell(w.GetAttrs()[(i%(len(w.GetAttrs())-1))+1], gowid.ColorDefault, app.GetColorMode())

			bar := pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{
					fill.NewSolidFromCell(
						gowid.Cell{},
					),
					gowid.RenderWithWeight{w.GetMax() - datum},
				},
				// Use fg as background because I'm using spaces
				&gowid.ContainerWidget{
					fill.NewSolidFromCell(
						gowid.MakeCell(
							' ',
							gowid.ColorNone,
							dataTCellColor,
							gowid.StyleNone),
					),
					gowid.RenderWithWeight{datum}},
			})

			cols[i] = &gowid.ContainerWidget{bar, weight1}
		}
		dataWidgets[dataIdx] = columns.New(cols)
	}

	var res gowid.IWidget = fill.NewSolidFromCell(
		gowid.MakeCell(
			' ',
			gowid.ColorNone,
			bgTCellColor,
			gowid.StyleNone),
	)

	for _, dataWidget := range dataWidgets {
		res = overlay.New(
			dataWidget,
			res,
			gowid.VAlignMiddle{}, gowid.RenderWithRatio{R: 1.0},
			gowid.HAlignMiddle{}, gowid.RenderWithRatio{R: 1.0},
		)
	}

	return res.Render(size, focus, app)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
