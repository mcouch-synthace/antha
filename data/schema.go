package data

import (
	"reflect"

	"github.com/pkg/errors"
)

type ColumnName string
type Index int
type Key []ColumnName

// Schema is intended as an immutable representation of table metadata
type Schema struct {
	Columns []Column
	byName  map[ColumnName][]int
}

// Column is Series metadata
// TODO: model nullability
type Column struct {
	Name ColumnName
	Type reflect.Type
}

// Size is number of columns
func (s Schema) Size() int {
	return len(s.Columns)
}

// Equal returns true if order, name and types match
func (s Schema) Equal(other Schema) bool {
	if s.Size() != other.Size() {
		return false
	}
	for i, c := range s.Columns {
		if c != other.Columns[i] {
			return false
		}
	}
	return true
}

// Col gets the column by name, first matched
func (s Schema) Col(col ColumnName) (Column, error) {
	cs, found := s.byName[col]
	if found {
		return s.Columns[cs[0]], nil
	}
	return Column{}, errors.Errorf("no such column: %v", col)
}

// TODO String()

func newSchema(series []*Series) Schema {
	schema := Schema{byName: map[ColumnName][]int{}}
	for c, s := range series {
		schema.Columns = append(schema.Columns, Column{Type: s.typ, Name: s.col})
		schema.byName[s.col] = append(schema.byName[s.col], c)
	}
	return schema
}
