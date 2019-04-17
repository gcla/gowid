// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A gowid test app which exercises the list, edit, columns and styled widgets.
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/vpadding"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func main() {

	f := examples.RedirectLogger("widgets6.log")
	defer f.Close()

	palette := gowid.Palette{
		"body":  gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorCyan),
		"fbody": gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
	}

	edits := make([]gowid.IWidget, 0)

	for i := 0; i < 5; i++ {
		w11 := edit.New(edit.Options{Caption: fmt.Sprintf("Cap%d:", i+1), Text: "abcde"})
		w1 := styled.NewExt(w11, gowid.MakePaletteRef("body"), gowid.MakePaletteRef("fbody"))
		w22 := checkbox.New(false)
		w2 := styled.NewExt(w22, gowid.MakePaletteRef("body"), gowid.MakePaletteRef("fbody"))
		colwids := make([]gowid.IContainerWidget, 0)
		colwids = append(colwids, &gowid.ContainerWidget{w1, gowid.RenderWithWeight{50}})
		colwids = append(colwids, &gowid.ContainerWidget{w2, gowid.RenderFixed{}})
		cols1 := columns.New(colwids)
		edits = append(edits, cols1)
	}

	walker := list.NewSimpleListWalker(edits)
	lbox := list.New(walker)
	lbox2 := vpadding.NewBox(lbox, 5)
	fr := framed.New(lbox2)

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    fr,
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
