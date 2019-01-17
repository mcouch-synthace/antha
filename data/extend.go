package data

import (
	"github.com/pkg/errors"
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

// TODO codegen below here

// Float64 adds a float64 col using float64 inputs.  Null on any null inputs.
func (e *ExtendOn) Float64(f func(v ...float64) float64) *Table {
	// TODO move from lazy to eager type validation
	return NewTable(append(e.extension.series, &Series{
		col: e.extension.newCol,
		typ: reflect.TypeOf(float64(0)),
		read: func(cache seriesIterCache) iterator {
			// Every series must be cast or converted
			colReader := make([]iterFloat64, len(e.inputs))
			var err error
			for i, ser := range e.inputs {
				iter := cache.Ensure(ser)
				colReader[i], err = ser.iterateFloat64(iter) // note colReader[i] is not itself in the cache!
				if err != nil {
					// TODO non-panic option?
					// TODO test coverage
					panic(errors.Wrapf(err, "when projecting new column %v", e.extension.newCol))
				}
			}
			return &extendInt64{f: f, source: colReader}
		}},
	))
}

var _ iterFloat64 = (*extendInt64)(nil)

type extendInt64 struct {
	f      func(v ...float64) float64
	source []iterFloat64
}

func (x *extendInt64) Next() bool {
	return true
}
func (x *extendInt64) Value() interface{} {
	v, ok := x.Float64()
	if !ok {
		return nil
	}
	return v
}
func (x *extendInt64) Float64() (float64, bool) {
	args := make([]float64, len(x.source))
	var ok bool
	for i, s := range x.source {
		args[i], ok = s.Float64()
		if !ok {
			return 0, false
		}
	}
	v := x.f(args...)
	return v, true
}
