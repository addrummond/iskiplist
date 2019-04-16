// sliceutils is an internal package providing various utility functions, some
// of which are used only in tests.
package sliceutils

import "fmt"

type elemType = int

func SliceInsert(a *[]elemType, index int, elem elemType) {
	if len(*a) == 0 && index == 0 {
		*a = append(*a, elem)
	} else {
		last := (*a)[len(*a)-1]
		for i := len(*a) - 1; i > index; i-- {
			(*a)[i] = (*a)[i-1]
		}
		(*a)[index] = elem
		*a = append(*a, last)
	}
}

func SliceRemove(a *[]elemType, index int) {
	for i := index; i < len(*a)-1; i++ {
		(*a)[i] = (*a)[i+1]
	}
	*a = (*a)[:len(*a)-1]
}

func SliceSwap(a *[]elemType, index1, index2 int) {
	(*a)[index1], (*a)[index2] = (*a)[index2], (*a)[index1]
}

type OpKind int

const (
	OpInsert = iota
	OpRemove
	OpSwap
)

type Op struct {
	Kind   OpKind
	Index1 int
	Index2 int
	Elem   elemType
}

func ApplyOpToSlice(op *Op, a *[]elemType) {
	switch op.Kind {
	case OpInsert:
		SliceInsert(a, op.Index1, op.Elem)
	case OpRemove:
		SliceRemove(a, op.Index1)
	case OpSwap:
		SliceSwap(a, op.Index1, op.Index2)
	}
}

func PrintOp(op *Op) string {
	switch op.Kind {
	case OpInsert:
		return fmt.Sprintf("Insert %v at index %v\n", op.Elem, op.Index1)
	case OpRemove:
		return fmt.Sprintf("Remove element at index %v\n", op.Index1)
	case OpSwap:
		return fmt.Sprintf("Swap element at index %v with element at index %v\n", op.Index1, op.Index2)
	default:
		panic("Unrecognized op")
	}
}
