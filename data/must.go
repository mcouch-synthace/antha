package data

import ()
import "github.com/apache/arrow/go/arrow/array"

/*
 * utility for wrapping error functions
 */

type MustCreate struct{}

// Must asserts no error on the objects it creates
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

func (m MustCreate) NewArrowSeries(col ColumnName, values array.Interface) *Series {
	ser, err := NewArrowSeries(col, values)
	m.handle(err)
	return ser
}

func (m MustCreate) NewArrowSeriesFromSlice(col ColumnName, values interface{}, mask []bool) *Series {
	ser, err := NewArrowSeriesFromSlice(col, values, mask)
	m.handle(err)
	return ser
}
