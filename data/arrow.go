package data

import (
	"reflect"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/pkg/errors"
)

// Data sets using Apache Arrow.
// For now, looks like it's easier to implement this in the same style as arrow/go library: i.e. no reflection, using template code instead

// --- Generic code ---

// NewArrowSeries converts a Arrow array to a new Series. Only closed list of Arrow data types is supported yet.
func NewArrowSeries(col ColumnName, values array.Interface) (*Series, error) {
	switch typedValues := values.(type) {
	case *array.Boolean:
		return NewArrowSeriesBool(col, typedValues), nil
	case *array.Int64:
		return NewArrowSeriesInt64(col, typedValues), nil
	case *array.Float64:
		return NewArrowSeriesFloat64(col, typedValues), nil
	case *array.String:
		return NewArrowSeriesString(col, typedValues), nil
	default:
		return nil, errors.Errorf("Arrow data type %v is not supported", values.DataType().ID())
	}
}

// NewArrowSeriesFromSlice converts a slice of scalars to a new (Arrow-based) Series
// mask denotes elements set to null; it is optional and can be set to nil
// Only a closed list of primitive data types is supported yet
func NewArrowSeriesFromSlice(col ColumnName, values interface{}, mask []bool) (*Series, error) {
	switch typedValues := values.(type) {
	case []bool:
		return NewArrowSeriesFromSliceBool(col, typedValues, mask), nil
	case []int64:
		return NewArrowSeriesFromSliceInt64(col, typedValues, mask), nil
	case []float64:
		return NewArrowSeriesFromSliceFloat64(col, typedValues, mask), nil
	case []string:
		return NewArrowSeriesFromSliceString(col, typedValues, mask), nil
	default:
		return nil, errors.Errorf("The data type %v is not supported, expecting slice of supported primitive types", reflect.TypeOf(values))
	}
}

// --- Type-specific code ---

// TODO:replace most of this type-specific code (now copied-pasted manually) with code generation

// bool

func NewArrowSeriesBool(col ColumnName, values *array.Boolean) *Series {
	metadata := &boolArrowSeriesMeta{values: values}
	return &Series{
		typ:  reflect.TypeOf(false),
		col:  col,
		read: metadata.read,
		meta: metadata,
	}
}

func NewArrowSeriesFromSliceBool(col ColumnName, values []bool, mask []bool) *Series {
	builder := array.NewBooleanBuilder(memory.DefaultAllocator)
	if len(values) > 0 {
		builder.AppendValues(values, mask)
	}
	arrowValues := builder.NewBooleanArray()
	return NewArrowSeriesBool(col, arrowValues)
}

type boolArrowSeriesMeta struct {
	values *array.Boolean
}

func (m *boolArrowSeriesMeta) ExactSize() int {
	return m.values.Len()
}
func (m *boolArrowSeriesMeta) MaxSize() int {
	return m.values.Len()
}

func (m *boolArrowSeriesMeta) read(_ seriesIterCache) iterator {
	return &boolArrowSeriesIter{
		boolArrowSeriesMeta: m,
		pos:                 -1,
	}
}

var _ Bounded = (*boolArrowSeriesMeta)(nil)

type boolArrowSeriesIter struct {
	*boolArrowSeriesMeta
	pos int
}

func (i *boolArrowSeriesIter) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *boolArrowSeriesIter) Bool() (bool, bool) {
	if !i.values.IsNull(i.pos) {
		return i.values.Value(i.pos), true
	} else {
		return false, false
	}
}

func (i *boolArrowSeriesIter) Value() interface{} {
	if val, ok := i.Bool(); ok {
		return val
	} else {
		return nil
	}
}

var _ iterator = (*boolArrowSeriesIter)(nil)
var _ iterBool = (*boolArrowSeriesIter)(nil)

// int64: constructor and iterator

func NewArrowSeriesInt64(col ColumnName, values *array.Int64) *Series {
	metadata := &int64ArrowSeriesMeta{values: values}
	return &Series{
		typ:  reflect.TypeOf(int64(0)),
		col:  col,
		read: metadata.read,
		meta: metadata,
	}
}

func NewArrowSeriesFromSliceInt64(col ColumnName, values []int64, mask []bool) *Series {
	builder := array.NewInt64Builder(memory.DefaultAllocator)
	if len(values) > 0 {
		builder.AppendValues(values, mask)
	}
	arrowValues := builder.NewInt64Array()
	return NewArrowSeriesInt64(col, arrowValues)
}

type int64ArrowSeriesMeta struct {
	values *array.Int64
}

func (m *int64ArrowSeriesMeta) ExactSize() int {
	return m.values.Len()
}
func (m *int64ArrowSeriesMeta) MaxSize() int {
	return m.values.Len()
}

