package data

import (
	"reflect"

	"github.com/pkg/errors"
)

// Schema is intended as an immutable representation of table metadata
type Schema struct {
	Columns []Column
	byName  map[ColumnName]int
}

// Column is Series metadata
type Column struct {
	Name ColumnName
	Type reflect.Type
}

// Col gets the column by name
func (s Schema) Col(col ColumnName) (Column, error) {
	c, found := s.byName[col]
	if found {
		return s.Columns[c], nil
	}
	return Column{}, errors.Errorf("no such column: %v", col)
}

// TODO String()

func newSchema(series []*Series) Schema {
	schema := Schema{byName: map[ColumnName]int{}}
	for c, s := range series {
		schema.Columns = append(schema.Columns, Column{Type: s.typ, Name: s.col})
		schema.byName[s.col] = c
	}
	return schema
}
