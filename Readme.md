# iskiplist

Indexable
[skip lists](https://en.wikipedia.org/wiki/Skip_list) for Go.

Skip lists are typically used to implement associative arrays (analogous to Go
maps). The `ISkipList` provided by this package is a sequence with O(log n)
access, insertion and removal of elements at a given index (analogous to Go
slices, but with different big O characteristics).

Each element of an `ISkipList` is an `int`. The idea is to use the `int` value
as an index into a slice of the data structure of your choice. If this isn't
feasible, you can modify `ElemType`, `elemToDist` and `distToElem`.

Each `ISkipList` maintains its own local PCG pseudorandom number generator
state.