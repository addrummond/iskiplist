// Package iskiplist provides a skip list based implementation of arrays with
// O(log n) indexing, insertion and removal. The array element type is 'int'.
// The idea is to use the int value as an index into a slice of the data
// structure of your choice. If this technique isn't applicable, it's easy
// to modify the code to use interface{} as the element type instead.
//
// Each ISkipList maintains its own pseudorandom number generator state. The
// algorithm used is PCG32. By default, seed initialization piggybacks on
// address space randomization by using the address of an ISkipList to generate
// a seed. A seed can be supplied manually via Seed() if more entropy is
// required.
//
// A cache is maintained of the index and set of nodes associated with the last
// element access. This increases the efficiency of common iteration patterns
// without introducing the complexities associated with explicit iterators.
// For example, if you iterate through every third element in an ISkipList by
// indexing using At(), then the search for each element at i+3 will begin at
// element i, not at the root of the skip list. The cache is automatically
// invalidated in the expected way by operations that mutate the ISkipList. For
// example, removing the element at i invalidates the cache for a preceding
// access of any element at index >= i.
//
// The fastest way to iterate through the elements of an ISkipList in sequence
// is to use Iterate(), IterateI(), IterateRange(), IterateRangeI(), ForAll(),
// ForAllI(), ForAllRange(), and ForAllRangeI(). These functions do the minimum
// work possible and do not update the cache.
//
// The behavior of the iteration methods mentioned in the preceding paragraph is
// unspecified if the ISkipList is mutated within the callback function.
// (Mutating the element itself is fine, you just can't insert or remove
// elements.) If you wish to mutate an ISkipList while iterating thorugh it, you
// should iterate by index.
//
// The most efficient way to build an ISkipList is to add elements sequentially
// using PushFront(). The next most efficient method is to add elements
// sequentially using PushBack(). Both run in constant time (the latter due
// to caching), but PushFront() has a lower constant overhead.
//
// Slices can often be faster in practice than more sophisticated data
// structures. The following cautionary notes should be borne in mind:
//
// 1) Inserting or removing an element in the middle of a slice is extremely
// fast for slices of a thousand elements or fewer. You will not necessarily
// see any benefit from using an ISkipList unless you are dealing with sequences
// of more than 1000 elements. Once you get up to around 10,000 elements,
// insertions and removals targetting the middle of an ISkipList are much
// faster. (The precise numbers of course depend on architecture and other
// factors.)
//
// 2) It takes much longer to create an ISkipList by sequentially appending
// elements than it does to do the same with a slice. Profiling suggests that
// most of the additional time is spent allocating the list nodes. Thus, if you
// are creating a list in sequence and then performing only a small number of
// insertion/removal operations on it, you might find that the total time is
// dominated by creation time.
//
package iskiplist

import (
	"fmt"
	"strings"
	// 'unsafe' is used only to get integer values from pointers, which is not
	// actually unsafe (so long as conversion isn't performed in the other
	// direction!)
	"unsafe"
)

// This is approximately 1/e, which according to the following article is the
// optimal value for a general purpose skip list.
// https://www.sciencedirect.com/science/article/pii/030439759400296U
const pWithUint32Denom = 1580030168

// We set a maximum number of levels just to guard against the possibility of
// the pseudorandom number generator going haywire. 30 levels is sufficient to
// ensure O(log n) indexing for any ISkipList of a realistic size. (e^30 is
// between 2^43 and 2^44.)
const maxLevels = 30

// In the interests of keeping small ISkipLists small, don't cache small
// indices.
const minIndexToCache = 8

/*
The optimal value of p for a general purpose skiplist is is approximately 1/e.
See https://github.com/sean-public/fast-skiplist and the following paper
that it references:
https://www.sciencedirect.com/science/article/pii/030439759400296U

Paste the following Python3 code into the repl to generate the table:

for _ in range(1): # dummy loop to create a block
    from math import *
    tot = 0
    for i in range(21):
        ip = round(pow(1/e, i) * (1 - 1/e) * (1 << 32))
        print(str(tot + ip) + ",")
        tot += ip

We can simulate up to 21 "coin tosses" (where heads has probability 1/e) using
a single unsigned 32-bit random number and a lookup table. If the random number
is >= the last value in the table, then a second random number has to be
generated.
*/
var pTable = [...]uint32{
	2714937127,
	3713706680,
	4081133465,
	4216302225,
	4266028033,
	4284321135,
	4291050791,
	4293526493,
	4294437253,
	4294772303,
	4294895561,
	4294940905,
	4294957586,
	4294963723,
	4294965981,
	4294966812,
	4294967118,
	4294967230,
	4294967271,
	4294967286,
	4294967292,
}

