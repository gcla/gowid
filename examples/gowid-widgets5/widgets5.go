// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// A gowid test app which exercises the edit and vpadding widgets.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/edit"
	"github.com/gcla/gowid/widgets/vpadding"
	log "github.com/sirupsen/logrus"
)

//======================================================================

func main() {

	f := examples.RedirectLogger("widgets5.log")
	defer f.Close()

	palette := gowid.Palette{}

	e1e := edit.New(edit.Options{Caption: "Name:", Text: "(1)abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzab(2)CDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCD(3)efghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdef(4)&^&^&^&^&GHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZ(5)sskdjfhskajfhskajfhksjadfhksjdfhksahdfksahdfkjsdhfkjsdhfkjshadfkshdf(6)87267823687268276382638263826382638263(7)xyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyxyx(8)ewewewewewewewewewewewewewewewewewewewew"})

	e1 := vpadding.New(e1e, gowid.VAlignTop{}, gowid.RenderWithUnits{U: 4})

	app, err := gowid.NewApp(gowid.AppArgs{
		View:    e1,
		Palette: &palette,
		Log:     log.StandardLogger(),
	})
	examples.ExitOnErr(err)

	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
