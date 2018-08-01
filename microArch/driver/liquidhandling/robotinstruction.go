// anthalib/driver/liquidhandling/robotinstruction.go: Part of the Antha language
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
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/antha/anthalib/wutil/text"
)

type RobotInstruction interface {
	Type() *InstructionType
	GetParameter(name InstructionParameter) interface{}
	Generate(ctx context.Context, policy *wtype.LHPolicyRuleSet, prms *LHProperties) ([]RobotInstruction, error)
	MaybeMerge(next RobotInstruction) RobotInstruction
	Check(lhpr wtype.LHPolicyRule) bool
}

type TerminalRobotInstruction interface {
	RobotInstruction
	OutputTo(driver LiquidhandlingDriver) error
}

var (
	TFR = NewInstructionType("TFR", "Transfer")
	TFB = NewInstructionType("TFB", "TransferBlock")
	SCB = NewInstructionType("SCB", "SingleChannelTransferBlock")
	MCB = NewInstructionType("MCB", "MultiChannelTransferBlock")
	SCT = NewInstructionType("SCT", "SingleChannelTransfer")
	MCT = NewInstructionType("MCT", "MultiChannelTransfer")
	CCC = NewInstructionType("CCC", "ChangeChannelCharacteristics")
	LDT = NewInstructionType("LDT", "LoadTipsMove")
	UDT = NewInstructionType("UDT", "UnloadTipsMove")
	RST = NewInstructionType("RST", "Reset")
	CHA = NewInstructionType("CHA", "ChangeAdaptor")
	ASP = NewInstructionType("ASP", "Aspirate")
	DSP = NewInstructionType("DSP", "Dispense")
	BLO = NewInstructionType("BLO", "Blowout")
	PTZ = NewInstructionType("PTZ", "ResetPistons")
	MOV = NewInstructionType("MOV", "Move")
	MRW = NewInstructionType("MRW", "MoveRaw")
	LOD = NewInstructionType("LOD", "LoadTips")
	ULD = NewInstructionType("ULD", "UnloadTips")
	SUK = NewInstructionType("SUK", "Suck")
	BLW = NewInstructionType("BLW", "Blow")
	SPS = NewInstructionType("SPS", "SetPipetteSpeed")
	SDS = NewInstructionType("SDS", "SetDriveSpeed")
	INI = NewInstructionType("INI", "Initialize")
	FIN = NewInstructionType("FIN", "Finalize")
	WAI = NewInstructionType("WAI", "Wait")
	LON = NewInstructionType("LON", "LightsOn")
	LOF = NewInstructionType("LOF", "LightsOff")
	OPN = NewInstructionType("OPN", "Open")
	CLS = NewInstructionType("CLS", "Close")
	LAD = NewInstructionType("LAD", "LoadAdaptor")
	UAD = NewInstructionType("UAD", "UnloadAdaptor")
	MMX = NewInstructionType("MMX", "MoveMix")
	MIX = NewInstructionType("MIX", "Mix")
	MSG = NewInstructionType("MSG", "Message")
	MAS = NewInstructionType("MAS", "MoveAspirate")
	MDS = NewInstructionType("MDS", "MoveDispense")
	MVM = NewInstructionType("MVM", "MoveMix")
	MBL = NewInstructionType("MBL", "MoveBlowout")
	RAP = NewInstructionType("RAP", "RemoveAllPlates")
	APT = NewInstructionType("APT", "AddPlateTo")
	RPA = NewInstructionType("RPA", "RemovePlateAt")
	SPB = NewInstructionType("SPB", "SplitBlock")
)

type InstructionType struct {
	Name      string `json:"Type"`
	HumanName string `json:"-"`
}

// This exists so that when InstructionType is embedded within other
// instructions, we can satisfy the RobotInstruction interface with a
// minimum amount of boilerplate.
func (it *InstructionType) Type() *InstructionType {
	return it
}

func NewInstructionType(machine, human string) *InstructionType {
	return &InstructionType{
		Name:      machine,
		HumanName: human,
	}
}

