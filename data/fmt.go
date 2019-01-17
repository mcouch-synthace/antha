package data

import (
	"fmt"
	"strings"
)

// Pretty printing Rows
var _ fmt.Stringer = Rows{}
var _ fmt.Stringer = Row{}

// String formats the rows as an ascii art table
func (r Rows) String() string {
	const sep = "|"
	const line = "-"
	const hdrSize = 2
	// row-major array with offset for header rows
	cellVals := make([][]interface{}, len(r.Data)+hdrSize)
	colMaxes := make([]int, len(r.Schema.Columns)+1)

	add := func(rownumWithOffset, colnumWithOffset int, value interface{}) {
		// TODO an appropriate format could be determined from the schema
		str := fmt.Sprintf("%+v", value)
		cellVals[rownumWithOffset] = append(cellVals[rownumWithOffset], str)
		if len(str) > colMaxes[colnumWithOffset] {
			colMaxes[colnumWithOffset] = len(str)
		}
	}
	// headers
	add(0, 0, "")
	add(1, 0, "")

	for c, col := range r.Schema.Columns {
		add(0, c+1, col.Name)
		add(1, c+1, col.Type)
	}

	for rownum, rr := range r.Data {
		add(rownum+hdrSize, 0, rr.Index)
		for c, o := range rr.Values {
			add(rownum+hdrSize, c+1, o.value)
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
		if idx == 1 {
			// divider
			builder.WriteString(strings.Repeat(line, len(rowStr)-1))
			builder.WriteString("\n")
		}
	}
	return fmt.Sprintf("%d Row(s):\n%s", len(cellVals)-hdrSize, builder.String())
}

// print a single row as if it's a 1-row table
// FIXME this is broken
func (r Row) String() string {
	return fmt.Sprintf("Row #%d:\n%v", r.Index, Rows{Data: []Row{r}})
}
