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
	meta seriesMeta
}

// seriesMeta captures differing series backend capabilities
type seriesMeta interface{}

type Bounded interface {
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

/*
 * 'iter<Type'> are iterator specializations for potentially no-copy, boxed values.
 *
 * The 'as<Type>' types are fallbacks for when the underlying series is dynamic.
 */

//TODO codegen below this point

// bool
type BoxBool interface {
	Bool() (bool, bool) // returns false = nil
}

type iterBool interface {
	advanceable
	BoxBool
}

// bridge
func (s *Series) iterateBool(iter iterator) (iterBool, error) {
	if cast, ok := iter.(iterBool); ok {
		return cast, nil
	}
	if err := s.assignableTo(reflect.TypeOf(false)); err != nil {
		return nil, err
	}
	return &asBool{iterator: iter}, nil
}

type asBool struct {
	iterator
}

func (a *asBool) Bool() (bool, bool) {
	v := a.iterator.Value()
	if v == nil {
		return false, false
	}
	return v.(bool), true
}

// int64
type BoxInt64 interface {
	Int64() (int64, bool) // returns false = nil
}

type iterInt64 interface {
	advanceable
	BoxInt64
}

// bridge
func (s *Series) iterateInt64(iter iterator) (iterInt64, error) {
	if cast, ok := iter.(iterInt64); ok {
		return cast, nil
	}
	if err := s.assignableTo(reflect.TypeOf(int64(0))); err != nil {
		return nil, err
	}
	return &asInt64{iterator: iter}, nil
}

type asInt64 struct {
	iterator
}

func (a *asInt64) Int64() (int64, bool) {
	v := a.iterator.Value()
	if v == nil {
		return 0, false
	}
	return v.(int64), true
}

// float64
type BoxFloat64 interface {
	Float64() (float64, bool)
}

type iterFloat64 interface {
	advanceable
	BoxFloat64
}

// bridge
func (s *Series) iterateFloat64(iter iterator) (iterFloat64, error) {
	if cast, ok := iter.(iterFloat64); ok {
		return cast, nil
	}
	if err := s.assignableTo(reflect.TypeOf(float64(0))); err != nil {
		return nil, err
	}
	return &asFloat64{iterator: iter}, nil
}

type asFloat64 struct {
	iterator
}

func (a *asFloat64) Float64() (float64, bool) {
	v := a.iterator.Value()
	if v == nil {
		return 0, false
	}
	return v.(float64), true
}

// string
type BoxString interface {
	String() (string, bool)
}

type iterString interface {
	advanceable
	BoxString
}

// bridge
func (s *Series) iterateString(iter iterator) (iterString, error) {
	if cast, ok := iter.(iterString); ok {
		return cast, nil
	}
	if err := s.assignableTo(reflect.TypeOf("")); err != nil {
		return nil, err
	}
	return &asString{iterator: iter}, nil
}

type asString struct {
	iterator
}

func (a *asString) String() (string, bool) {
	v := a.iterator.Value()
	if v == nil {
		return "", false
	}
	return v.(string), true
}
