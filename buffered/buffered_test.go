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

func applyOpToISkipList(op *sliceutils.Op, sl *iskiplist.ISkipList) {
	switch op.Kind {
	case sliceutils.OpInsert:
		sl.Insert(op.Index1, op.Elem)
	case sliceutils.OpRemove:
		sl.Remove(op.Index1)
	case sliceutils.OpSwap:
		sl.Swap(op.Index1, op.Index2)
	}
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

func benchmarkRandomOpSequenceWithISKipList(ops []sliceutils.Op, sl *iskiplist.ISkipList, l int) {
	fmt.Printf("START\n\n")
	for _, o := range ops {
		fmt.Printf("L %v\n", sl.Length())
		fmt.Printf("%v\n", sliceutils.PrintOp(&o))
		lbefore := sl.Length()
		applyOpToISkipList(&o, sl)
		lafter := sl.Length()
		if !((o.Kind == sliceutils.OpInsert && lafter == lbefore+1) || (o.Kind == sliceutils.OpRemove && lbefore == lafter+1) || (o.Kind == sliceutils.OpSwap && lbefore == lafter)) {
			panic("NOO!")
		}
	}
}

func benchmarkRandomOpSequenceWithBufferedISKipList(ops []sliceutils.Op, sl *BufferedISkipList, l int) {
	for _, o := range ops {
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
		ops := sliceutils.GenOps(nops, i)

		a := make([]iskiplist.ElemType, i)
		for j := 0; j < i; j++ {
			a[j] = intToElem(j)
		}
		b.Run(fmt.Sprintf("With slice [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				benchmarkRandomOpSequenceWithSlice(ops, a, nops)
			}
		})

		var sl iskiplist.ISkipList
		sl.Seed(randSeed1, randSeed2)
		for j := 0; j < i; j++ {
			if j%2 == 0 {
				sl.PushBack(intToElem(j))
			} else {
				sl.PushFront(intToElem(j))
			}
		}
		b.Run(fmt.Sprintf("With ISkipList [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				benchmarkRandomOpSequenceWithISKipList(ops, &sl, nops)
			}
		})

		var slb BufferedISkipList
		slb.Seed(randSeed1, randSeed2)
		for j := 0; j < i; j++ {
			if j%2 == 0 {
				slb.PushBack(intToElem(j))
			} else {
				slb.PushFront(intToElem(j))
			}
		}
		b.Run(fmt.Sprintf("With BufferedISkipList [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				benchmarkRandomOpSequenceWithBufferedISKipList(ops, &slb, nops)
			}
		})
	}
}

func TestRandomOpSequence2(t *testing.T) {
	const nops = 10 //500
	const n = 10

	for i := 0; i < 100000; i += 1000 {
		ops := sliceutils.GenOps(nops, i)

		for j := 0; j < n; j++ {
			a := make([]iskiplist.ElemType, i)
			for j := 0; j < i; j++ {
				a[j] = intToElem(j)
			}
			benchmarkRandomOpSequenceWithSlice(ops, a, nops)
		}

		for j := 0; j < n; j++ {
			var sl iskiplist.ISkipList
			sl.Seed(randSeed1, randSeed2)
			for j := 0; j < i; j++ {
				if j%2 == 0 {
					sl.PushBack(intToElem(j))
				} else {
					sl.PushFront(intToElem(j))
				}
			}

			fmt.Printf("i = %v, len=%v\n", i, sl.Length())
			benchmarkRandomOpSequenceWithISKipList(ops, &sl, nops)
		}

		for j := 0; j < n; j++ {
			var slb BufferedISkipList
			slb.Seed(randSeed1, randSeed2)
			for j := 0; j < i; j++ {
				if j%2 == 0 {
					slb.PushBack(intToElem(j))
				} else {
					slb.PushFront(intToElem(j))
				}
			}
			benchmarkRandomOpSequenceWithBufferedISKipList(ops, &slb, nops)
		}
	}
}
