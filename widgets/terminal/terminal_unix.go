// Copyright 2019-2022 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package terminal provides a widget that functions as a unix terminal. Like urwid, it emulates
// a vt220 (roughly). Mouse support is provided. See the terminal demo for more.
//
// +build !windows

package terminal

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gcla/gowid"
	log "github.com/sirupsen/logrus"
)

//======================================================================

// gcla later todo

// PlatformWidget is a widget that hosts a terminal-based application. The user provides the
// command to run, an optional environment in which to run it, and an optional hotKey. The hotKey is
// used to "escape" from the terminal (if using only the keyboard), and serves a similar role to the
// default ctrl-b in tmux. For example, to move focus to a widget to the right, the user could hit
// ctrl-b <right>. See examples/gowid-editor for a demo.
type PlatformWidget struct {
	Cmd    *exec.Cmd
	master *os.File
}

func (w *Widget) Connected() bool {
	return w.master != nil
}

func (w *Widget) SetUnderlyingTerminalSize(width, height int) error {
	spec := &terminalSizeSpec{
		Row: uint16(height),
		Col: uint16(width),
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		w.master.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(spec)),
	)

	var err error
	if errno != 0 {
		err = errno
	}

	return err
}

func (w *Widget) Write(p []byte) (n int, err error) {
	n, err = w.master.Write(p)
	return
}

func (w *Widget) Signal(sig syscall.Signal) error {
	var err error
	if w.Cmd != nil {
		err = w.Cmd.Process.Signal(sig)
	}
	return err
}

func (w *Widget) RequestTerminate() error {
	return w.Signal(syscall.SIGTERM)
}

func (w *Widget) StartCommand(app gowid.IApp, width, height int) error {
	w.Cmd = exec.Command(w.params.Command[0], w.params.Command[1:]...)
	var err error
	var tty *os.File
	w.master, tty, err = PtyStart1(w.Cmd)
	if err != nil {
		return err
	}
	defer tty.Close()

	err = w.SetTerminalSize(width, height)
	if err != nil {
		log.WithFields(log.Fields{
			"width":  width,
			"height": height,
			"error":  err,
		}).Warn("Could not set terminal size")
	}

	err = w.Cmd.Start()
	if err != nil {
		w.master.Close()
		return err
	}

	master := w.master

	w.SetUpCallbacks(app)

	go func() {
		var wg sync.WaitGroup

		data := make([]byte, 4096)

		for {
			//data := make([]byte, 4096)
			wg.Wait()

			n, err := master.Read(data)

			if n == 0 && err == io.EOF {
				w.Cmd.Wait()
				app.Run(&appRunExt{
					fn: func(app gowid.IApp) bool {
						gowid.RunWidgetCallbacks(w.Callbacks, ProcessExited{}, app, w)
						return false
					},
				})

				break
			} else if err != nil {
				w.Cmd.Wait()
				app.Run(&appRunExt{
					fn: func(app gowid.IApp) bool {
						gowid.RunWidgetCallbacks(w.Callbacks, ProcessExited{}, app, w)
						return false
					},
				})
				break
			}

			wg.Add(1)
			app.Run(&appRunExt{
				fn: func(app gowid.IApp) bool {
					render := false
					for _, b := range data[0:n] {
						if w.canvas.ProcessByteExt(b) {
							render = true
						}
					}
					defer wg.Done()
					return render
				},
			})
		}
	}()

	return nil
}

func (w *Widget) StopCommand() {
	if w.master != nil {
		w.master.Close()
	}
}

// PtyStart1 connects the supplied Cmd's stdin/stdout/stderr to a new tty
// object. The function returns the pty and tty, and also an error which is
// nil if the operation was successful.
func PtyStart1(c *exec.Cmd) (pty2, tty *os.File, err error) {
	pty2, tty, err = pty.Open()
	if err != nil {
		return nil, nil, err
	}
	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.Setctty = true
	c.SysProcAttr.Setsid = true
	return pty2, tty, err
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
