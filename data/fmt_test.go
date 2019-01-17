package data

import (
	"testing"
	// TODO "github.com/stretchr/testify/assert"
)

func TestFmt(t *testing.T) {
	tab := NewTable([]*Series{
		Must().NewSliceSeries("measure", []int64{1, 1000}),
		Must().NewSliceSeries("label", []string{"abcdef", "abcd"}),
	})
	formatted := tab.ToRows().String()
	if formatted != `2 Row(s):
| |measure| label|
| |  int64|string|
------------------
|0|      1|abcdef|
|1|   1000|  abcd|
` {
		t.Errorf("fmt: %s", formatted)
	}
}

func TestFmtEmpty(t *testing.T) {
	tab := NewTable([]*Series{
		Must().NewSliceSeries("A", []float64{}),
	})
	formatted := tab.ToRows().String()
	expected := `0 Row(s):
||      A|
||float64|
----------
`
	if formatted != expected {
		t.Errorf("fmt: %s", formatted)
	}
}
