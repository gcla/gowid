// Copyright 2019 Graham Clark. All rights reserved.  Use of this source
// code is governed by the MIT license that can be found in the LICENSE
// file.

// Package tree is a widget that displays a collection of widgets organized in a tree structure.
package tree

import (
	"fmt"
	"strings"

	"github.com/gcla/gowid"
	"github.com/gcla/gowid/gwutil"
	"github.com/gcla/gowid/widgets/list"
	lru "github.com/hashicorp/golang-lru"
)

//======================================================================

type IModel interface {
	Leaf() string
	Children() IIterator
	fmt.Stringer
}

type IIterator interface {
	Value() IModel
	Next() bool
}

type IExpandedCallback interface {
	Expanded(app gowid.IApp)
}

type ExpandedFunction func(app gowid.IApp)

func (f ExpandedFunction) Expanded(app gowid.IApp) {
	f(app)
}

type ICollapsedCallback interface {
	Collapsed(app gowid.IApp)
}

type CollapsedFunction func(app gowid.IApp)

func (f CollapsedFunction) Collapsed(app gowid.IApp) {
	f(app)
}

type ICollapsible interface {
	IModel
	IsCollapsed() bool
	SetCollapsed(gowid.IApp, bool)
}

//======================================================================

// For callback registration
type Collapsed struct{}
type Expanded struct{}

//======================================================================

type iterator struct {
	current int
	tree    *Tree
}

func (i *iterator) Value() IModel {
	return i.tree.theChildren[i.current]
}

func (i *iterator) Next() bool {
	i.current++
	return i.current < len(i.tree.theChildren)
}

//======================================================================

type Tree struct {
	theLeaf     string
	theChildren []IModel
}

func NewTree(leaf string, children []IModel) *Tree {
	return &Tree{leaf, children}
}

func (t *Tree) Leaf() string {
	return t.theLeaf
}

func (t *Tree) SetLeaf(s string) {
	t.theLeaf = s
}

func (t *Tree) Children() IIterator {
	return &iterator{current: -1, tree: t}
}

func (t *Tree) GetChildren() []IModel {
	return t.theChildren
}

func (t *Tree) SetChildren(c []IModel) {
	newC := make([]IModel, len(c))
	copy(newC, c)
	t.theChildren = newC
}

func (t *Tree) String() string {
	res := t.theLeaf
	var idx int
	var i IIterator
	for idx, i = 0, t.Children(); i.Next(); idx++ {
	}
	if idx > 0 {
		res = res + "+["
		st := make([]string, idx)
		for idx, i = 0, t.Children(); i.Next(); idx++ {
			st[idx] = i.Value().String()
		}
		res = res + strings.Join(st, ",") + "]"
	}
	return res
}

//======================================================================

type collapsibleIterator struct {
	sub  IIterator
	tree *Collapsible
}

func (i *collapsibleIterator) Value() IModel {
	return i.sub.Value()
}

func (i *collapsibleIterator) Next() bool {
	return !i.tree.IsCollapsed() && i.sub.Next()
}

//======================================================================

type Collapsible struct {
	*Tree
	collapsed bool
	Callbacks *gowid.Callbacks
}

func NewCollapsible(leaf string, children []IModel) *Collapsible {
	t := NewTree(leaf, children)
	return &Collapsible{
		Tree:      t,
		Callbacks: gowid.NewCallbacks(),
	}
}

func (t *Collapsible) String() string {
	var res string
	if t.IsCollapsed() {
		res = "<C>"
	} else {
		res = "< >"
	}
	res = fmt.Sprintf("%s%s", res, t.Tree.String())
	return res
}

func (t *Collapsible) IsCollapsed() bool {
	return t.collapsed
}

func (t *Collapsible) SetCollapsed(app gowid.IApp, collapsed bool) {
	t.collapsed = collapsed
	if t.IsCollapsed() {
		t.Callbacks.RunCallbacks(Collapsed{}, app)
	} else {
		t.Callbacks.RunCallbacks(Expanded{}, app)
	}
}