type InstructionParameter string

func (name InstructionParameter) String() string {
	return string(name)
}

const (
	BLOWOUT         InstructionParameter = "BLOWOUT"
	CHANNEL                              = "CHANNEL"
	COMPONENT                            = "COMPONENT"
	CYCLES                               = "CYCLES"
	DRIVE                                = "DRIVE"
	FPLATEWX                             = "FPLATEWX"
	FPLATEWY                             = "FPLATEWY"
	FROMPLATETYPE                        = "FROMPLATETYPE"
	HEAD                                 = "HEAD"
	INSTRUCTIONTYPE                      = "INSTRUCTIONTYPE"
	LIQUIDCLASS                          = "LIQUIDCLASS" // LIQUIDCLASS refers to the Component Type, This is currently used to look up the corresponding LHPolicy from an LHPolicyRuleSet
	LLF                                  = "LLF"
	MESSAGE                              = "MESSAGE"
	MULTI                                = "MULTI"
	NAME                                 = "NAME"
	NEWADAPTOR                           = "NEWADAPTOR"
	NEWSTATE                             = "NEWSTATE"
	OFFSETX                              = "OFFSETX"
	OFFSETY                              = "OFFSETY"
	OFFSETZ                              = "OFFSETZ"
	OLDADAPTOR                           = "OLDADAPTOR"
	OLDSTATE                             = "OLDSTATE"
	OVERSTROKE                           = "OVERSTROKE"
	PARAMS                               = "PARAMS"
	PLATE                                = "PLATE"
	PLATETYPE                            = "PLATETYPE"
	PLATFORM                             = "PLATFORM"
	PLT                                  = "PLT"
	POS                                  = "POS"
	POSFROM                              = "POSFROM"
	POSITION                             = "POSITION"
	POSTO                                = "POSTO"
	REFERENCE                            = "REFERENCE"
	SPEED                                = "SPEED"
	TIME                                 = "TIME"
	TIPTYPE                              = "TIPTYPE"
	TOPLATETYPE                          = "TOPLATETYPE"
	TPLATEWX                             = "TPLATEWX"
	TPLATEWY                             = "TPLATEWY"
	VOLUME                               = "VOLUME"
	VOLUNT                               = "VOLUNT"
	WELL                                 = "WELL"
	WELLFROM                             = "WELLFROM"
	WELLFROMVOLUME                       = "WELLFROMVOLUME"
	WELLTO                               = "WELLTO"
	WELLTOVOLUME                         = "WELLTOVOLUME" // WELLTOVOLUME refers to the volume of liquid already present in the well location for which a sample is due to be transferred to.
	WELLVOLUME                           = "WELLVOLUME"
	WHAT                                 = "WHAT"
	WHICH                                = "WHICH" // WHICH returns the Component IDs, i.e. representing the specific instance of an LHComponent not currently implemented.
)

// we want set semantics, so it's much nicer in Go to use a map for
// this with the empty value, than it is to use a slice.
type InstructionParameters map[InstructionParameter]struct{}

func (a InstructionParameters) clone() InstructionParameters {
	b := make(InstructionParameters, len(a))
	for param, v := range a {
		b[param] = v
	}
	return b
}

func (a InstructionParameters) merge(b InstructionParameters) InstructionParameters {
	result := a.clone()
	for param, v := range b {
		result[param] = v
	}
	return result
}

// convenience construct
func NewInstructionParameters(params ...InstructionParameter) InstructionParameters {
	empty := struct{}{}
	result := make(InstructionParameters, len(params))
	for _, param := range params {
		result[param] = empty
	}
	return result
}

var RobotParameters = NewInstructionParameters(
	HEAD, CHANNEL, LIQUIDCLASS, POSTO, WELLFROM, WELLTO, REFERENCE, VOLUME, VOLUNT,
	FROMPLATETYPE, WELLFROMVOLUME, POSFROM, WELLTOVOLUME, TOPLATETYPE, MULTI, WHAT,
	LLF, PLT, OFFSETX, OFFSETY, OFFSETZ, TIME, SPEED, MESSAGE, COMPONENT)

