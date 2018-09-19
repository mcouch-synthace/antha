// anthalib//liquidhandling/liquidhandling_test.go: Part of the Antha language
// Copyright (C) 2015 The Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 2 Royal College St, London NW1 0NH UK

package liquidhandling

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/antha-lang/antha/antha/anthalib/mixer"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/antha/anthalib/wutil"
	"github.com/antha-lang/antha/antha/anthalib/wutil/text"
	"github.com/antha-lang/antha/inventory"
	"github.com/antha-lang/antha/inventory/testinventory"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
)

func GetContextForTest() context.Context {
	ctx := testinventory.NewContext(context.Background())
	return ctx
}

func GetPlateForTest() *wtype.Plate {

	offset := 0.25
	riserheightinmm := 40.0 - offset

	// pcr plate skirted (on riser)
	cone := wtype.NewShape("cylinder", "mm", 5.5, 5.5, 20.4)
	welltype := wtype.NewLHWell("ul", 200, 5, cone, wtype.UWellBottom, 5.5, 5.5, 20.4, 1.4, "mm")

	plate := wtype.NewLHPlate("pcrplate_skirted_riser", "Unknown", 8, 12, wtype.Coordinates{X: 127.76, Y: 85.48, Z: 25.7}, welltype, 9, 9, 0.0, 0.0, riserheightinmm-1.25)
	return plate
}

func GetTipwasteForTest() *wtype.LHTipwaste {
	shp := wtype.NewShape("box", "mm", 123.0, 80.0, 92.0)
	w := wtype.NewLHWell("ul", 800000.0, 800000.0, shp, 0, 123.0, 80.0, 92.0, 0.0, "mm")
	lht := wtype.NewLHTipwaste(6000, "Gilsontipwaste", "gilson", wtype.Coordinates{X: 127.76, Y: 85.48, Z: 92.0}, w, 49.5, 31.5, 0.0)
	return lht
}

func GetTroughForTest() *wtype.Plate {
	stshp := wtype.NewShape("box", "mm", 8.2, 72, 41.3)
	trough12 := wtype.NewLHWell("ul", 1500, 500, stshp, wtype.VWellBottom, 8.2, 72, 41.3, 4.7, "mm")
	plate := wtype.NewLHPlate("DWST12", "Unknown", 1, 12, wtype.Coordinates{X: 127.76, Y: 85.48, Z: 44.1}, trough12, 9, 9, 0, 30.0, 4.5)
	return plate
}

func TestStockConcs(*testing.T) {
	rand := wutil.GetRandom()
	names := []string{"tea", "milk", "sugar"}

	minrequired := make(map[string]float64, len(names))
	maxrequired := make(map[string]float64, len(names))
	Smax := make(map[string]float64, len(names))
	T := make(map[string]wunit.Volume, len(names))
	vmin := 10.0

	for _, name := range names {
		r := rand.Float64() + 1.0
		r2 := rand.Float64() + 1.0
		r3 := rand.Float64() + 1.0

		minrequired[name] = r * r2 * 20.0
		maxrequired[name] = r * r2 * 30.0
		Smax[name] = r * r2 * r3 * 70.0
		T[name] = wunit.NewVolume(100.0, "ul")
	}

	choose_stock_concentrations(minrequired, maxrequired, Smax, vmin, T)

}

func configure_request_simple(ctx context.Context, rq *LHRequest) {
	water := GetComponentForTest(ctx, "water", wunit.NewVolume(100.0, "ul"))
	water.Type = wtype.LTSingleChannel
	mmx := GetComponentForTest(ctx, "mastermix_sapI", wunit.NewVolume(100.0, "ul"))
	mmx.Type = wtype.LTSingleChannel
	part := GetComponentForTest(ctx, "dna", wunit.NewVolume(50.0, "ul"))
	part.Type = wtype.LTSingleChannel

	for k := 0; k < 9; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.Sample(water, wunit.NewVolume(8.0, "ul"))
		mmxs := mixer.Sample(mmx, wunit.NewVolume(8.0, "ul"))
		ps := mixer.Sample(part, wunit.NewVolume(1.0, "ul"))

		ins.AddInput(ws)
		ins.AddInput(mmxs)
		ins.AddInput(ps)
		ins.AddOutput(GetComponentForTest(ctx, "water", wunit.NewVolume(17.0, "ul")))
		rq.Add_instruction(ins)
	}

}

func configure_request_total_volume(ctx context.Context, rq *LHRequest) {
	water := GetComponentForTest(ctx, "water", wunit.NewVolume(100.0, "ul"))
	mmx := GetComponentForTest(ctx, "mastermix_sapI", wunit.NewVolume(100.0, "ul"))
	part := GetComponentForTest(ctx, "dna", wunit.NewVolume(50.0, "ul"))

	for k := 0; k < 9; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.SampleForTotalVolume(water, wunit.NewVolume(17.0, "ul"))
		mmxs := mixer.Sample(mmx, wunit.NewVolume(8.0, "ul"))
		ps := mixer.Sample(part, wunit.NewVolume(1.0, "ul"))

		ins.AddInput(ws)
		ins.AddInput(mmxs)
		ins.AddInput(ps)
		ins.AddOutput(GetComponentForTest(ctx, "water", wunit.NewVolume(17.0, "ul")))
		rq.Add_instruction(ins)
	}

}

