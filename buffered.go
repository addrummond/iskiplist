package iskiplist

// If the BufferedISkipList is no longer than this, then we perform all
// operations directly on the slices when possible.
const noHoldsBarredMaxLength = 64

type BufferedISkipList struct {
	start     []ElemType // reverse order
	iskiplist ISkipList
	end       []ElemType
}

func (l *BufferedISkipList) Length() int {
	return len(l.start) + l.iskiplist.Length() + len(l.end)
}

func (l *BufferedISkipList) PushBack(elem ElemType) {
	l.end = append(l.end, elem)
}

func (l *BufferedISkipList) PushFront(elem ElemType) {
	l.start = append(l.start, elem)
}

func (l *BufferedISkipList) Insert(index int, elem ElemType) {

}
