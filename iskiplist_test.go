package iskiplist

import (
	"fmt"
	"testing"

	"github.com/addrummond/iskiplist/sliceutils"
)

const (
	randSeed1 = 12345
	randSeed2 = 67891
)

func TestCopy(t *testing.T) {
	var sl ISkipList
	sl.Seed(randSeed1, randSeed2)
	for i := 0; i < 10; i++ {
		sl.PushBack(distToElem(i))
	}
	sl2 := sl.Copy()
	t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
	t.Logf("%v\n", DebugPrintISkipList(sl2, 3))
}

func TestCopyNoops(t *testing.T) {
	var sl ISkipList
	sl.Seed(randSeed1, randSeed2)
	sl.PushBack(distToElem(1))
	sl.PushBack(distToElem(2))
	slice := make([]ElemType, 0)
	sl.CopyRange(1, 0)
	sl.CopyRangeToSlice(1, 0, slice)
}

func TestInsertAtBeginning(t *testing.T) {
	var sl ISkipList
	sl.Seed(12345, 67891) // not using randSeed1 and randSeed2 because this test depends on a particular value for the random seeds
	for i := 0; i < 10; i++ {
		sl.Insert(0, distToElem(i))
	}
	t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
	if sl.nLevels+1 != 3 {
		t.Errorf("Unexpected number of levels in result (expected 3, got %v)\n", sl.nLevels+1)
	}
}

func TestRemoveFromBeginning(t *testing.T) {
	var sl ISkipList
	sl.Seed(randSeed1, randSeed2)
	for i := 0; i < 20; i++ {
		t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
		sl.Insert(0, distToElem(i))
	}
	t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
	for i := 0; i < 20; i++ {
		if sl.Remove(0) != 20-i-1 {
			t.Errorf("Unexpected value removed.\n")
		}
		t.Logf("Removed an element:\n%v\n", DebugPrintISkipList(&sl, 3))
	}
	if sl.Length() != 0 || sl.length != 0 || sl.nLevels != 0 || sl.root != nil || sl.cache != nil {
		t.Errorf("Unexpected result following removals.\n")
	}
}

func TestRemoveAtTwo(t *testing.T) {
	var sl ISkipList
	sl.Seed(randSeed1, randSeed2)
	for i := 0; i < 20; i++ {
		t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
		sl.Insert(0, distToElem(i))
	}
	t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
	ev := 17
	for i := 0; i < 18; i++ {
		v := sl.Remove(2)
		if v != ev {
			t.Errorf("Unexpected value removed (%v vs %v).\n", v, ev)
		}
		t.Logf("Removed an element:\n%v\n", DebugPrintISkipList(&sl, 3))
		ev--
	}
	if sl.Length() != 2 {
		t.Errorf("Unexpected result following removals.\n")
	}
}

func TestTruncate(t *testing.T) {
	const l = 100000
	const tl1 = 10000
	const tl2 = 1000
	const tl3 = 100
	const tl4 = 32
	const tl5 = 2
	var sl ISkipList
	sl.Seed(12345, 67891) // not using randSeed1 and randSeed2 because this test depends on a particular value for the random seeds
	for i := 0; i < l; i++ {
		sl.PushFront(distToElem(0))
	}
	err := false
	t.Logf("Number of levels with %v elems: %v\n", l, sl.nLevels+1)
	if sl.nLevels != 10 {
		err = true
	}
	sl.Truncate(tl1)
	t.Logf("Number of levels with %v elems: %v\n", tl1, sl.nLevels+1)
	if sl.nLevels != 9 {
		err = true
	}
	sl.Truncate(tl2)
	t.Logf("Number of levels with %v elems: %v\n", tl2, sl.nLevels+1)
	if sl.nLevels != 6 {
		err = true
	}
	sl.Truncate(tl3)
	t.Logf("Number of levels with %v elems: %v\n", tl3, sl.nLevels+1)
	if sl.nLevels != 5 {
		err = true
	}
	sl.Truncate(tl4)
	t.Logf("Number of levels with %v elems: %v\n", tl4, sl.nLevels+1)
	if sl.nLevels != 5 {
		err = true
	}
	sl.Truncate(tl5)
	t.Logf("Number of levels with %v elems: %v\n", tl5, sl.nLevels+1)
	if sl.nLevels != 2 {
		err = true
	}

	t.Logf("%v\n", DebugPrintISkipList(&sl, 3))

	if err {
		t.Errorf("Unexpected number of levels.")
	}
}

