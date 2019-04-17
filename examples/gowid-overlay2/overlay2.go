// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A demonstration of gowid's overlay, fill, asciigraph and radio widgets.
package main

import (
	"math/rand"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/asciigraph"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/hpadding"
	"github.com/gcla/gowid/widgets/overlay"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/radio"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
	"github.com/gdamore/tcell"
	asc "github.com/guptarohit/asciigraph"
	log "github.com/sirupsen/logrus"
)

//======================================================================

var ov *overlay.Widget
var ovh, ovw int = 50, 50

//======================================================================

type handler struct{}

func (h handler) UnhandledInput(app gowid.IApp, ev interface{}) bool {
	handled := false
	if evk, ok := ev.(*tcell.EventKey); ok {
		handled = true
		if evk.Key() == tcell.KeyCtrlC || evk.Key() == tcell.KeyEsc || evk.Rune() == 'q' || evk.Rune() == 'Q' {
			app.Quit()
		} else if evk.Key() == tcell.KeyUp || evk.Rune() == 'u' {
			ovh = gwutil.Min(100, ovh+1)
			ov.SetHeight(gowid.RenderWithRatio{float64(ovh) / 100.0}, app)
		} else if evk.Key() == tcell.KeyDown || evk.Rune() == 'd' {
			ovh = gwutil.Max(0, ovh-1)
			ov.SetHeight(gowid.RenderWithRatio{float64(ovh) / 100.0}, app)
		} else if evk.Key() == tcell.KeyRight {
			ovw = gwutil.Min(100, ovw+1)
			ov.SetWidth(gowid.RenderWithRatio{float64(ovw) / 100.0}, app)
		} else if evk.Key() == tcell.KeyLeft {
			ovw = gwutil.Max(0, ovw-1)
			ov.SetWidth(gowid.RenderWithRatio{float64(ovw) / 100.0}, app)
		} else {
			handled = false
		}
	}
	return handled
}

//======================================================================

func main() {

	f := examples.RedirectLogger("overlay2.log")
	defer f.Close()

	palette := gowid.Palette{
		"red": gowid.MakePaletteEntry(gowid.ColorRed, gowid.ColorDefault),
	}

	fixed := gowid.RenderFixed{}

	rbgroup := make([]radio.IWidget, 0)
	rb1 := radio.New(&rbgroup)
	rbt1 := text.New(" option1 ")
	rb2 := radio.New(&rbgroup)
	rbt2 := text.New(" option2 ")
	rb3 := radio.New(&rbgroup)
	rbt3 := text.New(" option3 ")

	data := []float64{2, 1, 1, 2, -2, 5, 7, 11, 3, 7, 1, 4, 7, 2, 2, 9}
	data2 := []float64{9, 2, 2, 7, 4, 1, 7, 3, 11, 7, 5, -2, 2, 1, 1, 2}
	conf := []asc.Option{}
	graph := asciigraph.New(data, conf)

	callback := func(app gowid.IApp, target gowid.IWidget) {
		if rb1.IsChecked() {
			graph.SetData(data, app)
		}
		if rb2.IsChecked() {
			graph.SetData(data2, app)
		}
		if rb3.IsChecked() {
			data3 := make([]float64, 40)
			for i := 0; i < len(data3); i++ {
				data3[i] = gwutil.Round(rand.Float64() * 14)
			}
			graph.SetData(data3, app)
		}
	}

	rb1.OnClick(gowid.WidgetCallback{gowid.ClickCB{}, callback})
	rb2.OnClick(gowid.WidgetCallback{gowid.ClickCB{}, callback})
	rb3.OnClick(gowid.WidgetCallback{gowid.ClickCB{}, callback})

	c2cols := []gowid.IContainerWidget{
		&gowid.ContainerWidget{rb1, fixed},
		&gowid.ContainerWidget{rbt1, fixed},
		&gowid.ContainerWidget{rb2, fixed},
		&gowid.ContainerWidget{rbt2, fixed},
		&gowid.ContainerWidget{rb3, fixed},
		&gowid.ContainerWidget{rbt3, fixed},
	}
	cols := columns.New(c2cols)

	rows := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{cols, gowid.RenderWithUnits{U: 1}},
		&gowid.ContainerWidget{divider.NewUnicode(), gowid.RenderFlow{}},
		&gowid.ContainerWidget{graph, gowid.RenderWithWeight{1}},
	})

	fcols := framed.NewUnicodeAlt(framed.NewUnicodeAlt(rows))
	top := styled.New(fcols, gowid.MakePaletteRef("red"))
	bottom := vpadding.New(hpadding.New(checkbox.New(false), gowid.HAlignLeft{}, gowid.RenderFixed{}), gowid.VAlignTop{}, gowid.RenderFlow{})

	ov = overlay.New(top, bottom,
		gowid.VAlignMiddle{}, gowid.RenderWithRatio{0.5},
		gowid.HAlignMiddle{}, gowid.RenderWithRatio{0.5})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    ov,
		Palette: &palette,
		Log:     log.StandardLogger(),
	})
	examples.ExitOnErr(err)

	app.MainLoop(handler{})
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
