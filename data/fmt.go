package data

import (
	"fmt"
	"strings"
)

// Pretty printing Rows
var _ fmt.Stringer = (Rows)(nil)
var _ fmt.Stringer = Row{}

// TODO pretty print might be better on Table, since this fn can't print header
// for empty slice
func (r Rows) String() string {
	if len(r) == 0 {
		return "(no Rows)"
	}
	// note: lots of garbage
	const sep = "|"
	const line = "-"
	// row-major array
	cellVals := make([][]interface{}, len(r)+1)
	var colMaxes []int

	add := func(i, c int, value interface{}) {
		// TODO the format could be reflected from the schema
		str := fmt.Sprintf("%+v", value)
		cellVals[i] = append(cellVals[i], str)
		if len(str) > colMaxes[c] {
			colMaxes[c] = len(str)
		}
	}
	// TODO also print the row indices
	for idx, rr := range r {
		if idx == 0 {
			colMaxes = make([]int, len(rr.Values))
			// headers
			for c, o := range rr.Values {
				add(0, c, o.ColumnName())
			}
		}
		for c, o := range rr.Values {
			add(idx+1, c, o.value)
		}
	}
	fmtStrBuilder := strings.Builder{}

	for _, colMax := range colMaxes {
		// TODO pad string columns (only) on the right. etc
		fmtStrBuilder.WriteString(fmt.Sprintf("%s%%%ds", sep, colMax))
	}
	fmtStrBuilder.WriteString(sep + "\n")
	fmtStr := fmtStrBuilder.String()
	builder := strings.Builder{}
	for idx, cells := range cellVals {
		rowStr := fmt.Sprintf(fmtStr, cells...)
		builder.WriteString(rowStr)
		if idx == 0 {
			builder.WriteString(strings.Repeat(line, len(rowStr)))
			builder.WriteString("\n")
		}
	}
	return fmt.Sprintf("%d Row(s):\n%s", len(cellVals)-1, builder.String())
}

// print a single row as if it's a 1-row table
func (r Row) String() string {
	return fmt.Sprintf("Row #%d:\n%v", r.Index, Rows{r})
}