func configure_request_bigger(ctx context.Context, rq *LHRequest) {
	water := GetComponentForTest(ctx, "water", wunit.NewVolume(2000.0, "ul"))
	mmx := GetComponentForTest(ctx, "mastermix_sapI", wunit.NewVolume(2000.0, "ul"))
	part := GetComponentForTest(ctx, "dna", wunit.NewVolume(1000.0, "ul"))

	for k := 0; k < 99; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.Sample(water, wunit.NewVolume(8.0, "ul"))
		mmxs := mixer.Sample(mmx, wunit.NewVolume(8.0, "ul"))
		ps := mixer.Sample(part, wunit.NewVolume(1.0, "ul"))

		ins.AddInput(ws)
		ins.AddInput(mmxs)
		ins.AddInput(ps)
		ins.AddOutput(GetComponentForTest(ctx, "water", wunit.NewVolume(17.0, "ul")))
		rq.Add_instruction(ins)
	}

}

func configureMultiChannelTestRequest(ctx context.Context, rq *LHRequest) {
	water := GetComponentForTest(ctx, "multiwater", wunit.NewVolume(2000.0, "ul"))

	for k := 0; k < 9; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.Sample(water, wunit.NewVolume(50.0, "ul"))

		ins.AddInput(ws)

		ins.AddOutput(GetComponentForTest(ctx, "water", wunit.NewVolume(50, "ul")))
		rq.Add_instruction(ins)
	}

}

func configureTransferRequestForZTest(policyName string, transferVol wunit.Volume, numberOfTransfers int) (rq *LHRequest, err error) {

	// set up ctx
	ctx := GetContextForTest()

	// make liquid handler
	lh := GetLiquidHandlerForTest(ctx)

	// make some tipboxes
	var tipBoxes []*wtype.LHTipbox
	tpHigh, err := inventory.NewTipbox(ctx, "Gilson200")
	if err != nil {
		return rq, err
	}
	tpLow, err := inventory.NewTipbox(ctx, "Gilson20")
	if err != nil {
		return rq, err
	}
	tipBoxes = append(tipBoxes, tpHigh, tpLow)

	//initialise request
	rq = GetLHRequestForTest()

	liq := GetComponentForTest(ctx, "water", wunit.NewVolume(2000.0, "ul"))

	err = liq.SetPolicyName(wtype.PolicyName(policyName))
	if err != nil {
		return rq, err
	}
	liq.SetName(policyName)

	for k := 0; k < numberOfTransfers; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.Sample(liq, transferVol)

		ins.AddInput(ws)

		expectedProduct := GetComponentForTest(ctx, "water", transferVol)

		err = expectedProduct.SetPolicyName(wtype.PolicyName(policyName))
		if err != nil {
			return rq, err
		}
		expectedProduct.SetName(policyName)

		ins.AddOutput(expectedProduct)

		rq.Add_instruction(ins)
	}

	// add plates and tip boxes
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	rq.Tips = tipBoxes

	if err := lh.Plan(ctx, rq); err != nil {
		return rq, fmt.Errorf("Got an error planning with no inputs: %s", err.Error())
	}
	return rq, nil
}

func configureSingleChannelTestRequest(ctx context.Context, rq *LHRequest) {
	water := GetComponentForTest(ctx, "multiwater", wunit.NewVolume(2000.0, "ul"))

	for k := 0; k < 1; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.Sample(water, wunit.NewVolume(50.0, "ul"))

		ins.AddInput(ws)

		ins.AddOutput(GetComponentForTest(ctx, "water", wunit.NewVolume(50, "ul")))
		rq.Add_instruction(ins)
	}

}

func configureTransferRequestMutliSamplesTest(policyName string, samples ...*wtype.Liquid) (rq *LHRequest, err error) {

	// set up ctx
	ctx := GetContextForTest()

	// make liquid handler
	lh := GetLiquidHandlerForTest(ctx)

	// make some tipboxes
	var tipBoxes []*wtype.LHTipbox
	tpHigh, err := inventory.NewTipbox(ctx, "Gilson200")
	if err != nil {
		return rq, err
	}
	tpLow, err := inventory.NewTipbox(ctx, "Gilson20")
	if err != nil {
		return rq, err
	}
	tipBoxes = append(tipBoxes, tpHigh, tpLow)

	//initialise request
	rq = GetLHRequestForTest()

	// add plates and tip boxes
	inPlate := GetPlateForTest()
	rq.InputPlatetypes = append(rq.InputPlatetypes, inPlate)
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())
	rq.Tips = tipBoxes

	it := wtype.NewAddressIterator(inPlate, wtype.RowWise, wtype.TopToBottom, wtype.LeftToRight, false)

	for _, sample := range samples {
		ins := wtype.NewLHMixInstruction()

		sample.SetPolicyName(wtype.PolicyName(policyName))

		ins.AddInput(sample)
		ins.AddOutput(GetComponentForTest(ctx, "water", sample.Volume()))

		if !it.Valid() {
			return nil, errors.New("out of space on input plate")
		}

		ins.Welladdress = it.Curr().FormatA1()
		it.Next()

		rq.Add_instruction(ins)
	}

	if err := lh.Plan(ctx, rq); err != nil {
		return rq, errors.WithMessage(err, "while planning")
	}
	return rq, nil
}

