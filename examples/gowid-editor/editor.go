// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A poor-man's editor using gowid widgets - shows dialog, edit and vscroll.
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"net/http"
	_ "net/http"
	_ "net/http/pprof"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/columns"
	"github.com/gcla/gowid/widgets/dialog"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/holder"
	"github.com/gcla/gowid/widgets/hpadding"
	"github.com/gcla/gowid/widgets/pile"
	"github.com/gcla/gowid/widgets/styled"
	"github.com/gcla/gowid/widgets/text"
	"github.com/gcla/gowid/widgets/vscroll"
	"github.com/gcla/tcell"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

//======================================================================

var editWidget *edit.Widget
var footerContent []text.ContentSegment
var footerText *styled.Widget
var yesno *dialog.Widget
var viewHolder *holder.Widget

var app *gowid.App
var wg sync.WaitGroup
var updates chan string
var filename = kingpin.Arg("file", "File to edit.").Required().String()

//======================================================================

func updateStatusBar(message string) {
	app.Run(gowid.RunFunction(func(app gowid.IApp) {
		footerContent[1].Text = message
		footerWidget2 := text.NewFromContent(text.NewContent(footerContent))
		footerText.SetSubWidget(footerWidget2, app)
	}))
}

//======================================================================

type handler struct{}

func (h handler) UnhandledInput(app gowid.IApp, ev interface{}) bool {
	handled := false
	if evk, ok := ev.(*tcell.EventKey); ok {
		if evk.Key() == tcell.KeyEsc {
			handled = true
			app.Quit()
		}
		switch evk.Key() {
		case tcell.KeyCtrlC:
			handled = true
			msg := text.New("Do you want to quit?")
			yesno = dialog.New(
				framed.NewSpace(hpadding.New(msg, gowid.HAlignMiddle{}, gowid.RenderFixed{})),
				dialog.Options{
					Buttons: dialog.OkCancel,
				},
			)
			yesno.Open(viewHolder, gowid.RenderWithRatio{R: 0.5}, app)

		case tcell.KeyCtrlF:
			handled = true
			fi, err := os.Open(*filename)
			if err != nil {
				fi, err = os.Create(*filename)
				if err != nil {
					updates <- fmt.Sprintf("(FAILED to create %s)", *filename)
				}
			}
			if err == nil {
				defer fi.Close()

				_, err = io.Copy(&edit.Writer{editWidget, app}, fi)
				if err != nil {
					updates <- fmt.Sprintf("(FAILED to load %s)", *filename)
				} else {
					updates <- "(OPENED)"
				}
			}
		case tcell.KeyCtrlS:
			handled = true
			fo, err := os.Create(*filename)
			if err != nil {
				updates <- fmt.Sprintf("(FAILED to create %s)", *filename)
			} else {
				defer fo.Close()

				_, err = io.Copy(fo, editWidget)
				if err != nil {
					updates <- fmt.Sprintf("(FAILED to write %s)", *filename)
				} else {
					updates <- "(SAVED!)"
				}
			}
		}
	}
	return handled
}

//======================================================================

type EditWithScrollbar struct {
	*columns.Widget
	e        *edit.Widget
	sb       *vscroll.Widget
	goUpDown int // positive means down
	pgUpDown int // positive means down
}

func NewEditWithScrollbar(e *edit.Widget) *EditWithScrollbar {
	sb := vscroll.NewExt(vscroll.VerticalScrollbarUnicodeRunes)
	res := &EditWithScrollbar{
		columns.New([]gowid.IContainerWidget{
			&gowid.ContainerWidget{e, gowid.RenderWithWeight{W: 1}},
			&gowid.ContainerWidget{sb, gowid.RenderWithUnits{U: 1}},
		}),
		e, sb, 0, 0,
	}
	sb.OnClickAbove(gowid.WidgetCallback{"cb", res.clickUp})
	sb.OnClickBelow(gowid.WidgetCallback{"cb", res.clickDown})
	sb.OnClickUpArrow(gowid.WidgetCallback{"cb", res.clickUpArrow})
	sb.OnClickDownArrow(gowid.WidgetCallback{"cb", res.clickDownArrow})
	return res
}

func (e *EditWithScrollbar) clickUp(app gowid.IApp, w gowid.IWidget) {
	e.pgUpDown -= 1
}

func (e *EditWithScrollbar) clickDown(app gowid.IApp, w gowid.IWidget) {
	e.pgUpDown += 1
}

func (e *EditWithScrollbar) clickUpArrow(app gowid.IApp, w gowid.IWidget) {
	e.goUpDown -= 1
}

