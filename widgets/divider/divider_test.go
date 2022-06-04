// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package divider

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDivider1(t *testing.T) {
	w := New(Options{Chr: '-', Above: 0, Below: 0})

	c1 := w.Render(gowid.RenderFlowWith{C: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c1.String(), "-----")

	w2 := New(Options{Chr: 'x', Above: 1, Below: 2})

	c2 := w2.Render(gowid.RenderFlowWith{C: 5}, gowid.Focused, gwtest.D)
	assert.Equal(t, c2.String(), "     \nxxxxx\n     \n     ")

	assert.Panics(t, func() {
		w.Render(gowid.RenderBox{C: 5, R: 3}, gowid.Focused, gwtest.D)
	})

	assert.Panics(t, func() {
		w.Render(gowid.RenderFixed{}, gowid.Focused, gwtest.D)
	})

}

func TestCanvas20(t *testing.T) {
	widget1 := New(Options{Chr: '-', Above: 1, Below: 2})
	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 6}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget20 is %v", widget1)
	log.Infof("Canvas20 is %s", canvas1.String())
	res := strings.Join([]string{"      ", "------", "      ", "      "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
