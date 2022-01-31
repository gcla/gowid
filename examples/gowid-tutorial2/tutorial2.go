// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// The second example from the gowid tutorial.
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

var txt *text.Widget

func unhandled(app gowid.IApp, ev interface{}) bool {
	if evk, ok := ev.(*tcell.EventKey); ok {
		switch evk.Rune() {
		case 'q', 'Q':
			app.Quit()
		default:
			txt.SetText(fmt.Sprintf("hello world - %c", evk.Rune()), app)
		}
	}
	return true
}

func main() {
	txt = text.New("hello world")
	app, err := gowid.NewApp(gowid.AppArgs{View: txt})
	examples.ExitOnErr(err)
	app.MainLoop(gowid.UnhandledInputFunc(unhandled))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