// func HumanInstructionName(ins RobotInstruction) string {
// 	if ins == nil {
// 		return "no instruction"
// 	}
// 	if ret, ok := humanRobotInstructionNames[ins.InstructionType()]; ok {
// 		return ret
// 	}
// 	return "unknown"
// }

// option to feed into InsToString function
type printOption string

// Option to feed into InsToString function
// which prints key words of the instruction with coloured text.
// Designed for easier reading.
const colouredTerminalOutput printOption = "colouredTerminalOutput"

func ansiPrint(options ...printOption) bool {
	for _, option := range options {
		if option == colouredTerminalOutput {
			return true
		}
	}
	return false
}

/*
func printInstructionArray(inss []RobotInstruction) {
	for _, ins := range inss {
		fmt.Println(InsToString(ins))
	}
}
*/

func InsToString(ins RobotInstruction, ansiPrintOptions ...printOption) string {

	s := ins.Type().Name + " "

	if apt, ok := ins.(*AddPlateToInstruction); ok {
		s += fmt.Sprintf("NAME: %s POSITION: %s PLATE: %s", apt.Name, apt.Position, wtype.NameOf(apt.Plate))
		return s
	}

	var changeColour func(string) string

	if strings.TrimSpace(s) == "ASP" {
		changeColour = text.Green
	} else if strings.TrimSpace(s) == "DSP" {
		changeColour = text.Blue
	} else if strings.TrimSpace(s) == "MOV" {
		changeColour = text.Yellow
	} else {
		changeColour = text.White
	}
	if ansiPrint(ansiPrintOptions...) {
		s = changeColour(s)
	}
	for name := range RobotParameters {
		p := ins.GetParameter(name)

		if p == nil {
			continue
		}

		ss := ""

		switch p.(type) {
		case []wunit.Volume:
			if len(p.([]wunit.Volume)) == 0 {
				continue
			}
			ss = concatvolarray(p.([]wunit.Volume))

		case []string:
			if len(p.([]string)) == 0 {
				continue
			}
			ss = concatstringarray(p.([]string))
		case string:
			ss = p.(string)
		case []float64:
			if len(p.([]float64)) == 0 {
				continue
			}
			ss = concatfloatarray(p.([]float64))
		case float64:
			ss = fmt.Sprintf("%-6.4f", p.(float64))
		case []int:
			if len(p.([]int)) == 0 {
				continue
			}
			ss = concatintarray(p.([]int))
		case int:
			ss = fmt.Sprintf("%d", p.(int))
		case []bool:
			if len(p.([]bool)) == 0 {
				continue
			}
			ss = concatboolarray(p.([]bool))
		}
		if str := name.String(); ansiPrint(ansiPrintOptions...) {
			if name == WHAT {
				s += str + ": " + text.Yellow(ss) + " "
			} else if name == MULTI {
				s += text.Blue(str+": ") + ss + " "
			} else if name == OFFSETZ {
				s += str + ": " + changeColour(ss) + " "
			} else if name == TOPLATETYPE {
				s += str + ": " + text.Cyan(ss) + " "
			} else {
				s += str + ": " + ss + " "
			}
		} else {
			s += str + ": " + ss + " "
		}
	}

	return s
}

// StepSummary summarises the instruction for
// an Aspirate or Dispense instruction combined
// with the related Move instruction.
type StepSummary struct {
	Type         string // Asp or DSP
	LiquidType   string
	PlateType    string
	Multi        string
	OffsetZ      string
	WellToVolume string
	Volume       string
}

func mergeSummaries(a, b StepSummary, aspOrDsp string) (c StepSummary) {
	return StepSummary{
		Type:         aspOrDsp,
		LiquidType:   a.LiquidType + b.LiquidType,
		PlateType:    a.PlateType + b.PlateType,
		Multi:        a.Multi + b.Multi,
		OffsetZ:      a.OffsetZ + b.OffsetZ,
		WellToVolume: a.WellToVolume + b.WellToVolume,
		Volume:       a.Volume + b.Volume,
	}
}

