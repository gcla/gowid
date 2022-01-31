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

func tcellScreen() (tcell.Screen, error) {
	var tty tcell.Tty
	var err error

	tty, err = tcell.NewDevTtyFromDev(bestTty())
	if err != nil {
		return nil, WithKVs(err, map[string]interface{}{"GOWID_TTY": os.Getenv("GOWID_TTY")})
	}

	return tcell.NewTerminfoScreenFromTty(tty)
}

func bestTty() string {
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