func (t *Collapsible) AddOnCollapsed(name interface{}, cb ICollapsedCallback) {
	t.Callbacks.AddCallback(Collapsed{},
		gowid.Callback{name,
			gowid.CallbackFunction(
				func(args ...interface{}) {
					app := args[0].(gowid.IApp)
					cb.Collapsed(app)
				},
			),
		})
}

func (t *Collapsible) RemoveOnCollapsed(name interface{}) {
	t.Callbacks.RemoveCallback(Collapsed{}, gowid.CallbackID{Name: name})
}

func (t *Collapsible) AddOnExpanded(name interface{}, cb IExpandedCallback) {
	t.Callbacks.AddCallback(Expanded{},
		gowid.Callback{name,
			gowid.CallbackFunction(
				func(args ...interface{}) {
					app := args[0].(gowid.IApp)
					cb.Expanded(app)
				},
			),
		})
}

func (t *Collapsible) RemoveOnExpanded(name interface{}) {
	t.Callbacks.RemoveCallback(Expanded{}, gowid.CallbackID{Name: name})
}

func (t *Collapsible) Children() IIterator {
	return &collapsibleIterator{sub: t.Tree.Children(), tree: t}
}

//======================================================================

// IPos is the interface of a type that represents the position of a
// sub-tree or leaf in a tree.
//
// nil means invalid
// [] means the root of the tree
// [0] means 0th child of root
// [3] means 3rd child of root
// [1,2] means 2nd child of 1st child of root
//
type IPos interface {
	GetSubStructure(IModel) IModel
	Copy() IPos
	Indices() []int
	SetIndices([]int)
	Equal(list.IWalkerPosition) bool
	GreaterThan(list.IWalkerPosition) bool
	fmt.Stringer
}

func IsSubPosition(outer IPos, inner IPos) bool {
	oidc := outer.Indices()
	iidc := inner.Indices()
	if len(iidc) < len(oidc) {
		return false
	}
	for i := 0; i < len(oidc); i++ {
		if oidc[i] != iidc[i] {
			return false
		}
	}
	return true
}

//======================================================================

// TreePos is a simple implementation of IPos.
type TreePos struct {
	Pos []int
}

func NewPos() *TreePos {
	res := TreePos{}
	res.Pos = make([]int, 0)
	return &res
}

func NewPosExt(pos []int) *TreePos {
	res := TreePos{
		Pos: pos,
	}
	return &res
}

func (tp *TreePos) Equal(other list.IWalkerPosition) bool {
	switch o := other.(type) {
	case *TreePos:
		if (tp.Pos == nil) || (o.Pos == nil) {
			panic(gowid.InvalidTypeToCompare{LHS: tp.Pos, RHS: o.Pos})
		}

		if len(tp.Pos) != len(o.Pos) {
			return false
		}

		for i := range tp.Pos {
			if tp.Pos[i] != o.Pos[i] {
				return false
			}
		}

		return true
	default:
		panic(gowid.InvalidTypeToCompare{LHS: tp, RHS: other})
	}
}

func (tp *TreePos) GreaterThan(other list.IWalkerPosition) bool {
	switch o := other.(type) {
	case *TreePos:
		if (tp.Pos == nil) || (o.Pos == nil) {
			panic(gowid.InvalidTypeToCompare{LHS: tp.Pos, RHS: o.Pos})
		}
		for i := 0; i < gwutil.Min(len(tp.Pos), len(o.Pos)); i++ {
			// e.g. [3,4] > [3]
			if tp.Pos[i] > o.Pos[i] {
				return true
			} else if tp.Pos[i] < o.Pos[i] {
				return false
			}
		}
		if len(tp.Pos) > len(o.Pos) {
			return true
		}
		return false

	default:
		panic(gowid.InvalidTypeToCompare{LHS: tp, RHS: other})
	}
}

