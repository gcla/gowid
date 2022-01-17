// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package columns

import (
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/button"
	"github.com/gcla/gowid/widgets/checkbox"
	"github.com/gcla/gowid/widgets/fill"
	"github.com/gcla/gowid/widgets/selectable"
	"github.com/gcla/gowid/widgets/text"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestInterfaces1(t *testing.T) {
	var _ gowid.IWidget = (*Widget)(nil)
	var _ IWidget = (*Widget)(nil)
	var _ gowid.ICompositeMultipleDimensions = (*Widget)(nil)
	var _ gowid.ICompositeMultipleWidget = (*Widget)(nil)
}

func TestColumns1(t *testing.T) {
	w1 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithUnits{U: 2}},
	})
	c1 := w1.Render(gowid.RenderBox{C: 4, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "xxyy\nxxyy\nxxyy")

	w2 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithWeight{6}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithWeight{2}},
	})
	c2 := w2.Render(gowid.RenderBox{C: 4, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "xxxy\nxxxy\nxxxy")

	w3 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithRatio{0.75}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithRatio{0.35}},
	})
	c3 := w3.Render(gowid.RenderBox{C: 4, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c3.String(), "xxxy\nxxxy\nxxxy")

	w4 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{fill.New('x'), gowid.RenderWithRatio{0.5}},
		&gowid.ContainerWidget{fill.New('y'), gowid.RenderWithWeight{10}},
		&gowid.ContainerWidget{fill.New('z'), gowid.RenderWithWeight{5}},
	})
	c4 := w4.Render(gowid.RenderBox{C: 6, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c4.String(), "xxxyyz\nxxxyyz\nxxxyyz")

	w5 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{checkbox.New(false), gowid.RenderFixed{}},
		&gowid.ContainerWidget{checkbox.New(false), gowid.RenderFixed{}},
		&gowid.ContainerWidget{checkbox.New(false), gowid.RenderFixed{}},
	})

	idx := -1
	w5.OnFocusChanged(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		w2 := w.(*Widget)
		idx = w2.Focus()
	}})

	assert.Equal(t, w5.Focus(), 0)
	assert.Equal(t, idx, -1)
	w5.SetFocus(gwtest.D, 1)
	assert.Equal(t, w5.Focus(), 1)
	assert.Equal(t, idx, 1)

	w5.SetFocus(gwtest.D, 100)
	assert.Equal(t, w5.Focus(), 2)

	for _, w := range []gowid.IWidget{w1, w2, w3, w4} {
		gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
		gwtest.RenderFlowManyTimes(t, w, 0, 10)
	}
	gwtest.RenderFixedDoesNotPanic(t, w5)
}

func TestColumns2(t *testing.T) {
	w1 := New([]gowid.IContainerWidget{
		&gowid.ContainerWidget{&text.Widget1{0}, gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{&text.Widget1{1}, gowid.RenderWithUnits{U: 2}},
		&gowid.ContainerWidget{&text.Widget1{2}, gowid.RenderWithUnits{U: 2}},
	})
	sz := gowid.RenderBox{C: 6, R: 1}
	c1 := w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "0f1 2 ", c1.String())
	assert.Equal(t, 0, w1.Focus())

	evright := gwtest.CursorRight()
	w1.UserInput(evright, sz, gowid.Focused, gwtest.D)
	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "0 1f2 ", c1.String())
	assert.Equal(t, 1, w1.Focus())

	evlmx0y0 := tcell.NewEventMouse(0, 0, tcell.Button1, 0)
	evnonex0y0 := tcell.NewEventMouse(0, 0, tcell.ButtonNone, 0)

	w1.UserInput(evlmx0y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{true, false, false})
	w1.UserInput(evnonex0y0, sz, gowid.Focused, gwtest.D)
	gwtest.D.SetLastMouseState(gowid.MouseState{false, false, false})
	c1 = w1.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "0f1 2 ", c1.String())
}

func makec3(txt string) gowid.IWidget {
	return selectable.New(text.New(txt))
}

func makec3fixed(txt string) gowid.IContainerWidget {
	return &gowid.ContainerWidget{
		IWidget: makec3(txt),
		D:       gowid.RenderFixed{},
	}
}

func TestColumns3(t *testing.T) {
	w := NewFixed(makec3("111"), makec3("222"), makec3("333"))
	c := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "111222333", c.String())
	assert.Equal(t, 0, w.Focus())
	w.SetFocus(gwtest.D, 2)
	assert.Equal(t, 2, w.Focus())
	w.SetSubWidgets([]gowid.IWidget{
		makec3fixed("aaaa"),
		makec3fixed("bbbbb"),
	},
		gwtest.D,
	)
	c = w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, "aaaabbbbb", c.String())
	assert.Equal(t, 1, w.Focus())
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