func fastSeed(l *ISkipList) {
	l.rand = *newPCG32()

	// Use the address of the ISkipList to seed the RNG. This is not ideal,
	// but it's cheap. For any given execution of any given program,
	// there'll be more variation in the lower bits of the address
	// (excluding the lowest 2/4). On the other hand, the higher bits will
	// vary more between different executions. We alternate every 4 bits
	// when splitting the pointer into two seed values, and ignore the
	// lowest 2/4 bits.
	const PtrSize = 32 << uintptr(^uintptr(0)>>63)
	s := uint64(uintptr(unsafe.Pointer(l)))
	var seed1, seed2 uint64
	if PtrSize <= 32 {
		s >>= 2
		seed1 = (s & 7) | (((s >> 8) & 7) << 4) | (((s >> 16) & 7) << 8) | (((s >> 24) & 7) << 12)
		seed2 = ((s >> 4) & 7) | (((s >> 12) & 7) << 4) | (((s >> 20) & 7) << 8) | ((s >> 28) << 12)
	} else {
		s >>= 4
		seed1 = (s & 7) | (((s >> 8) & 7) << 4) | (((s >> 16) & 7) << 8) | (((s >> 24) & 7) << 12) | (((s >> 32) & 7) << 16) | (((s >> 40) & 7) << 20) | (((s >> 48) & 7) << 24) | (((s >> 56) & 7) << 28)
		seed2 = ((s >> 4) & 7) | (((s >> 12) & 7) << 4) | (((s >> 20) & 7) << 8) | (((s >> 28) & 7) << 12) | (((s >> 36) & 7) << 16) | (((s >> 44) & 7) << 20) | (((s >> 52) & 7) << 24)
	}
	l.Seed(seed1, seed2)
}

func nTosses(l *ISkipList) int {
	// The PCG state has to be odd, so we know that it's uninitialized if the
	// state is zero.
	if l.rand.state == 0 {
		fastSeed(l)
	}

	// Note that a binary search isn't the way to go here, since the value is
	// far more likely to be < one of the first few elements of pTable. A linear
	// search probably isn't quite the probabilistically optimal algorithm, but
	// it's simple and close enough.

	r := l.rand.Random()
	for i := 0; i < len(pTable); i++ {
		if r < pTable[i] {
			return int(i)
		}
	}
	r = l.rand.Random()
	for i := 0; i+len(pTable) < maxLevels; i++ {
		if r < pTable[i] {
			return i + len(pTable)
		}
	}
	return maxLevels
}

// ElemType is the type of an element in the skip list.
type ElemType = int

// ^^ ElemType can be any type that can be converted to and from an int without
// loss. Make corresponding modifications to 'elemToDist' and 'distToElem' if
// you modify this definition.

// convert something of ElemType to a distance represented as an int
func elemToDist(e ElemType) int {
	return e
}

// convert a distance represented as an int to ElemType
func distToElem(d int) ElemType {
	return d
}

type listNode struct {
	elem      ElemType // elem if on densest level; distance to next otherwise
	next      *listNode
	nextLevel *listNode // level lists start with the sparsest level first
}

type indexCache struct {
	index       int
	prevs       []*listNode
	prevIndices []int
}

func (c *indexCache) invalidate() {
	c.index = -1
	for i := range c.prevs {
		c.prevs[i] = nil // just to stop references to deleted nodes hanging around
	}
}

func (c *indexCache) isValid() bool {
	return c.index >= 0
}

// ISkipList is an indexable skip list. It behaves like an array or slice
// (elements sequenced and accessed by index) rather than a dictionary (elements
// not sequenced and accessed by key).
type ISkipList struct {
	length  int
	nLevels int32 // number of levels - 1; int32 is more than enough for this, saves a bit of space on archs that allow 4-byte align
	root    *listNode
	rand    pcg32
	cache   *indexCache
}

