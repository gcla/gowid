// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// The sixth example from the gowid tutorial.
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gdamore/tcell"
)

//======================================================================

func question() *pile.Widget {
	return pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			IWidget: edit.New(edit.Options{Caption: "What is your name?\n"}),
			D:       gowid.RenderFlow{},
		},
	})
}

func answer(name string) *gowid.ContainerWidget {
	return &gowid.ContainerWidget{
		IWidget: text.New(fmt.Sprintf("Nice to meet you, %s", name)),
		D:       gowid.RenderFlow{},
	}
}

type ConversationWidget struct {
	*list.Widget
}

func NewConversationWidget() *ConversationWidget {
	widgets := make([]gowid.IWidget, 1)
	widgets[0] = question()
	lb := list.New(list.NewSimpleListWalker(widgets))
	return &ConversationWidget{lb}
}

func (w *ConversationWidget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	res := false
	if evk, ok := ev.(*tcell.EventKey); ok && evk.Key() == tcell.KeyEnter {
		res = true
		focus := w.Walker().Focus()
		curw := w.Walker().At(focus)
		focusPile := curw.(*pile.Widget)
		pileSubWidgets := focusPile.SubWidgets()
		ed := pileSubWidgets[0].(*gowid.ContainerWidget).SubWidget().(*edit.Widget)
		focusPile.SetSubWidgets(append(pileSubWidgets[0:1], answer(ed.Text())), app)
		walker := w.Widget.Walker().(*list.SimpleListWalker)
		walker.Widgets = append(walker.Widgets, question())
		nextPos := walker.Next(focus)
		walker.SetFocus(nextPos, app)
		w.Widget.GoToBottom(app)
	} else {
		res = w.Widget.UserInput(ev, size, focus, app)
	}
	return res
}

func main() {
	app, err := gowid.NewApp(gowid.AppArgs{View: NewConversationWidget()})
	examples.ExitOnErr(err)
	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
