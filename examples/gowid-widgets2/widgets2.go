// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A gowid test app which exercises the columns, checkbox, edit and styled widgets.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func main() {

	f := examples.RedirectLogger("widgets2.log")
	defer f.Close()

	palette := gowid.Palette{
		"regular": gowid.MakePaletteEntry(gowid.ColorDefault, gowid.ColorDefault),
		"focus":   gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorWhite),
	}

	text1 := text.New("hello1")
	text2 := text.New("hello2 yahoo foo another one bar will it wrap")
	text3 := text.New("cols==7")
	text4 := styled.NewExt(text.New("len4"), gowid.MakePaletteRef("regular"), gowid.MakePaletteRef("focus"))
	text5 := selectable.New(text4)
	edit2 := edit.New(edit.Options{Caption: "Pass:", Text: "foobar", Mask: edit.MakeMask('*')})
	edit3 := edit.New(edit.Options{Caption: "E3:", Text: "something"})
	cb1 := checkbox.New(true)
	div1 := fill.New('|')

	fixed := gowid.RenderFixed{}
	units7 := gowid.RenderWithUnits{U: 7}
	weight1 := gowid.RenderWithWeight{1}
	units1 := gowid.RenderWithUnits{U: 1}

	c1 := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{text1, weight1},
		&gowid.ContainerWidget{div1, units1},
		&gowid.ContainerWidget{cb1, fixed},
		&gowid.ContainerWidget{div1, units1},
		&gowid.ContainerWidget{edit2, units7},
		&gowid.ContainerWidget{div1, units1},
		&gowid.ContainerWidget{text2, weight1},
		&gowid.ContainerWidget{div1, units1},
		&gowid.ContainerWidget{edit3, weight1},
		&gowid.ContainerWidget{div1, units1},
		&gowid.ContainerWidget{text5, fixed},
		&gowid.ContainerWidget{div1, units1},
		&gowid.ContainerWidget{text3, units7},
	})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    c1,
		Palette: &palette,
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
