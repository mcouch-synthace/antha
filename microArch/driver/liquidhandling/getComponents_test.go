package liquidhandling

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

type DistributeVolumesTest struct {
	Name            string
	Requested       []float64   // Volumes to request per channel
	AvailableByWell [][]float64 // Volumes available by well then channel - total number of floats should match requested
	Expected        []float64   // Expected volumes supplied to each channel
}

func (test *DistributeVolumesTest) Run(t *testing.T) {
	requested := make(wtype.ComponentVector, 0, len(test.Requested))
	for _, v := range test.Requested {
		requested = append(requested, &wtype.Liquid{Vol: v, Vunit: "ul"})
	}

	available := make(wtype.ComponentVector, 0, len(test.Requested))
	for w, wv := range test.AvailableByWell {
		loc := fmt.Sprintf("well_%d", w)
		for _, v := range wv {
			available = append(available, &wtype.Liquid{Vol: v, Vunit: "ul", Loc: loc})
		}
	}

	if len(requested) != len(available) { //test is bad, explode
		t.Fatalf("bad test: len(requested) != len(available): %d != %d", len(requested), len(available))
	}

	got := distributeVolumes(requested, available)
	gotVols := make([]float64, 0, len(got))
	for _, v := range got {
		gotVols = append(gotVols, v.Vol)
	}

	if !reflect.DeepEqual(gotVols, test.Expected) {
		t.Errorf("return didn't match expected:\ne: %v\ng: %v", test.Expected, gotVols)
	}
}

type DistributeVolumesTests []*DistributeVolumesTest

func (tests DistributeVolumesTests) Run(t *testing.T) {
	for _, test := range tests {
		t.Run(test.Name, test.Run)
	}
}

func TestDistrubuteVolumes(t *testing.T) {
	DistributeVolumesTests{
		{
			Name:            "all equal with excess",
			Requested:       []float64{100, 100, 100, 100},
			AvailableByWell: [][]float64{{500, 500, 500, 500}},
			Expected:        []float64{100 + 25, 100 + 25, 100 + 25, 100 + 25}, // value allocate by need + evenly distributed excess
		},
		{
			Name:            "all equal exact match",
			Requested:       []float64{100, 100, 100, 100},
			AvailableByWell: [][]float64{{400, 400, 400, 400}},
			Expected:        []float64{100, 100, 100, 100},
		},
		{
			Name:            "all equal with shortfall",
			Requested:       []float64{100, 100, 100, 100},
			AvailableByWell: [][]float64{{200, 200, 200, 200}},
			Expected:        []float64{50, 50, 50, 50},
		},
		{
			Name:            "mixed with excess",
			Requested:       []float64{20, 100, 0, 70},
			AvailableByWell: [][]float64{{500, 500, 500, 500}},
			Expected:        []float64{20 + 77.5, 100 + 77.5, 0 + 77.5, 70 + 77.5},
		},
		{
			Name:            "mixed exact match",
			Requested:       []float64{20, 100, 0, 70},
			AvailableByWell: [][]float64{{190, 190, 190, 190}},
			Expected:        []float64{20, 100, 0, 70},
		},
		{
			Name:            "mixed with shortfall",
			Requested:       []float64{20, 100, 0, 70},
			AvailableByWell: [][]float64{{100, 100, 100, 100}},
			Expected:        []float64{20, 40, 0, 40},
		},
		{
			Name:            "mixed with excess multiwell",
			Requested:       []float64{20, 100, 0, 70},
			AvailableByWell: [][]float64{{500, 500}, {500, 500}},
			Expected:        []float64{20 + (500-120)/2.0, 100 + (500-120)/2.0, 0 + (500-70)/2.0, 70 + (500-70)/2.0},
		},
		{
			Name:            "mixed exact match multiwell",
			Requested:       []float64{20, 100, 0, 70},
			AvailableByWell: [][]float64{{120, 120}, {70, 70}},
			Expected:        []float64{20, 100, 0, 70},
		},
		{
			Name:            "mixed with shortfall multiwell",
			Requested:       []float64{20, 100, 0, 70},
			AvailableByWell: [][]float64{{50, 50}, {50, 50}},
			Expected:        []float64{20, 30, 0, 50},
		},
	}.Run(t)
}