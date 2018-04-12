// Copyright 2010 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// A Left-Leaning Red-Black (LLRB) implementation of 2-3 balanced binary search trees,
// based on the following work:
//
//   http://www.cs.princeton.edu/~rs/talks/LLRB/08Penn.pdf
//   http://www.cs.princeton.edu/~rs/talks/LLRB/LLRB.pdf
//   http://www.cs.princeton.edu/~rs/talks/LLRB/Java/RedBlackBST.java
//
//  2-3 trees (and the run-time equivalent 2-3-4 trees) are the de facto standard BST
//  algoritms found in implementations of Python, Java, and other libraries. The LLRB
//  implementation of 2-3 trees is a recent improvement on the traditional implementation,
//  observed and documented by Robert Sedgewick.
//
package llrb

// Tree is a Left-Leaning Red-Black (LLRB) implementation of 2-3 trees
type LLRB struct {
	count int
	root  *Node
}

type Node struct {
	Item
	Left, Right   *Node // Pointers to left and right child nodes
	NLeft, NRight int
	Black         bool // If set, the color of the link (incoming from the parent) is black
	// In the LLRB, new nodes are always red, hence the zero-value for node
}

type Item interface {
	Less(than Item) bool
}

//
func less(x, y Item) bool {
	if x == pinf {
		return false
	}
	if x == ninf {
		return true
	}
	return x.Less(y)
}

// Inf returns an Item that is "bigger than" any other item, if sign is positive.
// Otherwise  it returns an Item that is "smaller than" any other item.
func Inf(sign int) Item {
	if sign == 0 {
		panic("sign")
	}
	if sign > 0 {
		return pinf
	}
	return ninf
}

var (
	ninf = nInf{}
	pinf = pInf{}
)

type nInf struct{}

func (nInf) Less(Item) bool {
	return true
}

type pInf struct{}

func (pInf) Less(Item) bool {
	return false
}

// New() allocates a new tree
func New() *LLRB {
	return &LLRB{}
}

// SetRoot sets the root node of the tree.
// It is intended to be used by functions that deserialize the tree.
func (t *LLRB) SetRoot(r *Node) {
	t.root = r
}

// Root returns the root node of the tree.
// It is intended to be used by functions that serialize the tree.
func (t *LLRB) Root() *Node {
	return t.root
}

// Len returns the number of nodes in the tree.
func (t *LLRB) Len() int { return t.count }

// Has returns true if the tree contains an element whose order is the same as that of key.
func (t *LLRB) Has(key Item) bool {
	return t.Get(key) != nil
}

// Get retrieves an element from the tree whose order is the same as that of key.
func (t *LLRB) Get(key Item) Item {
	h := t.root
	for h != nil {
		switch {
		case less(key, h.Item):
			h = h.Left
		case less(h.Item, key):
			h = h.Right
		default:
			return h.Item
		}
	}
	return nil
}

// Min returns the minimum element in the tree.
func (t *LLRB) Min() Item {
	h := t.root
	if h == nil {
		return nil
	}
	for h.Left != nil {
		h = h.Left
	}
	return h.Item
}

// Max returns the maximum element in the tree.
func (t *LLRB) Max() Item {
	h := t.root
	if h == nil {
		return nil
	}
	for h.Right != nil {
		h = h.Right
	}
	return h.Item
}

func (t *LLRB) ReplaceOrInsertBulk(items ...Item) {
	for _, i := range items {
		t.ReplaceOrInsert(i)
	}
}

func (t *LLRB) InsertNoReplaceBulk(items ...Item) {
	for _, i := range items {
		t.InsertNoReplace(i)
	}
}

// ReplaceOrInsert inserts item into the tree. If an existing element has the
// same order, it is removed from the tree and returned. Returns the replaced
// item, if any, and the item's position from the smallest item in the tree.
func (t *LLRB) ReplaceOrInsert(item Item) (Item, int) {
	if item == nil {
		panic("inserting nil item")
	}
	var replaced Item
	var pos int
	t.root, replaced, pos = t.replaceOrInsert(t.root, item, 0)
	t.root.Black = true
	if replaced == nil {
		t.count++
	}
	return replaced, pos
}

