// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A gowid test app which exercises the checkbox, columns and hpadding widgets.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/hpadding"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func main() {

	f := examples.RedirectLogger("widgets8.log")
	defer f.Close()

	palette := gowid.Palette{}

	fixed := gowid.RenderFixed{}
	cb1 := checkbox.New(false)
	cp1 := hpadding.New(cb1, gowid.HAlignLeft{}, fixed)

	cb2 := checkbox.New(false)
	cp2 := hpadding.New(cb2, gowid.HAlignLeft{}, fixed)

	view := columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{cp1, gowid.RenderWithWeight{10}},
		&gowid.ContainerWidget{cp2, gowid.RenderWithWeight{10}},
	})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    view,
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