func weight(w int) gowid.RenderWithWeight {
	return gowid.RenderWithWeight{W: w}
}

func makep(c rune) gowid.IWidget {
	return selectable.New(fill.New(c))
}

func TestColumns4(t *testing.T) {
	subs := []gowid.IContainerWidget{
		&gowid.ContainerWidget{makep('x'), gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{makep('y'), gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{makep('z'), gowid.RenderWithWeight{W: 1}},
	}
	w := New(subs)
	c := w.Render(gowid.RenderBox{C: 12, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, "xxxxyyyyzzzz", c.String())
	subs[2] = &gowid.ContainerWidget{makep('z'), renderWeightUpTo{gowid.RenderWithWeight{W: 1}, 2}}
	w = New(subs)
	c = w.Render(gowid.RenderBox{C: 12, R: 1}, gowid.Focused, gwtest.D)
	assert.Equal(t, "xxxxxyyyyyzz", c.String())
}

func TestColumns5(t *testing.T) {
	// None are selectable
	subs := []gowid.IContainerWidget{
		&gowid.ContainerWidget{text.New("x"), gowid.RenderWithWeight{W: 1}},
		&gowid.ContainerWidget{text.New("y"), gowid.RenderWithWeight{W: 1}},
	}
	w := New(subs)
	sz := gowid.RenderBox{C: 2, R: 1}
	c := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, "xy", c.String())

	evright := gwtest.CursorRight()
	acc := w.UserInput(evright, sz, gowid.Focused, gwtest.D)

	// Nothing in here should accept the input, so it should bubble back up
	assert.False(t, acc)
}

type renderWithUnitsMax struct {
	gowid.RenderWithUnits
	gowid.RenderMax
}

func TestColumns6(t *testing.T) {
	h := renderWithUnitsMax{
		RenderWithUnits: gowid.RenderWithUnits{1},
	}

	f := fill.New(' ')

	subs := []gowid.IContainerWidget{
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("1")), gowid.RenderWithWeight{W: 4}},
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("0.000000")), gowid.RenderWithWeight{W: 8}},
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("192.168.44.123")), gowid.RenderWithWeight{W: 14}},
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("192.168.44.213")), gowid.RenderWithWeight{W: 14}},
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("TFTP")), gowid.RenderWithWeight{W: 6}},
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("77")), gowid.RenderWithWeight{W: 7}},
		&gowid.ContainerWidget{f, h},
		&gowid.ContainerWidget{button.NewBare(text.New("Read Request, File: C:\\IBMTCPIP\\lccm.1, Transfer type: octet")), gowid.RenderWithWeight{W: 60}},
		&gowid.ContainerWidget{f, h},
	}
	w := New(subs)
	sz := gowid.RenderFlowWith{C: 158}
	c := w.Render(sz, gowid.Focused, gwtest.D)
	assert.Equal(t, " 1     0.000000    192.168.44.123      192.168.44.213      TFTP     77        Read Request, File: C:\\IBMTCPIP\\lccm.1, Transfer type: octet                    ", c.String())

}

func TestColumns7(t *testing.T) {
	d2 := weightupto(1, 2) // weight 1, max 2
	d3 := weightupto(1, 3) // weight 1, max 3
	d4 := weightupto(1, 4) // weight 1, max 4

	sz4 := gowid.RenderFlowWith{C: 4}
	sz6 := gowid.RenderFlowWith{C: 6}

	subs := []gowid.IContainerWidget{
		&gowid.ContainerWidget{text.New("aa"), d2},
		&gowid.ContainerWidget{text.New("bb"), d2},
	}

	w := New(subs)
	c := w.Render(sz4, gowid.Focused, gwtest.D)
	assert.Equal(t, "aabb", c.String())

	subs = []gowid.IContainerWidget{
		&gowid.ContainerWidget{text.New("aa"), d3},
		&gowid.ContainerWidget{text.New("bb"), d3},
	}
	w = New(subs)

	c = w.Render(sz4, gowid.Focused, gwtest.D)
	assert.Equal(t, "aabb", c.String())

	c = w.Render(sz6, gowid.Focused, gwtest.D)
	assert.Equal(t, "aa bb ", c.String())

	subs = []gowid.IContainerWidget{
		&gowid.ContainerWidget{text.New("aaaa"), d4},
		&gowid.ContainerWidget{text.New("bb"), d2},
	}
	w = New(subs)

	c = w.Render(sz6, gowid.Focused, gwtest.D)
	assert.Equal(t, "aaaabb", c.String())

	subs = []gowid.IContainerWidget{
		&gowid.ContainerWidget{text.New("aaaa"), weight(1)},
		&gowid.ContainerWidget{text.New("bb"), weight(1)},
	}
	w = New(subs)

	c = w.Render(sz6, gowid.Focused, gwtest.D)
	assert.Equal(t, "aaabb \na     ", c.String())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
