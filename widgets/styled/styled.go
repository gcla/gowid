// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package styled provides a colored styled widget.
package styled

import (
	"fmt"

	"github.com/gcla/gowid"
)

//======================================================================

// TODO - make a constructor to keep these fields unexported
type AttributeRange struct {
	Start  int
	End    int
	Styler gowid.ICellStyler
}

type Widget struct {
	gowid.IWidget
	focusRange    []AttributeRange
	notFocusRange []AttributeRange
	options       Options
	*gowid.Callbacks
	gowid.SubWidgetCallbacks
}

type Options struct {
	OverWrite bool // If true, then apply the style over any style below; if false, style underneath takes precedence
}

// Very simple way to color an entire widget
func New(inner gowid.IWidget, styler gowid.ICellStyler, opts ...Options) *Widget {
	res := NewWithRanges(
		inner,
		[]AttributeRange{AttributeRange{0, -1, styler}},
		[]AttributeRange{AttributeRange{0, -1, styler}},
		opts...,
	)
	var _ gowid.ICompositeWidget = res
	return res
}

func NewInvertedFocus(inner gowid.IWidget, styler gowid.ICellStyler, opts ...Options) *Widget {
	return NewExt(inner, styler, gowid.ColorInverter{styler}, opts...)
}

func NewFocus(inner gowid.IWidget, styler gowid.ICellStyler, opts ...Options) *Widget {
	return NewExt(inner, nil, styler, opts...)
}

func NewNoFocus(inner gowid.IWidget, styler gowid.ICellStyler, opts ...Options) *Widget {
	return NewExt(inner, styler, nil, opts...)
}

func NewExt(inner gowid.IWidget, notFocusStyler, focusStyler gowid.ICellStyler, opts ...Options) *Widget {
	res := NewWithRanges(
		inner,
		[]AttributeRange{AttributeRange{0, -1, notFocusStyler}},
		[]AttributeRange{AttributeRange{0, -1, focusStyler}},
		opts...,
	)
	var _ gowid.ICompositeWidget = res
	return res
}

func NewWithRanges(inner gowid.IWidget, notFocusRange []AttributeRange, focusRange []AttributeRange, opts ...Options) *Widget {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}
	res := &Widget{
		IWidget:       inner,
		focusRange:    focusRange,
		notFocusRange: notFocusRange,
		options:       opt,
	}
	res.SubWidgetCallbacks = gowid.SubWidgetCallbacks{CB: &res.Callbacks}
	var _ gowid.IWidget = res
	return res
}

func (w *Widget) String() string {
	return fmt.Sprintf("styler[%v]", w.SubWidget())
}

func (w *Widget) SubWidget() gowid.IWidget {
	return w.IWidget
}

func (w *Widget) SetSubWidget(inner gowid.IWidget, app gowid.IApp) {
	w.IWidget = inner
	gowid.RunWidgetCallbacks(w, gowid.SubWidgetCB{}, app, w)
}

func (w *Widget) SubWidgetSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderSize {
	return w.RenderSize(size, focus, app)
}

func (w *Widget) RenderSize(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.IRenderBox {
	return w.SubWidget().RenderSize(size, focus, app)
}

func (w *Widget) UserInput(ev interface{}, size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) bool {
	return gowid.UserInputIfSelectable(w.IWidget, ev, size, focus, app)
}

func (w *Widget) Render(size gowid.IRenderSize, focus gowid.Selector, app gowid.IApp) gowid.ICanvas {
	canvas := w.SubWidget().Render(size, focus, app)

	cols := canvas.BoxColumns()

	var attrSpecs []AttributeRange
	var f1 gowid.TCellColor
	var b1 gowid.TCellColor

	if focus.Focus {
		attrSpecs = w.focusRange
	} else {
		attrSpecs = w.notFocusRange
	}

	x := cols
	y := canvas.BoxRows()
	max := x * y

	if attrSpecs != nil {
		for _, attr := range attrSpecs {
			// TODO - bounds checks
			if attr.Styler != nil {
				f, b, s := attr.Styler.GetStyle(app)
				for i := attr.Start; true; i++ {
					if attr.End != -1 && i == attr.End {
						break
					}
					if i == max {
						break
					}
					col, row := i%cols, i/cols

					c := canvas.CellAt(col, row)
					c2 := c

					if f != nil {
						f1 = gowid.IColorToTCell(f, gowid.ColorNone, app.GetColorMode())
						c = c.WithForegroundColor(f1)
					}
					if b != nil {
						b1 = gowid.IColorToTCell(b, gowid.ColorNone, app.GetColorMode())
						c = c.WithBackgroundColor(b1)
					}

					if !w.options.OverWrite {
						c = c.WithStyle(s).MergeDisplayAttrsUnder(c2)
					} else {
						c = c2.MergeDisplayAttrsUnder(c.WithStyle(s))
					}
					canvas.SetCellAt(col, row, c)
				}
			}
		}
	}

	return canvas
}

//======================================================================

type ReverseIfSelectedForCopy struct{}

var _ gowid.IClipboardSelected = ReverseIfSelectedForCopy{}

func (r ReverseIfSelectedForCopy) AlterWidget(w gowid.IWidget, app gowid.IApp) gowid.IWidget {
	return New(w, gowid.MakeStyledAs(gowid.StyleReverse))
}

//======================================================================

type BoldIfSelectedForCopy struct{}

var _ gowid.IClipboardSelected = BoldIfSelectedForCopy{}

func (r BoldIfSelectedForCopy) AlterWidget(w gowid.IWidget, app gowid.IApp) gowid.IWidget {
	return New(w, gowid.MakeStyledAs(gowid.StyleBold))
}

//======================================================================

type BlinkIfSelectedForCopy struct{}

var _ gowid.IClipboardSelected = BlinkIfSelectedForCopy{}

func (r BlinkIfSelectedForCopy) AlterWidget(w gowid.IWidget, app gowid.IApp) gowid.IWidget {
	return New(w, gowid.MakeStyledAs(gowid.StyleBlink))
}

//======================================================================

type UsePaletteIfSelectedForCopy struct {
	Entry string
}

var _ gowid.IClipboardSelected = UsePaletteIfSelectedForCopy{}

func (r UsePaletteIfSelectedForCopy) AlterWidget(w gowid.IWidget, app gowid.IApp) gowid.IWidget {
	return New(w, gowid.MakePaletteRef(r.Entry))
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
