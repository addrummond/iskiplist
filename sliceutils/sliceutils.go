// Package sliceutils is an internal package providing various utility
// functions, some of which are used only in tests.
package sliceutils

import (
	"fmt"

	"github.com/addrummond/iskiplist/pcg"
)

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

const (
	randSeed1 = 12345
	randSeed2 = 67891
)

var randState *pcg.Pcg32

func GenOps(n int) []Op {
	if randState == nil {
		randState = pcg.NewPCG32()
		randState.Seed(randSeed1, randSeed2)
	}

	ops := make([]Op, n)
	slLen := 0
	for i := 0; i < n; i++ {
		r := randState.Random()
		if slLen == 0 || r < ^uint32(0)/3 {
			ops[i].Kind = OpInsert
			ops[i].Elem = int(r)
			if ops[i].Elem != 0 {
				ops[i].Elem %= 100
			}
			if slLen == 0 {
				ops[i].Index1 = 0
			} else {
				ops[i].Index1 = int(r) % slLen
			}
			slLen++
		} else if slLen >= 1 || r < (^uint32(0)/3)*2 {
			ops[i].Kind = OpSwap
			ops[i].Index1 = int(r) % slLen
			ops[i].Index2 = int(randState.Random()) % slLen
		} else {
			ops[i].Kind = OpRemove
			ops[i].Index1 = int(r) % slLen
			slLen--
		}
	}

	return ops
}
