// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// A demonstration of gowid's tree widget.
package main

import (
	"fmt"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/palettemap"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/tree"
	tcell "github.com/gdamore/tcell/v2"
	log "github.com/sirupsen/logrus"
)

var pos *tree.TreePos
var tb *list.Widget
var parent1 *tree.Collapsible
var walker *tree.TreeWalker

//======================================================================

func MakeDemoDecoration(pos tree.IPos, tr tree.IModel, wmaker tree.IWidgetMaker) gowid.IWidget {
	var res gowid.IWidget
	level := -1
	for cur := pos; cur != nil; cur = tree.ParentPosition(cur) {
		level += 1
	}
	pad := gwutil.StringOfLength(' ', level*10)
	cwidgets := make([]gowid.IContainerWidget, 0)
	cwidgets = append(cwidgets, &gowid.ContainerWidget{text.New(pad), gowid.RenderWithUnits{U: len(pad)}})
	if ct, ok := tr.(tree.ICollapsible); ok {
		var bn *button.Widget
		if ct.IsCollapsed() {
			bn = button.New(text.New("+"))
		} else {
			bn = button.New(text.New("-"))
		}

		// If I use one button with conditional logic in the callback, rather than make
		// a separate button depending on whether or not the tree is collapsed, it will
		// correctly work when the DecoratorMaker is caching the widgets i.e. it will
		// collapse or expand even when the widget is rendered from the cache
		bn.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
			// Run this outside current event loop because we are implicitly
			// adjusting the data structure behind the list walker, and it's
			// not prepared to handle that in the same pass of processing
			// UserInput. TODO.
			app.Run(gowid.RunFunction(func(app gowid.IApp) {
				ct.SetCollapsed(app, !ct.IsCollapsed())
			}))
		}})

		cwidgets = append(cwidgets, &gowid.ContainerWidget{
			framed.NewUnicode(
				styled.NewExt(
					bn,
					gowid.MakePaletteRef("body"),
					gowid.MakePaletteRef("fbody"),
				),
			),
			gowid.RenderFixed{},
		})
	}
	inner := wmaker.MakeWidget(pos, tr)
	cwidgets = append(cwidgets, &gowid.ContainerWidget{inner, gowid.RenderFixed{}})

	res = palettemap.New(
		columns.New(cwidgets),
		palettemap.Map{"body": "fbody"},
		palettemap.Map{},
	)

	return res
}

func MakeDemoWidget(pos tree.IPos, tr tree.IModel) gowid.IWidget {
	var res gowid.IWidget

	cbox := checkbox.New(false)
	cbox.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		log.Info("Clicked checkbox in tree")
	}})

	res = columns.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			framed.NewUnicode(cbox),
			gowid.RenderFixed{},
		},
		&gowid.ContainerWidget{
			styled.NewExt(
				framed.NewUnicode(
					selectable.New(
						text.NewFromContent(
							text.NewContent(
								[]text.ContentSegment{
									text.StyledContent(
										fmt.Sprintf("tr %s:%v", tr.Leaf(), pos.String()),
										gowid.MakePaletteRef("body"),
									),
								},
							),
						),
					),
				),
				gowid.MakePaletteRef("body"),
				gowid.MakePaletteRef("fbody"),
			),
			gowid.RenderFixed{},
		},
	})

	return res
}

//======================================================================

type handler struct{}

func (h handler) UnhandledInput(app gowid.IApp, ev interface{}) bool {
	handled := false
	if evk, ok := ev.(*tcell.EventKey); ok {
		handled = true
		if evk.Key() == tcell.KeyCtrlC || evk.Rune() == 'q' || evk.Rune() == 'Q' {
			app.Quit()
		} else if evk.Rune() == 'x' {
			f := walker.Focus()
			f2 := f.(tree.IPos)
			s := f2.GetSubStructure(parent1)
			if t2, ok := s.(tree.ICollapsible); ok {
				t2.SetCollapsed(app, true)
			}
		} else if evk.Rune() == 'z' {
			f := walker.Focus()
			f2 := f.(tree.IPos)
			s := f2.GetSubStructure(parent1)
			if t2, ok := s.(tree.ICollapsible); ok {
				t2.SetCollapsed(app, false)
			}
		} else {
			handled = false
		}
	}
	return handled
}

//======================================================================

func main() {

	f := examples.RedirectLogger("tree1.log")
	defer f.Close()

	palette := gowid.Palette{
		"title": gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
		"key":   gowid.MakePaletteEntry(gowid.ColorCyan, gowid.ColorBlack),
		"foot":  gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
		"body":  gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorCyan),
		"fbody": gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
	}

	body := gowid.MakePaletteRef("body")

	leaf1 := tree.NewTree("leaf1", []tree.IModel{})
	leaf2 := tree.NewTree("leaf2", []tree.IModel{})
	leaf3 := tree.NewTree("leaf3", []tree.IModel{})
	leaf4 := tree.NewTree("leaf4", []tree.IModel{})
	leaf5 := tree.NewTree("leaf5", []tree.IModel{})
	leaf21 := tree.NewTree("leaf21", []tree.IModel{})
	leaf22 := tree.NewTree("leaf22", []tree.IModel{})
	leaf23 := tree.NewTree("leaf23", []tree.IModel{})
	stree1 := tree.NewCollapsible("stree1", []tree.IModel{leaf4, leaf5})
	stree2 := tree.NewCollapsible("stree2", []tree.IModel{leaf21, leaf22, leaf23})
	parent1 = tree.NewCollapsible("parent1", []tree.IModel{leaf1, stree1, leaf2, stree2, leaf3})

	parent1.AddOnExpanded("exp", tree.ExpandedFunction(func(app gowid.IApp) {
		ch := parent1.GetChildren()
		newLeaf := tree.NewTree("foo", []tree.IModel{})
		parent1.SetChildren(append([]tree.IModel{newLeaf}, ch...))
	}))

	pos = tree.NewPos()
	walker = tree.NewWalker(parent1, pos,
		tree.NewCachingMaker(tree.WidgetMakerFunction(MakeDemoWidget)),
		tree.NewCachingDecorator(tree.DecoratorFunction(MakeDemoDecoration)))
	tb = tree.New(walker)
	tb.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		log.Infof("Focus changed - widget is now %v", w)
	}})
	view := styled.New(tb, body)

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    view,
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
