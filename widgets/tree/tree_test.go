// Copyright 2019 Graham Clark. All rights reserved.  Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package tree

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTree0(t *testing.T) {
	var _ IPos = (*TreePos)(nil)
}

func TestTreePos1(t *testing.T) {
	tp1 := TreePos{
		Pos: []int{3, 4, 5},
	}
	tp2 := TreePos{
		Pos: []int{3, 4, 3},
	}
	assert.Equal(t, true, tp1.GreaterThan(&tp2))
	assert.Equal(t, false, tp2.GreaterThan(&tp1))
}

func TestTreePos2(t *testing.T) {
	tp1 := TreePos{
		Pos: []int{3, 4, 5, 6},
	}
	tp2 := TreePos{
		Pos: []int{3, 4, 3},
	}
	assert.Equal(t, true, tp1.GreaterThan(&tp2))
	assert.Equal(t, false, tp2.GreaterThan(&tp1))
}

func TestTree1(t *testing.T) {

	// t <-> 0
	// l1,s1,l2,l3 <-> t
	// s1,l2,l3 <-> t,l1
	// l4,l5,l2,l3 <-> t,l1,s2

	leaf1 := &Tree{"leaf1", []IModel{}}
	leaf2 := &Tree{"leaf2", []IModel{}}
	leaf3 := &Tree{"leaf3", []IModel{}}
	leaf4 := &Tree{"leaf4", []IModel{}}
	leaf5 := &Tree{"leaf5", []IModel{}}
	stree1 := &Tree{"stree1", []IModel{leaf4, leaf5}}
	parent1 := &Tree{"parent1", []IModel{leaf1, stree1, leaf2, leaf3}}

	log.Infof("Tree is %s", parent1.String())

	var lpos, spos IPos
	for spos = NewPos(); spos != nil; spos = NextPosition(spos, parent1) {
		log.Infof("Cur pos is %v and tree is %v", spos.String(), spos.GetSubStructure(parent1).String())
		lpos = spos.Copy()
		log.Infof("Last pos loop was %v", lpos.String())
	}
	log.Infof("Last pos was %v", lpos.String())

	for spos = lpos; spos != nil; spos = PreviousPosition(spos, parent1) {
		log.Infof("Backwards Cur pos is %v and tree is %v", spos.String(), spos.GetSubStructure(parent1).String())
	}

	tp := NewPosExt([]int{0})
	tt := tp.GetSubStructure(parent1)

	assert.Equal(t, leaf1, tt)
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
