// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// The first example from the gowid tutorial.
package main

import (
	"github.com/gcla/gowid"
	"github.com/gcla/gowid/examples"
	"github.com/gcla/gowid/widgets/text"
)

//======================================================================

func main() {
	txt := text.New("hello world")
	app, err := gowid.NewApp(gowid.AppArgs{View: txt})
	examples.ExitOnErr(err)
	app.SimpleMainLoop()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