// Seed seeds the random number generator used for the ISkipList. If Seed is
// called, it should be called immediately following creation of the ISkipList.
// If Seed is not called, the random number generator is automatically seeded
// using the address of the ISkipList. This works fine, but may not be
// sufficiently random if the ISkipList could be the target of adversarial
// usage.
func (l *ISkipList) Seed(seed1 uint64, seed2 uint64) {
	seed1 |= 1 // pcg algo requires seed1 (= state) to be odd
	l.rand.Seed(seed1, seed2)
}

// SeedFrom sets the pseudorandom number generator state of an ISkipList by
// copying it from another ISkipList. If SeedFrom is called, it should be called
// immediately following creation of the ISkipList.
func (l *ISkipList) SeedFrom(l2 *ISkipList) {
	l.rand = l2.rand
}

func insertAfter(node *listNode, after *listNode) {
	after.next = node.next
	node.next = after
}

// Length returns the length of an ISkipList
func (l *ISkipList) Length() int {
	return l.length
}

// Clear empties an ISkipList
func (l *ISkipList) Clear() {
	l.length = 0
	l.nLevels = 0
	l.root = nil
	l.cache = nil
}

func first(l *ISkipList) ElemType {
	var r ElemType
	n := l.root
	for n != nil {
		r = n.elem
		n = n.nextLevel
	}
	return r
}

func getTo(node *listNode, index int) *listNode {
	li := 0
	for node.nextLevel != nil {
		d := elemToDist(node.elem)
		if index >= d && node.next != nil {
			index -= d
			node = node.next
		} else {
			node = node.nextLevel
			li++
		}
	}

	for index != 0 {
		index--
		node = node.next
	}

	return node
}

func getToWithPrevIndices(node *listNode, index int, prevs []*listNode, prevIndices []int) *listNode {
	li := 0
	i := 0
	for node.nextLevel != nil {
		prevs[li] = node
		prevIndices[li] = i
		d := elemToDist(node.elem)
		if index-i >= d && node.next != nil {
			i += d
			node = node.next
		} else {
			node = node.nextLevel
			li++
		}
	}

	for i < index {
		i++
		node = node.next
	}

	return node
}

func copyToCache(l *ISkipList, index int, prevs []*listNode, prevIndices []int) {
	if l.cache == nil {
		l.cache = &indexCache{
			index:       index,
			prevs:       make([]*listNode, len(prevs)),
			prevIndices: make([]int, len(prevIndices)),
		}
		copy(l.cache.prevs, prevs)
		copy(l.cache.prevIndices, prevIndices)
		return
	}

	dp := len(l.cache.prevs) - len(prevs)
	if dp < 0 {
		for i := dp; i < 0; i++ {
			l.cache.prevs = append(l.cache.prevs, nil)
		}
	} else if dp > 0 {
		l.cache.prevs = l.cache.prevs[:len(prevs)]
	}

	dpi := len(l.cache.prevIndices) - len(prevIndices)
	if dpi < 0 {
		for i := dpi; i < 0; i++ {
			l.cache.prevIndices = append(l.cache.prevIndices, 0)
		}
	} else if dpi > 0 {
		l.cache.prevIndices = l.cache.prevIndices[:len(prevIndices)]
	}

	l.cache.index = index
	copy(l.cache.prevs, prevs)
	copy(l.cache.prevIndices, prevIndices)
}

func retrieve(l *ISkipList, i int) *listNode {
	if i < minIndexToCache {
		return getTo(l.root, i)
	}

	// Some of the copying in subsequent code is in the service of ensuring
	// that these values are stack allocated. (We don't want to heap allocate
	// two arrays every time the list is indexed!)
	prevs := make([]*listNode, l.nLevels)
	prevIndices := make([]int, l.nLevels)

	var node *listNode
	if l.cache != nil && l.cache.isValid() && l.cache.index <= i {
		p := l.root
		pi := 0
		if len(l.cache.prevs) > 0 {
			p = l.cache.prevs[0]
			pi = l.cache.prevIndices[0]
		}
		node = getToWithPrevIndices(p, i-pi, prevs, prevIndices)

		for j := range prevIndices {
			prevIndices[j] += l.cache.prevIndices[0]
		}
	} else {
		node = getToWithPrevIndices(l.root, i, prevs, prevIndices)
	}

	copyToCache(l, i, prevs, prevIndices)

	return node
}