func (tp *TreePos) Copy() IPos {
	tpCopy := *tp
	tpCopy.Pos = make([]int, len(tp.Pos))
	copy(tpCopy.Pos, tp.Pos)
	return &tpCopy
}

func (tp *TreePos) Indices() []int {
	return tp.Pos
}

func (tp *TreePos) SetIndices(indices []int) {
	tp.Pos = make([]int, len(indices))
	copy(tp.Pos, indices)
}

func (tp *TreePos) String() string {
	return fmt.Sprintf("%v", tp.Pos)
}

//
// Returns nil if the treepos is invalid, or a tree pointer if valid
//

// GetSubStructure returns the (sub-)Tree at this position from the tree argument provided.
func (tp *TreePos) GetSubStructure(tree IModel) IModel {
	var res IModel
	indices := tp.Indices() // we won't modify any of these
	if len(indices) == 0 {
		res = tree
	} else {
		var idx int
		var it IIterator
		// Walk through immediate children of tree until we hit the sub-tree at the correct index
		for idx, it = -1, tree.Children(); idx < indices[0] && it.Next(); idx++ {
		}
		if idx == indices[0] {
			// tp with the current tree-level stripped off i.e. one deeper
			res = NewPosExt(indices[1:]).GetSubStructure(it.Value())
		}
	}
	return res
}

// ConfirmPosition returns true if there is a tree at position tp of argument tree.
func ConfirmPosition(tp IPos, tree IModel) bool {
	res := false
	if tp.GetSubStructure(tree) != nil {
		res = true
	}
	return res
}

// FirstChildPosition returns the IPos corresponding to the first child
// of tp, or nil if there are no children.
func FirstChildPosition(tp IPos, tree IModel) IPos {
	tpCopy := tp.Copy()
	tpCopy.SetIndices(append(tpCopy.Indices(), 0))
	if ConfirmPosition(tpCopy, tree) {
		return tpCopy
	} else {
		return nil
	}
}

// LastChildPosition returns the IPos corresponding to the last child
// of tp, or nil if there are no children.
func LastChildPosition(tp IPos, tree IModel) IPos {
	subTree := tp.GetSubStructure(tree)
	if subTree != nil {
		var idx int
		var i IIterator
		for idx, i = 0, subTree.Children(); i.Next(); idx++ {
		}
		if idx > 0 {
			tp2 := tp.Copy()
			tp2.SetIndices(append(tp2.Indices(), idx-1))
			return tp2
		}
	}
	return nil
}

// LastInDirection navigates the tree starting at tp and moving in the
// direction determined by the function f until the end is reached. The
// function returns the last position that was passed through.
func LastInDirection(tp IPos, tree IModel, f func(IPos, IModel) IPos) IPos {
	var cur IPos
	var nextPos IPos
	for nextPos = tp; nextPos != nil; nextPos = f(nextPos, tree) {
		cur = nextPos
	}
	return cur
}

// LastDescendant moves to the last child of the current level, and
// then to the last child of that node, and onwards until the end
// is reached.
func LastDescendant(tp IPos, tree IModel) IPos {
	var f = func(a IPos, t IModel) IPos {
		return LastChildPosition(a, t)
	}
	return LastInDirection(tp, tree, f)
}

// NextSiblingPosition returns the position of the next sibling relative to
// the position tp.
func NextSiblingPosition(tp IPos, tree IModel) IPos {
	var res IPos
	tpCopy := tp.Copy()
	indices := tpCopy.Indices()
	if len(indices) > 0 {
		indices = append(indices[:len(indices)-1], indices[len(indices)-1]+1)
		tpCopy.SetIndices(indices)
		if ConfirmPosition(tpCopy, tree) {
			res = tpCopy
		}
	}
	return res
}

