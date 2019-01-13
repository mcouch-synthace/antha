package data

import (
	"reflect"
)

type extendSeries struct {
	f         func(r Row) interface{}
	tableIter iterator
}

func (i *extendSeries) Next() bool {
	return i.tableIter.Next()
}
func (i *extendSeries) Value() interface{} {
	row := i.tableIter.Value().(Row)
	return i.f(row)
}

func extendTable(f func(r Row) interface{}, newCol ColumnName, newType reflect.Type, table *Table) *Table {
	series := table.series
	return NewTable(append(series, &Series{
		col: newCol,
		typ: newType,
		read: func(*Series) iterator {
			return &extendSeries{f: f, tableIter: table.read(table)}
		}},
	))
}