// Copy copies the ISkipList. It does not rerandomize. Seed() and SeedFrom()
// may be called on the result prior to any other operations. The cache of the
// ISkipList is not copied.
func (l *ISkipList) Copy() *ISkipList {
	oldLRoot := l.root
	var newRoot *listNode
	var aboveN, oldAboveN *listNode
	for oldLRoot != nil { // one for each level
		oldn := oldLRoot
		var newn, prevNewn, newL *listNode

		for oldn != nil {
			cp := *oldn
			newn = &cp

			if newRoot == nil {
				newRoot = newn
			}
			if newL == nil {
				newL = newn
			}

			if prevNewn != nil {
				prevNewn.next = newn
			}
			prevNewn = newn

			if oldAboveN != nil && oldAboveN.nextLevel == oldn {
				aboveN.nextLevel = newn
				aboveN = aboveN.next
				oldAboveN = oldAboveN.next
			}

			oldn = oldn.next
		}

		aboveN = newL
		oldAboveN = oldLRoot
		oldLRoot = oldLRoot.nextLevel
		newL = nil
	}

	return &ISkipList{
		length:  l.length,
		nLevels: l.nLevels,
		root:    newRoot,
	}
}

// CopyRange creates a new ISkipList whose contents are equal to a range of
// the original ISkipList. The 'from' argument must be >= 0 and < the length of
// the ISkipList. The 'to' argument must be >= 0 and <= the length of the
// ISkipList. If neither 'from' nor 'to' is out of bounds, but to <= from, then
// this is a no-op.
func (l *ISkipList) CopyRange(from, to int) *ISkipList {
	// TODO: This should be replaced with a specialized implementation, as for
	// Copy above.

	var nw ISkipList
	for i := to - 1; i >= from; i-- {
		nw.PushFront(l.At(i))
	}
	return &nw
}

// At retrieves the element at the specified index.
func (l *ISkipList) At(i int) ElemType {
	if i < 0 || i >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", i, l))
	}

	return retrieve(l, i).elem
}

// PtrAt retrieves a pointer to the element at the specified index. This pointer
// remains valid following any subsequent operations on the ISkipList. Keeping
// a pointer to a deleted element will prevent full garbage collection of the
// associated skip list nodes.
func (l *ISkipList) PtrAt(i int) *ElemType {
	if i < 0 || i >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", i, l))
	}

	return &retrieve(l, i).elem
}

// Set updates the element at the specified index.
func (l *ISkipList) Set(i int, v ElemType) {
	if i < 0 || i >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", i, l))
	}

	retrieve(l, i).elem = v
}

// Update applies an update function to the element at the specified index.
func (l *ISkipList) Update(i int, upd func(ElemType) ElemType) {
	if i < 0 || i >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", i, l))
	}

	node := retrieve(l, i)
	node.elem = upd(node.elem)
}

// CopyRangeToSlice copies a range of the ISkipList to a slice. The 'from'
// argument must be >= 0 and < the length of the ISkipList. The 'to' argument
// must be >= 0 and <= the length of the ISkipList. If neither 'from' nor 'to'
// is out of bounds, but to <= from, then this is a no-op.
func (l *ISkipList) CopyRangeToSlice(from, to int, slice []ElemType) {
	if from < 0 || from >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", from, l))
	}
	if to < 0 || to > l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", to, l))
	}

	// Returning early for this case saves the cost of finding the 'from' node.
	if to <= from {
		return
	}

	node := retrieve(l, from)
	dist := to - from
	for i := 0; i < dist; i++ {
		slice[i] = node.elem
		node = node.next
	}
}

// CopyToSlice is a shorthand for l.CopyRangeToSlice(0, l.Length())
func (l *ISkipList) CopyToSlice(slice []ElemType) {
	l.CopyRangeToSlice(0, l.length, slice)
}

