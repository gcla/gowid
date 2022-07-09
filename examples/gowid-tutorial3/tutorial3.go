// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// The third example from the gowid tutorial.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

//======================================================================

func main() {
	palette := gowid.Palette{
		"banner": gowid.MakePaletteEntry(gowid.ColorBlack, gowid.NewUrwidColor("light gray")),
		"streak": gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorRed),
		"bg":     gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorDarkBlue),
	}

	txt := text.NewFromContentExt(
		text.NewContent([]text.ContentSegment{
			text.StyledContent("hello world", gowid.MakePaletteRef("banner")),
		}), text.Options{
			Align: gowid.HAlignMiddle{},
		})

	map1 := styled.New(txt, gowid.MakePaletteRef("streak"))
	vert := vpadding.New(map1, gowid.VAlignMiddle{}, gowid.RenderFlow{})
	map2 := styled.New(vert, gowid.MakePaletteRef("bg"))
	app, err := gowid.NewApp(gowid.AppArgs{
		View:    map2,
		Palette: palette,
	})
	examples.ExitOnErr(err)
	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
