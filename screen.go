// Copyright 2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.
//
// +build !windows

package gowid

import (
	"os"

	tcell "github.com/gdamore/tcell/v2"
)

//======================================================================

func tcellScreen(ttys string) (tcell.Screen, error) {
	var tty tcell.Tty
	var err error

	tty, err = tcell.NewDevTtyFromDev(bestTty(ttys))
	if err != nil {
		return nil, WithKVs(err, map[string]interface{}{"tty": ttys})
	}

	return tcell.NewTerminfoScreenFromTty(tty)
}

func bestTty(tty string) string {
	if tty != "" {
		return tty
	}
	gwtty := os.Getenv("GOWID_TTY")
	if gwtty != "" {
		return gwtty
	}
	return "/dev/tty"
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
