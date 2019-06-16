# Terminal User Interface Widgets in Go

Gowid provides widgets and a framework for making terminal user interfaces. It's written in Go and inspired by [urwid](http://urwid.org). 

Widgets out-of-the-box include:

 - input components like button, checkbox and an editable text field with support for passwords
 - layout components for arranging widgets in columns, rows and a grid
 - structured components - a tree, an infinite list and a table
 - pre-canned widgets - a progress bar, a modal dialog, a bar graph and a menu
 - a VT220-compatible terminal widget, heavily cribbed from urwid :smiley:

All widgets support interaction with the mouse when the terminal allows.

Gowid is built on top of the fantastic [tcell](https://github.com/gdamore/tcell) package.

There are many alternatives to gowid - see [Similar Projects](#similar-projects)

The most developed gowid application is currently [termshark](https://termshark.io), a terminal UI for tshark.
## Installation

```bash
go get github.com/gcla/gowid/...
```
## Examples

Make sure `$GOPATH/bin` is in your PATH (or `~/go/bin` if `GOPATH` isn't set), then tab complete "gowid-" e.g.

```bash
gowid-fib
```

Here is a port of urwid's [palette](https://github.com/urwid/urwid/blob/master/examples/palette_test.py) example:

<a href="https://drive.google.com/uc?export=view&id=1wENPAEOOdPp6eeHvpH0TvYOYnl4Gmy9Q"><img src="https://drive.google.com/uc?export=view&id=1wENPAEOOdPp6eeHvpH0TvYOYnl4Gmy9Q" style="width: 50px; max-width: 10%; height: auto" title="Click for the larger version." /></a>

Here is urwid's [graph](https://github.com/urwid/urwid/blob/master/examples/graph.py) example:

<a href="https://drive.google.com/uc?export=view&id=16p1NFrc3X3ReD-wz7bPXeYF8pCap3U-y"><img src="https://drive.google.com/uc?export=view&id=16p1NFrc3X3ReD-wz7bPXeYF8pCap3U-y" style="width: 50px; max-width: 10%; height: auto" title="Click for the larger version." /></a>

And urwid's [fibonacci](https://github.com/urwid/urwid/blob/master/examples/fib.py) example:

<a href="https://drive.google.com/uc?export=view&id=1fPVYOWt7EMUP18ZQL78OFY7IXwmeeqUO"><img src="https://drive.google.com/uc?export=view&id=1fPVYOWt7EMUP18ZQL78OFY7IXwmeeqUO" style="width: 500px; max-width: 100%; height: auto" title="Click for the larger version." /></a>

A demonstration of gowid's terminal widget, a port of urwid's [terminal widget](https://github.com/urwid/urwid/blob/master/examples/terminal.py):

<a href="https://drive.google.com/uc?export=view&id=1bRtgHoXcy0UESmKZK6JID8FIlkf5T7aL"><img src="https://drive.google.com/uc?export=view&id=1bRtgHoXcy0UESmKZK6JID8FIlkf5T7aL" style="width: 500px; max-width: 100%; height: auto" title="Click for the larger version." /></a>

Finally, here is an animation of termshark in action:

<a href="https://drive.google.com/uc?export=view&id=1vDecxjqwJrtMGJjOObL-LLvi-1pBVByt"><img src="https://drive.google.com/uc?export=view&id=1vDecxjqwJrtMGJjOObL-LLvi-1pBVByt" style="width: 500px; max-width: 100%; height: auto" title="Click for the larger version." /></a>

## Hello World

This example is an attempt to mimic urwid's ["Hello World"](http://urwid.org/tutorial/index.html) example.

```go
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/widgets/divider"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vpadding"
)

//======================================================================

func main() {

	palette := gowid.Palette{
		"banner":  gowid.MakePaletteEntry(gowid.ColorWhite, gowid.MakeRGBColor("#60d")),
		"streak":  gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#60a")),
		"inside":  gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#808")),
		"outside": gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#a06")),
		"bg":      gowid.MakePaletteEntry(gowid.ColorNone, gowid.MakeRGBColor("#d06")),
	}

	div := divider.NewBlank()
	outside := styled.New(div, gowid.MakePaletteRef("outside"))
	inside := styled.New(div, gowid.MakePaletteRef("inside"))

	helloworld := styled.New(
		text.NewFromContentExt(
			text.NewContent([]text.ContentSegment{
				text.StyledContent("Hello World", gowid.MakePaletteRef("banner")),
			}),
			text.Options{
				Align: gowid.HAlignMiddle{},
			},
		),
		gowid.MakePaletteRef("streak"),
	)

	f := gowid.RenderFlow{}

	view := styled.New(
		vpadding.New(
			pile.New([]gowid.IContainerWidget{
				&gowid.ContainerWidget{IWidget: outside, D: f},
				&gowid.ContainerWidget{IWidget: inside, D: f},
				&gowid.ContainerWidget{IWidget: helloworld, D: f},
				&gowid.ContainerWidget{IWidget: inside, D: f},
				&gowid.ContainerWidget{IWidget: outside, D: f},
			}),
			gowid.VAlignMiddle{},
			f),
		gowid.MakePaletteRef("bg"),
	)

	app, _ := gowid.NewApp(gowid.AppArgs{
		View:    view,
		Palette: &palette,
	})
    
	app.SimpleMainLoop()
}
```

Running the example above displays this:

<a href="https://drive.google.com/uc?export=view&id=1P2kjWagHJmhtWLV0hPQti0fXKidr_WMB"><img src="https://drive.google.com/uc?export=view&id=1P2kjWagHJmhtWLV0hPQti0fXKidr_WMB" style="width: 500px; max-width: 100%; height: auto" title="Click for the larger version." /></a>

## Documentation

 - The beginnings of a [tutorial](docs/Tutorial.md)
 - A list of most of the [widgets](docs/Widgets.md)
 - Some [FAQs](docs/FAQ.md) (which I guessed at...)
 
## Similar Projects

Gowid is late to the TUI party. There are many options from which to choose - please read https://appliedgo.net/tui/ for a nice summary for the Go language. Here is a selection:

 - [urwid](http://urwid.org/) - one of the oldest, for those working in python
 - [tview](https://github.com/rivo/tview) - active, polished, concise, lots of widgets, Go
 - [termui](https://github.com/gizak/termui) - focus on graphing and dataviz, Go
 - [gocui](https://github.com/jroimartin/gocui) - focus on layout, good input options, mouse support, Go
 - [clui](https://github.com/VladimirMarkelov/clui) - active, many widgets, mouse support, Go
 - [tui-go](https://github.com/marcusolsson/tui-go) - QT-inspired, experimental, nice examples, Go

## Dependencies

Gowid depends on these great open-source packages:

- [urwid](http://urwid.org) - not a Go-dependency, but the model for most of gowid's design
- [tcell](https://github.com/gdamore/tcell) - a cell based view for text terminals, like xterm, inspired by termbox
- [asciigraph](https://github.com/guptarohit/asciigraph) - lightweight ASCII line-graphs for Go
- [logrus](https://github.com/sirupsen/logrus) - structured pluggable logging for Go
- [testify](github.com/stretchr/testify) - tools for testifying that your code will behave as you intend

## Contact

- The author - Graham Clark (grclark@gmail.com)

## License

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

