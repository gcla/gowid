// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package pile

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/framed"
	"github.com/gcla/gowid/widgets/list"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// Test that a mouse wheel down inside the region of the list within a pile
// is correctly translated and passed to the list.
func TestPile5(t *testing.T) {
	bws := make([]gowid.IWidget, 50)
	for i := 0; i < len(bws); i++ {
		bws[i] = button.New(text.New(fmt.Sprintf("%03d", i)))
	}

	walker := list.NewSimpleListWalker(bws)
	lb := list.New(walker)
	// Framed is needed because it validates the mouse y coordinate before passing it on
	// to its subwidget.
	flb := framed.New(lb)

	pws := make([]gowid.IContainerWidget, 3)
	pws[0] = &gowid.ContainerWidget{button.New(text.New("top  ")), gowid.RenderWithUnits{U: 1}}
	pws[1] = &gowid.ContainerWidget{flb, gowid.RenderWithWeight{W: 1}} // WEIGHT!!
	pws[2] = &gowid.ContainerWidget{button.New(text.New("bot  ")), gowid.RenderWithUnits{U: 1}}

	pl := New(pws)

	sta := make([]string, 0)
	sta = append(sta, "<top  >")
	sta = append(sta, "-------")
	for i := 0; i < 6; i++ {
		sta = append(sta, fmt.Sprintf("|<%03d>|", i))
	}
	sta = append(sta, "-------")
	sta = append(sta, "<bot  >")

	csize := gowid.RenderBox{C: 7, R: 10}
	c := pl.Render(csize, gowid.Focused, gwtest.D)
	assert.Equal(t, strings.Join(sta, "\n"), c.String())

	assert.Equal(t, 0, pl.Focus())

	evmdown := tcell.NewEventMouse(1, 4, tcell.WheelDown, 0)

	pl.UserInput(evmdown, csize, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, pl.Focus()) // Now at list widget
	assert.Equal(t, 0, lb.Walker().Focus().(list.ListPos).ToInt())

	pl.UserInput(evmdown, csize, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, pl.Focus()) // Now at list widget
	assert.Equal(t, 1, lb.Walker().Focus().(list.ListPos).ToInt())

	for i := 0; i < 40; i++ {
		pl.UserInput(evmdown, csize, gowid.Focused, gwtest.D)
	}

	assert.Equal(t, 1, pl.Focus()) // Now at list widget
	assert.Equal(t, 41, lb.Walker().Focus().(list.ListPos).ToInt())
}

func TestPile2(t *testing.T) {
	btns := make([]gowid.IContainerWidget, 0)
	//clicks := make([]*gwtest.ButtonTester, 0)

	for i := 0; i < 3; i++ {
		btn := button.New(text.New("abc"))
		click := &gwtest.ButtonTester{Gotit: false}
		btn.OnClick(click)
		btns = append(btns, &gowid.ContainerWidget{btn, gowid.RenderFixed{}})
		//clicks = append(clicks, click)
	}

	pl := New(btns)

	st1 := "<abc>\n<abc>\n<abc>"
	st2 := "<abc> \n<abc> \n<abc> "

	c := pl.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c.String(), st1)

	c2 := pl.Render(gowid.RenderFlowWith{C: 6}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), st2)

	assert.Equal(t, pl.Focus(), 0)

	// evright := tcell.NewEventKey(tcell.KeyRight, ' ', tcell.ModNone)
	// evleft := tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
	// evdown := tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
	// evspace := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	evmdown := tcell.NewEventMouse(1, 1, tcell.WheelDown, 0)
	evmup := tcell.NewEventMouse(1, 1, tcell.WheelUp, 0)
	// evmright := tcell.NewEventMouse(1, 1, tcell.WheelRight, 0)
	// evmleft := tcell.NewEventMouse(1, 1, tcell.WheelLeft, 0)

	cbcalled := false

	pl.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		assert.Equal(t, w, pl)
		cbcalled = true
	}})

	pl.UserInput(evmdown, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, pl.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	pl.UserInput(evmdown, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, 2, pl.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	pl.UserInput(evmdown, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, 2, pl.Focus())
	assert.Equal(t, false, cbcalled)

	pl.UserInput(evmup, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, 1, pl.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	pl.UserInput(evmup, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, 0, pl.Focus())
	assert.Equal(t, true, cbcalled)
	cbcalled = false
	pl.UserInput(evmup, gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, 0, pl.Focus())
	assert.Equal(t, false, cbcalled)
}

func TestPile1(t *testing.T) {
	w1 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithUnits{U: 2}},
	})
	c1 := w1.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "xxx\nxxx\nyyy\nyyy")

	w2 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithUnits{U: 1}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithUnits{U: 2}},
	})
	c2 := w2.Render(gowid.RenderFlowWith{C: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "xxx\nyyy\nyyy")

	w3 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithWeight{1}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithWeight{2}},
	})
	assert.Panics(t, func() {
		w3.Render(gowid.RenderFlowWith{C: 3}, gowid.Focused, gwtest.D)
	})

	w4 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithRatio{0.25}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithRatio{0.5}},
	})

	c4 := w4.Render(gowid.RenderBox{C: 3, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c4.String(), "xxx\nyyy\nyyy")

	c41 := w4.Render(gowid.RenderBox{C: 3, R: 4}, gowid.Focused, gwtest.D)
	assert.Equal(t, c41.String(), "xxx\nyyy\nyyy\n   ")

	for _, w := range []gowid.IWidget{w1, w2, w4} {
		gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	}
	gwtest.RenderFlowManyTimes(t, w2, 0, 10)
}

