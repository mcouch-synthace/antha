package main

import (
	"fmt"

	"github.com/antha-lang/antha/data"
)

func ExampleSliceSeries() {
	// In this example we're creating a bounded table
	// from scalar slices.  This means reflection
	// but no copying; the representation is dense
	// and nullable series are not supported.

	// create the raw data.
	column1 := data.Must().NewSliceSeries("measure", []int64{10, 10, 30, 0, 5})
	column2 := data.Must().NewSliceSeries("label", []string{"", "aa", "xx", "aa", ""})

	// populate the Table.
	tab := data.NewTable([]*data.Series{column1, column2})

	// just print Table as a whole.
	fmt.Println("before filter\n", tab.ToRows())

	// iterate over the entire Table.
	for record := range tab.Iter() {
		fmt.Println(record)
		m, _ := record.Observation("measure")
		fmt.Println(m.MustInt())
	}
	// TODO...
	// // produce a new Table by filtering dynamically.
	// smallerTab := tab.Filter(data.Eq{"label", "aa"})

	// // add a column
	// extended := tab.ExtendBy(func(r Row) {}, "colInt", reflect.Int32)

	// // get all the records.
	// recordsFiltered := smallerTab.ToRows()
	// fmt.Println("after filter\n", recordsFiltered)
}

func main() {
	ExampleSliceSeries()
}
