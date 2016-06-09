package lib

import (
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/bvendor/golang.org/x/net/context"
	"github.com/antha-lang/antha/execute"
	"github.com/antha-lang/antha/inject"
)

// Input parameters for this protocol

// Data which is returned from this protocol

// Physical inputs to this protocol

// Physical outputs from this protocol

func _Units_SumRequirements() {

}

// Actions to perform before protocol itself
func _Units_SumSetup(_ctx context.Context, _input *Units_SumInput) {

}

// Core process of the protocol: steps to be performed for each input
func _Units_SumSteps(_ctx context.Context, _input *Units_SumInput, _output *Units_SumOutput) {

	var sumofSIValues float64
	var siUnit string

	sumofSIValues = _input.MyVolume.SIValue() + _input.MyOtherVolume.SIValue()

	siUnit = _input.MyVolume.Unit().BaseSISymbol()

	// or a less safe but simpler way would be
	// siUnit = "l"

	_output.SumOfVolumes = wunit.NewVolume(sumofSIValues, siUnit)

}

// Actions to perform after steps block to analyze data
func _Units_SumAnalysis(_ctx context.Context, _input *Units_SumInput, _output *Units_SumOutput) {

}

func _Units_SumValidation(_ctx context.Context, _input *Units_SumInput, _output *Units_SumOutput) {

}
func _Units_SumRun(_ctx context.Context, input *Units_SumInput) *Units_SumOutput {
	output := &Units_SumOutput{}
	_Units_SumSetup(_ctx, input)
	_Units_SumSteps(_ctx, input, output)
	_Units_SumAnalysis(_ctx, input, output)
	_Units_SumValidation(_ctx, input, output)
	return output
}

func Units_SumRunSteps(_ctx context.Context, input *Units_SumInput) *Units_SumSOutput {
	soutput := &Units_SumSOutput{}
	output := _Units_SumRun(_ctx, input)
	if err := inject.AssignSome(output, &soutput.Data); err != nil {
		panic(err)
	}
	if err := inject.AssignSome(output, &soutput.Outputs); err != nil {
		panic(err)
	}
	return soutput
}

func Units_SumNew() interface{} {
	return &Units_SumElement{
		inject.CheckedRunner{
			RunFunc: func(_ctx context.Context, value inject.Value) (inject.Value, error) {
				input := &Units_SumInput{}
				if err := inject.Assign(value, input); err != nil {
					return nil, err
				}
				output := _Units_SumRun(_ctx, input)
				return inject.MakeValue(output), nil
			},
			In:  &Units_SumInput{},
			Out: &Units_SumOutput{},
		},
	}
}

var (
	_ = execute.MixInto
	_ = wunit.Make_units
)

type Units_SumElement struct {
	inject.CheckedRunner
}

type Units_SumInput struct {
	MyOtherVolume wunit.Volume
	MyVolume      wunit.Volume
}

type Units_SumOutput struct {
	SumOfVolumes wunit.Volume
}

type Units_SumSOutput struct {
	Data struct {
		SumOfVolumes wunit.Volume
	}
	Outputs struct {
	}
}

func init() {
	addComponent(Component{Name: "Units_Sum",
		Constructor: Units_SumNew,
		Desc: ComponentDesc{
			Desc: "",
			Path: "antha/component/an/AnthaAcademy/Lesson0_Units/units_Sum.an",
			Params: []ParamDesc{
				{Name: "MyOtherVolume", Desc: "", Kind: "Parameters"},
				{Name: "MyVolume", Desc: "", Kind: "Parameters"},
				{Name: "SumOfVolumes", Desc: "", Kind: "Data"},
			},
		},
	})
}
