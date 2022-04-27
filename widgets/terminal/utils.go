// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Based heavily on vterm.py from urwid

package terminal

import (
	"fmt"

	"github.com/gcla/gowid"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/terminfo"
	log "github.com/sirupsen/logrus"
)

//======================================================================

type EventNotSupported struct {
	Event interface{}
}

var _ error = EventNotSupported{}

func (e EventNotSupported) Error() string {
	return fmt.Sprintf("Terminal input event %v of type %T not supported yet", e.Event, e.Event)
}

func pasteStart(ti *terminfo.Terminfo) []byte {
	if ti.PasteStart != "" {
		return []byte(ti.PasteStart)
	} else {
		return []byte("\x1b[200~")
	}
}

func pasteEnd(ti *terminfo.Terminfo) []byte {
	if ti.PasteEnd != "" {
		return []byte(ti.PasteEnd)
	} else {
		return []byte("\x1b[201~")
	}
}

func enablePaste(ti *terminfo.Terminfo) []byte {
	if ti.EnablePaste != "" {
		return []byte(ti.EnablePaste)
	} else {
		return []byte("\x1b[?2004h")
	}
}

func disablePaste(ti *terminfo.Terminfo) []byte {
	if ti.DisablePaste != "" {
		return []byte(ti.DisablePaste)
	} else {
		return []byte("\x1b[?2004l")
	}
}