type stepType string

// Aspirate designates a step is an aspirate step
const Aspirate stepType = "Aspirate"

// Dispense designates a step is a dispense step
const Dispense stepType = "Dispense"

// MakeAspOrDspSummary returns a summary of the key parameters involved in a Dispense or Aspirate step.
// It requires two consecutive instructions to do this, a Move instruction followed by a dispense of aspirate instruction.
// An error is returned if this is not the case.
func MakeAspOrDspSummary(moveInstruction, dspOrAspInstruction RobotInstruction) (StepSummary, error) {
	step1summary, err := summarise(moveInstruction)

	if err != nil {
		return StepSummary{}, err
	}

	step2summary, err := summarise(dspOrAspInstruction)

	if err != nil {
		return StepSummary{}, err
	}

	if moveInstruction.Type() != MOV {
		return StepSummary{}, fmt.Errorf("first instruction is not a move instruction: found %s", moveInstruction.Type().Name)
	}

	if dspOrAspInstruction.Type() == ASP {
		return mergeSummaries(step1summary, step2summary, string(Aspirate)), nil
	} else if dspOrAspInstruction.Type() == DSP {
		return mergeSummaries(step1summary, step2summary, string(Dispense)), nil
	} else {
		return StepSummary{}, fmt.Errorf("second instruction is not an aspirate or dispense: found %s", dspOrAspInstruction.Type().Name)
	}

}

func summarise(ins RobotInstruction) (StepSummary, error) {

	var summaryOfMoveOperation StepSummary

	for name := range RobotParameters {
		p := ins.GetParameter(name)

		if p == nil {
			continue
		}

		ss := ""

		switch p.(type) {
		case []wunit.Volume:
			if len(p.([]wunit.Volume)) == 0 {
				continue
			}
			ss = concatvolarray(p.([]wunit.Volume))

		case []string:
			if len(p.([]string)) == 0 {
				continue
			}
			ss = concatstringarray(p.([]string))
		case string:
			ss = p.(string)
		case []float64:
			if len(p.([]float64)) == 0 {
				continue
			}
			ss = concatfloatarray(p.([]float64))
		case float64:
			ss = fmt.Sprintf("%-6.4f", p.(float64))
		case []int:
			if len(p.([]int)) == 0 {
				continue
			}
			ss = concatintarray(p.([]int))
		case int:
			ss = fmt.Sprintf("%d", p.(int))
		case []bool:
			if len(p.([]bool)) == 0 {
				continue
			}
			ss = concatboolarray(p.([]bool))
		}
		if name == WHAT {
			summaryOfMoveOperation.LiquidType = ss
		} else if name == MULTI {
			summaryOfMoveOperation.Multi = ss
		} else if name == OFFSETZ {
			summaryOfMoveOperation.OffsetZ = ss
		} else if name == TOPLATETYPE {
			summaryOfMoveOperation.PlateType = ss
		} else if name == WELLTOVOLUME {
			summaryOfMoveOperation.WellToVolume = ss
		} else if name == VOLUME {
			summaryOfMoveOperation.Volume = ss
		}
	}

	return summaryOfMoveOperation, nil
}

func InsToString2(ins RobotInstruction) string {
	// IS THIS IT?!
	b, _ := json.Marshal(ins)
	return string(b)
}

func concatstringarray(a []string) string {
	r := ""

	for i, s := range a {
		r += s
		if i < len(a)-1 {
			r += ","
		}
	}

	return r
}

func concatvolarray(a []wunit.Volume) string {
	r := ""
	for i, s := range a {
		r += s.ToString()
		if i < len(a)-1 {
			r += ","
		}
	}

	return r

}

func concatfloatarray(a []float64) string {
	r := ""

	for i, s := range a {
		r += fmt.Sprintf("%-6.4f", s)
		if i < len(a)-1 {
			r += ","
		}
	}

	return r

}

