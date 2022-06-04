// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A port of urwid's fib.py example using gowid widgets.
package main

import (
	"fmt"
	"math/big"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/hpadding"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/palettemap"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	log "github.com/sirupsen/logrus"
)

//======================================================================

type FibWalker struct {
	a, b big.Int
}

var _ list.IWalker = (*FibWalker)(nil)
var _ list.IWalkerPosition = FibWalker{}
var _ list.IWalkerHome = FibWalker{}

func (f FibWalker) First() list.IWalkerPosition {
	return FibWalker{*big.NewInt(0), *big.NewInt(1)}
}

func (f FibWalker) Equal(other list.IWalkerPosition) bool {
	switch o := other.(type) {
	case FibWalker:
		return (f.a.Cmp(&o.a) == 0) && (f.b.Cmp(&o.b) == 0)
	default:
		return false
	}
}

func (f FibWalker) GreaterThan(other list.IWalkerPosition) bool {
	switch o := other.(type) {
	case FibWalker:
		return (f.b.Cmp(&o.b) > 0)
	default:
		panic(fmt.Errorf("Invalid type to compare against FibWalker - %T", other))
	}
}

func getWidget(f FibWalker) gowid.IWidget {
	var res gowid.IWidget
	res = selectable.New(
		palettemap.New(
			hpadding.New(
				text.NewFromContent(
					text.NewContent([]text.ContentSegment{
						text.StyledContent(f.b.Text(10), gowid.MakePaletteRef("body")),
					}),
				),
				gowid.HAlignRight{}, gowid.RenderFlow{},
			),
			palettemap.Map{"body": "fbody"},
			palettemap.Map{},
		),
	)
	return res
}

func (f *FibWalker) At(pos list.IWalkerPosition) gowid.IWidget {
	f2 := pos.(FibWalker)
	return getWidget(f2)
}

func (f *FibWalker) Focus() list.IWalkerPosition {
	return *f
}

func (f *FibWalker) SetFocus(pos list.IWalkerPosition, app gowid.IApp) {
	*f = pos.(FibWalker)
}

func (f *FibWalker) Next(pos list.IWalkerPosition) list.IWalkerPosition {
	fc := pos.(FibWalker)
	var sum big.Int
	sum.Add(&fc.a, &fc.b)
	fn := FibWalker{fc.b, sum}
	return fn
}

func (f *FibWalker) Previous(pos list.IWalkerPosition) list.IWalkerPosition {
	fc := pos.(FibWalker)
	var diff big.Int
	diff.Sub(&fc.b, &fc.a)
	fn := FibWalker{diff, fc.a}
	return fn
}

//======================================================================

func main() {

	f := examples.RedirectLogger("fib.log")
	defer f.Close()

	palette := gowid.Palette{
		"title": gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
		"key":   gowid.MakePaletteEntry(gowid.ColorCyan, gowid.ColorBlack),
		"foot":  gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
		"body":  gowid.MakePaletteEntry(gowid.ColorBlack, gowid.ColorCyan),
		"fbody": gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
	}

	key := gowid.MakePaletteRef("key")
	foot := gowid.MakePaletteRef("foot")
	title := gowid.MakePaletteRef("title")
	body := gowid.MakePaletteRef("body")

	footerContent := []text.ContentSegment{
		text.StyledContent("Fibonacci Set Viewer", title),
		text.StringContent("    "),
		text.StyledContent("UP", key),
		text.StringContent(", "),
		text.StyledContent("DOWN", key),
		text.StringContent(", "),
		text.StyledContent("PAGE_UP", key),
		text.StringContent(", "),
		text.StyledContent("PAGE_DOWN", key),
		text.StringContent(", "),
		text.StyledContent("HOME", key),
		text.StringContent(", "),
		text.StyledContent("CTRL-L", key),
		text.StringContent(" move view  "),
		text.StyledContent("Q", key),
		text.StringContent(" exits. Try the mouse wheel."),
	}

	footerText := styled.New(text.NewFromContent(text.NewContent(footerContent)), foot)

	walker := FibWalker{*big.NewInt(0), *big.NewInt(1)}
	lb := list.New(&walker)
	styledLb := styled.New(lb, body)

	lb.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		log.Infof("Focus changed - widget is now %p", w)
	}})

	view := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			IWidget: styledLb,
			D:       gowid.RenderWithWeight{W: 1},
		},
		&gowid.ContainerWidget{
			IWidget: footerText,
			D:       gowid.RenderFlow{},
		},
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
