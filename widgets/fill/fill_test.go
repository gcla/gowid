// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package fill

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSolidFill1(t *testing.T) {
	w := New('G')

	c1 := w.Render(gowid.RenderFlowWith{C: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "GGGGG")

	c2 := w.Render(gowid.RenderBox{C: 3, R: 3}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "GGG\nGGG\nGGG")

	assert.Panics(t, func() {
		w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	})

	gwtest.RenderBoxManyTimes(t, w, 0, 10, 0, 10)
}

func TestCanvas21(t *testing.T) {
	widget1 := New('x')
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is %s", canvas1.String())
	res := strings.Join([]string{"xxxxxx"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas22(t *testing.T) {
	widget1 := New('x')
	canvas1 := widget1.Render(gowid.RenderBox{C: 6, R: 3}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is %s", canvas1.String())
	res := strings.Join([]string{"xxxxxx", "xxxxxx", "xxxxxx"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas29(t *testing.T) {
	widget1 := New('x')
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	canvas2 := gowid.NewCanvasOfSize(6, 1)
	canvas2.Lines[0][1] = canvas2.Lines[0][1].WithRune('#')
	canvas1.MergeUnder(canvas2, 0, 0, false)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is %s", canvas1.String())
	res := strings.Join([]string{"x#xxxx"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
