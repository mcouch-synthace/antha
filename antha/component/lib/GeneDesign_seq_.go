package lib

import (
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/enzymes"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/enzymes/lookup"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/export"
	"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/sequences"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	//"github.com/antha-lang/antha/antha/AnthaStandardLibrary/Packages/sequences/entrez"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/bvendor/golang.org/x/net/context"
	"github.com/antha-lang/antha/execute"
	"github.com/antha-lang/antha/inject"
	"strconv"
)

// input seq

// output parts with correct overhangs

func _GeneDesign_seqRequirements() {
}

func _GeneDesign_seqSetup(_ctx context.Context, _input *GeneDesign_seqInput) {
}

func _GeneDesign_seqSteps(_ctx context.Context, _input *GeneDesign_seqInput, _output *GeneDesign_seqOutput) {
	PartDNA := make([]wtype.DNASequence, 0)

	// Retrieve part seqs from entrez
	for i, part := range _input.Parts {
		DNA := wtype.MakeLinearDNASequence("part"+strconv.Itoa(i), part)
		PartDNA = append(PartDNA, DNA)
	}

	// look up vector sequence
	VectorSeq := wtype.MakePlasmidDNASequence("Vector", _input.Vector)

	// Look up the restriction enzyme
	EnzymeInf, _ := lookup.TypeIIsLookup(_input.RE)

	// Add overhangs
	_output.PartsWithOverhangs = enzymes.MakeScarfreeCustomTypeIIsassemblyParts(PartDNA, VectorSeq, EnzymeInf)

	// validation
	assembly := enzymes.Assemblyparameters{"NewConstruct", _input.RE, VectorSeq, _output.PartsWithOverhangs}
	_output.SimulationStatus, _, _, _, _ = enzymes.Assemblysimulator(assembly)

	// check if sequence meets requirements for synthesis
	_output.ValiadationStatus, _output.Validated = sequences.ValidateSynthesis(_output.PartsWithOverhangs, _input.Vector, _input.SynthesisProvider)

	// export sequence to fasta
	export.Makefastaserial2("NewConstruct", _output.PartsWithOverhangs)

}

func _GeneDesign_seqAnalysis(_ctx context.Context, _input *GeneDesign_seqInput, _output *GeneDesign_seqOutput) {

}

func _GeneDesign_seqValidation(_ctx context.Context, _input *GeneDesign_seqInput, _output *GeneDesign_seqOutput) {

}
func _GeneDesign_seqRun(_ctx context.Context, input *GeneDesign_seqInput) *GeneDesign_seqOutput {
	output := &GeneDesign_seqOutput{}
	_GeneDesign_seqSetup(_ctx, input)
	_GeneDesign_seqSteps(_ctx, input, output)
	_GeneDesign_seqAnalysis(_ctx, input, output)
	_GeneDesign_seqValidation(_ctx, input, output)
	return output
}

func GeneDesign_seqRunSteps(_ctx context.Context, input *GeneDesign_seqInput) *GeneDesign_seqSOutput {
	soutput := &GeneDesign_seqSOutput{}
	output := _GeneDesign_seqRun(_ctx, input)
	if err := inject.AssignSome(output, &soutput.Data); err != nil {
		panic(err)
	}
	if err := inject.AssignSome(output, &soutput.Outputs); err != nil {
		panic(err)
	}
	return soutput
}

func GeneDesign_seqNew() interface{} {
	return &GeneDesign_seqElement{
		inject.CheckedRunner{
			RunFunc: func(_ctx context.Context, value inject.Value) (inject.Value, error) {
				input := &GeneDesign_seqInput{}
				if err := inject.Assign(value, input); err != nil {
					return nil, err
				}
				output := _GeneDesign_seqRun(_ctx, input)
				return inject.MakeValue(output), nil
			},
			In:  &GeneDesign_seqInput{},
			Out: &GeneDesign_seqOutput{},
		},
	}
}

var (
	_ = execute.MixInto
	_ = wunit.Make_units
)

type GeneDesign_seqElement struct {
	inject.CheckedRunner
}

type GeneDesign_seqInput struct {
	Parts             []string
	RE                string
	SynthesisProvider string
	Vector            string
}

type GeneDesign_seqOutput struct {
	PartsWithOverhangs []wtype.DNASequence
	Sequence           string
	SimulationStatus   string
	ValiadationStatus  string
	Validated          bool
}

type GeneDesign_seqSOutput struct {
	Data struct {
		PartsWithOverhangs []wtype.DNASequence
		Sequence           string
		SimulationStatus   string
		ValiadationStatus  string
		Validated          bool
	}
	Outputs struct {
	}
}

func init() {
	addComponent(Component{Name: "GeneDesign_seq",
		Constructor: GeneDesign_seqNew,
		Desc: ComponentDesc{
			Desc: "",
			Path: "antha/component/an/Data/DNA/GeneDesign/GeneDesign_seq.an",
			Params: []ParamDesc{
				{Name: "Parts", Desc: "", Kind: "Parameters"},
				{Name: "RE", Desc: "", Kind: "Parameters"},
				{Name: "SynthesisProvider", Desc: "", Kind: "Parameters"},
				{Name: "Vector", Desc: "", Kind: "Parameters"},
				{Name: "PartsWithOverhangs", Desc: "output parts with correct overhangs\n", Kind: "Data"},
				{Name: "Sequence", Desc: "input seq\n", Kind: "Data"},
				{Name: "SimulationStatus", Desc: "", Kind: "Data"},
				{Name: "ValiadationStatus", Desc: "", Kind: "Data"},
				{Name: "Validated", Desc: "", Kind: "Data"},
			},
		},
	})
}