func (e *EditWithScrollbar) clickDownArrow(app gowid.IApp, w gowid.IWidget) {
	e.goUpDown += 1
}

// gcdoc - do this so columns navigation e.g. ctrl-f doesn't get passed to columns
func (w *EditWithScrollbar) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	// Stop these keys moving focus in the columns used by this widget. C-f is used to
	// open a file.
	if evk, ok := ev.(*tcell.EventKey); ok {
		switch evk.Key() {
		case tcell.KeyCtrlF, tcell.KeyCtrlB:
			return false
		}
	}

	box, _ := size.(gowid.IRenderBox)
	w.sb.Top, w.sb.Middle, w.sb.Bottom = w.e.CalculateTopMiddleBottom(gowid.MakeRenderBox(box.BoxColumns()-1, box.BoxRows()))

	res := w.Widget.UserInput(ev, size, focus, app)
	if res {
		w.Widget.SetFocus(app, 0)
	}
	return res
}

func (w *EditWithScrollbar) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	box, _ := size.(gowid.IRenderBox)
	ecols := box.BoxColumns() - 1
	ebox := gowid.MakeRenderBox(ecols, box.BoxRows())
	if w.goUpDown != 0 || w.pgUpDown != 0 {
		w.e.SetLinesFromTop(gwutil.Max(0, w.e.LinesFromTop()+w.goUpDown+(w.pgUpDown*box.BoxRows())), app)
		txt := w.e.MakeText()
		layout := text.MakeTextLayout(txt.Content(), ecols, txt.Wrap(), gowid.HAlignLeft{})
		_, y := text.GetCoordsFromCursorPos(w.e.CursorPos(), ecols, layout, w.e)
		if y < w.e.LinesFromTop() {
			for i := y; i < w.e.LinesFromTop(); i++ {
				w.e.DownLines(ebox, false, app)
			}
		} else if y >= w.e.LinesFromTop()+box.BoxRows() {
			for i := w.e.LinesFromTop() + box.BoxRows(); i <= y; i++ {
				w.e.UpLines(ebox, false, app)
			}
		}

	}
	w.goUpDown = 0
	w.pgUpDown = 0
	w.sb.Top, w.sb.Middle, w.sb.Bottom = w.e.CalculateTopMiddleBottom(ebox)

	canvas := gowid.Render(w.Widget, size, focus, app)

	return canvas
}

//======================================================================

func main() {
	var err error

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	kingpin.Parse()

	if _, err := os.Stat(*filename); os.IsNotExist(err) {
		fmt.Printf("Requested file \"%v\" does not exist.", *filename)
		os.Exit(1)
	}

	f := examples.RedirectLogger("editor.log")
	defer f.Close()

	palette := gowid.Palette{
		"mainpane": gowid.MakePaletteEntry(gowid.ColorBlack, gowid.NewUrwidColor("light gray")),
		"cyan":     gowid.MakePaletteEntry(gowid.ColorCyan, gowid.ColorBlack),
		"inv":      gowid.MakePaletteEntry(gowid.ColorWhite, gowid.ColorBlack),
	}

	editWidget = edit.New()

	footerContent = []text.ContentSegment{
		text.StyledContent("Gowid Text Editor. ", gowid.MakePaletteRef("inv")),
		text.StyledContent("", gowid.MakePaletteRef("cyan")), //		emptyContent,
		text.StyledContent(" C-c to exit. C-s to save. C-f to open test file.", gowid.MakePaletteRef("inv")),
	}

	footerText = styled.New(
		text.NewFromContent(
			text.NewContent(footerContent),
		),
		gowid.MakePaletteRef("inv"),
	)

	mainView := pile.New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{
			styled.New(
				framed.NewUnicode(
					NewEditWithScrollbar(editWidget),
				),
				gowid.MakePaletteRef("mainpane"),
			),
			gowid.RenderWithWeight{1},
		},
		&gowid.ContainerWidget{
			footerText,
			gowid.RenderFlow{},
		},
	})

	viewHolder = holder.New(mainView)

	updates = make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			status := <-updates
			if status == "done" {
				break
			} else if status == "clear" {
				updateStatusBar("")
			} else {
				updateStatusBar(status)
				go func() {
					timer := time.NewTimer(time.Second)
					<-timer.C
					updates <- "clear"
				}()
			}
		}
	}()

	logger := logrus.New()
	logger.Out = ioutil.Discard

	app, err = gowid.NewApp(gowid.AppArgs{
		View:    viewHolder,
		Palette: &palette,
		Log:     logger,
	})
	examples.ExitOnErr(err)

	app.MainLoop(handler{})

	updates <- "done"
	wg.Wait()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