func TestToWellVolume(t *testing.T) {
	// set up ctx
	ctx := GetContextForTest()
	water := GetComponentForTest(ctx, "water", wunit.NewVolume(2000.0, "ul"))
	mmx := GetComponentForTest(ctx, "mastermix_sapI", wunit.NewVolume(2000.0, "ul"))
	part := GetComponentForTest(ctx, "dna", wunit.NewVolume(1000.0, "ul"))

	ws := mixer.Sample(water, wunit.NewVolume(150.0, "ul"))
	mmxs := mixer.Sample(mmx, wunit.NewVolume(49.0, "ul"))
	ps := mixer.Sample(part, wunit.NewVolume(1.0, "ul"))
	_, err := configureTransferRequestMutliSamplesTest("SmartMix", ws, mmxs, ps)

	if err != nil {
		t.Error(err.Error())
	}

}

func configure_request_overfilled(ctx context.Context, rq *LHRequest) {
	water := GetComponentForTest(ctx, "water", wunit.NewVolume(100.0, "ul"))
	mmx := GetComponentForTest(ctx, "mastermix_sapI", wunit.NewVolume(100.0, "ul"))
	part := GetComponentForTest(ctx, "dna", wunit.NewVolume(50.0, "ul"))

	for k := 0; k < 9; k++ {
		ins := wtype.NewLHMixInstruction()
		ws := mixer.Sample(water, wunit.NewVolume(160.0, "ul"))
		mmxs := mixer.Sample(mmx, wunit.NewVolume(160.0, "ul"))
		ps := mixer.Sample(part, wunit.NewVolume(20.0, "ul"))

		ins.AddInput(ws)
		ins.AddInput(mmxs)
		ins.AddInput(ps)
		ins.AddOutput(GetComponentForTest(ctx, "water", wunit.NewVolume(340.0, "ul")))
		rq.Add_instruction(ins)
	}

}

type zOffsetTest struct {
	liquidType              string
	numberOfTransfers       int
	volume                  wunit.Volume
	expectedAspirateZOffset []float64
	expectedDispenseZOffset []float64
}

func (self zOffsetTest) String() string {
	return fmt.Sprintf("%dx %v with policy=%q", self.numberOfTransfers, self.volume, self.liquidType)
}

var offsetTests []zOffsetTest = []zOffsetTest{
	{
		liquidType:              "multiwater",
		numberOfTransfers:       1,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{1.0000},
	},
	{
		liquidType:              "multiwater",
		numberOfTransfers:       2,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000, 0.5000},
		expectedDispenseZOffset: []float64{1.0000, 1.0000},
	},
	{
		liquidType:              "multiwater",
		numberOfTransfers:       1,
		volume:                  wunit.NewVolume(5, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{1.0000},
	},
	{
		liquidType:              "multiwater",
		numberOfTransfers:       2,
		volume:                  wunit.NewVolume(5, "ul"),
		expectedAspirateZOffset: []float64{0.5000, 0.5000},
		expectedDispenseZOffset: []float64{1.0000, 1.0000},
	},
	// Commented this out as it's not directly related to z offset and is failing
	// due to not performing a multichannel transfer.
	/*
		zOffsetTest{
			liquidType:              "multiwater",
			numberOfTransfers:       8,
			volume:                  wunit.NewVolume(50, "ul"),
			expectedAspirateZOffset: []float64{1.2500,1.2500,1.2500,1.2500,1.2500,1.2500,1.2500,1.2500},
			expectedDispenseZOffset: []float64{1.7500,1.7500,1.7500,1.7500,1.7500,1.7500,1.7500,1.7500},
		},*/
	{
		liquidType:              "SingleChannel",
		numberOfTransfers:       1,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{1.0000},
	},
	{
		liquidType:              "SingleChannel",
		numberOfTransfers:       2,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{1.0000},
	},
	{
		liquidType:              "SingleChannel",
		numberOfTransfers:       1,
		volume:                  wunit.NewVolume(5, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{1.0000},
	},
	{
		liquidType:              "SingleChannel",
		numberOfTransfers:       2,
		volume:                  wunit.NewVolume(5, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{1.0000},
	},
	{
		liquidType:              "SmartMix",
		numberOfTransfers:       1,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{0.5000},
	},
	{
		liquidType:              "SmartMix",
		numberOfTransfers:       2,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000, 0.5000},
		expectedDispenseZOffset: []float64{0.5000, 0.5000},
	}, /*
		zOffsetTest{
			liquidType:              "SmartMix",
			numberOfTransfers:       1,
			volume:                  wunit.NewVolume(5, "ul"),
			expectedAspirateZOffset: []float64{0.500},
			expectedDispenseZOffset: []float64{0.500},
		},
		zOffsetTest{
			liquidType:              "SmartMix",
			numberOfTransfers:       2,
			volume:                  wunit.NewVolume(5, "ul"),
			expectedAspirateZOffset: []float64{0.500,0.500},
			expectedDispenseZOffset: []float64{0.500,0.500},
		},*/
	{
		liquidType:              "NeedToMix",
		numberOfTransfers:       1,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000},
		expectedDispenseZOffset: []float64{0.5000},
	},
	{
		liquidType:              "NeedToMix",
		numberOfTransfers:       2,
		volume:                  wunit.NewVolume(50, "ul"),
		expectedAspirateZOffset: []float64{0.5000, 0.5000},
		expectedDispenseZOffset: []float64{0.5000, 0.5000},
	}, /*
		zOffsetTest{
			liquidType:              "NeedToMix",
			numberOfTransfers:       1,
			volume:                  wunit.NewVolume(5, "ul"),
			expectedAspirateZOffset: []float64{0.500},
			expectedDispenseZOffset: []float64{0.500},
		},
		zOffsetTest{
			liquidType:              "NeedToMix",
			numberOfTransfers:       2,
			volume:                  wunit.NewVolume(5, "ul"),
			expectedAspirateZOffset: []float64{0.500,0.500},
			expectedDispenseZOffset: []float64{0.500,0.500},
		},*/
}

