// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A simple gowid directory browser.
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/boxadapter"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/tree"
	"github.com/gcla/tcell"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var pos *tree.TreePos
var tb *list.Widget
var parent1 *DirTree
var walker *tree.TreeWalker
var dirname = kingpin.Arg("dir", "Directory to scan.").Required().String()

var expandedCache map[string]int

//======================================================================

func IsDir(name string) bool {
	res := false
	if f, err := os.OpenFile(name, syscall.O_NONBLOCK|os.O_RDONLY, 0); err == nil {
		defer f.Close()
		if st, err2 := f.Stat(); err2 == nil {
			if st.IsDir() {
				res = true
			}
		}
	}
	return res
}

func NodeState(name string, cache map[string]int) int {
	state := notDir
	if lookup, ok := cache[name]; ok {
		state = lookup
	} else {
		if IsDir(name) {
			state = collapsed
		} else {
			state = notDir
		}
		cache[name] = state
	}
	return state
}

//======================================================================

type NoIterator struct{}

func (d *NoIterator) Next() bool {
	return false
}

func (d *NoIterator) Value() tree.IModel {
	panic("Do not call")
}

//======================================================================

const (
	notDir = iota
	collapsed
	expanded
)

type DirIterator struct {
	init   bool
	parent tree.ICollapsible
	files  []os.FileInfo
	pos    int
	cache  *map[string]int // shared - track whether any path is collapsed or expanded
}

func (d *DirIterator) FullPath() string {
	return d.parent.(*DirTree).FullPath() + "/" + d.files[d.pos].Name()
}

func (d *DirIterator) Next() bool {
	if !d.init {
		// Ignore the error because if there's a problem, we'll simply have files thus no children
		d.files, _ = ioutil.ReadDir(d.parent.(*DirTree).FullPath())
		d.init = true
	}
	d.pos++
	return !d.parent.IsCollapsed() && d.pos < len(d.files)
}

func (d *DirIterator) Value() tree.IModel {
	var res tree.IModel
	state := NodeState(d.FullPath(), *d.cache)
	switch state {
	case notDir:
		res = &DirTree{
			name:   d.files[d.pos].Name(),
			cache:  d.cache,
			parent: d.parent,
			notDir: true,
		}
	default:
		res = &DirTree{
			name:   d.files[d.pos].Name(),
			cache:  d.cache,
			parent: d.parent,
		}
	}

	return res
}

//======================================================================

type DirTree struct {
	parent tree.IModel
	name   string
	iter   *DirIterator
	notDir bool
	cache  *map[string]int // shared - track whether any path is collapsed or expanded
}

var _ tree.ICollapsible = (*DirTree)(nil)

func (d *DirTree) FullPath() string {
	if d.parent == nil {
		return d.name
	} else {
		return d.parent.(*DirTree).FullPath() + "/" + d.name
	}
}

func (d *DirTree) Leaf() string {
	return d.name
}

func (d *DirTree) String() string {
	return d.name
}

func (d *DirTree) Children() tree.IIterator {
	var res tree.IIterator
	res = &NoIterator{}
	if d.iter != nil {
		d.iter.pos = -1
		res = d.iter
	} else {
		state := NodeState(d.FullPath(), *d.cache)
		if state != notDir {
			d.iter = &DirIterator{
				init:   false,
				pos:    -1,
				cache:  d.cache,
				parent: d,
			}
			res = d.iter
		}
	}
	return res
}

func (d *DirTree) IsCollapsed() bool {
	fp := d.FullPath()
	if v, res := (*d.cache)[fp]; res {
		return (v == collapsed)
	} else {
		return true
	}
}

func (d *DirTree) SetCollapsed(app gowid.IApp, isCollapsed bool) {
	fp := d.FullPath()
	if isCollapsed {
		(*d.cache)[fp] = collapsed
	} else {
		(*d.cache)[fp] = expanded
	}
}

//======================================================================

// DirButton is a button widget that provides its own ID function, rather than relying on the default
// address of an embedded struct. This is in case the button widget isn't the exact same object when
// the mouse is clicked and when the mouse is released. Instead, the pathname is used as the ID, ensuring
// that if the user clicks on "/tmp/foo" then releases the mouse button on "/tmp/foo", the button's
// Click() method will be called. We need to provide a wrapper for UserInput(), else UserInput() will come
// from the embedded *button.Widget, which will then result in the IButtonWidget interface being
// built from *button.Widget and not DirButton.
//
type DirButton struct {
	*button.Widget
	id string
}

func (d *DirButton) ID() interface{} {
	return d.id
}

func (d *DirButton) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return button.UserInput(d, ev, size, focus, app)
}

//======================================================================

