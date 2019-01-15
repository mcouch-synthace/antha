package data

import (
	"github.com/pkg/errors"
	"reflect"
)

// Lazy data sets

// Table is an immutable container of Series
// It can optionally be keyed.
type Table struct {
	series []*Series
	// this must return Row
	read func([]*Series) *tableIterator
}

// NewTable gives lowlevel access.  TODO semantics if jagged columns, duplicates etc
func NewTable(series []*Series) *Table {
	return &Table{
		series: series,
		read:   newTableIterator,
	}
}

// Schema returns the type information for the Table
func (t *Table) Schema() Schema {
	return newSchema(t.series)
}

// TODO do we need this
func (t *Table) Series() []*Series {
	return nil
}

func (t *Table) seriesMap() map[ColumnName]*Series {

	schema := t.Schema()
	byName := map[ColumnName]*Series{}
	for n, i := range schema.byName {
		byName[n] = t.series[i]
	}
	return byName
}

// TODO do we need this
// func (t *Table) ColumnNames() []ColumnName {
// 	return nil
// }

// Iter iterates over the entire table, no buffer (so blocking)
// TODO leaking a goroutine here?  need a control channel.
func (t *Table) Iter() <-chan Row {
	channel := make(chan Row)
	iter := t.read(t.series)
	go func() {
		for iter.Next() {
			rowRaw := iter.Value()
			row := rowRaw.(Row)
			channel <- row
		}
		close(channel)
	}()
	return channel
}

// ToRows materializes data: may be very expensive
func (t *Table) ToRows() Rows {
	rr := make(Rows, 0)
	for r := range t.Iter() {
		rr = append(rr, r)
	}
	return rr
}

// Slice TODO semantics,  This should probably materialize lazy tables.
func (t *Table) Slice(start, endExclusive int) *Table {
	return nil
}

// Key returns the key columns
func (t *Table) Key() Key {
	return nil
}

// WithKey sets the sort key (but does not sort).
func (t *Table) WithKey(key Key) *Table {
	return nil
}

// Sort produces a sorted Table using the Key.
func (t *Table) Sort(asc ...bool) *Table {
	return nil
}

// Project reorders and/or takes a subset of columns
func (t *Table) Project(columns []ColumnName) *Table {
	s := make([]*Series, len(columns))
	byName := t.seriesMap()
	for i, n := range columns {
		if ser, found := byName[n]; !found {
			panic(errors.Errorf("cannot project %v, no such column '%s'", t.Schema(), n))
		} else {
			s[i] = ser
		}
	}
	return NewTable(s)
}

// ProjectAllBut discards the named columns
func (t *Table) ProjectAllBut(columns []ColumnName) *Table {
	return nil
}

// Filter selects some rows lazily
func (t *Table) Filter(f FilterSpec) *Table {
	return lazyFilterTable(f, t)
}

// Join is a natural join on tables with the same Key
// TODO dedup series (?)
func (t *Table) Join(other Joinable) *Table {
	return nil
}

// ExtendBy adds a column by applying f.
// TODO the implicit dependency on the table schema for t (via Row) is a bit ugly here
func (t *Table) ExtendBy(f func(r Row) interface{}, newCol ColumnName, newType reflect.Type) *Table {
	return extendTable(f, newCol, newType, t)
}

// Copy gives a new table, optionally with duplicate Series data
// TODO semantics
// func (t *Table) Copy(deep bool) *Table {
// 	return nil
// }

var _ Joinable = (*Table)(nil)

type Joinable interface {
	//?
	Key() Key
	Series() []*Series
	// TODO bridge interfaces for more efficient backend join ops
}
