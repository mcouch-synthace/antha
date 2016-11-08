// Protocol for resuspending freeze dried DNA with a diluent
package lib

import

// we need to import the wtype package to use the LHComponent type
// the mixer package is required to use the Sample function
(
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/component"
	"github.com/antha-lang/antha/execute"
	"github.com/antha-lang/antha/inject"
	"golang.org/x/net/context"
)

// Input parameters for this protocol (data)

func _ResuspendDNA_MultipleRequirements() {
}

func _ResuspendDNA_MultipleSetup(_ctx context.Context, _input *ResuspendDNA_MultipleInput) {
}

func _ResuspendDNA_MultipleSteps(_ctx context.Context, _input *ResuspendDNA_MultipleInput, _output *ResuspendDNA_MultipleOutput) {

	_output.ResuspendedDNAMap = make(map[string]*wtype.LHComponent)

	for _, part := range _input.Parts {

		result := ResuspendDNARunSteps(_ctx, &ResuspendDNAInput{DNAMass: _input.PartMassMap[part],
			TargetConc:      _input.TargetConc,
			MolecularWeight: _input.PartMolecularWeightMap[part],
			Well:            _input.PartLocationsMap[part],

			Diluent:  _input.Diluent,
			DNAPlate: _input.DNAPlate},
		)
		_output.ResuspendedDNAMap[part] = result.Outputs.ResuspendedDNA
	}
}

func _ResuspendDNA_MultipleAnalysis(_ctx context.Context, _input *ResuspendDNA_MultipleInput, _output *ResuspendDNA_MultipleOutput) {
}

func _ResuspendDNA_MultipleValidation(_ctx context.Context, _input *ResuspendDNA_MultipleInput, _output *ResuspendDNA_MultipleOutput) {
}
func _ResuspendDNA_MultipleRun(_ctx context.Context, input *ResuspendDNA_MultipleInput) *ResuspendDNA_MultipleOutput {
	output := &ResuspendDNA_MultipleOutput{}
	_ResuspendDNA_MultipleSetup(_ctx, input)
	_ResuspendDNA_MultipleSteps(_ctx, input, output)
	_ResuspendDNA_MultipleAnalysis(_ctx, input, output)
	_ResuspendDNA_MultipleValidation(_ctx, input, output)
	return output
}

func ResuspendDNA_MultipleRunSteps(_ctx context.Context, input *ResuspendDNA_MultipleInput) *ResuspendDNA_MultipleSOutput {
	soutput := &ResuspendDNA_MultipleSOutput{}
	output := _ResuspendDNA_MultipleRun(_ctx, input)
	if err := inject.AssignSome(output, &soutput.Data); err != nil {
		panic(err)
	}
	if err := inject.AssignSome(output, &soutput.Outputs); err != nil {
		panic(err)
	}
	return soutput
}

func ResuspendDNA_MultipleNew() interface{} {
	return &ResuspendDNA_MultipleElement{
		inject.CheckedRunner{
			RunFunc: func(_ctx context.Context, value inject.Value) (inject.Value, error) {
				input := &ResuspendDNA_MultipleInput{}
				if err := inject.Assign(value, input); err != nil {
					return nil, err
				}
				output := _ResuspendDNA_MultipleRun(_ctx, input)
				return inject.MakeValue(output), nil
			},
			In:  &ResuspendDNA_MultipleInput{},
			Out: &ResuspendDNA_MultipleOutput{},
		},
	}
}

var (
	_ = execute.MixInto
	_ = wunit.Make_units
)

type ResuspendDNA_MultipleElement struct {
	inject.CheckedRunner
}

type ResuspendDNA_MultipleInput struct {
	DNAPlate               *wtype.LHPlate
	Diluent                *wtype.LHComponent
	PartLocationsMap       map[string]string
	PartMassMap            map[string]wunit.Mass
	PartMolecularWeightMap map[string]float64
	Parts                  []string
	TargetConc             wunit.Concentration
}

type ResuspendDNA_MultipleOutput struct {
	ResuspendedDNAMap map[string]*wtype.LHComponent
}

type ResuspendDNA_MultipleSOutput struct {
	Data struct {
	}
	Outputs struct {
		ResuspendedDNAMap map[string]*wtype.LHComponent
	}
}

func init() {
	if err := addComponent(component.Component{Name: "ResuspendDNA_Multiple",
		Constructor: ResuspendDNA_MultipleNew,
		Desc: component.ComponentDesc{
			Desc: "Protocol for resuspending freeze dried DNA with a diluent\n",
			Path: "antha/component/an/Liquid_handling/ResuspendDNA/ResuspendDNAFromPlate.an",
			Params: []component.ParamDesc{
				{Name: "DNAPlate", Desc: "", Kind: "Inputs"},
				{Name: "Diluent", Desc: "", Kind: "Inputs"},
				{Name: "PartLocationsMap", Desc: "", Kind: "Parameters"},
				{Name: "PartMassMap", Desc: "", Kind: "Parameters"},
				{Name: "PartMolecularWeightMap", Desc: "", Kind: "Parameters"},
				{Name: "Parts", Desc: "", Kind: "Parameters"},
				{Name: "TargetConc", Desc: "", Kind: "Parameters"},
				{Name: "ResuspendedDNAMap", Desc: "", Kind: "Outputs"},
			},
		},
	}); err != nil {
		panic(err)
	}
}
