// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package progress

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//======================================================================

var (
	pcb1 int
)

func testProgressCallback1(app gowid.IApp, w gowid.IWidget) {
	pcb1++
}

func TestCallbacks2(t *testing.T) {
	widget1 := New(Options{gowid.EmptyPalette{}, gowid.EmptyPalette{}, 100, 0})
	widget1.OnSetProgress(gowid.WidgetCallback{"cb", testProgressCallback1})
	assert.Equal(t, pcb1, 0)
	widget1.SetProgress(gwtest.D, 50)
	assert.Equal(t, pcb1, 1)
}

func TestCanvas23(t *testing.T) {
	widget1 := New(Options{gowid.EmptyPalette{}, gowid.EmptyPalette{}, 100, 0})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 10}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is %s", canvas1.String())
	res := strings.Join([]string{"    0 %   "}, "\n")
	log.Infof("FOOFOO1: res is '%v'", res)
	log.Infof("FOOFOO2: c1 is '%v'", canvas1)
	log.Infof("FOOFOO3: c1s is '%v'", canvas1.String())
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
	widget1.SetProgress(gwtest.D, 50)
	canvas1 = widget1.Render(gowid.RenderFlowWith{C: 10}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is %s", canvas1.String())
	res = strings.Join([]string{"   50 %   "}, "\n")
	log.Infof("FOOFOO1: res is '%v'", res)
	log.Infof("FOOFOO2: c1 is '%v'", canvas1)
	log.Infof("FOOFOO3: c1s is '%v'", canvas1.String())
	//res = strings.Join([]string{"█████     "}, "\n")
	//res = strings.Join([]string{"xxxxx     "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
	widget1.SetProgress(gwtest.D, 100)
	canvas1 = widget1.Render(gowid.RenderFlowWith{C: 10}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget is %v", widget1)
	log.Infof("Canvas is %s", canvas1.String())
	res = strings.Join([]string{"   100 %  "}, "\n")
	log.Infof("FOOFOO1: res is '%v'", res)
	log.Infof("FOOFOO2: c1 is '%v'", canvas1)
	log.Infof("FOOFOO3: c1s is '%v'", canvas1.String())
	//res = strings.Join([]string{"xxxxx     "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