// TestCreateAndIter creates some ISkipLists using Insert and runs some simple
// tests of the basic ISkipList operations.
func TestCreateAndIter(t *testing.T) {
	type insert struct {
		index int
		value ElemType
	}
	type tst struct {
		inserts []insert
		result  []ElemType
	}

	tsts := []tst{
		{[]insert{{0, distToElem(0)}, {1, distToElem(1)}, {2, distToElem(2)}, {3, distToElem(3)}, {4, distToElem(4)}, {5, distToElem(5)}, {6, distToElem(6)}, {7, distToElem(7)}, {8, distToElem(8)}, {9, distToElem(9)}, {10, distToElem(10)}},
			[]ElemType{distToElem(0), distToElem(1), distToElem(2), distToElem(3), distToElem(4), distToElem(5), distToElem(6), distToElem(7), distToElem(8), distToElem(9), distToElem(10)},
		},
		{[]insert{{0, distToElem(10)}, {0, distToElem(9)}, {0, distToElem(8)}, {0, distToElem(7)}, {0, distToElem(6)}, {0, distToElem(5)}, {0, distToElem(4)}, {0, distToElem(3)}, {0, distToElem(2)}, {0, distToElem(1)}, {0, distToElem(0)}},
			[]ElemType{distToElem(0), distToElem(1), distToElem(2), distToElem(3), distToElem(4), distToElem(5), distToElem(6), distToElem(7), distToElem(8), distToElem(9), distToElem(10)},
		},
	}

	for _, ts := range tsts {
		var sl ISkipList
		sl.Seed(randSeed1, randSeed2)

		for _, ins := range ts.inserts {
			t.Logf("Inserting %v at %v\n", ins.value, ins.index)
			sl.Insert(ins.index, ins.value)
		}

		if sl.Length() != len(ts.result) {
			t.Errorf("Error mismatch: ISKipList has length %v; expected result has length %v\n", sl.Length(), len(ts.result))
		}

		// Test iterating through the list using At()
		for i, v := range ts.result {
			t.Logf("At %v\n", i)
			slv := sl.At(i)
			if slv != v {
				t.Errorf("ISkipList[%v] = %v, expectedResult[%v] = %v\n", i, slv, i, v)
			}
		}

		// Test iterating through the by copying it to a slice.
		cpy := make([]ElemType, len(ts.result))
		sl.CopyToSlice(cpy)
		for i, v := range cpy {
			if v != ts.result[i] {
				t.Errorf("sliceCopy[%v] = %v, expectedResult[%v] = %v\n", i, v, i, ts.result[i])
			}
		}

		// Test iterating through part of the list by copying it to a slice.
		middle := make([]ElemType, len(ts.result)-4)
		sl.CopyRangeToSlice(2, sl.Length()-2, middle)
		for i, v := range middle {
			if v != ts.result[i+2] {
				t.Errorf("middle[%v] = %v, expectedResult[%v] = %v\n", i, v, i+2, ts.result[i+2])
			}
		}

		// Test iterating through the list using Iterate()
		i := 0
		sl.Iterate(func(e *ElemType) bool {
			if *e != ts.result[i] {
				t.Errorf("Expected value %v in iteration, got %v at index %v\n", ts.result[i], *e, i)
			}
			i++
			return true
		})
		i = 0
		sl.IterateI(func(j int, e *ElemType) bool {
			if *e != ts.result[i] {
				t.Errorf("Expected value %v in iteration, got %v at index %v\n", ts.result[i], *e, i)
			}
			if i != j {
				t.Errorf("Unexpected index in iteration: %v vs. %v\n", i, j)
			}
			i++
			return true
		})
	}
}

