package data

import (
	"reflect"
)

/*
 * calculated columns
 */

// TODO figure out how to avoid recalculating state, without user having to cache

// Extension is the fluent interface for adding calculated columns
type Extension struct {
	// the new column to add
	newCol ColumnName
	// all table series
	series []*Series
}

// By allows dynamic access to observations
func (e *Extension) By(f func(r Row) interface{}, newType reflect.Type) *Table {
	// TODO: either reflectively infer newType, or assert/verify the f return type

	series := append(e.series, &Series{
		col: e.newCol,
		typ: newType,
		read: func(cache seriesIterCache) iterator {
			// virtual table will not be used to advance
			source := &readRow{iteratorCache: cache}
			// go get the series iterators we need from the cache
			source.fill(e.series)
			return &extendRowSeries{f: f, source: source}
		}},
	)
	newT := NewTable(series)
	return newT
}

type extendRowSeries struct {
	f      func(r Row) interface{}
	source *readRow
}

func (i *extendRowSeries) Next() bool {
	return true //only exhausted when underlying iterators are
}

func (i *extendRowSeries) Value() interface{} {
	row := i.source.Value().(Row)
	v := i.f(row)
	return v
}

// On is for operations on homogeneous columns of static type
func (e *Extension) On(cols ...ColumnName) *ExtendOn {
	schema := newSchema(e.series)
	on := &ExtendOn{extension: e}
	for _, c := range cols {
		// TODO panic here, need test
		on.inputs = append(on.inputs, e.series[schema.byName[c][0]])
	}
	return on
}

type ExtendOn struct {
	extension *Extension
	inputs    []*Series
}
