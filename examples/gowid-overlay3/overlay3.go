// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A demonstration of gowid's overlay and fill widgets.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/overlay"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
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
		} else {
			handled = false
		}
	}
	return handled
}

//======================================================================

func main() {

	f := examples.RedirectLogger("overlay1.log")
	defer f.Close()

	palette := gowid.Palette{
		"red": gowid.MakePaletteEntry(gowid.ColorDefault, gowid.ColorRed),
	}

	top := styled.New(
		framed.NewUnicode(
			text.New("hello"),
		),
		gowid.MakePaletteRef("red"),
	)
	bottom := fill.New(' ')

	ov = overlay.New(top, bottom,
		gowid.VAlignMiddle{}, gowid.RenderFixed{},
		gowid.HAlignMiddle{}, gowid.RenderFixed{},
	)

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
