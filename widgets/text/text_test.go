// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package text

import (
	"io"
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//======================================================================

var testl gowid.PaletteRef
var testl2 gowid.PaletteRef

func init() {
	testl = gowid.MakePaletteRef("test")
	testl2 = gowid.MakePaletteRef("test2")
}

func TestAdd1(t *testing.T) {
	m1 := StyledContent("hello world", testl)
	m2 := StyledContent("foobar", testl2)

	t1 := NewContent([]ContentSegment{m1})
	t1.AddAt(5, m2)
	assert.Equal(t, "hellofoobar world", t1.String())

	t1.DeleteAt(3, 7)
	assert.Equal(t, "helr world", t1.String())
}

func TestLayout1(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 5, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 1)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 5, Clipped: true})

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 5, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, len(l2.Lines), 3)
	assert.Equal(t, l2.Lines[0], LineLayout{Start: 0, End: 5})
	assert.Equal(t, l2.Lines[1], LineLayout{Start: 5, End: 10})
	assert.Equal(t, l2.Lines[2], LineLayout{Start: 10, End: 11})
}

func TestLayout2(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 20, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 1)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 11})
}

func TestLayout3(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\nsome more", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 2)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 7, Clipped: true})
	assert.Equal(t, l1.Lines[1], LineLayout{Start: 12, End: 19, Clipped: true})

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 7, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, len(l2.Lines), 4)
	assert.Equal(t, l2.Lines[0], LineLayout{Start: 0, End: 7})
	assert.Equal(t, l2.Lines[1], LineLayout{Start: 7, End: 11})
	assert.Equal(t, l2.Lines[2], LineLayout{Start: 12, End: 19})
	assert.Equal(t, l2.Lines[3], LineLayout{Start: 19, End: 21})
}

func TestLayout31(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("he\nsome more", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 2)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 2})
	assert.Equal(t, l1.Lines[1], LineLayout{Start: 3, End: 10, Clipped: true})
}

func TestLayout4(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\nsome", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 2)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 7, Clipped: true})
	assert.Equal(t, l1.Lines[1], LineLayout{Start: 12, End: 16})
}

func TestLayout5(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\n\nsome", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 3)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 7, Clipped: true})
	assert.Equal(t, l1.Lines[1], LineLayout{Start: 12, End: 12})
	assert.Equal(t, l1.Lines[2], LineLayout{Start: 13, End: 17})

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 7, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, len(l2.Lines), 4)
	assert.Equal(t, l2.Lines[0], LineLayout{Start: 0, End: 7})
	assert.Equal(t, l2.Lines[1], LineLayout{Start: 7, End: 11})
	assert.Equal(t, l2.Lines[2], LineLayout{Start: 12, End: 12})
	assert.Equal(t, l2.Lines[3], LineLayout{Start: 13, End: 17})
}

func TestLayout6(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\n", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 1)
	assert.Equal(t, l1.Lines[0], LineLayout{Start: 0, End: 7, Clipped: true})
}

