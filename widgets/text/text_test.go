// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
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

	assert.Equal(t, 1, len(l1.Lines))
	assert.Equal(t, LineLayout{StartWidth: 0, StartLength: 0, EndLength: 5, EndWidth: 5, Clipped: true}, l1.Lines[0])

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 5, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, 3, len(l2.Lines))
	assert.Equal(t, LineLayout{StartWidth: 0, StartLength: 0, EndLength: 5, EndWidth: 5}, l2.Lines[0])
	assert.Equal(t, LineLayout{StartWidth: 5, StartLength: 5, EndLength: 10, EndWidth: 10}, l2.Lines[1])
	assert.Equal(t, LineLayout{StartWidth: 10, StartLength: 10, EndLength: 11, EndWidth: 11}, l2.Lines[2])
}

func TestLayoutW1(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hell现o world", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 5, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, 1, len(l1.Lines))
	assert.Equal(t, LineLayout{StartWidth: 0, StartLength: 0, EndLength: 4, EndWidth: 4, Clipped: true}, l1.Lines[0])

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 5, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, 3, len(l2.Lines))
	assert.Equal(t, LineLayout{StartWidth: 0, StartLength: 0, EndLength: 4, EndWidth: 4}, l2.Lines[0])
	assert.Equal(t, LineLayout{StartWidth: 4, StartLength: 4, EndLength: 8, EndWidth: 9}, l2.Lines[1])
	assert.Equal(t, LineLayout{StartWidth: 9, StartLength: 8, EndLength: 12, EndWidth: 13}, l2.Lines[2])
}

func TestLayout2(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 20, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 1)
	assert.Equal(t, l1.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 11, EndWidth: 11})
}

func TestLayout3(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\nsome more", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 2)
	assert.Equal(t, l1.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 7, Clipped: true, EndWidth: 7})
	assert.Equal(t, l1.Lines[1], LineLayout{StartWidth: 12, StartLength: 12, EndLength: 19, Clipped: true, EndWidth: 19})

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 7, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, len(l2.Lines), 4)
	assert.Equal(t, l2.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 7, EndWidth: 7})
	assert.Equal(t, l2.Lines[1], LineLayout{StartWidth: 7, StartLength: 7, EndLength: 11, EndWidth: 11})
	assert.Equal(t, l2.Lines[2], LineLayout{StartWidth: 12, StartLength: 12, EndLength: 19, EndWidth: 19})
	assert.Equal(t, l2.Lines[3], LineLayout{StartWidth: 19, StartLength: 19, EndLength: 21, EndWidth: 21})
}

func TestLayout31(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("he\nsome more", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 2)
	assert.Equal(t, l1.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 2, EndWidth: 2})
	assert.Equal(t, l1.Lines[1], LineLayout{StartWidth: 3, StartLength: 3, EndLength: 10, Clipped: true, EndWidth: 10})
}

func TestLayout4(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\nsome", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 2)
	assert.Equal(t, l1.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 7, Clipped: true, EndWidth: 7})
	assert.Equal(t, l1.Lines[1], LineLayout{StartWidth: 12, StartLength: 12, EndLength: 16, EndWidth: 16})
}

func TestLayout5(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\n\nsome", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, len(l1.Lines), 3)
	assert.Equal(t, l1.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 7, Clipped: true, EndWidth: 7})
	assert.Equal(t, l1.Lines[1], LineLayout{StartWidth: 12, StartLength: 12, EndLength: 12, EndWidth: 12})
	assert.Equal(t, l1.Lines[2], LineLayout{StartWidth: 13, StartLength: 13, EndLength: 17, EndWidth: 17})

	t2 := NewContent(tm1)
	l2 := MakeTextLayout(t2, 7, WrapAny, gowid.HAlignLeft{})

	assert.Equal(t, len(l2.Lines), 4)
	assert.Equal(t, l2.Lines[0], LineLayout{StartWidth: 0, StartLength: 0, EndLength: 7, EndWidth: 7})
	assert.Equal(t, l2.Lines[1], LineLayout{StartWidth: 7, StartLength: 7, EndLength: 11, EndWidth: 11})
	assert.Equal(t, l2.Lines[2], LineLayout{StartWidth: 12, StartLength: 12, EndLength: 12, EndWidth: 12})
	assert.Equal(t, l2.Lines[3], LineLayout{StartWidth: 13, StartLength: 13, EndLength: 17, EndWidth: 17})
}