// IterateRange iterates over a range of the ISkipList and passes the supplied
// function a pointer to each element visited. The iteration is halted if the
// function returns false. The 'from' argument must be >= 0 and < the length of
// the ISkipList. The 'to' argument must be >= 0 and <= the length of the
// ISkipList. If neither 'from' nor 'to' is out of bounds, but to <= from, then
// this is a no-op. Element pointers remain valid following any subsequent
// operations on the ISkipList. Keeping a pointer to a deleted element will
// prevent full garbage collection of the associated skip list nodes.
func (l *ISkipList) IterateRange(from, to int, f func(*ElemType) bool) {
	if from < 0 || from >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", from, l))
	}
	if to < 0 || to > l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", to, l))
	}

	// Returning early for this case saves the cost of finding the 'from' node.
	if to <= from {
		return
	}

	node := retrieve(l, from)
	dist := to - from
	for i := 0; i < dist; i++ {
		if !f(&node.elem) {
			return
		}
		node = node.next
	}
}

// IterateRangeI iterates over a range of the ISkipList and passes to the
// supplied function the index of each visited element and a pointer to it. The
// iteration is halted if the function returns false. The 'from' argument must
// be >= 0 and < the length of the ISkipList. The 'to' argument must be >= 0 and
// <= the length of the ISkipList. If neither 'from' nor 'to' is out of bounds,
// but to <= from, then this is a no-op. Element pointers remain valid following
// any subsequent operations on the ISkipList. Keeping a pointer to a deleted
// element will prevent full garbage collection of the associated skip list
// nodes.
func (l *ISkipList) IterateRangeI(from, to int, f func(int, *ElemType) bool) {
	if from < 0 || from >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", from, l))
	}
	if to < 0 || to > l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", to, l))
	}

	// Returning early for this case saves the cost of finding the 'from' node.
	if to <= from {
		return
	}

	node := retrieve(l, from)
	dist := to - from
	index := from
	for i := 0; i < dist; i++ {
		if !f(index, &node.elem) {
			return
		}
		node = node.next
		index++
	}
}

// Iterate is a shorthand for l.IterateRange(0, l.Length(), f)
func (l *ISkipList) Iterate(f func(*ElemType) bool) {
	l.IterateRange(0, l.length, f)
}

// IterateI is a shorthand for l.IterateRangeI(0, l.Length(), f)
func (l *ISkipList) IterateI(f func(int, *ElemType) bool) {
	l.IterateRangeI(0, l.length, f)
}

// ForAllRange is like IterateRange except that the iteration always continues
// to the end of the specified range. This saves the bother of adding a boolean
// return value to the iteration function. Element pointers remain valid
// following any subsequent operations on the ISkipList. Keeping a pointer to a
// deleted element will prevent garbage collection of the associated skip list
// nodes.
func (l *ISkipList) ForAllRange(from, to int, f func(*ElemType)) {
	l.IterateRange(from, to, func(e *ElemType) bool {
		f(e)
		return true
	})
}

// ForAllRangeI is like IterateRangeI except that the iteration always continues
// to the end of the specified range. This saves the bother of adding a boolean
// return value to the iteration function. Element pointers remain valid
// following any subsequent operations on the ISkipList. Keeping a pointer to a
// deleted element will prevent garbage collection of the associated skip list
// nodes.
func (l *ISkipList) ForAllRangeI(from, to int, f func(int, *ElemType)) {
	l.IterateRangeI(from, to, func(i int, e *ElemType) bool {
		f(i, e)
		return true
	})
}

// ForAll is a shorthand for l.ForAllRange(0, l.Length(), f)
func (l *ISkipList) ForAll(f func(*ElemType)) {
	l.ForAllRange(0, l.length, f)
}

// ForAllI is a shorthand for l.ForAllI(0, l.Length(), f)
func (l *ISkipList) ForAllI(f func(int, *ElemType)) {
	l.ForAllRangeI(0, l.length, f)
}

// assumes that list is of length >= 2
func removeFirst(l *ISkipList) ElemType {
	// Find the highest level of the next node along, and get the densest level
	// of the root node.
	nLevels := l.nLevels
	n := l.root
	for ; n.nextLevel != nil; n = n.nextLevel {
		if n.next == nil || elemToDist(n.elem) != 1 {
			nLevels--
		}
	}

	// 'n' is now the densest level of the root
	l.root = n.next

	if l.root != nil {
		addNRootLevels(l, int(l.nLevels-nLevels))
	}

	return n.elem
}