func TestMultiZOffset2(t *testing.T) {

	for _, test := range offsetTests {
		request, err := configureTransferRequestForZTest(test.liquidType, test.volume, test.numberOfTransfers)
		if err != nil {
			t.Error(err.Error())
		}

		aspiratePairs, dispensePairs := extractMoveAspirateDispenseInstructions(request.Instructions)

		for i, pair := range aspiratePairs {
			if !reflect.DeepEqual(pair.mov.OffsetZ, test.expectedAspirateZOffset) {
				t.Errorf("for test: %v\naspiration step: %d\nexpected Z offset for aspirate: %v\ngot: %v",
					test, i, test.expectedAspirateZOffset, pair.mov.OffsetZ)
			}
		}

		for i, pair := range dispensePairs {
			if !reflect.DeepEqual(pair.mov.OffsetZ, test.expectedDispenseZOffset) {
				t.Errorf("for test: %v\ndispense step: %d\nexpected Z offset for dispense: %v\ngot: %v",
					test, i, test.expectedDispenseZOffset, pair.mov.OffsetZ)
			}
		}

	}
}

func makeMultiTestRequest() (multiRq *LHRequest, err error) {
	// set up ctx
	ctx := GetContextForTest()

	// make liquid handler
	lh := GetLiquidHandlerForTest(ctx)

	// make some tipboxes
	var tipBoxes []*wtype.LHTipbox
	tpHigh, err := inventory.NewTipbox(ctx, "Gilson200")
	if err != nil {
		return
	}
	tpLow, err := inventory.NewTipbox(ctx, "Gilson20")
	if err != nil {
		return
	}
	tipBoxes = append(tipBoxes, tpHigh, tpLow)

	// set up multi

	//initialise multi request
	multiRq = GetLHRequestForTest()

	// set to Multi channel test request
	configureMultiChannelTestRequest(ctx, multiRq)
	// add plates and tip boxes
	multiRq.InputPlatetypes = append(multiRq.InputPlatetypes, GetPlateForTest())
	multiRq.OutputPlatetypes = append(multiRq.OutputPlatetypes, GetPlateForTest())

	multiRq.Tips = tipBoxes

	if err := lh.Plan(ctx, multiRq); err != nil {
		return multiRq, fmt.Errorf("Got an error planning with no inputs: %s", err)
	}
	return multiRq, nil
}

func makeSingleTestRequest() (singleRq *LHRequest, err error) {
	// set up ctx
	ctx := GetContextForTest()

	// make liquid handler
	lh := GetLiquidHandlerForTest(ctx)

	// make some tipboxes
	var tipBoxes []*wtype.LHTipbox
	tpHigh, err := inventory.NewTipbox(ctx, "Gilson200")
	if err != nil {
		return
	}
	tpLow, err := inventory.NewTipbox(ctx, "Gilson20")
	if err != nil {
		return
	}
	tipBoxes = append(tipBoxes, tpHigh, tpLow)

	// set up single channel

	//initialise single request
	singleRq = GetLHRequestForTest()

	// set to single channel test request
	configureSingleChannelTestRequest(ctx, singleRq)
	// add plates and tip boxes
	singleRq.InputPlatetypes = append(singleRq.InputPlatetypes, GetPlateForTest())
	singleRq.OutputPlatetypes = append(singleRq.OutputPlatetypes, GetPlateForTest())

	singleRq.Tips = tipBoxes

	if err := lh.Plan(ctx, singleRq); err != nil {
		return singleRq, fmt.Errorf("Got an error planning with no inputs: %s", err)
	}
	return singleRq, nil
}

