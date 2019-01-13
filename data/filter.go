package data

import (
//	"fmt"
)

// Filter specs

// Eq matches Rows that have the given column equal to this value
type Eq struct {
	Col ColumnName
	Val interface{}
}

// TODO Eq specialization methods to more efficiently filter known scalar types

func (w Eq) Columns() []ColumnName {
	return []ColumnName{w.Col}
}

func (w Eq) MatchValues(v ...interface{}) bool {
	// TODO if possible, cast to the matched type
	// if not possible, we should panic or set an error field
	return v[0] == w.Val
}

type FilterSpec interface {
	Columns() []ColumnName
	MatchValues(...interface{}) bool
	// TODO bridge interfaces for more efficient backend filter ops
}

type filterIter struct {
	filterSpec FilterSpec
	raw        iterator
	state      Row
	index      int
}

func (i *filterIter) ok() bool {
	rowRaw := i.raw.Value()
	// TODO we could save some garbage here by indexing columns by position
	row := rowRaw.(Row)
	i.state = row
	colVals := []interface{}{}
	for _, cName := range i.filterSpec.Columns() {
		o, _ := row.Observation(cName)
		colVals = append(colVals, o.value)
	}
	return i.filterSpec.MatchValues(colVals...)
}
func (i *filterIter) Next() bool {
	for i.raw.Next() {
		if i.ok() {
			i.index++
			return true
		}
	}
	return false
}
func (i *filterIter) Value() interface{} {
	row := i.state
	row.Index = Index(i.index)
	return row
}

func lazyFilterTable(filterSpec FilterSpec, table *Table) *Table {
	newTable := &Table{series: table.series}
	newTable.read = func(t *Table) iterator {
		return &filterIter{index: -1, filterSpec: filterSpec, raw: table.read(t)}
	}
	return newTable
}
