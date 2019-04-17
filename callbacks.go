// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package gowid

import (
	"sync"
)

//======================================================================

type ClickCB struct{}
type KeyPressCB struct{}
type SubWidgetCB struct{}
type SubWidgetsCB struct{}
type DimensionsCB struct{}
type FocusCB struct{}
type VAlignCB struct{}
type HAlignCB struct{}
type HeightCB struct{}
type WidthCB struct{}

// ICallback represents any object that can provide a way to be compared to others,
// and that can be called with an arbitrary number of arguments returning no result.
// The comparison is expected to be used by having the callback object provide a name
// to identify the callback operation e.g. "buttonclicked", so that it can later
// be removed.
type ICallback interface {
	IIdentity
	Call(args ...interface{})
}

type CallbackFunction func(args ...interface{})

type CallbackID struct {
	Name interface{}
}

// Callback is a simple implementation of ICallback.
type Callback struct {
	Name interface{}
	CallbackFunction
}

func (f CallbackFunction) Call(args ...interface{}) {
	f(args...)
}

func (f CallbackID) ID() interface{} {
	return f.Name
}

func (f Callback) ID() interface{} {
	return f.Name
}

type Callbacks struct {
	sync.Mutex
	callbacks map[interface{}][]ICallback
}

type ICallbacks interface {
	RunCallbacks(name interface{}, args ...interface{})
	AddCallback(name interface{}, cb ICallback)
	RemoveCallback(name interface{}, cb IIdentity) bool
}

func NewCallbacks() *Callbacks {
	cb := &Callbacks{}
	cb.callbacks = make(map[interface{}][]ICallback)

	var _ ICallbacks = cb

	return cb
}

// CopyOfCallbacks is used when callbacks are run - they are copied
// so that any callers modifying the callbacks themselves can do so
// safely with the modifications taking effect after all callbacks
// are run.
func (c *Callbacks) CopyOfCallbacks(name interface{}) ([]ICallback, bool) {
	c.Lock()
	defer c.Unlock()
	cbs, ok := c.callbacks[name]
	if ok {
		cbscopy := make([]ICallback, len(cbs))
		copy(cbscopy, cbs)
		return cbscopy, true
	}
	return []ICallback{}, false
}

func (c *Callbacks) RunCallbacks(name interface{}, args ...interface{}) {
	if cbs, ok := c.CopyOfCallbacks(name); ok {
		for _, cb := range cbs {
			if cb != nil {
				cb.Call(args...)
			}
		}
	}
}

func (c *Callbacks) AddCallback(name interface{}, cb ICallback) {
	c.Lock()
	defer c.Unlock()
	cbs := c.callbacks[name]
	cbs = append(cbs, cb)
	c.callbacks[name] = cbs
}

func (c *Callbacks) RemoveCallback(name interface{}, cb IIdentity) bool {
	c.Lock()
	defer c.Unlock()
	cbs, ok := c.callbacks[name]
	if ok {
		idxs := make([]int, 0)
		ok = false
		for i, cb2 := range cbs {
			if cb.ID() == cb2.ID() {
				//delete(c.callbacks, name)
				// Append backwards for easier deletion later
				idxs = append([]int{i}, idxs...)
			}
		}
		if len(idxs) > 0 {
			ok = true
			for _, j := range idxs {
				cbs = append(cbs[:j], cbs[j+1:]...)
			}
			if len(cbs) == 0 {
				delete(c.callbacks, name)
			} else {
				c.callbacks[name] = cbs
			}
		}
	}
	return ok
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