// PreviousSiblingPosition returns the position of the previous sibling
// relative to the position tp.
func PreviousSiblingPosition(tp IPos, tree IModel) IPos {
	var res IPos
	indices := tp.Indices()
	if len(indices) > 0 && indices[len(indices)-1] > 0 {
		tpCopy := tp.Copy()
		indices := tpCopy.Indices()
		indices[len(indices)-1]--
		tpCopy.SetIndices(indices)
		res = tpCopy
	}
	return res
}

// ParentPosition returns the position of the parent of position tp or
// nil if tp is the root node.
func ParentPosition(tp IPos) IPos {
	indices := tp.Indices()
	if len(indices) > 1 {
		tpCopy := tp.Copy()
		indices := tpCopy.Indices()
		indices = indices[:len(indices)-1]
		tpCopy.SetIndices(indices)
		return tpCopy
	} else if len(indices) == 1 {
		return NewPos()
	} else {
		// [] means root of tree, and there's only one root
		return nil
	}
}

// NextOfKin returns the position of the sibling of the parent of tp if
// that position exists, and if not, the sibling of the parent's parent,
// and on upwards. If the next of kin is not found, nil is returned.
func NextOfKin(tp IPos, tree IModel) IPos {
	var res IPos
	parent := ParentPosition(tp)
	if parent != nil {
		res = NextSiblingPosition(parent, tree)
		if res == nil {
			res = NextOfKin(parent, tree)
		}
	}
	return res
}

// NextPosition is used to navigate the tree in a depth first
// manner. Starting at the current position, the first child is
// selected. If there isn't one, then the current node's sibling is
// selected. If there isn't one, then the "next of kin" is selected.
func NextPosition(tp IPos, tree IModel) IPos {
	var res IPos = FirstChildPosition(tp, tree)
	if res == nil {
		res = NextSiblingPosition(tp, tree)
		if res == nil {
			res = NextOfKin(tp, tree)
		}
	}
	return res
}

// PreviousPosition is used to navigate the tree backwards in a depth first
// manner.
func PreviousPosition(tp IPos, tree IModel) IPos {
	var res IPos

	prevSib := PreviousSiblingPosition(tp, tree)
	if prevSib != nil {
		res = LastDescendant(prevSib, tree)
	} else {
		res = ParentPosition(tp)
	}
	return res
}

//======================================================================

type ISearchPred interface {
	CheckNode(IModel, IPos) bool
}

type SearchPred func(IModel, IPos) bool

func (s SearchPred) CheckNode(tree IModel, pos IPos) bool {
	return s(tree, pos)
}

func DepthFirstSearch(tree IModel, fn ISearchPred) IPos {
	pos := NewPos()
	return depthFirstSearchImpl(tree, pos, fn)
}

func depthFirstSearchImpl(tree IModel, pos *TreePos, fn ISearchPred) IPos {
	if tree == nil {
		return nil
	}
	if fn.CheckNode(tree, pos) {
		return pos
	}
	cs := tree.Children()
	tpos := pos.Copy().(*TreePos)
	tpos.Pos = append(tpos.Pos, 0)
	i := 0
	for cs.Next() {
		tpos.Pos[len(tpos.Pos)-1] = i
		rpos := depthFirstSearchImpl(cs.Value(), tpos, fn)
		if rpos != nil {
			return rpos
		}
		i += 1
	}
	return nil
}

//======================================================================

type IWidgetMaker interface {
	MakeWidget(pos IPos, tree IModel) gowid.IWidget
}

type WidgetMakerFunction func(pos IPos, tree IModel) gowid.IWidget

func (f WidgetMakerFunction) MakeWidget(pos IPos, tree IModel) gowid.IWidget {
	return f(pos, tree)
}

type IDecorator interface {
	MakeDecoration(pos IPos, tree IModel, wmaker IWidgetMaker) gowid.IWidget
}

type DecoratorFunction func(pos IPos, tree IModel, wmaker IWidgetMaker) gowid.IWidget

