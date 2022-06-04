// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// The fifth example from the gowid tutorial.
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
)

//======================================================================

func main() {

	ask := edit.New(edit.Options{Caption: "What is your name?\n"})
	reply := text.New("")
	btn := button.New(text.New("Exit"))
	sbtn := styled.New(btn, gowid.MakeStyledAs(gowid.StyleReverse))
	div := divider.NewBlank()

	btn.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		app.Quit()
	}})

	ask.OnTextSet(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		if ask.Text() == "" {
			reply.SetText("", app)
		} else {
			reply.SetText(fmt.Sprintf("Nice to meet you, %s", ask.Text()), app)
		}
	}})

	f := gowid.RenderFlow{}

	view := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{IWidget: ask, D: f},
		&gowid.ContainerWidget{IWidget: div, D: f},
		&gowid.ContainerWidget{IWidget: reply, D: f},
		&gowid.ContainerWidget{IWidget: div, D: f},
		&gowid.ContainerWidget{IWidget: sbtn, D: f},
	})

	app, err := gowid.NewApp(gowid.AppArgs{View: view})
	examples.ExitOnErr(err)

	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