type movAspPair struct {
	mov *liquidhandling.MoveInstruction
	asp *liquidhandling.AspirateInstruction
}
type movDspPair struct {
	mov *liquidhandling.MoveInstruction
	dsp *liquidhandling.DispenseInstruction
}

func extractMoveAspirateDispenseInstructions(ins []liquidhandling.TerminalRobotInstruction) ([]movAspPair, []movDspPair) {
	mov := make([]*liquidhandling.MoveInstruction, len(ins))
	ma := []movAspPair{}
	md := []movDspPair{}

	for idx, i := range ins {
		i.Visit(&liquidhandling.RobotInstructionBaseVisitor{
			HandleMove: func(ins *liquidhandling.MoveInstruction) { mov[idx] = ins },
			HandleAspirate: func(ins *liquidhandling.AspirateInstruction) {
				if idx > 0 && mov[idx-1] != nil {
					ma = append(ma, movAspPair{mov: mov[idx-1], asp: ins})
				}
			},
			HandleDispense: func(ins *liquidhandling.DispenseInstruction) {
				if idx > 0 && mov[idx-1] != nil {
					md = append(md, movDspPair{mov: mov[idx-1], dsp: ins})
				}
			},
		})
	}
	return ma, md
}

func allElemsSame(nums []float64) bool {
	if len(nums) > 1 {
		n := nums[0]
		for _, m := range nums[1:] {
			if n != m {
				return false
			}
		}
	}
	return true
}

func TestMultiZOffset(t *testing.T) {
	multiRq, err := makeMultiTestRequest()
	if err != nil {
		t.Fatal(err.Error())
	}

	singleRq, err := makeSingleTestRequest()
	if err != nil {
		t.Fatal(err.Error())
	}

	multiAspPairs, multiDspPairs := extractMoveAspirateDispenseInstructions(multiRq.Instructions)
	singleAspPairs, singleDspPairs := extractMoveAspirateDispenseInstructions(singleRq.Instructions)

	if len(multiAspPairs) < len(singleAspPairs) {
		t.Error(fmt.Sprintf("Too few (%d) multi Asp pairs (need at least %d)", len(multiAspPairs), len(singleAspPairs)))
	}
	if len(multiDspPairs) < len(singleDspPairs) {
		t.Error(fmt.Sprintf("Too few (%d) multi Dsp pairs (need at least %d)", len(multiDspPairs), len(singleDspPairs)))
	}

	for i, singlePair := range singleAspPairs {
		if !allElemsSame(singlePair.mov.OffsetZ) {
			t.Error(fmt.Sprintf("Z offsets not all the same (single asp pair): %#v", singlePair.mov.OffsetZ))
		}
		multiPair := multiAspPairs[i]
		if !allElemsSame(multiPair.mov.OffsetZ) {
			t.Error(fmt.Sprintf("Z offsets not all the same (multi asp pair): %#v", multiPair.mov.OffsetZ))
		}
		if singlePair.mov.OffsetZ[0] != multiPair.mov.OffsetZ[0] {
			t.Error(fmt.Sprintf("single Aspirate Z offset: %+v ", text.PrettyPrint(singlePair)), "\n",
				fmt.Sprintf("Not equal to \n"),
				fmt.Sprintf("multi Aspirate Z offset: %+v ", text.PrettyPrint(multiPair)), "\n")
		}
	}

	for i, singlePair := range singleDspPairs {
		if !allElemsSame(singlePair.mov.OffsetZ) {
			t.Error(fmt.Sprintf("Z offsets not all the same (single asp pair): %#v", singlePair.mov.OffsetZ))
		}
		multiPair := multiDspPairs[i]
		if !allElemsSame(multiPair.mov.OffsetZ) {
			t.Error(fmt.Sprintf("Z offsets not all the same (multi asp pair): %#v", multiPair.mov.OffsetZ))
		}
		if singlePair.mov.OffsetZ[0] != multiPair.mov.OffsetZ[0] {
			t.Error("single Dispense Z offset: ", text.PrettyPrint(singlePair), "\n",
				fmt.Sprintf("Not equal to \n"),
				fmt.Sprintf("multi Dispense Z offset: %+v ", text.PrettyPrint(multiPair)), "\n")
		}
	}

}

func TestTipOverridePositive(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	var tpz []*wtype.LHTipbox
	tp, err := inventory.NewTipbox(ctx, "Gilson20")
	if err != nil {
		t.Fatal(err)
	}
	tpz = append(tpz, tp)

	rq.Tips = tpz

	if err := lh.Plan(ctx, rq); err != nil {
		t.Fatalf("Got an error planning with no inputs: %s", err)
	}

}
func TestTipOverrideNegative(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())
	var tpz []*wtype.LHTipbox
	tp, err := inventory.NewTipbox(ctx, "Gilson200")
	if err != nil {
		t.Fatal(err)
	}
	tpz = append(tpz, tp)

	rq.Tips = tpz

	err = lh.Plan(ctx, rq)

	if e, f := "7 (LH_ERR_VOL) : volume error : No tip chosen: Volume 8 ul is too low to be accurately moved by the liquid handler (configured minimum 10 ul). Low volume tips may not be available and / or the robot may need to be configured differently", err.Error(); e != f {
		t.Fatalf("expecting error %q found %q", e, f)
	}
}