func TestLayout6(t *testing.T) {
	tm1 := []ContentSegment{StyledContent("hello world\n", testl)}

	t1 := NewContent(tm1)
	l1 := MakeTextLayout(t1, 7, WrapClip, gowid.HAlignLeft{})

	assert.Equal(t, 2, len(l1.Lines))
	assert.Equal(t, LineLayout{StartWidth: 0, StartLength: 0, EndLength: 7, Clipped: true, EndWidth: 7}, l1.Lines[0])
	assert.Equal(t, LineLayout{StartWidth: 12, StartLength: 12, EndLength: 12, Clipped: false, EndWidth: 12}, l1.Lines[1])
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
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas1(t *testing.T) {
	widget1 := New("hello world")
	canvas1 := widget1.Render(gowid.RenderBox{C: 20, R: 1}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget1 is %v", widget1)
	log.Infof("Canvas1 is %s", canvas1.String())
	res := strings.Join([]string{"hello world         "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas1b(t *testing.T) {
	widget1 := New("hello world this is a test")
	canvas1 := widget1.Render(gowid.RenderBox{C: 7, R: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget1 is %v", widget1)
	log.Infof("Canvas1 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld th", "is is a"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas2(t *testing.T) {
	widget1 := New("hello world")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget2 is %v", widget1)
	log.Infof("Canvas2 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld   "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas3(t *testing.T) {
	widget1 := New("hello world")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget3 is %v", widget1)
	log.Infof("Canvas3 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld   "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas4(t *testing.T) {
	widget1 := New("hello world every day")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget4 is %v", widget1)
	log.Infof("Canvas4 is %s", canvas1.String())
	res := strings.Join([]string{"hello w", "orld ev", "ery day"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas5(t *testing.T) {
	widget1 := New("hello world every day")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget5 is %v", widget1)
	log.Infof("Canvas5 is %s", canvas1.String())
	res := strings.Join([]string{"hel", "lo ", "wor", "ld ", "eve", "ry ", "day"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas6(t *testing.T) {
	widget1 := New("hello transubstantiation")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget6 is %v", widget1)
	log.Infof("Canvas6 is %s", canvas1.String())
	res := strings.Join([]string{"hello tr", "ansubsta", "ntiation"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas6b(t *testing.T) {
	widget1 := New("hello transubstantiation good")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget6 is %v", widget1)
	log.Infof("Canvas6 is %s", canvas1.String())
	res := strings.Join([]string{"hello tr", "ansubsta", "ntiation", " good   "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas7(t *testing.T) {
	widget1 := New("hello transubstantiation good")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget7 is %v", widget1)
	log.Infof("Canvas7 is %s", canvas1.String())
	res := strings.Join([]string{"hello trans", "ubstantiati", "on good    "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas8(t *testing.T) {
	widget1 := New("hello transubstantiation boy")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget8 is %v", widget1)
	log.Infof("Canvas8 is %s", canvas1.String())
	res := strings.Join([]string{"hello trans", "ubstantiati", "on boy     "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas11(t *testing.T) {
	widget1 := New("hello transubstantiation good")
	//fwidget1 := NewFramed(widget1)
	canvas1 := widget1.Render(gowid.RenderBox{C: 5, R: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget11 is %v", widget1)
	log.Infof("Canvas11 is %s", canvas1.String())
	res := strings.Join([]string{"hello", " tran", "subst"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas16(t *testing.T) {
	widget1 := New("line 1 line 2 line 3")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 8}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget16 is %v", widget1)
	log.Infof("Canvas16 is %s", canvas1.String())
	res := strings.Join([]string{"line 1 l", "ine 2 li", "ne 3    "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas24(t *testing.T) {
	widget1 := New("line 1line 2line 3")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	canvas1.Truncate(1, 0)
	log.Infof("Widget18 is %v", widget1)
	log.Infof("Canvas18 is %v", canvas1)
	log.Infof("Canvas18 is %s", canvas1.String())
	res := strings.Join([]string{"line 2", "line 3"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas25(t *testing.T) {
	widget1 := New("line 1line 2line 3")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	canvas1.Truncate(0, 2)
	log.Infof("Widget18 is %v", widget1)
	log.Infof("Canvas18 is %s", canvas1.String())
	res := strings.Join([]string{"line 1"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas27(t *testing.T) {
	widget1 := New("")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 0}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is '%s'", canvas1.String())
	res := strings.Join([]string{""}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas28(t *testing.T) {
	widget1 := New("the needs of the many outweigh\nthe needs of the few.\nOr the one.", Options{
		Wrap: WrapClip,
	})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 10}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is '%s'", canvas1.String())
	res := strings.Join([]string{"the needs ", "the needs ", "Or the one"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestTextAlign1(t *testing.T) {
	widget1 := New("hel你o\nworld\nf你o\nba")
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{"hel你o ", "world  ", "f你o   ", "ba     "}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestTextAlign2(t *testing.T) {
	widget1 := New("hello\nworld\nfoo\nba", Options{
		Align: gowid.HAlignRight{},
	})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{"  hello", "  world", "    foo", "     ba"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestTextAlign3(t *testing.T) {
	widget1 := New("hello\nworld\nfoo\nba", Options{
		Align: gowid.HAlignMiddle{},
	})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 7}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{" hello ", " world ", "  foo  ", "  ba   "}, "\n")
	assert.Equal(t, res, canvas1.String())
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

func TestChinese1(t *testing.T) {
	w := New("|你|好|，|世|界|")
	c1 := w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	// Each width-2 rune takes up 2 screen cells
	assert.Equal(t, "|你|好|，|世|界|", c1.String())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
