// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// An example of the gowid asciigraph widget which relies upon the
// asciigraph package at github.com/guptarohit/asciigraph.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/asciigraph"
	asc "github.com/guptarohit/asciigraph"
	log "github.com/sirupsen/logrus"
)

//======================================================================

var app *gowid.App

//======================================================================

func main() {
	var err error

	f := examples.RedirectLogger("asciigraph.log")
	defer f.Close()

	palette := gowid.Palette{}

	data := []float64{2, 1, 1, 2, -2, 5, 7, 11, 3, 7, 1}

	graph := asciigraph.New(data, []asc.Option{})

	app, err = gowid.NewApp(gowid.AppArgs{
		View:    graph,
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
