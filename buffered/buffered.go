package bufferediskiplist

import (
	"github.com/addrummond/iskiplist"
	"github.com/addrummond/iskiplist/sliceutils"
)

type BufferedISkipList struct {
	start     []iskiplist.ElemType // reverse order
	iskiplist iskiplist.ISkipList
	end       []iskiplist.ElemType
}

// If a slice is no longer than this, then we perform all operations directly on
// the slice when possible.
const noHoldsBarredMaxLength = 64

// We don't let either 'start' or 'end' grow longer than maxSliceLength.
// This is to prevent counterintuitive performance characteristics. For example,
// imagine that a BufferedISkipList of one million elements is constructed by
// repeated use of PushBack, and that all of these elements are appended to the
// 'end' slice. A value is now inserted in the middle of the list. One would
// expect this to be a fast O(log n) operation. But in fact, the first half of
// the 'end' slice, containing half a million elements, will first have to be
// pushed onto the ISkipList. We can avoid this kind of situation by not letting
// 'start' or 'end' grow too big.
const maxSliceLength = 128

func checkStartSliceGrowth(l *BufferedISkipList) {
	if len(l.start) >= maxSliceLength {
		for _, v := range l.start { // remember that 'start' is reversed
			l.iskiplist.PushFront(v)
		}
		l.start = nil
	}
}

func checkEndSliceGrowth(l *BufferedISkipList) {
	if len(l.end) >= maxSliceLength {
		for _, v := range l.end {
			l.iskiplist.PushBack(v)
		}
		l.end = nil
	}
}

func (l *BufferedISkipList) Length() int {
	return len(l.start) + l.iskiplist.Length() + len(l.end)
}

func (l *BufferedISkipList) PushBack(elem iskiplist.ElemType) {
	checkEndSliceGrowth(l)
	l.end = append(l.end, elem)
}

func (l *BufferedISkipList) PushFront(elem iskiplist.ElemType) {
	checkStartSliceGrowth(l)
	l.start = append(l.start, elem)
}

func (l *BufferedISkipList) Swap(index1, index2 int) {
	upToEnd := len(l.start) + l.iskiplist.Length()
	if index1 >= len(l.start) && index1 < upToEnd && index2 >= len(l.start) && index2 < upToEnd {
		l.iskiplist.Swap(index1-len(l.start), index2-len(l.start))
		return
	}

	var val1, val2 *iskiplist.ElemType
	if index1 < len(l.start) {
		val1 = &l.start[len(l.start)-index1-1]
	} else if index1 < upToEnd {
		val1 = l.iskiplist.PtrAt(index1 - len(l.start))
	} else {
		val1 = &l.end[index1-upToEnd]
	}
	if index2 < len(l.start) {
		val2 = &l.start[len(l.start)-index2-1]
	} else if index2 < upToEnd {
		val2 = l.iskiplist.PtrAt(index2 - len(l.start))
	} else {
		val2 = &l.end[index2-upToEnd]
	}

	*val1, *val2 = *val2, *val1
}

func (l *BufferedISkipList) Remove(index int) {
	if index < 0 || index >= l.Length() {
		panic("Index out of range in call to 'Remove'")
	}

	if index < len(l.start) {
		sliceutils.SliceRemove(&l.start, len(l.start)-index-1)
		return
	}

	if index < len(l.start)+l.iskiplist.Length() {
		l.iskiplist.Remove(index - len(l.start))
		return
	}

	sliceutils.SliceRemove(&l.end, index-len(l.start)-l.iskiplist.Length())
}

func (l *BufferedISkipList) Insert(index int, elem iskiplist.ElemType) {
	length := l.Length()
	if index < 0 || index > length {
		panic("Index out of range in call to 'Insert'")
	}

	checkStartSliceGrowth(l)
	checkEndSliceGrowth(l)

	// trivial case: prepend
	if index == 0 {
		l.PushFront(elem)
		return
	}

	// trivial case: append
	if index == length {
		l.PushBack(elem)
		return
	}

	// insertion within 'start' where 'start' is small
	if index < len(l.start) && len(l.start) <= noHoldsBarredMaxLength {
		sliceutils.SliceInsert(&l.start, len(l.start)-index, elem)
		return
	}

	// insertion within 'end' where 'end' is small
	if index >= len(l.start)+l.iskiplist.Length() && len(l.end) <= noHoldsBarredMaxLength {
		sliceutils.SliceInsert(&l.start, index-len(l.start)-l.iskiplist.Length(), elem)
		return
	}

	// insertion within the iskiplist
	if index > len(l.start) && index < len(l.start)+l.iskiplist.Length() {
		l.iskiplist.Insert(index-len(l.start), elem)
		return
	}

	// insertion within 'start' where 'start' is large
	if index < len(l.start) {
		l.iskiplist.PushFront(elem)
		for i, j := len(l.start)-1, 0; j < index; i, j = i-1, j+1 { // remember that 'start' is reversed
			l.iskiplist.PushFront(l.start[i])
		}
		l.start = l.start[len(l.start)-index:]
		return
	}

	// insertion within 'end' where 'end' is large
	if index < len(l.start)+l.iskiplist.Length() {
		panic("Internal error in 'Insert'")
	}
	for i := 0; i < index-len(l.start)-l.iskiplist.Length(); i++ {
		l.iskiplist.PushBack(l.end[i])

	}
	l.iskiplist.PushBack(elem)
	l.end = l.end[index-len(l.start)-l.iskiplist.Length():]
}