// TCellEventToBytes converts TCell's representation of a terminal event to
// the string of bytes that would be the equivalent event according to the
// supplied Terminfo object. It returns a tuple of the byte slice
// representing the terminal event (if successful), and a bool (denoting
// success or failure). This function is used by the TerminalWidget. Its
// subprocess is connected to a tty controlled by gowid. Events from the
// user are parsed by gowid via TCell - they are then translated by this
// function before being written to the TerminalWidget subprocess's tty.
func TCellEventToBytes(ev interface{}, mouse IMouseSupport, last gowid.MouseState, paster IPaste, ti *terminfo.Terminfo) ([]byte, bool) {
	res := make([]byte, 0)
	res2 := false

	switch ev := ev.(type) {
	case *tcell.EventPaste:
		res2 = true
		if paster.PasteState() {
			// Already saw start
			res = append(res, pasteEnd(ti)...)
			paster.PasteState(false)
		} else {
			res = append(res, pasteStart(ti)...)
			paster.PasteState(true)
		}
	case *tcell.EventKey:
		if ev.Key() < ' ' {
			str := []rune{rune(ev.Key())}
			res = append(res, string(str)...)
			res2 = true
		} else {
			res2 = true
			switch ev.Key() {
			case tcell.KeyRune:
				str := []rune{ev.Rune()}
				res = append(res, string(str)...)
			case tcell.KeyCR:
				str := []rune{rune(tcell.KeyCR)}
				res = append(res, string(str)...)
			case tcell.KeyF1:
				res = append(res, ti.KeyF1...)
			case tcell.KeyF2:
				res = append(res, ti.KeyF2...)
			case tcell.KeyF3:
				res = append(res, ti.KeyF3...)
			case tcell.KeyF4:
				res = append(res, ti.KeyF4...)
			case tcell.KeyF5:
				res = append(res, ti.KeyF5...)
			case tcell.KeyF6:
				res = append(res, ti.KeyF6...)
			case tcell.KeyF7:
				res = append(res, ti.KeyF7...)
			case tcell.KeyF8:
				res = append(res, ti.KeyF8...)
			case tcell.KeyF9:
				res = append(res, ti.KeyF9...)
			case tcell.KeyF10:
				res = append(res, ti.KeyF10...)
			case tcell.KeyF11:
				res = append(res, ti.KeyF11...)
			case tcell.KeyF12:
				res = append(res, ti.KeyF12...)
			case tcell.KeyF13:
				res = append(res, ti.KeyF13...)
			case tcell.KeyF14:
				res = append(res, ti.KeyF14...)
			case tcell.KeyF15:
				res = append(res, ti.KeyF15...)
			case tcell.KeyF16:
				res = append(res, ti.KeyF16...)
			case tcell.KeyF17:
				res = append(res, ti.KeyF17...)
			case tcell.KeyF18:
				res = append(res, ti.KeyF18...)
			case tcell.KeyF19:
				res = append(res, ti.KeyF19...)
			case tcell.KeyF20:
				res = append(res, ti.KeyF20...)
			case tcell.KeyF21:
				res = append(res, ti.KeyF21...)
			case tcell.KeyF22:
				res = append(res, ti.KeyF22...)
			case tcell.KeyF23:
				res = append(res, ti.KeyF23...)
			case tcell.KeyF24:
				res = append(res, ti.KeyF24...)
			case tcell.KeyF25:
				res = append(res, ti.KeyF25...)
			case tcell.KeyF26:
				res = append(res, ti.KeyF26...)
			case tcell.KeyF27:
				res = append(res, ti.KeyF27...)
			case tcell.KeyF28:
				res = append(res, ti.KeyF28...)
			case tcell.KeyF29:
				res = append(res, ti.KeyF29...)
			case tcell.KeyF30:
				res = append(res, ti.KeyF30...)
			case tcell.KeyF31:
				res = append(res, ti.KeyF31...)
			case tcell.KeyF32:
				res = append(res, ti.KeyF32...)
			case tcell.KeyF33:
				res = append(res, ti.KeyF33...)
			case tcell.KeyF34:
				res = append(res, ti.KeyF34...)
			case tcell.KeyF35:
				res = append(res, ti.KeyF35...)
			case tcell.KeyF36:
				res = append(res, ti.KeyF36...)
			case tcell.KeyF37:
				res = append(res, ti.KeyF37...)
			case tcell.KeyF38:
				res = append(res, ti.KeyF38...)
			case tcell.KeyF39:
				res = append(res, ti.KeyF39...)
			case tcell.KeyF40:
				res = append(res, ti.KeyF40...)
			case tcell.KeyF41:
				res = append(res, ti.KeyF41...)
			case tcell.KeyF42:
				res = append(res, ti.KeyF42...)
			case tcell.KeyF43:
				res = append(res, ti.KeyF43...)
			case tcell.KeyF44:
				res = append(res, ti.KeyF44...)
			case tcell.KeyF45:
				res = append(res, ti.KeyF45...)
			case tcell.KeyF46:
				res = append(res, ti.KeyF46...)
			case tcell.KeyF47:
				res = append(res, ti.KeyF47...)
			case tcell.KeyF48:
				res = append(res, ti.KeyF48...)
			case tcell.KeyF49:
				res = append(res, ti.KeyF49...)
			case tcell.KeyF50:
				res = append(res, ti.KeyF50...)
			case tcell.KeyF51:
				res = append(res, ti.KeyF51...)
			case tcell.KeyF52:
				res = append(res, ti.KeyF52...)
			case tcell.KeyF53:
				res = append(res, ti.KeyF53...)
			case tcell.KeyF54:
				res = append(res, ti.KeyF54...)
			case tcell.KeyF55:
				res = append(res, ti.KeyF55...)
			case tcell.KeyF56:
				res = append(res, ti.KeyF56...)
			case tcell.KeyF57:
				res = append(res, ti.KeyF57...)
			case tcell.KeyF58:
				res = append(res, ti.KeyF58...)
			case tcell.KeyF59:
				res = append(res, ti.KeyF59...)
			case tcell.KeyF60:
				res = append(res, ti.KeyF60...)
			case tcell.KeyF61:
				res = append(res, ti.KeyF61...)
			case tcell.KeyF62:
				res = append(res, ti.KeyF62...)
			case tcell.KeyF63:
				res = append(res, ti.KeyF63...)
			case tcell.KeyF64:
				res = append(res, ti.KeyF64...)
			case tcell.KeyInsert:
				res = append(res, ti.KeyInsert...)
			case tcell.KeyDelete:
				res = append(res, ti.KeyDelete...)
			case tcell.KeyHome:
				res = append(res, ti.KeyHome...)
			case tcell.KeyEnd:
				res = append(res, ti.KeyEnd...)
			case tcell.KeyHelp:
				res = append(res, ti.KeyHelp...)
			case tcell.KeyPgUp:
				res = append(res, ti.KeyPgUp...)
			case tcell.KeyPgDn:
				res = append(res, ti.KeyPgDn...)
			case tcell.KeyUp:
				res = append(res, ti.KeyUp...)
			case tcell.KeyDown:
				res = append(res, ti.KeyDown...)
			case tcell.KeyLeft:
				res = append(res, ti.KeyLeft...)
			case tcell.KeyRight:
				res = append(res, ti.KeyRight...)
			case tcell.KeyBacktab:
				res = append(res, ti.KeyBacktab...)
			case tcell.KeyExit:
				res = append(res, ti.KeyExit...)
			case tcell.KeyClear:
				res = append(res, ti.KeyClear...)
			case tcell.KeyPrint:
				res = append(res, ti.KeyPrint...)
			case tcell.KeyCancel:
				res = append(res, ti.KeyCancel...)
			case tcell.KeyDEL:
				res = append(res, ti.KeyBackspace...)
			case tcell.KeyBackspace:
				res = append(res, ti.KeyBackspace...)
			default:
				res2 = false
				panic(EventNotSupported{Event: ev})
			}
		}
	case *tcell.EventMouse:
		if mouse.MouseEnabled() {
			var data string

			btnind := 0
			switch ev.Buttons() {
			case tcell.Button1:
				btnind = 0
			case tcell.Button2:
				btnind = 1
			case tcell.Button3:
				btnind = 2
			case tcell.WheelUp:
				btnind = 64
			case tcell.WheelDown:
				btnind = 65
			}

			lastind := 0
			if last.LeftIsClicked() {
				lastind = 0
			} else if last.MiddleIsClicked() {
				lastind = 1
			} else if last.RightIsClicked() {
				lastind = 2
			}

			switch ev.Buttons() {
			case tcell.Button1, tcell.Button2, tcell.Button3, tcell.WheelUp, tcell.WheelDown:
				mx, my := ev.Position()
				btn := btnind
				if (last.LeftIsClicked() && (ev.Buttons() == tcell.Button1)) ||
					(last.MiddleIsClicked() && (ev.Buttons() == tcell.Button2)) ||
					(last.RightIsClicked() && (ev.Buttons() == tcell.Button3)) {
					// assume the mouse pointer has been moved with button down, a "drag"
					btn += 32
				}
				if mouse.MouseIsSgr() {
					data = fmt.Sprintf("\033[<%d;%d;%dM", btn, mx+1, my+1)
				} else {
					data = fmt.Sprintf("\033[M%c%c%c", btn+32, mx+33, my+33)
				}
				res = append(res, data...)
				res2 = true
			case tcell.ButtonNone:
				// TODO - how to report no press?
				mx, my := ev.Position()

				if last.LeftIsClicked() || last.MiddleIsClicked() || last.RightIsClicked() {
					// 0 means left mouse button, m means released
					if mouse.MouseIsSgr() {
						data = fmt.Sprintf("\033[<%d;%d;%dm", lastind, mx+1, my+1)
					} else if mouse.MouseReportAny() {
						data = fmt.Sprintf("\033[M%c%c%c", 35, mx+33, my+33)
					}
				} else if mouse.MouseReportAny() {
					if mouse.MouseIsSgr() {
						// +32 for motion, +3 for no button
						data = fmt.Sprintf("\033[<35;%d;%dm", mx+1, my+1)
					} else {
						data = fmt.Sprintf("\033[M%c%c%c", 35+32, mx+33, my+33)
					}
				}
				res = append(res, data...)
				res2 = true
			}
		}
	default:
		log.WithField("event", ev).Info("Event not implemented")
	}
	return res, res2
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
