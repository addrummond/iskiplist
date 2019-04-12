# iskiplist

Indexable
[skip lists](https://en.wikipedia.org/wiki/Skip_list) for Go.

Skip lists are typically used to implement associative arrays (analogous to Go
maps). The `ISkipList` provided by this library is a sequence with O(log n)
access, insertion and removal of elements at a given index (analogous to Go
slices, but with different big O characteristics).

The elements of an `ISkipList` are `int`s. The idea is to use the `int` value as
an index into a slice of the data structure of your choice. If this isn't
feasible, you can modify `ElemType`, `elemToDist` and `distToElem`.