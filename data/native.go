package data

import (
	"github.com/pkg/errors"
	"reflect"
)

// Data sets using go native types.  This can be slow!

// NewSliceSeries convert a slice of scalars to a new Series
// reflectively supports arbitrary slice types.
// TODO it would be possible to have a faster path for common scalar types
func NewSliceSeries(col ColumnName, values interface{}) (*Series, error) {
	rValue := reflect.ValueOf(values)
	if rValue.Kind() != reflect.Slice {
		return nil, errors.Errorf("can't use input of type %v, expecting slice", rValue.Kind())
	}
	len := rValue.Len()
	typ := rValue.Type().Elem()
	ser := &Series{
		typ: typ,
		col: col,
	}
	// index into the slice reflectively (slow)
	ser.read = func(s *Series) iterator {
		return &sliceIter{
			ser:    s,
			rValue: rValue,
			len:    len,
		}
	}
	return ser, nil
}

type sliceIter struct {
	ser       *Series
	rValue    reflect.Value
	cursorPos int
	len       int
}

// slice iterator is never in error state
func (i *sliceIter) State() error {
	return nil
}

func (i *sliceIter) Next() (interface{}, error) {
	i.cursorPos++
	if i.cursorPos > i.len {
		// avoid panic
		return nil, endIteration
	}
	return i.rValue.Index(i.cursorPos - 1).Interface(), nil
}

// FromRows constructs a new Table
func FromRows(Rows) (*Table, error) {
	return nil, nil
}

// ToStructSlice reflectively copies to the given struct fields
func (r Rows) ToStructSlice(structsPtr interface{}) error {
	return nil
}
func (r Row) ToStruct(structPtr interface{}) error {
	return nil
}

func FromStructSlice(structs interface{}) (Rows, error) {
	return nil, nil
}
func FromStruct(struc interface{}) (Row, error) {
	return Row{}, nil
}

type MustCreate struct{}

func Must() MustCreate {
	return MustCreate{}
}

func (m MustCreate) handle(err error) {
	if err != nil {
		panic(err)
	}
}
func (m MustCreate) NewSliceSeries(col ColumnName, values interface{}) *Series {
	ser, err := NewSliceSeries(col, values)
	m.handle(err)
	return ser
}
