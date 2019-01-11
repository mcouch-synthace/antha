package data

import (
	"reflect"
)

// Lazy data sets

// Table is an immutable container of Series
// It can optionally be keyed.
type Table struct {
	series []*Series
	// this must return Row
	read func(*Table) iterator
}

// NewTable gives lowlevel access.  TODO semantics if jagged columns, duplicates etc
func NewTable(series []*Series) *Table {
	return &Table{
		series: series,
		read:   newLzyIter,
	}
}
func (t *Table) Series() []*Series {
	return nil
}

func (t *Table) ColumnNames() []ColumnName {
	return nil
}

// Iter iterates over the entire table, no buffer (so blocking)
// TODO leaking a goroutine here?  need a control channel.
func (t *Table) Iter() <-chan Row {
	channel := make(chan Row)
	iter := t.read(t)
	go func() {
		for {
			rowRaw, err := iter.Next()
			if err != nil {
				close(channel)
				if err == endIteration {
					return
				}
				// TODO panic here?
				return
			}
			row := rowRaw.(Row)
			channel <- row

		}
	}()
	return channel
}

type lazyRowIter struct {
	cols       []*ColumnName
	index      int
	seriesRead []iterator
}

func newLzyIter(t *Table) iterator {
	i := &lazyRowIter{}
	for _, ser := range t.series {
		i.seriesRead = append(i.seriesRead, ser.read(ser))
		i.cols = append(i.cols, &ser.col)
	}
	return i
}

func (i *lazyRowIter) State() error {
	return nil
}

func (i *lazyRowIter) Next() (interface{}, error) {
	r := Row{Index: Index(i.index)}
	for idx, sRead := range i.seriesRead {
		value, err := sRead.Next()
		if err != nil {
			return nil, err
		}
		r.Values = append(r.Values, Observation{col: i.cols[idx], value: value})
	}
	i.index++
	return r, nil
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
func (t *Table) Slice(start, endExclusive int) (*Table, error) {
	return nil, nil
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

// Project takes the named columns
func (t *Table) Project(columns []ColumnName) *Table {
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
func (t *Table) ExtendBy(f func(r Row) (interface{}, error), newColumn ColumnName, newType reflect.Type) *Table {
	return nil
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