func remove(l *ISkipList, node *listNode, prevs []*listNode) {
	nn := node.next.next // node.next can't be nil because it precedes the element to be removed
	node.next = nn
	for i := len(prevs) - 1; i >= 0; i-- { // from densest to sparsest
		p := prevs[i]
		if p.next != nil {
			d := elemToDist(p.elem) // if it's in prevs, we know it's not on the densest level, so elem is the distance
			if d == 1 {
				p.elem = p.next.elem
				pnn := p.next.next
				p.next = pnn
			} else {
				p.elem = distToElem(elemToDist(p.elem) - 1)
			}
		}
	}

	maybeShrink(l)
}

// Remove removes the element at the specified index.
func (l *ISkipList) Remove(index int) ElemType {
	if index < 0 || index >= l.length {
		panic("Index out of range in call to 'Remove'")
	}

	if l.cache != nil && l.cache.index >= index {
		l.cache.invalidate()
	}

	l.length--
	if l.length == 0 {
		v := l.root.elem
		l.root = nil
		return v
	}

	if index == 0 {
		return removeFirst(l)
	}

	prevs := make([]*listNode, l.nLevels)
	prevIndices := make([]int, l.nLevels)
	node := getToWithPrevIndices(l.root, index-1, prevs, prevIndices)
	remove(l, node, prevs)
	copyToCache(l, index-1, prevs, prevIndices)

	return node.elem
}

func singleton(elem ElemType) *listNode {
	return &listNode{
		elem: elem,
	}
}

func distance(from *listNode, to *listNode) int {
	d := 0
	for from != to {
		if from.nextLevel == nil {
			d++
		} else {
			d += elemToDist(from.elem)
		}

		if from.next != nil {
			from = from.next
		} else {
			panic("Internal error: could not find 'to' node")
		}
	}
	return d
}

func addNRootLevels(l *ISkipList, n int) {
	for i := 0; i < n; i++ {
		clone := *l.root
		l.root.nextLevel = &clone
		l.root.next = nil
		// We don't set l.root.elem, as its value (which is the distance to the
		// next node for nodes on levels other than the densest) is considered
		// meaningless when 'next' is nil.
	}
}

func addSparserLevel(l *ISkipList, prevAtLevel, node *listNode, level, index int) *listNode {
	// Make sure level exists at root
	nLevels := int(l.nLevels)
	if level > int(l.nLevels) {
		if l.cache != nil {
			l.cache.invalidate()
		}
		addNRootLevels(l, level-nLevels)
		l.nLevels = int32(level)
	}

	clone := *node
	clone.nextLevel = node
	if prevAtLevel == nil {
		l.root.next = &clone
		l.root.elem = distToElem(index)
		clone.next = nil
	} else {
		oldNext := prevAtLevel.next
		clone.next = oldNext
		prevAtLevel.next = &clone

		d := distance(prevAtLevel.nextLevel, node)
		if oldNext != nil {
			clone.elem = distToElem(elemToDist(prevAtLevel.elem) - d + 1)
		}
		prevAtLevel.elem = distToElem(d)
	}

	return &clone
}

// Whenever we remove an element from the list, we remove a level from the list
// with a certain probability.
func shrinkBy(l *ISkipList) int32 {
	levs := int32(nTosses(l))

	levelsToRemove := levs - l.nLevels + 1
	if levelsToRemove < 0 {
		return 0
	}
	return levelsToRemove
}

func maybeShrink(l *ISkipList) {
	levelsToRemove := shrinkBy(l)

	var i int32
	for i = 0; i < levelsToRemove; i++ {
		l.root = l.root.nextLevel
	}

	l.nLevels -= levelsToRemove
}

