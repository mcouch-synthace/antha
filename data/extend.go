package data

import (
	"reflect"
)

type extendSeries struct {
	f      func(r Row) interface{}
	source *readRow
}

func (i *extendSeries) Next() bool {
	return true //only exhausted when underlying iterators are
}
func (i *extendSeries) Value() interface{} {
	row := i.source.Value().(Row)
	v := i.f(row)
	return v
}

func extendTable(f func(r Row) interface{}, newCol ColumnName, newType reflect.Type, table *Table) *Table {
	series := append(table.series, &Series{
		col: newCol,
		typ: newType,
		read: func(cache seriesIterCache) iterator {
			// virtual table will not be used to advance
			source := &readRow{iteratorCache: cache}
			// go get the series iterators we need from the cache
			source.fill(table.series)

			return &extendSeries{f: f, source: source}
		}},
	)
	newT := NewTable(series)
	return newT
}