func (t *LLRB) replaceOrInsert(h *Node, item Item, n int) (*Node, Item, int) {
	if h == nil {
		return newNode(item), nil, n
	}

	h = walkDownRot23(h)

	var replaced Item
	var pos int
	if less(item, h.Item) { // BUG
		h.Left, replaced, pos = t.replaceOrInsert(h.Left, item, n)
	} else if less(h.Item, item) {
		h.Right, replaced, pos = t.replaceOrInsert(h.Right, item, n+1+h.NLeft)
	} else {
		replaced, h.Item, pos = h.Item, item, n
	}

	h = walkUpRot23(h)

	return h, replaced, pos
}

// InsertNoReplace inserts item into the tree. If an existing element has the
// same order, both elements remain in the tree. Returns the position of the
// inserted item from the smallest item in the tree.
func (t *LLRB) InsertNoReplace(item Item) int {
	if item == nil {
		panic("inserting nil item")
	}
	var pos int
	t.root, pos = t.insertNoReplace(t.root, item, 0)
	t.root.Black = true
	t.count++
	return pos
}

func (t *LLRB) insertNoReplace(h *Node, item Item, n int) (*Node, int) {
	if h == nil {
		return newNode(item), n
	}

	h = walkDownRot23(h)

	var pos int
	if less(item, h.Item) {
		h.Left, pos = t.insertNoReplace(h.Left, item, n)
		h.NLeft++
	} else {
		h.Right, pos = t.insertNoReplace(h.Right, item, n+1+h.NLeft)
		h.NRight++
	}

	return walkUpRot23(h), pos
}

// Rotation driver routines for 2-3 algorithm

func walkDownRot23(h *Node) *Node { return h }

func walkUpRot23(h *Node) *Node {
	if isRed(h.Right) && !isRed(h.Left) {
		h = rotateLeft(h)
	}

	if isRed(h.Left) && isRed(h.Left.Left) {
		h = rotateRight(h)
	}

	if isRed(h.Left) && isRed(h.Right) {
		flip(h)
	}

	return h
}

// Rotation driver routines for 2-3-4 algorithm

func walkDownRot234(h *Node) *Node {
	if isRed(h.Left) && isRed(h.Right) {
		flip(h)
	}

	return h
}

func walkUpRot234(h *Node) *Node {
	if isRed(h.Right) && !isRed(h.Left) {
		h = rotateLeft(h)
	}

	if isRed(h.Left) && isRed(h.Left.Left) {
		h = rotateRight(h)
	}

	return h
}

// DeleteMin deletes the minimum element in the tree and returns the
// deleted item or nil otherwise.
func (t *LLRB) DeleteMin() Item {
	var deleted Item
	t.root, deleted = deleteMin(t.root)
	if t.root != nil {
		t.root.Black = true
	}
	if deleted != nil {
		t.count--
	}
	return deleted
}

// deleteMin code for LLRB 2-3 trees
func deleteMin(h *Node) (*Node, Item) {
	if h == nil {
		return nil, nil
	}
	if h.Left == nil {
		return nil, h.Item
	}

	if !isRed(h.Left) && !isRed(h.Left.Left) {
		h = moveRedLeft(h)
	}

	var deleted Item
	h.Left, deleted = deleteMin(h.Left)
	if deleted != nil {
		h.NLeft--
	}

	return fixUp(h), deleted
}

// DeleteMax deletes the maximum element in the tree and returns
// the deleted item or nil otherwise
func (t *LLRB) DeleteMax() Item {
	var deleted Item
	t.root, deleted = deleteMax(t.root)
	if t.root != nil {
		t.root.Black = true
	}
	if deleted != nil {
		t.count--
	}
	return deleted
}

func deleteMax(h *Node) (*Node, Item) {
	if h == nil {
		return nil, nil
	}
	if isRed(h.Left) {
		h = rotateRight(h)
	}
	if h.Right == nil {
		return nil, h.Item
	}
	if !isRed(h.Right) && !isRed(h.Right.Left) {
		h = moveRedRight(h)
	}
	var deleted Item
	h.Right, deleted = deleteMax(h.Right)
	if deleted != nil {
		h.NRight--
	}

	return fixUp(h), deleted
}

