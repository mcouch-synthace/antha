package data

import ()

type slicer struct {
	wrapped         iterator
	start, end, pos Index
}

func (iter *slicer) Next() bool {
	for iter.pos+1 < iter.start {
		if !iter.wrapped.Next() {
			return false
		}
		iter.pos++
	}
	if !iter.wrapped.Next() {
		// exhausted the underlying series
		return false
	}
	iter.pos++
	if iter.pos == iter.end {
		return false
	}
	return true
}

func (iter *slicer) Value() interface{} {
	return iter.wrapped.Value()
}
