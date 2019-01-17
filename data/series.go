package data

import (
	"reflect"
)

type iterator interface {
	Next() bool // false = end iteration
	Value() interface{}
}

// Series is a named sequence of values. for larger datasets this sequence may
// be loaded lazily (eg memory map) or may even be unbounded
type Series struct {
	col ColumnName
	// typically a scalar type
	typ  reflect.Type
	read func(seriesIterCache) iterator
	// TODO specialized iterator types for known scalars
}
