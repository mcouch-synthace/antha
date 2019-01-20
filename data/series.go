package data

import (
	"reflect"

	"github.com/pkg/errors"
)

type advanceable interface {
	Next() bool // false = end iteration
}

// the generic value iterator.  Pays the cost of interface pointer on each value
type iterator interface {
	advanceable
	Value() interface{} // always must be implemented
}

// Series is a named sequence of values. for larger datasets this sequence may
// be loaded lazily (eg memory map) or may even be unbounded
type Series struct {
	col ColumnName
	// typically a scalar type
	typ  reflect.Type
	read func(seriesIterCache) iterator
	meta SeriesMeta
}

// SeriesMeta captures differing series backend capabilities
type SeriesMeta interface {
	// IsBounded = true if the Series is bounded
	IsBounded() bool
	// IsMaterialized = true if the Series is bounded and not lazy
	IsMaterialized() bool
}

// Bounded is implemented by bounded series metadata
type Bounded interface {
	SeriesMeta
	// ExactSize can return -1 if size is not known
	ExactSize() int
	// MaxSize should always return >=0
	MaxSize() int
}

// TODO ... for efficiently indexable backend
type Sliceable interface {
	Slice(start, end Index) *Series
}

/*
 * casting - not conversion
 */
func (s *Series) assignableTo(typ reflect.Type) error {
	if !s.typ.AssignableTo(typ) {
		return errors.Errorf("column %s of type %v cannot be iterated as %v", s.col, s.typ, typ)
	}
	return nil
}

// Copy converts a (possibly) lazy series to one that is fully materialized (currently Arrow)
func (s *Series) Copy() (*Series, error) {
	return NewArrowSeriesFromSeries(s)
}

// Cache does the same as Copy except it is in-place
func (s *Series) Cache() (*Series, error) {
	copy, err := s.Copy()
	if err != nil {
		return nil, err
	}
	s.read = copy.read
	s.meta = copy.meta
	return s, nil
}
