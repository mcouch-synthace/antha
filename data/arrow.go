package data

import (
	"reflect"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/pkg/errors"
)

// Data sets using Apache Arrow.
// For now, looks like it's easier to implement this in the same style as arrow/go library: i.e. no reflection, using template code instead
// TODO: replace the code copy-pasted manually with code generation

// TODO:make arrow series implement 'iter<Type>' statically typed iterators

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

// bool: constructor and iterator

func NewArrowSeriesBool(col ColumnName, values *array.Boolean) *Series {
	return &Series{
		typ: reflect.TypeOf(false),
		col: col,
		read: func(_ seriesIterCache) iterator {
			return &boolIterator{
				values: values,
				pos:    -1,
			}
		},
	}
}

type boolIterator struct {
	values *array.Boolean
	pos    int
}

func (i *boolIterator) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *boolIterator) Value() interface{} {
	if i.values.IsNull(i.pos) {
		return nil
	}
	return i.values.Value(i.pos)
}

// int64: constructor and iterator

func NewArrowSeriesInt64(col ColumnName, values *array.Int64) *Series {
	return &Series{
		typ: reflect.TypeOf(int64(0)),
		col: col,
		read: func(_ seriesIterCache) iterator {
			return &int64Iterator{
				values: values,
				pos:    -1,
			}
		},
	}
}

type int64Iterator struct {
	values *array.Int64
	pos    int
}

func (i *int64Iterator) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *int64Iterator) Value() interface{} {
	if i.values.IsNull(i.pos) {
		return nil
	}
	return i.values.Value(i.pos)
}

// float64: constructor and iterator

func NewArrowSeriesFloat64(col ColumnName, values *array.Float64) *Series {
	return &Series{
		typ: reflect.TypeOf(float64(0)),
		col: col,
		read: func(_ seriesIterCache) iterator {
			return &float64Iterator{
				values: values,
				pos:    -1,
			}
		},
	}
}

type float64Iterator struct {
	values *array.Float64
	pos    int
}

func (i *float64Iterator) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *float64Iterator) Value() interface{} {
	if i.values.IsNull(i.pos) {
		return nil
	}
	return i.values.Value(i.pos)
}

// string: constructor and iterator

func NewArrowSeriesString(col ColumnName, values *array.String) *Series {
	return &Series{
		typ: reflect.TypeOf(""),
		col: col,
		read: func(_ seriesIterCache) iterator {
			return &stringIterator{
				values: values,
				pos:    -1,
			}
		},
	}
}

type stringIterator struct {
	values *array.String
	pos    int
}

func (i *stringIterator) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *stringIterator) Value() interface{} {
	if i.values.IsNull(i.pos) {
		return nil
	}
	return i.values.Value(i.pos)
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

func NewArrowSeriesFromSliceBool(col ColumnName, values []bool, mask []bool) *Series {
	builder := array.NewBooleanBuilder(memory.DefaultAllocator)
	builder.AppendValues(values, mask)
	arrowValues := builder.NewBooleanArray()
	return NewArrowSeriesBool(col, arrowValues)
}

func NewArrowSeriesFromSliceInt64(col ColumnName, values []int64, mask []bool) *Series {
	builder := array.NewInt64Builder(memory.DefaultAllocator)
	builder.AppendValues(values, mask)
	arrowValues := builder.NewInt64Array()
	return NewArrowSeriesInt64(col, arrowValues)
}

func NewArrowSeriesFromSliceFloat64(col ColumnName, values []float64, mask []bool) *Series {
	builder := array.NewFloat64Builder(memory.DefaultAllocator)
	builder.AppendValues(values, mask)
	arrowValues := builder.NewFloat64Array()
	return NewArrowSeriesFloat64(col, arrowValues)
}

func NewArrowSeriesFromSliceString(col ColumnName, values []string, mask []bool) *Series {
	builder := array.NewStringBuilder(memory.DefaultAllocator)
	builder.AppendValues(values, mask)
	arrowValues := builder.NewStringArray()
	return NewArrowSeriesString(col, arrowValues)
}
