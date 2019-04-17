// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A gowid test app which exercises the button, grid, progress and radio widgets.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/clicktracker"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/grid"
	"github.com/gcla/gowid/widgets/holder"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/radio"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func main() {

	f := examples.RedirectLogger("widgets3.log")
	defer f.Close()

	styles := gowid.Palette{
		"streak":        gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorRed),
		"test1focus":    gowid.MakePaletteEntry(gowid.ColorBlue, gowid.ColorBlack),
		"test1notfocus": gowid.MakePaletteEntry(gowid.ColorGreen, gowid.ColorBlack),
		"test2focus":    gowid.MakePaletteEntry(gowid.ColorMagenta, gowid.ColorBlack),
		"test2notfocus": gowid.MakePaletteEntry(gowid.ColorCyan, gowid.ColorBlack),
	}

	text1 := text.New("something")
	text2 := styled.NewWithRanges(text1,

		[]styled.AttributeRange{styled.AttributeRange{0, 2, gowid.MakePaletteRef("test1notfocus")}}, []styled.AttributeRange{styled.AttributeRange{0, -1, gowid.MakePaletteRef("test1focus")}})

	text3 := styled.NewWithRanges(text2,

		[]styled.AttributeRange{styled.AttributeRange{0, 4, gowid.MakePaletteRef("test2notfocus")}}, []styled.AttributeRange{styled.AttributeRange{0, -1, gowid.MakePaletteRef("test2focus")}})

	dv1 := divider.NewAscii()
	bw1 := clicktracker.New(
		button.NewDecorated(
			text3,
			button.Decoration{"[==", "==]"},
		),
	)
	bw2 := vpadding.New(bw1, gowid.VAlignMiddle{}, gowid.RenderWithUnits{U: 10})

	fixed := gowid.RenderFixed{}
	flow := gowid.RenderFlow{}

	cb1 := checkbox.NewDecorated(false,
		checkbox.Decoration{button.Decoration{"[[", "]]"}, " X "})
	cbt1 := text.New(" Are you sure?")
	cols1 := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{cb1, fixed},
		&gowid.ContainerWidget{cbt1, fixed},
	})

	cb1.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if _, ok := w.(*checkbox.Widget); !ok {
			panic("Widget was unexpected type!")
		}
		log.Infof("Checkbox clicked!")
	}})

	rbgroup := make([]radio.IWidget, 0)
	rb1 := radio.New(&rbgroup)
	rbt1 := text.New(" option1 ")
	rb2 := radio.New(&rbgroup)
	rbt2 := text.New(" option2 ")
	rb3 := radio.New(&rbgroup)
	rbt3 := text.New(" option3 ")
	c2cols := []gowid.IContainerWidget{
		&gowid.ContainerWidget{rb1, fixed},
		&gowid.ContainerWidget{rbt1, fixed},
		&gowid.ContainerWidget{rb2, fixed},
		&gowid.ContainerWidget{rbt2, fixed},
		&gowid.ContainerWidget{rb3, fixed},
		&gowid.ContainerWidget{rbt3, fixed},
	}
	cols2 := columns.New(c2cols)

	text4 := text.New("abcde")
	text4h := holder.New(text4)
	text4s := styled.NewWithRanges(text4h,
		[]styled.AttributeRange{styled.AttributeRange{0, -1, gowid.MakePaletteRef("test1notfocus")}}, []styled.AttributeRange{styled.AttributeRange{0, -1, gowid.MakePaletteRef("streak")}})

	text4btn := button.New(text4s)
	gfwids := []gowid.IWidget{text4btn, text4btn, text4btn, text4btn, text4btn, text4btn, text4btn, text4btn}
	grid1 := grid.New(gfwids, 20, 3, 1, gowid.HAlignMiddle{})

	bw1.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		m := text1.Content()
		m.AddAt(m.Length(), text.StringContent("x"))
		rb := radio.New(&rbgroup)
		cols2.SetSubWidgets(append(cols2.SubWidgets(),
			&gowid.ContainerWidget{
				IWidget: rb,
				D:       fixed,
			}), app)
	}})

	text4btn.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		text4h.IWidget = text.New("edcba")
	}})

	rb1.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if _, ok := w.(*radio.Widget); !ok {
			panic("Widget was unexpected type!")
		}
		log.Infof("Radio button 1 checked/unchecked!")
	}})

	rb3.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if _, ok := w.(*radio.Widget); !ok {
			panic("Widget was unexpected type!")
		}
		log.Infof("Radio button 3 checked/unchecked!")
	}})

	pw := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{text3, fixed},
		&gowid.ContainerWidget{dv1, flow},
		&gowid.ContainerWidget{bw2, fixed},
		&gowid.ContainerWidget{dv1, flow},
		&gowid.ContainerWidget{cols1, fixed},
		&gowid.ContainerWidget{dv1, flow},
		&gowid.ContainerWidget{cols2, fixed},
		&gowid.ContainerWidget{dv1, flow},
		&gowid.ContainerWidget{grid1, flow},
		&gowid.ContainerWidget{dv1, flow},
	})
	pw2 := vpadding.New(pw, gowid.VAlignMiddle{}, gowid.RenderFlow{})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    pw2,
		Palette: &styles,
		Log:     log.StandardLogger(),
	})
	examples.ExitOnErr(err)

	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