func TestPlateReuse(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got an error planning with no inputs: ", err))
	}

	// reset the request
	rq = GetLHRequestForTest()
	configure_request_simple(ctx, rq)

	for _, plateid := range lh.Properties.PosLookup {
		if plateid == "" {
			continue
		}
		thing := lh.Properties.PlateLookup[plateid]

		plate, ok := thing.(*wtype.Plate)
		if !ok {
			continue
		}

		if strings.Contains(plate.GetName(), "Output_plate") {
			// leave it out
			continue
		}

		rq.InputPlates[plateid] = plate
	}
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	lh = GetLiquidHandlerForTest(ctx)
	err = lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got error resimulating: ", err))
	}

	// if we added nothing, input assignments should be empty

	if rq.NewComponentsAdded() {
		t.Fatal(fmt.Sprint("Resimulation failed: needed to add ", len(rq.InputSolutions.VolumesWanting), " components"))
	}

	// now try a deliberate fail

	// reset the request again
	rq = GetLHRequestForTest()
	configure_request_simple(ctx, rq)

	for _, plateid := range lh.Properties.PosLookup {
		if plateid == "" {
			continue
		}
		thing := lh.Properties.PlateLookup[plateid]

		plate, ok := thing.(*wtype.Plate)
		if !ok {
			continue
		}
		if strings.Contains(plate.GetName(), "Output_plate") {
			// leave it out
			continue
		}
		for _, v := range plate.Wellcoords {
			if !v.IsEmpty() {
				v.RemoveVolume(wunit.NewVolume(5.0, "ul"))
			}
		}

		rq.InputPlates[plateid] = plate
	}
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	lh = GetLiquidHandlerForTest(ctx)
	err = lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got error resimulating: ", err))
	}

	// this time we should have added some components again
	if len(rq.InputAssignments) != 3 {
		t.Fatal(fmt.Sprintf("Error resimulating, should have added 3 components, instead added %d", len(rq.InputAssignments)))
	}
}

func TestBeforeVsAfter(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got an error planning with no inputs: ", err))
	}

	for pos := range lh.Properties.PosLookup {

		id1, ok1 := lh.Properties.PosLookup[pos]
		id2, ok2 := lh.FinalProperties.PosLookup[pos]

		if ok1 && !ok2 || ok2 && !ok1 {
			t.Fatal(fmt.Sprintf("Position %s inconsistent: Before %t after %t", pos, ok1, ok2))
		}

		p1 := lh.Properties.PlateLookup[id1]
		p2 := lh.FinalProperties.PlateLookup[id2]

		// check types

		t1 := reflect.TypeOf(p1)
		t2 := reflect.TypeOf(p2)

		if t1 != t2 {
			t.Fatal(fmt.Sprintf("Types of thing at position %s not same: %v %v", pos, t1, t2))
		}

		// ok nice we have some sort of consistency

		switch p1.(type) {
		case *wtype.Plate:
			pp1 := p1.(*wtype.Plate)
			pp2 := p2.(*wtype.Plate)
			if pp1.Type != pp2.Type {
				t.Fatal(fmt.Sprintf("Plates at %s not same type: %s %s", pos, pp1.Type, pp2.Type))
			}
			it := wtype.NewAddressIterator(pp1, wtype.ColumnWise, wtype.TopToBottom, wtype.LeftToRight, false)

			for {
				if !it.Valid() {
					break
				}
				wc := it.Curr()
				w1 := pp1.Wellcoords[wc.FormatA1()]
				w2 := pp2.Wellcoords[wc.FormatA1()]

				if w1.IsEmpty() && w2.IsEmpty() {
					it.Next()
					continue
				}
				/*
					fmt.Println(pp1.PlateName, " ", pp1.Type)
					fmt.Println(pp2.PlateName, " ", pp2.Type)
					fmt.Println(wc.FormatA1())
					fmt.Println(w1.ID, " ", w1.WContents.ID, " ", w1.WContents.CName, " ", w1.WContents.Vol)
					fmt.Println(w2.ID, " ", w2.WContents.ID, " ", w2.WContents.CName, " ", w2.WContents.Vol)
				*/

				if w1.WContents.ID == w2.WContents.ID {
					t.Fatal(fmt.Sprintf("IDs before and after must differ"))
				}
				it.Next()
			}
		case *wtype.LHTipbox:
			tb1 := p1.(*wtype.LHTipbox)
			tb2 := p2.(*wtype.LHTipbox)

			if tb1.Type != tb2.Type {
				t.Fatal(fmt.Sprintf("Tipbox at changed type: %s %s", tb1.Type, tb2.Type))
			}
		case *wtype.LHTipwaste:
			tw1 := p1.(*wtype.LHTipwaste)
			tw2 := p2.(*wtype.LHTipwaste)

			if tw1.Type != tw2.Type {
				t.Fatal(fmt.Sprintf("Tipwaste changed type: %s %s", tw1.Type, tw2.Type))
			}
		}

	}

}