func (f DecoratorFunction) MakeDecoration(pos IPos, tree IModel, wmaker IWidgetMaker) gowid.IWidget {
	return f(pos, tree, wmaker)
}

type ITreeWalker interface {
	Tree() IModel
	Maker() IWidgetMaker
	Decorator() IDecorator
	Focus() list.IWalkerPosition
}

type TreeWalker struct {
	tree      IModel
	pos       IPos
	maker     IWidgetMaker
	decorator IDecorator
	*gowid.Callbacks
	gowid.FocusCallbacks
}

var _ ITreeWalker = (*TreeWalker)(nil)
var _ list.IWalker = (*TreeWalker)(nil)

func NewWalker(tree IModel, pos IPos, maker IWidgetMaker, dec IDecorator) *TreeWalker {
	cb := gowid.NewCallbacks()
	res := &TreeWalker{
		tree:      tree,
		pos:       pos,
		maker:     maker,
		decorator: dec,
		Callbacks: cb,
	}
	res.FocusCallbacks = gowid.FocusCallbacks{CB: &res.Callbacks}
	return res
}

func (f *TreeWalker) Tree() IModel {
	return f.tree
}

func (f *TreeWalker) Maker() IWidgetMaker {
	return f.maker
}

func (f *TreeWalker) Decorator() IDecorator {
	return f.decorator
}

// list.IWalker
func (f *TreeWalker) At(pos list.IWalkerPosition) gowid.IWidget {
	if pos == nil {
		return nil
	}
	return WidgetAt(f, pos.(IPos))
}

func WidgetAt(walker ITreeWalker, pos IPos) gowid.IWidget {
	stree := pos.GetSubStructure(walker.Tree())
	return walker.Decorator().MakeDecoration(pos, stree, walker.Maker())
}

// list.IWalker
func (f *TreeWalker) Focus() list.IWalkerPosition {
	return f.pos
}

func (f *TreeWalker) SetFocus(pos list.IWalkerPosition, app gowid.IApp) {
	old := f.pos
	f.pos = pos.(IPos)

	if !old.Equal(f.pos) {
		// the new widget in focus
		gowid.RunWidgetCallbacks(f.Callbacks, gowid.FocusCB{}, app, f)
	}
}

type IWalkerCallback interface {
	gowid.IIdentity
	Changed(app gowid.IApp, tree ITreeWalker, data ...interface{})
}

type walkerCallbackProxy struct {
	IWalkerCallback
}

func (p walkerCallbackProxy) Call(args ...interface{}) {
	t := args[0].(gowid.IApp)
	w := args[1].(ITreeWalker)
	p.IWalkerCallback.Changed(t, w, args[2:]...)
}

type WalkerFunction func(app gowid.IApp, tree ITreeWalker)

func (f WalkerFunction) Changed(app gowid.IApp, tree ITreeWalker, data ...interface{}) {
	f(app, tree)
}

// WidgetCallback is a simple struct with a name field for IIdentity and
// that embeds a WidgetChangedFunction to be issued as a callback when a widget
// property changes.
type WalkerCallback struct {
	Name interface{}
	WalkerFunction
}

func MakeCallback(name interface{}, fn WalkerFunction) WalkerCallback {
	return WalkerCallback{
		Name:           name,
		WalkerFunction: fn,
	}
}

func (f WalkerCallback) ID() interface{} {
	return f.Name
}

func RunWalkerCallbacks(c gowid.ICallbacks, name interface{}, app gowid.IApp, data ...interface{}) {
	data2 := append([]interface{}{app}, data...)
	c.RunCallbacks(name, data2...)
}

func AddWalkerCallback(c gowid.ICallbacks, name interface{}, cb IWalkerCallback) {
	c.AddCallback(name, walkerCallbackProxy{cb})
}

func RemoveWalkerCallback(c gowid.ICallbacks, name interface{}, id gowid.IIdentity) {
	c.RemoveCallback(name, id)
}