func insertAtBeginning(l *ISkipList, elem ElemType) {
	// We have to be careful with levels when inserting a node at the beginning
	// of the list. The first node must have nLevels levels. But if we
	// repeatedly insert elements at the beginning of the list, we don't want
	// to end up with every element having the same number of levels. To address
	// this, we in effect pretend that the newly inserted node was always the
	// root node, and that the old root node has just been inserted. Thus, we
	// randomly choose again the number of levels for the old root node.

	if l.length == 0 {
		l.root = singleton(elem)
		return
	}

	if l.cache != nil {
		l.cache.invalidate()
	}

	// The new root node
	var rt = &listNode{}
	for i := 0; i < int(l.nLevels); i++ {
		rt = &listNode{
			nextLevel: rt,
		}
	}

	// Figure out how many levels the previous root node should have now.
	oldl := nTosses(l)

	r := l.root
	n := rt
	for i := 0; i < int(l.nLevels)-oldl; i++ {
		n.next = r.next
		n.elem = distToElem(elemToDist(r.elem) + 1)
		r = r.nextLevel
		n = n.nextLevel
	}
	for n.nextLevel != nil {
		n.next = r
		n.elem = distToElem(1)
		r = r.nextLevel
		n = n.nextLevel
	}
	n.next = r
	n.elem = elem

	l.root = rt

	if oldl > int(l.nLevels) {
		toAdd := oldl - int(l.nLevels)
		addNRootLevels(l, toAdd)
		l.nLevels = int32(l.nLevels + int32(toAdd))
	}
}

// PushFront adds an element to the beginning of the ISkipList. PushFront runs
// in constant time.
func (l *ISkipList) PushFront(elem ElemType) {
	insertAtBeginning(l, elem)
}

// PopFront removes the first element of the list and returns it. The second
// return value is true if the list was non-empty prior to the pop. PopFront
// runs in constant time.
func (l *ISkipList) PopFront() (r ElemType, ok bool) {
	if l.length == 0 {
		return
	}
	ok = true
	r = l.Remove(0)
	return
}

// PushBack adds an element to the end of the ISkipList. PushFront should be
// preferred where applicable.
func (l *ISkipList) PushBack(elem ElemType) {
	index := l.length

	l.length++

	if index == 0 {
		insertAtBeginning(l, elem)
		return
	}

	prevs := make([]*listNode, l.nLevels)
	prevIndices := make([]int, l.nLevels)

	var node *listNode
	if l.cache != nil && l.cache.isValid() && l.cache.index <= index-1 {
		p := l.root
		pi := 0
		if len(l.cache.prevs) > 0 {
			p = l.cache.prevs[0]
			pi = l.cache.prevIndices[0]
		}

		node = getToWithPrevIndices(p, index-1-pi, prevs, prevIndices)

		for j := range prevIndices {
			prevIndices[j] += l.cache.prevIndices[0]
		}
	} else {
		node = getToWithPrevIndices(l.root, index-1, prevs, prevIndices)
	}

	if index-1 >= minIndexToCache {
		copyToCache(l, index-1, prevs, prevIndices)
	}

	if node == nil {
		panic("Internal error in 'PushBack'")
	}

	after := &listNode{
		elem: elem,
	}

	insertAfter(node, after)

	n := after
	prevsI := len(prevs) - 1
	nlev := nTosses(l)
	for i := 1; i < maxLevels && i <= nlev; i++ {
		var p *listNode
		if prevsI >= 0 {
			p = prevs[prevsI]
			prevsI--
		}
		n = addSparserLevel(l, p, n, i, index)
	}

	for ; prevsI >= 0; prevsI-- {
		prevs[prevsI].elem = distToElem(elemToDist(prevs[prevsI].elem) + 1)
	}
}

// PopBack removes the last element of the list and returns it. The second
// return value is true if the list was non-empty prior to the pop. PopFront
// should be preferred where applicable.
func (l *ISkipList) PopBack() (r ElemType, ok bool) {
	if l.length == 0 {
		return
	}
	ok = true
	r = l.Remove(l.length - 1)
	return
}