// TestInsertAndSwap runs a simple test of the Insert() and Swap() methods.
func TestInsertAndSwap(t *testing.T) {
	var expected = []ElemType{
		distToElem(0), distToElem(1), distToElem(99), distToElem(99), distToElem(4), distToElem(88), distToElem(2), distToElem(3), distToElem(88), distToElem(5), distToElem(6), distToElem(7), distToElem(8), distToElem(9),
	}

	var sl ISkipList
	sl.Seed(randSeed1, randSeed2)
	for i := 0; i < 10; i++ {
		t.Logf("Inserting %v\n", i)
		sl.Insert(i, distToElem(i))
		t.Logf("%s\n", DebugPrintISkipList(&sl, 3))
	}
	for i := 0; i < 2; i++ {
		t.Logf("Inserting 99\n")
		sl.Insert(2, distToElem(99))
		t.Logf("%s\n", DebugPrintISkipList(&sl, 3))
	}
	for i := 0; i < 2; i++ {
		t.Logf("Inserting 88\n")
		sl.Insert(4, distToElem(88))
		t.Logf("%s\n", DebugPrintISkipList(&sl, 3))
	}

	sl.Swap(4, 8)
	t.Logf("%s\n", DebugPrintISkipList(&sl, 3))

	if sl.Length() != len(expected) {
		t.Errorf("Expected length %v, actual length %v\n", len(expected), sl.Length())
	}

	t.Logf("Length %v\n", sl.Length())
	for i := 0; i < sl.Length(); i++ {
		t.Logf("Elem at %v: %v\n", i, sl.At(i))
		if sl.At(i) != expected[i] {
			t.Errorf("Expected value %v at index %v, got %v\n", expected[i], i, sl.At(i))
		}
	}
}

// This test creates random sequences of Insert, Swap and Remove operations and
// then applies these operations to both an ISkipList and a slice. The end
// results should match.
func TestRandomOpSequences(t *testing.T) {
	const nops = 1000
	const niters = 20

	var sl ISkipList
	sl.Seed(randSeed1, randSeed2)
	for i := 0; i < niters; i++ {
		t.Logf("----- Generating random sequence of %v operations -----\n", nops)
		ops := sliceutils.GenOps(nops, 0)
		sl.Clear()
		a := make([]ElemType, 0)
		for _, o := range ops {
			t.Logf("%s\n", sliceutils.PrintOp(&o))
			sliceutils.ApplyOpToSlice(&o, &a)
			applyOpToISkipList(&o, &sl)
			t.Logf("%v\n", DebugPrintISkipList(&sl, 3))
			t.Logf("%+v\n", a)

			t.Logf("Reported lengths: %v %v\n", sl.Length(), len(a))

			if len(a) != sl.Length() {
				t.Errorf("ISkipList has wrong length (%v instead of %v)\n", sl.Length(), len(a))
			}

			// Equality check by looping over indices.
			t.Logf("Testing result via index loop...\n")
			for i, v := range a {
				e := sl.At(i)
				t.Logf("Checking %v\n", i)
				if v != e {
					t.Errorf("Expected value %v at index %v, got %v instead (index loop).\n", v, i, e)
				}
			}
		}

		// Equality check using ForAllI
		t.Logf("Testing result via ForAllI()...")
		sl.ForAllI(func(i int, v *ElemType) {
			t.Logf("Checking %v\n", i)
			if *v != a[i] {
				t.Errorf("Expected value %v at index %v, got %v instead (ForAllI).\n", a[i], i, *v)
			}
		})

		// Copy and then check copy has expected elements using ForAllI.
		cp := sl.Copy()
		cp.ForAllI(func(i int, v *ElemType) {
			t.Logf("Checking %v\n", i)
			if *v != a[i] {
				t.Errorf("Expected value %v at index %v, got %v instead (ForAllI).\n", a[i], i, *v)
			}
		})
	}
}

