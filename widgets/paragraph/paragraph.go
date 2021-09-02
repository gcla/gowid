// Copyright 2021 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package paragraph provides a simple text widget that neatly splits
// over multiple lines.
package paragraph

import (
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
)

//======================================================================

// Widget can be used to display text on the screen, with words broken
// cleanly as the output width changes.
type Widget struct {
	text     string
	words    []string
	lastcols int
	lastrows int
	gowid.RejectUserInput
	gowid.NotSelectable
}

var _ gowid.IWidget = (*Widget)(nil)

func New(text string) *Widget {
	return &Widget{
		text:  text,
		words: strings.Fields(text),
	}
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	switch size := size.(type) {
	case gowid.IRenderFlowWith:
		res := gowid.NewCanvas()
		var lc gowid.ICanvas
		var pos int
		var curWord string
		i := 0
		space := false
		for {
			if i >= len(w.words) {
				if lc != nil {
					res.AppendBelow(lc, false, false)
				}
				break
			}
			if curWord == "" {
				curWord = w.words[i]
			}
			if lc == nil {
				lc = gowid.NewCanvasOfSize(size.FlowColumns(), 1)
				pos = 0
			}
			if space {
				if pos < size.FlowColumns()-1 {
					pos++
				} else {
					if lc != nil {
						res.AppendBelow(lc, false, false)
						lc = gowid.NewCanvasOfSize(size.FlowColumns(), 1)
						pos = 0
					}
				}
				space = false
			} else {
				// No space left in current line
				if len(curWord) > size.FlowColumns()-pos {
					// No space on this line, but it will fit on next line. If it
					// doesn't even fit on a line by itself, well just split it
					if pos > 0 && lc != nil && len(curWord) <= size.FlowColumns() {
						res.AppendBelow(lc, false, false)
						lc = gowid.NewCanvasOfSize(size.FlowColumns(), 1)
						pos = 0
					}
					take := gwutil.Min(size.FlowColumns()-pos, len(curWord))
					for j := 0; j < take; j++ {
						lc.SetCellAt(pos, 0, gowid.CellFromRune(rune(curWord[j])))
						pos++
					}
					if pos >= size.FlowColumns() {
						res.AppendBelow(lc, false, false)
						lc = nil
					}
					if take == len(curWord) {
						i++
						curWord = ""
						space = true
					} else {
						// Must be less
						curWord = curWord[take:]
					}
				} else {
					for j := 0; j < len(curWord); j++ {
						lc.SetCellAt(pos, 0, gowid.CellFromRune(rune(curWord[j])))
						pos++
					}
					i++
					curWord = ""
					space = true
				}
			}

		}
		return res
	default:
		panic(gowid.WidgetSizeError{Widget: w, Size: size, Required: "gowid.IRenderFlow"})
	}
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	if w.lastcols != 0 {
		return gowid.RenderBox{C: w.lastcols, R: w.lastrows}
	}

	res := gowid.CalculateRenderSizeFallback(w, size, focus, app)
	w.lastcols = res.C
	w.lastrows = res.R
	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
