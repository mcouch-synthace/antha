package data

import (
	"reflect"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/pkg/errors"
)

// Data sets using Apache Arrow.
// For now, looks like it's easier to implement this in the same style as arrow/go library: i.e. no reflection, using template code instead

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