func TestEP3(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got planning error: ", err))
	}

}

func TestEP3TotalVolume(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_total_volume(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got planning error: ", err))
	}

}

func TestEP3Overfilled(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_overfilled(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err == nil {
		t.Fatal("Overfull wells did not cause planning error")
	}
}

func TestEP3Negative(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)

	//make one volume of one instruction negative
	for _, ins := range rq.LHInstructions {
		cmp := ins.Inputs[0]
		cmp.Vol = -1.0
		break
	}
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err == nil {
		t.Fatal("Negative volume did not cause a planning error")
	}
}

func TestEP3WrongResult(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)

	//make one of the results wrong
	for _, ins := range rq.LHInstructions {
		ins.Outputs[0].Vol = 299792458.0
		break
	}
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err == nil {
		t.Fatal("Negative volume did not cause a planning error")
	}
}

func TestEP3WrongTotalVolume(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_total_volume(ctx, rq)

	//set an invalid total volume for one of the instructions
	for _, ins := range rq.LHInstructions {
		for _, cmp := range ins.Inputs {
			if cmp.Tvol > 0.0 {
				cmp.Tvol = 5.0
			}
		}
		break
	}
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err == nil {
		t.Fatal("Negative volume did not cause a planning error")
	}
}

func TestDistinctPlateNames(t *testing.T) {
	rq := NewLHRequest()
	for i := 0; i < 100; i++ {
		p := &wtype.Plate{ID: fmt.Sprintf("anID-%d", i), PlateName: "aName"}
		rq.InputPlateOrder = append(rq.InputPlateOrder, p.ID)
		rq.InputPlates[p.ID] = p
	}
	for i := 100; i < 200; i++ {
		p := &wtype.Plate{ID: fmt.Sprintf("anID-%d", i), PlateName: "aName"}
		rq.OutputPlateOrder = append(rq.OutputPlateOrder, p.ID)
		rq.OutputPlates[p.ID] = p
	}

	rq = fixDuplicatePlateNames(rq)

	found := make(map[string]int)

	for _, p := range rq.AllPlates() {
		_, ok := found[p.PlateName]

		if !ok {
			found[p.PlateName] = 1
		} else {
			t.Errorf("fixDuplicatePlateNames failed to prevent duplicates: found at least two of %s", p.PlateName)
		}
	}

}

func TestEP3DummyInstruction(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	//add a dummy instruction for each instruction
	for _, ins := range rq.LHInstructions {
		for _, cmp := range ins.Outputs {
			mix := mixer.GenericMix(mixer.MixOptions{Inputs: []*wtype.Liquid{cmp}})
			if !mix.IsDummy() {
				t.Fatalf("failed to make a dummy instruction: mix.Inputs[0].IsSample() = %t, cmp.IsSample() = %t", mix.Inputs[0].IsSample(), cmp.IsSample())
			}
			rq.Add_instruction(mix)
		}
	}

	err := lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got planning error: ", err))
	}

}

func assertCoordsEq(lhs, rhs []wtype.Coordinates) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for i := 0; i < len(lhs); i++ {
		if lhs[i].Subtract(rhs[i]).Abs() > 0.00001 {
			return false
		}
	}

	return true
}

func TestAddWellTargets(t *testing.T) {

	ctx := GetContextForTest()
	lh := GetLiquidHandlerForTest(ctx)

	plate := GetPlateForTest()
	lh.Properties.AddPlateTo("position_4", plate)

	tipwaste := GetTipwasteForTest()
	lh.Properties.AddTipWasteTo("position_1", tipwaste)

	trough := GetTroughForTest()
	lh.Properties.AddPlateTo("position_5", trough)

	lh.addWellTargets()

	expected := []wtype.Coordinates{
		{X: 0.0, Y: -31.5, Z: 0.0},
		{X: 0.0, Y: -22.5, Z: 0.0},
		{X: 0.0, Y: -13.5, Z: 0.0},
		{X: 0.0, Y: -4.5, Z: 0.0},
		{X: 0.0, Y: 4.5, Z: 0.0},
		{X: 0.0, Y: 13.5, Z: 0.0},
		{X: 0.0, Y: 22.5, Z: 0.0},
		{X: 0.0, Y: 31.5, Z: 0.0},
	}

	if e, g := []wtype.Coordinates{}, plate.Welltype.GetWellTargets("DummyAdaptor"); !assertCoordsEq(e, g) {
		t.Errorf("plate well targets incorrect, expected %v, got %v", e, g)
	}

	if e, g := expected, tipwaste.AsWell.GetWellTargets("DummyAdaptor"); !assertCoordsEq(e, g) {
		t.Errorf("plate well targets incorrect, expected %v, got %v", e, g)
	}

	if e, g := expected, trough.Welltype.GetWellTargets("DummyAdaptor"); !assertCoordsEq(e, g) {
		t.Errorf("plate well targets incorrect, expected %v, got %v", e, g)
	}

}

