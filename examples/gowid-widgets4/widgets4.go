// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A gowid test app which exercises the columns, list and framed widgets.
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/palettemap"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func main() {

	f := examples.RedirectLogger("widgets4.log")
	defer f.Close()

	styles := gowid.Palette{
		"red":    gowid.MakePaletteEntry(gowid.ColorRed, gowid.ColorBlack),
		"invred": gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorRed),
	}

	widgets := make([]gowid.IWidget, 0)
	widgets2 := make([]gowid.IWidget, 0)
	widgets3 := make([]gowid.IWidget, 0)
	widgets4 := make([]gowid.IWidget, 0)

	wid8 := gowid.RenderWithUnits{U: 8}
	wid10 := gowid.RenderWithUnits{U: 10}
	wid12 := gowid.RenderWithUnits{U: 12}
	wid14 := gowid.RenderWithUnits{U: 14}

	nl := gowid.MakePaletteRef

	for i := 0; i < 23; i++ {
		t := text.NewContent([]text.ContentSegment{
			text.StyledContent(fmt.Sprintf("abc%dd", i), nl("invred")),
		})
		mt := text.NewFromContent(t)
		mta := selectable.New(palettemap.New(mt, palettemap.Map{}, palettemap.Map{"invred": "red"}))
		widgets = append(widgets, mta)

		t2 := text.NewContent([]text.ContentSegment{
			text.StyledContent(fmt.Sprintf("abc%ddefghi", i), nl("invred")),
		})
		mt2 := text.NewFromContent(t2)
		mta2 := selectable.New(palettemap.New(mt2, palettemap.Map{}, palettemap.Map{"invred": "red"}))
		widgets2 = append(widgets2, mta2)

		t3 := text.NewContent([]text.ContentSegment{
			text.StyledContent(fmt.Sprintf("%d%d%d-1-2-3-4-5-6-7-8-9-10-11-12-13-14-abcdefghijklmn", i, i, i), nl("invred")),
		})
		mt3 := text.NewFromContent(t3)
		mta3 := selectable.New(palettemap.New(mt3, palettemap.Map{}, palettemap.Map{"invred": "red"}))
		widgets3 = append(widgets3, mta3)

		t4 := text.NewContent([]text.ContentSegment{
			text.StyledContent(fmt.Sprintf("%d%d%d-1-2-3-4-5-6-7-8-9-10-11-12-13-14-abcdefghijklmn", i, i, i), nl("invred")),
		})
		mt4 := text.NewFromContent(t4)
		mta4 := selectable.New(palettemap.New(mt4, palettemap.Map{}, palettemap.Map{"invred": "red"}))
		widgets4 = append(widgets4, mta4)

	}

	walker := list.NewSimpleListWalker(widgets)
	lb := list.New(walker)
	lbb := vpadding.NewBox(lb, 7)
	fr := framed.New(lbb)

	walker2 := list.NewSimpleListWalker(widgets2)
	lb2 := list.New(walker2)
	lbb2 := vpadding.NewBox(lb2, 7)
	fr2 := framed.New(lbb2)

	walker3 := list.NewSimpleListWalker(widgets3)
	lb3 := list.New(walker3)
	lbb3 := vpadding.NewBox(lb3, 7)
	fr3 := framed.New(lbb3)

	walker4 := list.NewSimpleListWalker(widgets4)
	lb4 := list.New(walker4)
	lbb4 := vpadding.NewBox(lb4, 7)
	fr4 := framed.New(lbb4)

	c1 := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fr, wid10},
		&gowid.ContainerWidget{fr2, wid12},
		&gowid.ContainerWidget{fr3, wid14},
		&gowid.ContainerWidget{fr4, wid8},
	})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    c1,
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
