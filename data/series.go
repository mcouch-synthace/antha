package data

import (
	"errors"
	"reflect"
)

type ColumnName string
type Index int
type Key []ColumnName

// Row represents a materialized record. Every time you make a row you are
// copying data so this should be minimized
type Row struct {
	Index  Index
	Values []Observation
}
type Rows []Row

func (r Row) Observation(c ColumnName) (Observation, error) {
	// TODO more efficiently access schema
	for _, o := range r.Values {
		if o.ColumnName() == c {
			return o, nil
		}
	}
	return Observation{}, errors.New("no column " + string(c))
}

type Observation struct {
	col   *ColumnName
	value interface{} // TODO ptr?
}

func (o Observation) ColumnName() ColumnName {
	return *o.col
}

// dynamic read
func (o Observation) IsNull() bool {
	// TODO.
	return o.value == nil
}

func (o Observation) MustInt64() int64 {
	// panic on err
	return o.value.(int64)
}

func (o Observation) MustInt() int {
	return int(o.MustInt64())
}

func (o Observation) MustString() string {
	return o.value.(string)
}

//etcetera

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
}

// func (s *Series) ColumnName() ColumnName {
// 	return s.col
// }

// func (s *Series) Slice(start, stop Index) []Observation {
// 	return nil
// }

func (s *Series) copy() *Series {
	// TODO semantics? may not be supported everywhere
	return nil
}