func TestMultiline1(t *testing.T) {
	msg := `hello
world
this
is
cool`
	widget1 := New(msg)
	canvas1 := widget1.Render(gowid.RenderBox{C: 16, R: 6}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget1 is %v", widget1)
	log.Infof("Canvas1 is %s", canvas1.String())
	res := strings.Join([]string{
		"hello           ",
		"world           ",
		"this            ",
		"is              ",
		"cool            ",
		"                ",
	}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas1(t *testing.T) {
	widget1 := New("hello world")
	canvas1 := widget1.Render(gowid.RenderBox{C: 20, R: 1}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget1 is %v", widget1)
	log.Infof("Canvas1 is %s", canvas1.String())
	res := strings.Join([]string{"hello world         "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas1b(t *testing.T) {
	widget1 := New("hello world this is a test")
	canvas1 := widget1.Render(gowid.RenderBox{C: 7, R: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget1 is %v", widget1)
	log.Infof("Canvas1 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld th", "is is a"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas2(t *testing.T) {
	widget1 := New("hello world")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget2 is %v", widget1)
	log.Infof("Canvas2 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld   "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas3(t *testing.T) {
	widget1 := New("hello world")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget3 is %v", widget1)
	log.Infof("Canvas3 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld   "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas4(t *testing.T) {
	widget1 := New("hello world every day")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget4 is %v", widget1)
	log.Infof("Canvas4 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld ev", "ery day"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas5(t *testing.T) {
	widget1 := New("hello world every day")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget5 is %v", widget1)
	log.Infof("Canvas5 is %s", canvas1.String())
	res := strings.Join([]string{"hel", "lo ", "wor", "ld ", "eve", "ry ", "day"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas6(t *testing.T) {
	widget1 := New("hello transubstantiation")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget6 is %v", widget1)
	log.Infof("Canvas6 is %s", canvas1.String())
	res := strings.Join([]string{"hello tr", "ansubsta", "ntiation"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas6b(t *testing.T) {
	widget1 := New("hello transubstantiation good")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget6 is %v", widget1)
	log.Infof("Canvas6 is %s", canvas1.String())
	res := strings.Join([]string{"hello tr", "ansubsta", "ntiation", " good   "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas7(t *testing.T) {
	widget1 := New("hello transubstantiation good")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget7 is %v", widget1)
	log.Infof("Canvas7 is %s", canvas1.String())
	res := strings.Join([]string{"hello trans", "ubstantiati", "on good    "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas8(t *testing.T) {
	widget1 := New("hello transubstantiation boy")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget8 is %v", widget1)
	log.Infof("Canvas8 is %s", canvas1.String())
	res := strings.Join([]string{"hello trans", "ubstantiati", "on boy     "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas11(t *testing.T) {
	widget1 := New("hello transubstantiation good")
	//fwidget1 := NewFramed(widget1)
	canvas1 := widget1.Render(gowid.RenderBox{C: 5, R: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget11 is %v", widget1)
	log.Infof("Canvas11 is %s", canvas1.String())
	res := strings.Join([]string{"hello", " tran", "subst"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas16(t *testing.T) {
	widget1 := New("line 1 line 2 line 3")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget16 is %v", widget1)
	log.Infof("Canvas16 is %s", canvas1.String())
	res := strings.Join([]string{"line 1 l", "ine 2 li", "ne 3    "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas24(t *testing.T) {
	widget1 := New("line 1line 2line 3")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	canvas1.Truncate(1, 0)
	log.Infof("Widget18 is %v", widget1)
	log.Infof("Canvas18 is %v", canvas1)
	log.Infof("Canvas18 is %s", canvas1.String())
	res := strings.Join([]string{"line 2", "line 3"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas25(t *testing.T) {
	widget1 := New("line 1line 2line 3")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	canvas1.Truncate(0, 2)
	log.Infof("Widget18 is %v", widget1)
	log.Infof("Canvas18 is %s", canvas1.String())
	res := strings.Join([]string{"line 1"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas27(t *testing.T) {
	widget1 := New("")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 0}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is '%s'", canvas1.String())
	res := strings.Join([]string{""}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas28(t *testing.T) {
	widget1 := New("the needs of the many outweigh\nthe needs of the few.\nOr the one.", Options{
		Wrap: WrapClip,
	})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 10}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is '%s'", canvas1.String())
	res := strings.Join([]string{"the needs ", "the needs ", "Or the one"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestTextAlign1(t *testing.T) {
	widget1 := New("hello\nworld\nfoo\nba")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{"hello  ", "world  ", "foo    ", "ba     "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestTextAlign2(t *testing.T) {
	widget1 := New("hello\nworld\nfoo\nba", Options{
		Align: gowid.HAlignRight{},
	})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{"  hello", "  world", "    foo", "     ba"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestTextAlign3(t *testing.T) {
	widget1 := New("hello\nworld\nfoo\nba", Options{
		Align: gowid.HAlignMiddle{},
	})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{" hello ", " world ", "  foo  ", "  ba   "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestText1(t *testing.T) {
	w := New("hello world")
	c1 := w.Render(gowid.RenderFlowWith{C: 20}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "hello world         ")

	tset := false
	w.OnContentSet(gowid.WidgetCallback{"cb", func(app gowid.IApp, w gowid.IWidget) {
		tset = true
	}})

	_, err := io.Copy(&Writer{w, gwtest.D}, strings.NewReader("goodbye everyone"))
	assert.NoError(t, err)
	assert.Equal(t, tset, true)
	tset = false
	c2 := w.Render(gowid.RenderFlowWith{C: 20}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "goodbye everyone    ")

	_, err = io.Copy(&Writer{w, gwtest.D}, strings.NewReader("multi\nline\ntest"))
	assert.NoError(t, err)
	assert.Equal(t, tset, true)
	tset = false
	c3 := w.Render(gowid.RenderFlowWith{C: 10}, gowid.Focused, gwtest.D)
	assert.Equal(t, c3.String(), "multi     \nline      \ntest      ")
}

func TestText2(t *testing.T) {
	w := New("hello world")
	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "hello world")
}

func TestText3(t *testing.T) {
	w := New("hello yo")

	// TODO - make 2nd last arg 0
	gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
	gwtest.RenderFlowManyTimes(t, w, 0, 10)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