func concatintarray(a []int) string {
	r := ""

	for i, s := range a {
		r += fmt.Sprintf("%d", s)
		if i < len(a)-1 {
			r += ","
		}
	}

	return r

}

func concatboolarray(a []bool) string {
	r := ""

	for i, s := range a {
		r += fmt.Sprintf("%t", s)
		if i < len(a)-1 {
			r += ","
		}
	}

	return r

}

type BaseRobotInstruction struct {
	Ins RobotInstruction `json:"-"`
}

func NewBaseRobotInstruction(ins RobotInstruction) BaseRobotInstruction {
	return BaseRobotInstruction{
		Ins: ins,
	}
}

func (bri BaseRobotInstruction) Check(rule wtype.LHPolicyRule) bool {
	for _, vcondition := range rule.Conditions {
		// todo - this cast to InstructionParameter is gross, but we're
		// going to have to tidy types with LHPolicy work later on.
		v := bri.Ins.GetParameter(InstructionParameter(vcondition.TestVariable))
		vrai := vcondition.Condition.Match(v)
		if !vrai {
			return false
		}
	}
	return true
}

// fall-through implementation to simplify instructions that have no parameters
func (bri BaseRobotInstruction) GetParameter(p InstructionParameter) interface{} {
	switch p {
	case INSTRUCTIONTYPE:
		return bri.Ins.Type()
	default:
		return nil
	}
}

func (bri BaseRobotInstruction) MaybeMerge(next RobotInstruction) RobotInstruction {
	return bri.Ins
}

/*
func printPolicyForDebug(ins RobotInstruction, rules []wtype.LHPolicyRule, pol wtype.LHPolicy) {
 	fmt.Println("*****")
 	fmt.Println("Policy for instruction ", InsToString(ins))
 	fmt.Println()
 	fmt.Println("Active Rules:")
 	fmt.Println("\t Default")
 	for _, r := range rules {
 		fmt.Println("\t", r.Name)
 	}
 	fmt.Println()
 	itemset := wtype.MakePolicyItems()
 	fmt.Println("Full output")
 	for _, s := range itemset.OrderedList() {
 		if pol[s] == nil {
 			continue
 		}
 		fmt.Println("\t", s, ": ", pol[s])
 	}
 	fmt.Println("_____")

}
*/

// ErrInvalidLiquidType is returned when no matching liquid policy is found.
type ErrInvalidLiquidType struct {
	PolicyNames      []string
	ValidPolicyNames []string
}

func (err ErrInvalidLiquidType) Error() string {
	return fmt.Sprintf("invalid LiquidType specified.\nValid Liquid Policies found: \n%s \n invalid LiquidType specified in instruction: %v \n ", strings.Join(err.ValidPolicyNames, " \n"), err.PolicyNames)
}

var (
	// ErrNoMatchingRules is returned when no matching LHPolicyRules are found when evaluating a rule set against a RobotInsturction.
	ErrNoMatchingRules = errors.New("no matching rules found")
	// ErrNoLiquidType is returned when no liquid policy is found.
	ErrNoLiquidType = errors.New("no LiquidType in instruction")
)

func matchesLiquidClass(rule wtype.LHPolicyRule) (match bool) {
	if len(rule.Conditions) > 0 {
		for i := range rule.Conditions {
			if rule.Conditions[i].TestVariable == "LIQUIDCLASS" {
				return true
			}
		}
	}
	return false
}

// GetDefaultPolicy currently returns the default policy
// this REALLY should not be necessary... ever
func GetDefaultPolicy(lhpr *wtype.LHPolicyRuleSet, ins RobotInstruction) (wtype.LHPolicy, error) {
	defaultPolicy := wtype.DupLHPolicy(lhpr.Policies["default"])
	return defaultPolicy, nil
}

