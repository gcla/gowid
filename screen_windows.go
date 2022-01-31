// Copyright 2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.
//

package gowid

import (
	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

func tcellScreen() (tcell.Screen, error) {
	return tcell.NewScreen()
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