func MakeDecoration(pos tree.IPos, tr tree.IModel, wmaker tree.IWidgetMaker) gowid.IWidget {
	var res gowid.IWidget
	level := -1
	for cur := pos; cur != nil; cur = tree.ParentPosition(cur) {
		level += 1
	}
	pad := gwutil.StringOfLength(' ', level*3)
	lineWidgets := make([]gowid.IContainerWidget, 0)
	// indentation
	lineWidgets = append(lineWidgets, &gowid.ContainerWidget{
		IWidget: text.New(pad),
		D:       gowid.RenderWithUnits{U: len(pad)},
	})
	if ctree2, ok := tr.(tree.ICollapsible); ok {
		ctree := ctree2.(*DirTree)
		if !ctree.notDir {
			btnChr := gwutil.If(ctree.IsCollapsed(), "+", "-").(string)
			dirButton := &DirButton{button.New(text.New(btnChr)), pos.String()}

			// If I use one button with conditional logic in the callback, rather than make
			// a separate button depending on whether or not the tree is collapsed, it will
			// correctly work when the DecoratorMaker is caching the widgets i.e. it will
			// collapse or expand even when the widget is rendered from the cache
			dirButton.OnClick(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
				// Note that I don't change the button widget itself ([+]/[-]) - just the underlying model, from which
				// the widget will be recreated
				app.Run(gowid.RunFunction(func(app gowid.IApp) {
					ctree.SetCollapsed(app, !ctree.IsCollapsed())
				}))
			}})

			coloredDirButton := styled.NewExt(dirButton, gowid.MakePaletteRef("body"), gowid.MakePaletteRef("selected"))
			// [+] / [-] if the widget is a directory
			lineWidgets = append(lineWidgets, &gowid.ContainerWidget{
				IWidget: coloredDirButton,
				D:       gowid.RenderFixed{},
			})
			lineWidgets = append(lineWidgets, &gowid.ContainerWidget{
				IWidget: text.New(" "),
				D:       gowid.RenderFixed{},
			})
		}
	}

	inner := wmaker.MakeWidget(pos, tr)
	// filename/dirname
	lineWidgets = append(lineWidgets, &gowid.ContainerWidget{
		IWidget: inner,
		D:       gowid.RenderFixed{},
	})

	line := columns.New(lineWidgets)
	res = line
	res = boxadapter.New(res, 1) // force the widget to be on one line

	return res
}

func MakeWidget(pos tree.IPos, tr tree.IModel) gowid.IWidget {
	var res gowid.IWidget

	pr := "body"
	ctree := tr.(*DirTree)
	if !ctree.notDir {
		pr = "dirbody"
	}

	cwidgets := make([]gowid.IContainerWidget, 1)

	cwidgets[0] = &gowid.ContainerWidget{
		IWidget: styled.NewExt(
			selectable.New(
				styled.NewExt(
					text.New(
						tr.Leaf(),
					),
					gowid.MakePaletteRef(pr), gowid.MakePaletteRef("selected"),
				),
			),
			gowid.MakePaletteRef("body"), gowid.MakePaletteRef("selected"),
		),
		D: gowid.RenderFixed{},
	}
	res = columns.New(cwidgets)

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
			pos := walker.Focus()
			tpos := pos.(tree.IPos)
			itree := tpos.GetSubStructure(parent1)
			if ctree, ok := itree.(tree.ICollapsible); ok {
				ctree.SetCollapsed(app, true)
			}
		} else if evk.Rune() == 'z' {
			pos := walker.Focus()
			tpos := pos.(tree.IPos)
			itree := tpos.GetSubStructure(parent1)
			if ctree, ok := itree.(tree.ICollapsible); ok {
				ctree.SetCollapsed(app, false)
			}
		} else {
			handled = false
		}
	}
	return handled
}

//======================================================================

func main() {

	kingpin.Parse()

	if _, err := os.Stat(*dirname); os.IsNotExist(err) {
		fmt.Printf("Directory \"%v\" does not exist.", *dirname)
		os.Exit(1)
	}

	f := examples.RedirectLogger("dir.log")
	defer f.Close()

	palette := gowid.Palette{
		"body":     gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
		"dirbody":  gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorCyan),
		"selected": gowid.MakePaletteEntry(gowid.ColorDefault, gowid.ColorWhite),
	}

	expandedCache := make(map[string]int)
	// We start expanded at the top level
	expandedCache[*dirname] = expanded

	body := gowid.MakePaletteRef("body")

	parent1 = &DirTree{
		name:  *dirname,
		cache: &expandedCache,
	}

	pos = tree.NewPos()

	walker = tree.NewWalker(parent1, pos,
		tree.NewCachingMaker(tree.WidgetMakerFunction(MakeWidget)),
		tree.NewCachingDecorator(tree.DecoratorFunction(MakeDecoration)))
	tb = tree.New(walker)
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