func benchmarkRandomOpSequenceWithISKipList(ops []sliceutils.Op, sl *ISkipList, l int) {
	for _, o := range ops {
		applyOpToISkipList(&o, sl)
	}
}

func benchmarkRandomOpSequenceWithSlice(ops []sliceutils.Op, a []ElemType, l int) {
	for _, o := range ops {
		sliceutils.ApplyOpToSlice(&o, &a)
	}
}

func BenchmarkRandomOpSequence(b *testing.B) {
	const nops = 500

	ops := sliceutils.GenOps(nops, 0)

	for i := 0; i < 100000; i += 1000 {
		b.Run(fmt.Sprintf("With slice [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				a := make([]ElemType, i)
				for j := 0; j < i; j++ {
					a[j] = j
				}

				benchmarkRandomOpSequenceWithSlice(ops, a, nops)
			}
		})

		b.Run(fmt.Sprintf("With ISkipList [initial length=%v, n_ops=%v]", i, nops), func(b *testing.B) {
			var sl ISkipList
			sl.Seed(randSeed1, randSeed2)
			for j := 0; j < i; j++ {
				sl.PushFront(distToElem(j))
			}

			for j := 0; j < b.N; j++ {
				benchmarkRandomOpSequenceWithISKipList(ops, &sl, nops)
			}
		})
	}
}

func BenchmarkStartInsert(b *testing.B) {
	for i := 0; i < 5000; i += 100 {
		b.Run(fmt.Sprintf("Creating ISkipList of length %v using start insert", i), func(b *testing.B) {
			var sl ISkipList
			for j := 0; j < b.N; j++ {
				sl.Clear()
				sl.Seed(randSeed1, randSeed2)

				for k := 0; k < i; k++ {
					sl.Insert(0, k)
				}
			}
			//b.Logf("Levels: %v\n", sl.nLevels+1)
		})
	}
}

func BenchmarkEndInsert(b *testing.B) {
	for i := 0; i < 5000; i += 100 {
		b.Run(fmt.Sprintf("Creating ISkipList of length %v using end insert", i), func(b *testing.B) {
			var sl ISkipList
			for j := 0; j < b.N; j++ {
				sl.Clear()
				sl.Seed(randSeed1, randSeed2)

				for k := 0; k < i; k++ {
					sl.Insert(k, k)
				}
			}
			//b.Logf("Levels: %v\n", sl.nLevels+1)
		})
	}
}

func BenchmarkCreationMethods(b *testing.B) {
	for i := 0; i < 100000; i += 1000 {
		b.Run(fmt.Sprintf("Creating slice of length %v", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				a := make([]ElemType, i, i)
				for k := 0; k < len(a); k++ {
					a[k] = k
				}
			}
		})

		b.Run(fmt.Sprintf("Creating ISkipList of length %v using PushFront", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				var sl ISkipList
				for k := 0; k < i; k++ {
					sl.PushFront(k)
				}
			}
		})

		b.Run(fmt.Sprintf("Creating ISkipList of length %v using PushBack", i), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				var sl ISkipList
				for k := 0; k < i; k++ {
					sl.PushBack(k)
				}
			}
		})
	}
}

func applyOpToISkipList(op *sliceutils.Op, sl *ISkipList) {
	switch op.Kind {
	case sliceutils.OpInsert:
		sl.Insert(op.Index1, op.Elem)
	case sliceutils.OpRemove:
		sl.Remove(op.Index1)
	case sliceutils.OpSwap:
		sl.Swap(op.Index1, op.Index2)
	}
}