// Insert inserts an element after the element at the specified index, or at
// the end of the list if the index is equal to the length of the ISkipList.
func (l *ISkipList) Insert(index int, elem ElemType) {
	if index > l.length {
		panic("Index out of range in call to 'Insert'")
	}

	if l.cache != nil && l.cache.index > index {
		l.cache.invalidate()
	}

	l.length++

	if index == 0 {
		insertAtBeginning(l, elem)
		return
	}

	prevs := make([]*listNode, l.nLevels)
	prevIndices := make([]int, l.nLevels)

	var node *listNode
	if l.cache != nil && l.cache.isValid() && l.cache.index <= index-1 {
		p := l.root
		pi := 0
		if len(l.cache.prevs) > 0 {
			p = l.cache.prevs[0]
			pi = l.cache.prevIndices[0]
		}

		node = getToWithPrevIndices(p, index-1-pi, prevs, prevIndices)

		for j := range prevIndices {
			prevIndices[j] += l.cache.prevIndices[0]
		}
	} else {
		node = getToWithPrevIndices(l.root, index-1, prevs, prevIndices)
	}

	if index-1 >= minIndexToCache {
		copyToCache(l, index-1, prevs, prevIndices)
	}

	if node == nil {
		panic("Internal error in 'Insert'")
	}

	after := &listNode{
		elem: elem,
	}

	insertAfter(node, after)

	n := after
	prevsI := len(prevs) - 1
	nlev := nTosses(l)
	for i := 1; i < maxLevels && i <= nlev; i++ {
		var p *listNode
		if prevsI >= 0 {
			p = prevs[prevsI]
			prevsI--
		}
		n = addSparserLevel(l, p, n, i, index)
	}

	for ; prevsI >= 0; prevsI-- {
		prevs[prevsI].elem = distToElem(elemToDist(prevs[prevsI].elem) + 1)
	}
}

// Swap swaps the values of the elements at the specified indices.
func (l *ISkipList) Swap(index1, index2 int) {
	if index1 < 0 || index1 >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", index1, l))
	}
	if index2 < 0 || index2 >= l.length {
		panic(fmt.Sprintf("Out of bounds index %v into ISkipList %+v", index2, l))
	}

	if index1 == index2 {
		return
	}
	if index1 > index2 {
		index1, index2 = index2, index1
	}

	if l.cache != nil && l.cache.index > index1 {
		l.cache.invalidate()
	}

	prevs := make([]*listNode, l.nLevels)
	prevIndices := make([]int, l.nLevels)
	node1 := getToWithPrevIndices(l.root, index1, prevs, prevIndices)
	if index1 >= minIndexToCache {
		copyToCache(l, index1, prevs, prevIndices)
	}

	p := l.root
	pi := 0
	if len(prevs) > 0 {
		p = prevs[0]
		pi = prevIndices[0]
	}
	node2 := getTo(p, index2-pi)
	node1.elem, node2.elem = node2.elem, node1.elem
}

func debugPrintList(node *listNode, pointerDigits int) string {
	if node == nil {
		return "(empty)"
	}

	var s strings.Builder

	for n := node; ; n = n.next {
		s.WriteString(fmt.Sprintf("%2d", n.elem))
		if n.next != nil && n.nextLevel == nil {
			s.WriteString(fmt.Sprintf("[%2d]", elemToDist(n.elem)))
		} else {
			s.WriteString("[  ]")
		}
		p := fmt.Sprintf("%016x", uintptr(unsafe.Pointer(n)))[16-pointerDigits:]
		s.WriteString(p)

		if n.next == nil {
			break
		}

		d := 1
		if n.nextLevel != nil {
			d = elemToDist(n.elem)
		}
		for i := 1; i < d; i++ {
			s.WriteString("        ")
			for i := 0; i < pointerDigits; i++ {
				s.WriteString(" ")
			}
		}
		s.WriteString("  ")
	}
	if pointerDigits > 0 {
		s.WriteString("\n")
		for n := node; ; n = n.next {
			s.WriteString("      ")
			p := fmt.Sprintf("%016x", uintptr(unsafe.Pointer(n.nextLevel)))[16-pointerDigits:]
			s.WriteString(p)

			if n.next == nil {
				break
			}

			d := 1
			if n.nextLevel != nil {
				d = elemToDist(n.elem)
			}
			for i := 1; i < d; i++ {
				s.WriteString("        ")
				for i := 0; i < pointerDigits; i++ {
					s.WriteString(" ")
				}
			}
			s.WriteString("  ")
		}
	}

	return s.String()
}

func debugPrintISkipList(l *ISkipList, pointerDigits int) string {
	var s strings.Builder

	s.WriteString("ISkipList:\n")

	levelRoot := l.root
	for levelRoot != nil {
		s.WriteString(debugPrintList(levelRoot, pointerDigits))
		s.WriteString("\n\n")

		levelRoot = levelRoot.nextLevel
	}

	return s.String()
}