// GetPolicyFor will return a matching LHPolicy for a RobotInstruction.
// If a common policy cannot be found for instances of the instruction then an error will be returned.
func GetPolicyFor(lhpr *wtype.LHPolicyRuleSet, ins RobotInstruction) (wtype.LHPolicy, error) {
	// find the set of matching rules
	rules := make([]wtype.LHPolicyRule, 0, len(lhpr.Rules))
	var lhpolicyFound bool

	for _, rule := range lhpr.Rules {

		if ins.Check(rule) {
			if matchesLiquidClass(rule) {
				lhpolicyFound = true
			}
			rules = append(rules, rule)
		}
	}

	// sort rules by priority
	sort.Sort(wtype.SortableRules(rules))

	// we might prefer to just merge this in

	ppl := wtype.DupLHPolicy(lhpr.Policies["default"])

	for _, rule := range rules {
		ppl.MergeWith(lhpr.Policies[rule.Name])
	}
	if len(rules) == 0 {
		return ppl, ErrNoMatchingRules
	}

	policy := ins.GetParameter(LIQUIDCLASS)
	var invalidPolicyNames []string
	if policies, ok := policy.([]string); ok {
		for _, policy := range policies {
			if _, found := lhpr.Policies[policy]; !found && policy != "" {
				invalidPolicyNames = append(invalidPolicyNames, policy)
			}
		}

	} else if policyString, ok := policy.(string); ok {
		if _, found := lhpr.Policies[policyString]; !found && policyString != "" {
			invalidPolicyNames = append(invalidPolicyNames, policyString)
		}
	}

	if len(invalidPolicyNames) > 0 {
		var validPolicies []string
		for key := range lhpr.Policies {
			validPolicies = append(validPolicies, key)
		}

		sort.Strings(validPolicies)
		return ppl, ErrInvalidLiquidType{PolicyNames: invalidPolicyNames, ValidPolicyNames: validPolicies}
	}

	if !lhpolicyFound {
		return ppl, ErrNoLiquidType
	}
	//printPolicyForDebug(ins, rules, ppl)
	return ppl, nil
}

type SetOfRobotInstructions struct {
	RobotInstructions []RobotInstruction
}

func (sori *SetOfRobotInstructions) UnmarshalJSON(b []byte) error {
	// first stage -- find the instructions

	soj := struct {
		RobotInstructions []json.RawMessage
	}{}

	if err := json.Unmarshal(b, &soj); err != nil {
		return err
	}

	// second stage -- unpack into an array
	sori.RobotInstructions = make([]RobotInstruction, len(soj.RobotInstructions))
	for i, raw := range soj.RobotInstructions {
		tId := struct {
			Type string
		}{}
		if err := json.Unmarshal(raw, &tId); err != nil {
			return err
		}

		if tId.Type == "" {
			return fmt.Errorf("Malformed instruction - no Type field field")
		}

		//motherofallswitches ugh

		var ins RobotInstruction

		switch tId.Type {
		case "RAP":
			ins = NewRemoveAllPlatesInstruction()
		case "APT":
			ins = NewAddPlateToInstruction("", "", nil)
		case "INI":
			ins = NewInitializeInstruction()
		case "ASP":
			ins = NewAspirateInstruction()
		case "DSP":
			ins = NewDispenseInstruction()
		case "MIX":
			ins = NewMixInstruction()
		case "SPS":
			ins = NewSetPipetteSpeedInstruction()
		case "SDS":
			ins = NewSetDriveSpeedInstruction()
		case "BLO":
			ins = NewBlowoutInstruction()
		case "LOD":
			ins = NewLoadTipsInstruction()
		case "MOV":
			ins = NewMoveInstruction()
		case "PTZ":
			ins = NewPTZInstruction()
		case "ULD":
			ins = NewUnloadTipsInstruction()
		case "MSG":
			ins = NewMessageInstruction(nil)
		case "WAI":
			ins = NewWaitInstruction()
		case "FIN":
			ins = NewFinalizeInstruction()
		default:
			return fmt.Errorf("Unknown instruction type: %s", tId.Type)
		}

		// finally unmarshal

		if err := json.Unmarshal(raw, ins); err != nil {
			return err
		}

		// add to array

		sori.RobotInstructions[i] = ins
	}

	return nil
}