func (m *int64ArrowSeriesMeta) read(_ seriesIterCache) iterator {
	return &int64ArrowSeriesIter{
		int64ArrowSeriesMeta: m,
		pos:                  -1,
	}
}

var _ Bounded = (*int64ArrowSeriesMeta)(nil)

type int64ArrowSeriesIter struct {
	*int64ArrowSeriesMeta
	pos int
}

func (i *int64ArrowSeriesIter) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *int64ArrowSeriesIter) Int64() (int64, bool) {
	if !i.values.IsNull(i.pos) {
		return i.values.Value(i.pos), true
	} else {
		return int64(0), false
	}
}

func (i *int64ArrowSeriesIter) Value() interface{} {
	if val, ok := i.Int64(); ok {
		return val
	} else {
		return nil
	}
}

var _ iterator = (*int64ArrowSeriesIter)(nil)
var _ iterInt64 = (*int64ArrowSeriesIter)(nil)

// float64: constructor and iterator

func NewArrowSeriesFloat64(col ColumnName, values *array.Float64) *Series {
	metadata := &float64ArrowSeriesMeta{values: values}
	return &Series{
		typ:  reflect.TypeOf(float64(0)),
		col:  col,
		read: metadata.read,
		meta: metadata,
	}
}

func NewArrowSeriesFromSliceFloat64(col ColumnName, values []float64, mask []bool) *Series {
	builder := array.NewFloat64Builder(memory.DefaultAllocator)
	if len(values) > 0 {
		builder.AppendValues(values, mask)
	}
	arrowValues := builder.NewFloat64Array()
	return NewArrowSeriesFloat64(col, arrowValues)
}

type float64ArrowSeriesMeta struct {
	values *array.Float64
}

func (m *float64ArrowSeriesMeta) ExactSize() int {
	return m.values.Len()
}
func (m *float64ArrowSeriesMeta) MaxSize() int {
	return m.values.Len()
}

func (m *float64ArrowSeriesMeta) read(_ seriesIterCache) iterator {
	return &float64ArrowSeriesIter{
		float64ArrowSeriesMeta: m,
		pos:                    -1,
	}
}

var _ Bounded = (*float64ArrowSeriesMeta)(nil)

type float64ArrowSeriesIter struct {
	*float64ArrowSeriesMeta
	pos int
}

func (i *float64ArrowSeriesIter) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *float64ArrowSeriesIter) Float64() (float64, bool) {
	if !i.values.IsNull(i.pos) {
		return i.values.Value(i.pos), true
	} else {
		return float64(0), false
	}
}

func (i *float64ArrowSeriesIter) Value() interface{} {
	if val, ok := i.Float64(); ok {
		return val
	} else {
		return nil
	}
}

var _ iterator = (*float64ArrowSeriesIter)(nil)
var _ iterFloat64 = (*float64ArrowSeriesIter)(nil)

// string: constructor and iterator

func NewArrowSeriesString(col ColumnName, values *array.String) *Series {
	metadata := &stringArrowSeriesMeta{values: values}
	return &Series{
		typ:  reflect.TypeOf(""),
		col:  col,
		read: metadata.read,
		meta: metadata,
	}
}

func NewArrowSeriesFromSliceString(col ColumnName, values []string, mask []bool) *Series {
	builder := array.NewStringBuilder(memory.DefaultAllocator)
	if len(values) > 0 {
		builder.AppendValues(values, mask)
	}
	arrowValues := builder.NewStringArray()
	return NewArrowSeriesString(col, arrowValues)
}

type stringArrowSeriesMeta struct {
	values *array.String
}

func (m *stringArrowSeriesMeta) ExactSize() int {
	return m.values.Len()
}
func (m *stringArrowSeriesMeta) MaxSize() int {
	return m.values.Len()
}

func (m *stringArrowSeriesMeta) read(_ seriesIterCache) iterator {
	return &stringArrowSeriesIter{
		stringArrowSeriesMeta: m,
		pos:                   -1,
	}
}

var _ Bounded = (*stringArrowSeriesMeta)(nil)

type stringArrowSeriesIter struct {
	*stringArrowSeriesMeta
	pos int
}

func (i *stringArrowSeriesIter) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *stringArrowSeriesIter) String() (string, bool) {
	if !i.values.IsNull(i.pos) {
		return i.values.Value(i.pos), true
	} else {
		return "", false
	}
}

func (i *stringArrowSeriesIter) Value() interface{} {
	if val, ok := i.String(); ok {
		return val
	} else {
		return nil
	}
}

var _ iterator = (*stringArrowSeriesIter)(nil)
var _ iterString = (*stringArrowSeriesIter)(nil)
