// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

package styled

import (
	"strings"
	"testing"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwtest"
	"github.com/gcla/gowid/widgets/text"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func TestCanvas12(t *testing.T) {
	widget1a := text.New("hello transubstantiation boy")
	widget1 := NewWithRanges(widget1a,

		[]AttributeRange{AttributeRange{0, 5, gowid.MakePaletteRef("test1notfocus")}}, []AttributeRange{AttributeRange{0, 5, gowid.MakePaletteRef("test1focus")}})

	canvas1 := widget1.Render(gowid.RenderFlowWith{C: 11}, gowid.NotSelected, gwtest.D)
	log.Infof("Widget12 is %v", widget1)
	log.Infof("Canvas12 is %s", canvas1.String())
	res := strings.Join([]string{"hello trans", "ubstantiati", "on boy     "}, "\n")
	if res != canvas1.String() {
		t.Errorf("Failed")
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
