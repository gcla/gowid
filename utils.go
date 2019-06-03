// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gowid

import (
	"fmt"
	"strings"

	"github.com/gcla/tcell"
)

//======================================================================

type Direction int

const (
	Forwards  = Direction(1)
	Backwards = Direction(-1)
)

//======================================================================

// Unit is a one-valued type used to send a message over a channel.
type Unit struct{}

//======================================================================

type InvalidTypeToCompare struct {
	LHS interface{}
	RHS interface{}
}

var _ error = InvalidTypeToCompare{}

func (e InvalidTypeToCompare) Error() string {
	return fmt.Sprintf("Cannot compare RHS %v of type %T with LHS %v of type %T", e.RHS, e.RHS, e.LHS, e.LHS)
}

//======================================================================

type KeyValueError struct {
	Base    error
	KeyVals map[string]interface{}
}

var _ error = KeyValueError{}

func (e KeyValueError) Error() string {
	kvs := make([]string, 0, len(e.KeyVals))
	for k, v := range e.KeyVals {
		kvs = append(kvs, fmt.Sprintf("%v: %v", k, v))
	}
	return fmt.Sprintf("%s [%s]", e.Cause().Error(), strings.Join(kvs, ","))
}

func (e KeyValueError) Cause() error {
	return e.Base
}

func WithKVs(err error, kvs map[string]interface{}) error {
	return KeyValueError{
		Base:    err,
		KeyVals: kvs,
	}
}

//======================================================================

// TranslatedMouseEvent is supplied with a tcell event and an x and y
// offset - it returns a tcell mouse event that represents a horizontal and
// vertical translation.
func TranslatedMouseEvent(ev interface{}, x, y int) interface{} {
	if ev3, ok := ev.(*tcell.EventMouse); ok {
		x2, y2 := ev3.Position()
		evTr := tcell.NewEventMouse(x2+x, y2+y, ev3.Buttons(), ev3.Modifiers())
		return evTr
	} else {
		return ev
	}
}

//======================================================================

func posInMap(value string, m map[string]int) int {
	i, ok := m[value]
	if ok {
		return i
	} else {
		return -1
	}
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