func TestPile3(t *testing.T) {
	w1 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{text.New("y"), gowid.RenderFlow{}},
	})
	// Test that a pile can render in flow mode with a single embedded flow widget
	c1 := w1.Render(gowid.RenderFlowWith{C: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "xxx\nxxx\ny  ")

	w1 = New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{text.New("y"), gowid.RenderWithWeight{1}},
	})
	// Test that a pile can render in flow mode with a single embedded flow widget
	c1 = w1.Render(gowid.RenderFlowWith{C: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "xxx\nxxx\ny  ")

	w1 = New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{text.New("y"), gowid.RenderWithWeight{1}},
		&gowid.ContainerWidget{text.New("z"), gowid.RenderWithWeight{1}},
	})
	// Two weight widgets don't work in flow mode, how do you restrict their vertical ratio?
	assert.Panics(t, func() {
		w1.Render(gowid.RenderFlowWith{C: 3}, gowid.Focused, gwtest.D)
	})

}

func makep(c rune) gowid.IWidget {
	return selectable.New(fill.New(c))
}

func makepfixed(c rune) gowid.IContainerWidget {
	return &gowid.ContainerWidget{
		IWidget: makep(c),
		D:       gowid.RenderFixed{},
	}
}

type renderWeightUpTo struct {
	gowid.RenderWithWeight
	max int
}

func (s renderWeightUpTo) MaxUnits() int {
	return s.max
}

func weightupto(w int, max int) renderWeightUpTo {
	return renderWeightUpTo{gowid.RenderWithWeight{W: w}, max}
}

func TestPile4(t *testing.T) {
	subs := []gowid.IContainerWidget{
		&gowid.ContainerWidget{makep('x'), gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{makep('y'), gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{makep('z'), gowid.RenderWithWeight{W: 1}},
	}
	w := New(subs)
	c := w.Render(gowid.RenderBox{C: 1, R: 12}, gowid.Focused, gwtest.D)
	assert.Equal(t, `
x
x
x
x
y
y
y
y
z
z
z
z`[1:], c.String())
	subs[2] = &gowid.ContainerWidget{makep('z'), renderWeightUpTo{gowid.RenderWithWeight{W: 1}, 2}}
	w = New(subs)
	c = w.Render(gowid.RenderBox{C: 1, R: 12}, gowid.Focused, gwtest.D)
	assert.Equal(t, `
x
x
x
x
x
y
y
y
y
y
z
z`[1:], c.String())

}

func TestPile6(t *testing.T) {
	subs := []gowid.IContainerWidget{
		&gowid.ContainerWidget{text.New("foo"), gowid.RenderFixed{}},
		&gowid.ContainerWidget{text.New("bar"), gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{text.New("baz"), gowid.RenderFixed{}},
	}
	w := New(subs)
	c := w.Render(gowid.RenderBox{C: 3, R: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, `
foo
bar
   
   
baz`[1:], c.String())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
