package data

import (
	"github.com/pkg/errors"
	"reflect"
)

// Data sets using go native types.  This can be slow!

// NewSliceSeries convert a slice of scalars to a new Series.
// reflectively supports arbitrary slice types.
// TODO it would be possible to have a faster path for common scalar types
func NewSliceSeries(col ColumnName, values interface{}) (*Series, error) {
	rValue := reflect.ValueOf(values)
	if rValue.Kind() != reflect.Slice {
		return nil, errors.Errorf("can't use input of type %v, expecting slice", rValue.Kind())
	}
	m := &nativeSliceSerMeta{
		rValue: rValue,
		len:    rValue.Len(),
	}
	return &Series{
		typ:  rValue.Type().Elem(),
		col:  col,
		read: m.read,
		meta: m,
	}, nil
}

type nativeSliceSerMeta struct {
	rValue reflect.Value
	len    int
}

func (m *nativeSliceSerMeta) ExactSize() int {
	return m.len
}
func (m *nativeSliceSerMeta) MaxSize() int {
	return m.len
}

func (m *nativeSliceSerMeta) read(_ seriesIterCache) iterator {
	return &nativeSliceSerIter{
		nativeSliceSerMeta: m,
		pos:                -1,
	}
}

var _ Bounded = (*nativeSliceSerMeta)(nil)

type nativeSliceSerIter struct {
	*nativeSliceSerMeta
	pos int
}

func (i *nativeSliceSerIter) Next() bool {
	i.pos++
	return i.pos < i.len
}

// index into the slice reflectively (slow)
func (i *nativeSliceSerIter) Value() interface{} {
	return i.rValue.Index(i.pos).Interface()
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