// Delete deletes an item from the tree whose key equals key. Returns the
// deleted item, if any matches, and its position from the smallest item in the
// tree.
func (t *LLRB) Delete(key Item) (deleted Item, pos int) {
	t.root, deleted, pos = t.delete(t.root, key, 0)
	if t.root != nil {
		t.root.Black = true
	}
	if deleted != nil {
		t.count--
	}
	return deleted, pos
}

func (t *LLRB) delete(h *Node, item Item, n int) (*Node, Item, int) {
	var deleted Item
	if h == nil {
		return nil, nil, n
	}
	var pos int
	if less(item, h.Item) {
		if h.Left == nil { // item not present. Nothing to delete
			return h, nil, -1
		}
		if !isRed(h.Left) && !isRed(h.Left.Left) {
			h = moveRedLeft(h)
		}
		h.Left, deleted, pos = t.delete(h.Left, item, n)
		if deleted != nil {
			h.NLeft--
		}
	} else {
		if isRed(h.Left) {
			h = rotateRight(h)
		}
		// If @item equals @h.Item and no right children at @h
		if !less(h.Item, item) && h.Right == nil {
			return nil, h.Item, n + h.NLeft
		}
		// PETAR: Added 'h.Right != nil' below
		if h.Right != nil && !isRed(h.Right) && !isRed(h.Right.Left) {
			h = moveRedRight(h)
		}
		// If @item equals @h.Item, and (from above) 'h.Right != nil'
		if !less(h.Item, item) {
			var subDeleted Item
			h.Right, subDeleted = deleteMin(h.Right)
			if subDeleted == nil {
				panic("logic")
			}
			deleted, h.Item, pos = h.Item, subDeleted, n+h.NLeft
		} else { // Else, @item is bigger than @h.Item
			h.Right, deleted, pos = t.delete(h.Right, item, n+1+h.NLeft)
		}
		if deleted != nil {
			h.NRight--
		}
	}

	return fixUp(h), deleted, pos
}

// Internal node manipulation routines

func newNode(item Item) *Node { return &Node{Item: item} }

func isRed(h *Node) bool {
	if h == nil {
		return false
	}
	return !h.Black
}

func rotateLeft(h *Node) *Node {
	x := h.Right
	if x.Black {
		panic("rotating a black link")
	}
	h.Right = x.Left
	x.Left = h
	x.Black = h.Black
	h.Black = false

	h.NRight = h.Right.Len()
	x.NLeft = x.Left.Len()

	return x
}

func rotateRight(h *Node) *Node {
	x := h.Left
	if x.Black {
		panic("rotating a black link")
	}
	h.Left = x.Right
	x.Right = h
	x.Black = h.Black
	h.Black = false

	h.NLeft = h.Left.Len()
	x.NRight = x.Right.Len()

	return x
}

func (h *Node) Len() int {
	if h == nil {
		return 0
	}
	return h.NLeft + 1 + h.NRight
}

// REQUIRE: Left and Right children must be present
func flip(h *Node) {
	h.Black = !h.Black
	h.Left.Black = !h.Left.Black
	h.Right.Black = !h.Right.Black
}

// REQUIRE: Left and Right children must be present
func moveRedLeft(h *Node) *Node {
	flip(h)
	if isRed(h.Right.Left) {
		h.Right = rotateRight(h.Right)
		h = rotateLeft(h)
		flip(h)
	}
	return h
}

// REQUIRE: Left and Right children must be present
func moveRedRight(h *Node) *Node {
	flip(h)
	if isRed(h.Left.Left) {
		h = rotateRight(h)
		flip(h)
	}
	return h
}

func fixUp(h *Node) *Node {
	if isRed(h.Right) {
		h = rotateLeft(h)
	}

	if isRed(h.Left) && isRed(h.Left.Left) {
		h = rotateRight(h)
	}

	if isRed(h.Left) && isRed(h.Right) {
		flip(h)
	}

	return h
}
