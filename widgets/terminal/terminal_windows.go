// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package terminal provides a widget that functions as a unix terminal. Like urwid, it emulates
// a vt220 (roughly). Mouse support is provided. See the terminal demo for more.
//
// +build windows

package terminal

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/UserExistsError/conpty"
	"github.com/gcla/gowid"

	log "github.com/sirupsen/logrus"
)

//======================================================================

// Widget is a widget that hosts a terminal-based application. The user provides the
// command to run, an optional environment in which to run it, and an optional hotKey. The hotKey is
// used to "escape" from the terminal (if using only the keyboard), and serves a similar role to the
// default ctrl-b in tmux. For example, to move focus to a widget to the right, the user could hit
// ctrl-b <right>. See examples/gowid-editor for a demo.
type PlatformWidget struct {
	Pty      *conpty.ConPty
	Ctx      context.Context
	CancelFn context.CancelFunc
}

func (w *Widget) Connected() bool {
	return w.Pty != nil
}

func (w *Widget) Write(p []byte) (n int, err error) {
	n, err = w.Pty.Write(p)
	return
}

func (w *Widget) SetUnderlyingTerminalSize(width, height int) error {
	return w.Pty.Resize(width, height)
}

func (w *Widget) RequestTerminate() error {
	if w.CancelFn == nil {
		return fmt.Errorf("Terminal widget asked to terminate but has not been started")
	}
	w.CancelFn()
	return nil
}

func (w *Widget) StartCommand(app gowid.IApp, width, height int) error {
	var err error
	w.Pty, err = conpty.Start(strings.Join(w.params.Command, " ")) // gcla later todo
	if err != nil {
		fmt.Printf("Failed to spawn a pty:  %v", err)
		return err
	}

	err = w.SetTerminalSize(width, height)
	if err != nil {
		log.WithFields(log.Fields{
			"width":  width,
			"height": height,
			"error":  err,
		}).Warn("Could not set terminal size")
	}

	w.SetUpCallbacks(app)

	w.Ctx, w.CancelFn = context.WithCancel(context.Background())

	go func() {
		// Wait for the process to exit and return the exit code. If context is canceled,
		// Wait() will return STILL_ACTIVE and an error indicating the context was canceled.
		_, err := w.Pty.Wait(w.Ctx)

		err = w.Pty.Close()
		if err != nil {
			log.Warnf("Error closing process %v: %v", w.params.Command, err)
		}

		app.Run(&appRunExt{
			fn: func(app gowid.IApp) bool {
				gowid.RunWidgetCallbacks(w.Callbacks, ProcessExited{}, app, w)
				return false
			},
		})
	}()

	go func() {
		f, err := os.OpenFile("codes.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}

		defer f.Close()

		for {
			data := make([]byte, 4096)

			n, err := w.Pty.Read(data)

			if err == nil {
				f.WriteString(string(data[0:n]))
			}

			if n == 0 && err == io.EOF {
				w.CancelFn()
				break
			} else if err != nil {
				w.CancelFn()
				break
			}

			app.Run(&appRunExt{
				fn: func(app gowid.IApp) bool {
					render := false
					for _, b := range data[0:n] {
						if w.canvas.ProcessByteExt(b) {
							render = true
						}
					}
					return render
				},
			})
		}
	}()

	return nil
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
