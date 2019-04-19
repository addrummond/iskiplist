package bufferediskiplist

import (
	"fmt"
	"testing"

	"github.com/addrummond/iskiplist"
	"github.com/addrummond/iskiplist/sliceutils"
)

const (
	randSeed1 = 12345
	randSeed2 = 67891
)

func intToElem(v int) iskiplist.ElemType {
	return v
}

func applyOpToBufferedISkipList(op *sliceutils.Op, sl *BufferedISkipList) {
	switch op.Kind {
	case sliceutils.OpInsert:
		sl.Insert(op.Index1, op.Elem)
	case sliceutils.OpRemove:
		sl.Remove(op.Index1)
	case sliceutils.OpSwap:
		sl.Swap(op.Index1, op.Index2)
	}
}

func TestCopyRange(t *testing.T) {
	const l = 1000

	var sl BufferedISkipList
	sl.CopyRange(0, 0) // ensure that copying an empty range of an empty BufferedISkipList is ok

	for i := 0; i < l; i++ {
		sl.PushFront(intToElem(i))
	}

	for i := 0; i < l/2; i++ {
		t.Logf("Copying range (%v, %v)\n", i, sl.Length()-i)
		sl.CopyRange(i, i) // check that empty range copy is ok
		cp := sl.CopyRange(i, sl.Length()-i)
		for j := 0; j < sl.Length()-i*2; j++ {
			if sl.At(j+i) != cp.At(j) {
				t.Errorf("Values don't match")
			}
		}
	}
}

func TestCopyRangeToSlice(t *testing.T) {
	const l = 1000

	var sl BufferedISkipList

	emptySlice := make([]iskiplist.ElemType, 0)
	sl.CopyRangeToSlice(0, 0, emptySlice)

	for i := 0; i < l; i++ {
		sl.PushFront(intToElem(i))
	}

	slice := make([]iskiplist.ElemType, l)
	for i := 0; i < l/2; i++ {
		sl.CopyRangeToSlice(i, i, slice) // check that empty range copy is ok
		t.Logf("Copying range (%v, %v) to slice\n", i, sl.Length()-i)
		sl.CopyRangeToSlice(i, sl.Length()-i, slice)
		for j := 0; j < sl.Length()-i*2; j++ {
			if sl.At(j+i) != slice[j] {
				t.Errorf("Values don't match")
			}
		}
	}
}

// This test creates random sequences of Insert, Swap and Remove operations and
// then applies these operations to both an ISkipList and a slice. The end
// results should match.
func TestRandomOpSequences(t *testing.T) {
	const nops = 1000
	const niters = 20

	var sl BufferedISkipList
	sl.Seed(randSeed1, randSeed2)
	for i := 1; i < niters; i++ {
		t.Logf("----- Generating random sequence of %v operations -----\n", nops)
		ops := sliceutils.GenOpsWithLotsOfPushing(nops, 0)
		sl.Clear()
		a := make([]int, 0)
		for _, o := range ops {
			t.Logf("%s\n", sliceutils.PrintOp(&o))
			sliceutils.ApplyOpToSlice(&o, &a)
			applyOpToBufferedISkipList(&o, &sl)

			t.Logf("Lengths %v %v\n", len(a), sl.Length())
			if len(a) != sl.Length() {
				t.Errorf("ISkipList has wrong length (%v instead of %v)\n", sl.Length(), len(a))
			}

			// Equality check by looping over indices.
			t.Logf("Testing result via index loop...\n")
			for i, v := range a {
				t.Logf("Checking %v\n", i)
				e := sl.At(i)
				if v != e {
					t.Errorf("Expected value %v at index %v, got %v instead (index loop).\n", v, i, e)
				}
			}
		}

		t.Logf("Reported lengths: %v %v\n", sl.Length(), len(a))

		// Equality check using ForAllI
		t.Logf("Testing result via ForAllI()...")
		sl.ForAllI(func(i int, v *iskiplist.ElemType) {
			t.Logf("Checking %v\n", i)
			if *v != a[i] {
				t.Errorf("Expected value %v at index %v, got %v instead (ForAllI).\n", a[i], i, *v)
			}
		})

		// Copy and then check copy has expected elements using ForAllI.
		cp := sl.Copy()
		cp.ForAllI(func(i int, v *iskiplist.ElemType) {
			t.Logf("Checking %v\n", i)
			if *v != a[i] {
				t.Errorf("Expected value %v at index %v, got %v instead (ForAllI).\n", a[i], i, *v)
			}
		})
	}
}

func benchmarkRandomOpSequenceWithBufferedISKipList(ops []sliceutils.Op, sl *BufferedISkipList, l int) {
	for _, o := range ops {
		fmt.Printf("OP len=%v %v\n", sl.Length(), sliceutils.PrintOp(&o))
		applyOpToBufferedISkipList(&o, sl)
	}
}

func benchmarkRandomOpSequenceWithSlice(ops []sliceutils.Op, a []int, l int) {
	for _, o := range ops {
		sliceutils.ApplyOpToSlice(&o, &a)
	}
}

func BenchmarkRandomOpSequence(b *testing.B) {
	const nops = 500

	for i := 0; i < 100000; i += 1000 {
		fmt.Printf("GEN\n")
		ops := sliceutils.GenOps(nops, i)

		var sl BufferedISkipList
		sl.Seed(randSeed1, randSeed2)
		for j := 0; j < i; j++ {
			if j%2 == 0 {
				sl.PushBack(intToElem(j))
			} else {
				sl.PushFront(intToElem(j))
			}
		}
		b.Run(fmt.Sprintf("With BufferedISkipList [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				benchmarkRandomOpSequenceWithBufferedISKipList(ops, &sl, nops)
			}
		})
		fmt.Printf("SL LEN %v\n", sl.Length())

		a := make([]int, i)
		for j := 0; j < i; j++ {
			a[j] = j
		}
		b.Run(fmt.Sprintf("With slice [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {

			for j := 0; j < b.N; j++ {
				benchmarkRandomOpSequenceWithSlice(ops, a, nops)
			}
		})
	}
}
