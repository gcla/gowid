// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package framed

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/text"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//======================================================================

func TestCanvas8(t *testing.T) {
	widget1 := text.New("hello")
	opts := Options{
		Frame: FrameRunes{'你', '你', '你', '你', '-', '-', '你', '你'},
	}
	fwidget1 := New(widget1, opts)
	canvas1 := fwidget1.Render(gowid.RenderFixed{}, gowid.NotSelected, gwtest.D)
	res := strings.Join([]string{"你-----你", "你hello你", "你-----你"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas9(t *testing.T) {
	widget1 := text.New("hello transubstantiation good")
	fwidget1 := New(widget1)
	canvas1 := fwidget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget9 is %v", fwidget1)
	log.Infof("Canvas9 is %s", canvas1.String())
	res := strings.Join([]string{"-----------", "|hello tra|", "|nsubstant|", "|iation go|", "|od       |", "-----------"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

func TestCanvas10(t *testing.T) {
	widget1 := text.New("hello transubstantiation good")
	fwidget1 := New(widget1)
	canvas1 := fwidget1.Render(gowid.RenderBox{C: 7, R: 5}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget10 is %v", fwidget1)
	log.Infof("Canvas10 is %s", canvas1.String())
	res := strings.Join([]string{"-------", "|hello|", "| tran|", "|subst|", "-------"}, "\n")
	assert.Equal(t, res, canvas1.String())
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
