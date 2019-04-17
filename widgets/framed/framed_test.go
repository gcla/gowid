// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package framed

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/text"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func TestCanvas9(t *testing.T) {
	widget1 := text.New("hello transubstantiation good")
	fwidget1 := New(widget1)
	canvas1 := fwidget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget9 is %v", fwidget1)
	log.Infof("Canvas9 is %s", canvas1.String())
	res := strings.Join([]string{"-----------", "|hello tra|", "|nsubstant|", "|iation go|", "|od       |", "-----------"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

func TestCanvas10(t *testing.T) {
	widget1 := text.New("hello transubstantiation good")
	fwidget1 := New(widget1)
	canvas1 := fwidget1.Render(gowid.RenderBox{C: 7, R: 5}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget10 is %v", fwidget1)
	log.Infof("Canvas10 is %s", canvas1.String())
	res := strings.Join([]string{"-------", "|hello|", "| tran|", "|subst|", "-------"}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
