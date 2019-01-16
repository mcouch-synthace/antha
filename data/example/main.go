package main

import (
	"fmt"
	"reflect"

	"github.com/antha-lang/antha/data"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
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
		m, _ := record.Observation("measure")
		fmt.Println("int measure value:", m.MustInt())
	}
	// produce a new Table by filtering
	smallerTab := tab.Filter(data.Eq("label", "aa"))
	fmt.Println("after filter\n", smallerTab.ToRows())
	// note the exact type matching required here
	fmt.Println("after filter 2\n", smallerTab.Filter(data.Eq("measure", int64(0))).ToRows())

	// add a column
	extended := tab.ExtendBy(func(r data.Row) interface{} {
		measure, _ := r.Observation("measure")
		return float64(measure.MustInt()) * float64(2.5)
	}, "multiplied", reflect.TypeOf(float64(0)))
	fmt.Println("extended\n", extended.ToRows())
}

func ExampleArrowSeries() {
	// This example is the same as the previous one, but:
	// 1) the Series are Arrow-based (instead of slice-based ones)
	// 2) nulls are allowed

	// create the arrow series from slices and validity masks
	column1 := data.NewArrowSeriesFromSliceInt64("measure", []int64{10, 10, 30, 0, 5}, []bool{true, true, true, false, true})
	column2 := data.NewArrowSeriesFromSliceString("label", []string{"", "aa", "xx", "aa", ""}, nil)

	// create the Arrow series manually using builder
	builder := array.NewFloat64Builder(memory.DefaultAllocator)
	builder.Append(5.1)
	builder.AppendValues([]float64{2, 3, 0}, []bool{true, true, false})
	builder.AppendNull()
	arrowArray := builder.NewFloat64Array()

	column3 := data.NewArrowSeriesFloat64("float_measure", arrowArray)

	// populate the Table.
	tab := data.NewTable([]*data.Series{column1, column2, column3})

	// just print Table as a whole.
	fmt.Println("before filter\n", tab.ToRows())

	// iterate over the entire Table.
	for record := range tab.Iter() {
		m, _ := record.Observation("measure")
		if !m.IsNull() {
			fmt.Println("int measure value:", m.MustInt())
		}
	}
	// produce a new Table by filtering
	smallerTab := tab.Filter(data.Eq("label", "aa"))
	fmt.Println("after filter\n", smallerTab.ToRows())
	// note the exact type matching required here
	fmt.Println("after filter 2\n", smallerTab.Filter(data.Eq("measure", int64(0))).ToRows())

	// add a column
	extended := tab.ExtendBy(func(r data.Row) interface{} {
		measure, _ := r.Observation("measure")
		if measure.IsNull() {
			return nil
		}
		return float64(measure.MustInt()) * float64(2.5)
	}, "multiplied", reflect.TypeOf(float64(0)))
	fmt.Println("extended\n", extended.ToRows())
}

func main() {
	//ExampleSliceSeries()
	ExampleArrowSeries()
}
