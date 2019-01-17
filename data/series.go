package data

import (
	"reflect"
)

type advanceable interface {
	Next() bool // false = end iteration
}

// the generic value iterator.  Pays the cost of interface pointers
type iterator interface {
	advanceable
	Value() interface{}
}

// Series is a named sequence of values. for larger datasets this sequence may
// be loaded lazily (eg memory map) or may even be unbounded
type Series struct {
	col ColumnName
	// typically a scalar type
	typ  reflect.Type
	read func(seriesIterCache) iterator
	meta seriesMeta
}

// Size returns the exact length of the series, or -1 if not known
func (s *Series) Size() int {
	if b, ok := s.meta.(Bounded); ok {
		return b.ExactSize()
	}
	return -1
}

// seriesMeta captures differing series backend capabilities
type seriesMeta interface{}

type Bounded interface {
	// ExactSize can return -1 if size is not known
	ExactSize() int
	// MaxSize should always return >=0
	MaxSize() int
}

// TODO ...
type Sliceable interface {
	Slice(start, end Index) *Series
}

/* TODO specializations for more efficient value access

type valueInt64 interface {
	Null() bool
	Int64() int64 // panic if null
}

type Int64Iter interface {
	advanceable
	valueInt64
}
//*/
