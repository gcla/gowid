// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A port of urwid's "Hello World" example from the tutorial, using gowid widgets.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

//======================================================================

func main() {

	palette := gowid.Palette{
		"banner":  gowid.MakePaletteEntry(gowid.ColorWhite, gowid.MakeRGBColor("#60d")),
		"streak":  gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#60a")),
		"inside":  gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#808")),
		"outside": gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#a06")),
		"bg":      gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#d06")),
	}

	div := divider.NewBlank()
	outside := styled.New(div, gowid.MakePaletteRef("outside"))
	inside := styled.New(div, gowid.MakePaletteRef("inside"))

	helloworld := styled.New(
		text.NewFromContentExt(
			text.NewContent([]text.ContentSegment{
				text.StyledContent("Hello World", gowid.MakePaletteRef("banner")),
			}),
			text.Options{
				Align: gowid.HAlignMiddle{},
			},
		),
		gowid.MakePaletteRef("streak"),
	)

	sf := gowid.RenderFlow{}

	view := styled.New(
		vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: outside, D: sf},
				&gowid.ContainerWidget{IWidget: inside, D: sf},
				&gowid.ContainerWidget{IWidget: helloworld, D: sf},
				&gowid.ContainerWidget{IWidget: inside, D: sf},
				&gowid.ContainerWidget{IWidget: outside, D: sf},
			}),
			gowid.VAlignMiddle{},
			sf),
		gowid.MakePaletteRef("bg"),
	)

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    view,
		Palette: &palette,
	})
	examples.ExitOnErr(err)

	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