func (t *TreeWalker) OnFocusChanged(f IWalkerCallback) {
	AddWalkerCallback(t, gowid.FocusCB{}, f)
}

func (t *TreeWalker) RemoveOnFocusChanged(f gowid.IIdentity) {
	RemoveWalkerCallback(t, gowid.FocusCB{}, f)
}

// list.IWalker
func (f *TreeWalker) Next(pos list.IWalkerPosition) list.IWalkerPosition {
	return WalkerNext(f, pos)
}

// list.IWalker
func (f *TreeWalker) Previous(pos list.IWalkerPosition) list.IWalkerPosition {
	return WalkerPrevious(f, pos)
}

//======================================================================

func WalkerNext(f ITreeWalker, pos list.IWalkerPosition) list.IWalkerPosition {
	fc := pos.(IPos)
	np := NextPosition(fc, f.Tree())
	if np != nil {
		return np
	}
	return nil
}

func WalkerPrevious(f ITreeWalker, pos list.IWalkerPosition) list.IWalkerPosition {
	fc := pos.(IPos)
	np := PreviousPosition(fc, f.Tree())
	if np != nil {
		return np
	}
	return nil
}

//======================================================================
// Could consider something more sophisticated e.g. https://godoc.org/github.com/hashicorp/golang-lru
//

// CacheKey tracks content for a given tree and its expansion state. Note tree position isn't tracked
// in case objects are inserted into the tree in such a way that the position moves.
type CacheKey struct {
	Tree IModel
	Exp  bool
}

type CachingMaker struct {
	IWidgetMaker
	cache *lru.Cache
}

type CachingMakerOptions struct {
	CacheSize int
}

func NewCachingMaker(dec IWidgetMaker, opts ...CachingMakerOptions) *CachingMaker {
	var opt CachingMakerOptions
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = CachingMakerOptions{4096}
	}

	cache, err := lru.New(opt.CacheSize)
	if err != nil {
		panic(err)
	}
	return &CachingMaker{dec, cache}
}

func (d *CachingMaker) MakeWidget(pos IPos, tree IModel) gowid.IWidget {
	exp := true
	if ct, ok := tree.(ICollapsible); ok {
		exp = !ct.IsCollapsed()
	}
	key := CacheKey{tree, exp}
	if w, ok := d.cache.Get(key); ok {
		return w.(gowid.IWidget)
	} else {
		w = d.IWidgetMaker.MakeWidget(pos, tree)
		if w != nil {
			d.cache.Add(key, w)
		}
		return w.(gowid.IWidget)
	}
}

//======================================================================

type CachingDecorator struct {
	IDecorator
	cache *lru.Cache
}

type CachingDecoratorOptions struct {
	CacheSize int
}

func NewCachingDecorator(dec IDecorator, opts ...CachingDecoratorOptions) *CachingDecorator {
	var opt CachingDecoratorOptions
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = CachingDecoratorOptions{4096}
	}

	cache, err := lru.New(opt.CacheSize)
	if err != nil {
		panic(err)
	}
	return &CachingDecorator{dec, cache}
}

func (d *CachingDecorator) MakeDecoration(pos IPos, tree IModel, wmaker IWidgetMaker) gowid.IWidget {
	exp := true
	if ct, ok := tree.(ICollapsible); ok {
		exp = !ct.IsCollapsed()
	}
	key := CacheKey{tree, exp}
	var w gowid.IWidget
	if wc, ok := d.cache.Get(key); ok {
		return wc.(gowid.IWidget)
	} else {
		w = d.IDecorator.MakeDecoration(pos, tree, wmaker)
		if w != nil {
			d.cache.Add(key, w)
		}
		return w
	}
}

//======================================================================

func New(walker list.IWalker) *list.Widget {
	res := list.New(walker)
	var _ gowid.IWidget = res

	return res
}

//======================================================================
// Local Variables:
// mode: Go
// fill-column: 110
// End:
