// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package examples

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// RedirectLogger sets the global logger to write to a file in append mode, the filename
// specified by the argument to the function. This is a convenience method used in a few
// of the example programs to avoid polluting the tty used to display the application.
func RedirectLogger(path string) *os.File {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	log.SetOutput(f)
	return f
}

func ExitOnErr(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