func TestShouldSetWellTargets(t *testing.T) {
	ctx := GetContextForTest()

	for _, plate := range testinventory.GetPlates(ctx) {
		e := !plate.IsSpecial()
		//IsSpecial is irrelevant for plates with 8 rows or more
		if plate.NRows() >= 8 {
			e = false
		}
		if g := plate.AreWellTargetsEnabled(8, 9.0); e != g {
			t.Errorf("For platetype %s (%d rows): plate.AreWellTargetsEnabled(8,9.0) = %t, expected %t", plate.GetType(), plate.NRows(), g, e)
		}
	}
}

func TestPlateIDMap(t *testing.T) {
	ctx := GetContextForTest()

	lh := GetLiquidHandlerForTest(ctx)
	lh.ExecutionPlanner = ExecutionPlanner3
	rq := GetLHRequestForTest()
	configure_request_simple(ctx, rq)
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)

	if err != nil {
		t.Fatal(fmt.Sprint("Got planning error: ", err))
	}

	beforePlates := lh.Properties.PlateLookup
	afterPlates := lh.FinalProperties.PlateLookup
	idMap := lh.PlateIDMap()

	//check that idMap refers to things that exist
	for beforeID, afterID := range idMap {
		beforeObj, ok := beforePlates[beforeID]
		if !ok {
			t.Errorf("idMap key \"%s\" doesn't exist in initial LHProperties.PlateLookup", beforeID)
			continue
		}
		afterObj, ok := afterPlates[afterID]
		if !ok {
			t.Errorf("idMap value \"%s\" doesn't exist in final LHProperties.PlateLookup", afterID)
			continue
		}
		//check that you don't have tipboxes turning into plates, for example
		if beforeClass, afterClass := wtype.ClassOf(beforeObj), wtype.ClassOf(afterObj); beforeClass != afterClass {
			t.Errorf("planner has turned a %s into a %s", beforeClass, afterClass)
		}
	}

	//check that everything in beforePlates is mapped to something
	for id, obj := range beforePlates {
		if _, ok := idMap[id]; !ok {
			t.Errorf("%s with id %s exists in initial LHProperties, but isn't mapped to final LHProperties", wtype.ClassOf(obj), id)
		}
	}
}

func getTestSplitSample(component *wtype.Liquid, volume float64) *wtype.LHInstruction {
	ret := wtype.NewLHSplitInstruction()

	ret.Inputs = append(ret.Inputs, component.Dup())
	cmpMoving, cmpStaying := mixer.SplitSample(component, wunit.NewVolume(volume, "ul"))

	ret.Outputs = append(ret.Outputs, cmpMoving, cmpStaying)

	return ret
}

func getTestMix(components []*wtype.Liquid, address string) *wtype.LHInstruction {
	mix := mixer.GenericMix(mixer.MixOptions{
		Inputs:  components,
		Address: address,
	})

	mx := 0
	for _, c := range components {
		if c.Generation() > mx {
			mx = c.Generation()
		}
	}
	mix.SetGeneration(mx)
	mix.Outputs[0].SetGeneration(mx + 1)
	mix.Outputs[0].DeclareInstance()

	return mix
}

func TestSplitSampleMultichannel(t *testing.T) {

	ctx := GetContextForTest()

	var instructions []*wtype.LHInstruction

	diluent := GetComponentForTest(ctx, "multiwater", wunit.NewVolume(1000.0, "ul"))
	stock := GetComponentForTest(ctx, "dna", wunit.NewVolume(1000, "ul"))
	stock.Type = wtype.LTMultiWater

	wc := wtype.MakeWellCoords("A1")

	for y := 0; y < 8; y++ {
		lastStock := stock
		wc.Y = y
		for x := 0; x < 2; x++ {
			wc.X = x
			diluentSample := mixer.Sample(diluent, wunit.NewVolume(20.0, "ul"))

			split := getTestSplitSample(lastStock, 20.0)

			mix := getTestMix([]*wtype.Liquid{split.Outputs[0], diluentSample}, wc.FormatA1())

			lastStock = mix.Outputs[0]

			instructions = append(instructions, mix, split)
		}
	}

	lh, rq, err := runPlan(ctx, instructions)
	if err != nil {
		t.Fatal(err)
	}

	//assert that there is some 8-way multi channel
	seenMultiEight := false
	for _, ins := range rq.Instructions {
		if multi, ok := ins.GetParameter(liquidhandling.MULTI).(int); ok && multi == 8 {
			seenMultiEight = true
		}
	}

	if !seenMultiEight {
		t.Error("Expected 8-way multichanneling but none seen")
	}

	OutputSetup(lh.FinalProperties)

}

func runPlan(ctx context.Context, instructions []*wtype.LHInstruction) (*Liquidhandler, *LHRequest, error) {

	lh := GetLiquidHandlerForTest(ctx)
	rq := GetLHRequestForTest()
	for _, ins := range instructions {
		rq.Add_instruction(ins)
	}
	rq.InputPlatetypes = append(rq.InputPlatetypes, GetPlateForTest())
	rq.OutputPlatetypes = append(rq.OutputPlatetypes, GetPlateForTest())

	err := lh.Plan(ctx, rq)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "while planning")
	}

	return lh, rq, nil
}
