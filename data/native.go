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
	ser.read = func(_ seriesIterCache) iterator {
		return &sliceIter{
			ser:    ser,
			rValue: rValue,
			len:    len,
			pos:    -1,
		}
	}
	return ser, nil
}

type sliceIter struct {
	ser    *Series
	rValue reflect.Value
	pos    int
	len    int
}

func (i *sliceIter) Next() bool {
	i.pos++
	return i.pos < i.len
}

func (i *sliceIter) Value() interface{} {
	return i.rValue.Index(i.pos).Interface()
}

// FromRows constructs a new Table
func FromRows(Rows) (*Table, error) {
	return nil, nil
}

/*
// ToStructSlice reflectively copies to the given struct fields
func (r *Rows) ToStructSlice(structsPtr interface{}) error {
	return nil
}
func (r Row) ToStruct(structPtr interface{}) error {
	return nil
}

func FromStructSlice(structs interface{}) (*Rows, error) {
	return nil, nil
}
func FromStruct(struc interface{}) (Row, error) {
	return Row{}, nil
}
*/
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
