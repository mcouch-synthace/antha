package liquidhandling

import (
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
)

func getTestBlowout(ch *wtype.LHChannelParameter, multi int, tipType string) RobotInstruction {
	bi := NewBlowInstruction()
	bi.Multi = multi
	bi.TipType = tipType
	for i := 0; i < multi; i++ {
		bi.What = append(bi.What, "soup")
		bi.PltTo = append(bi.PltTo, "position_4")
		bi.WellTo = append(bi.WellTo, "A1")
		bi.Volume = append(bi.Volume, wunit.NewVolume(10.0, "ul"))
		bi.TPlateType = append(bi.TPlateType, "pcrplate_skirted_riser40")
		bi.TVolume = append(bi.TVolume, wunit.NewVolume(5.0, "ul"))
	}
	bi.Prms = ch
	bi.Head = ch.Head
	return bi
}

func TestBlowMixing(t *testing.T) {

	tenUl := wunit.NewVolume(10.0, "ul")

	tests := []*PolicyTest{
		{
			Name: "blow no tip change single channel",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"POST_MIX":        5,
						"POST_MIX_VOLUME": 10.0,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,DSP,MOV,MIX,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 5, //the Mix
					Values: map[string]interface{}{
						"CYCLES": []int{5},
						"VOLUME": []wunit.Volume{tenUl},
					},
				},
			},
		},
		{
			Name: "blow no tip change eight channel",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"POST_MIX":        5,
						"POST_MIX_VOLUME": 10.0,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 8, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,DSP,MOV,MIX,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 5, //the Mix
					Values: map[string]interface{}{
						"CYCLES": []int{5, 5, 5, 5, 5, 5, 5, 5},
						"VOLUME": []wunit.Volume{tenUl, tenUl, tenUl, tenUl, tenUl, tenUl, tenUl, tenUl},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.Run(t)
	}
}

func TestBlowNoMixing(t *testing.T) {

	tests := []*PolicyTest{
		{
			Name: "z-offset single channel",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPZOFFSET": 5.4,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,DSP,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 0,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
				{
					Instruction: 1,
					Values: map[string]interface{}{
						"DRIVE": "Z",
						"SPEED": defaultZSpeed,
					},
				},
				{
					Instruction: 2, //the move before the dispense
					Values: map[string]interface{}{
						"OFFSETZ": []float64{5.4},
					},
				},
			},
		},
		{
			Name: "entry speed single channel",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPENTRYSPEED": 50.0,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,SDS,MOV,DSP,SDS,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 0,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
				{
					Instruction: 1,
					Values: map[string]interface{}{
						"DRIVE": "Z",
						"SPEED": defaultZSpeed,
					},
				},
				{
					Instruction: 3,
					Values: map[string]interface{}{
						"DRIVE": "Z",
						"SPEED": 50.0,
					},
				},
				{
					Instruction: 4,
					Values: map[string]interface{}{
						"OFFSETZ": []float64{defaultZOffset},
					},
				},
			},
		},
	}

	for _, test := range tests {
		test.Run(t)
	}
}

func TestPipetSpeed(t *testing.T) {

	tests := []*PolicyTest{
		{
			Name: "blow pipette speed OK",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPSPEED": 1.5,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,SPS,DSP,SPS,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 0,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
				{
					Instruction: 1,
					Values: map[string]interface{}{
						"DRIVE": "Z",
						"SPEED": defaultZSpeed,
					},
				},
				{
					Instruction: 2, //the move before the dispense
					Values: map[string]interface{}{
						"OFFSETZ": []float64{defaultZOffset},
					},
				},
				{
					Instruction: 3,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   1.5,
					},
				},
				{
					Instruction: 5,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
			},
		},
		{
			Name: "blow pipette speed too large, made safe",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPSPEED":             LVMaxRate + 1.0,
						"OVERRIDEPIPETTESPEED": true,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,SPS,DSP,SPS,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 0,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
				{
					Instruction: 1,
					Values: map[string]interface{}{
						"DRIVE": "Z",
						"SPEED": defaultZSpeed,
					},
				},
				{
					Instruction: 2, //the move before the dispense
					Values: map[string]interface{}{
						"OFFSETZ": []float64{defaultZOffset},
					},
				},
				{
					Instruction: 3,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   LVMaxRate,
					},
				},
				{
					Instruction: 5,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
			},
		},
		{
			Name: "blow pipette speed too small, made safe",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPSPEED":             LVMinRate * 0.5,
						"OVERRIDEPIPETTESPEED": true,
					},
				},
			},
			Instruction:          getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Robot:                nil,
			ExpectedInstructions: "[SPS,SDS,MOV,SPS,DSP,SPS,MOV,BLO]",
			Assertions: []*InstructionAssertion{
				{
					Instruction: 0,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
				{
					Instruction: 1,
					Values: map[string]interface{}{
						"DRIVE": "Z",
						"SPEED": defaultZSpeed,
					},
				},
				{
					Instruction: 2, //the move before the dispense
					Values: map[string]interface{}{
						"OFFSETZ": []float64{defaultZOffset},
					},
				},
				{
					Instruction: 3,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   LVMinRate,
					},
				},
				{
					Instruction: 5,
					Values: map[string]interface{}{
						"HEAD":    1,
						"CHANNEL": -1,
						"SPEED":   defaultPipetteSpeed,
					},
				},
			},
		},
		{
			Name: "blow pipette speed too large, error",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPSPEED":             4.75,
						"OVERRIDEPIPETTESPEED": false,
					},
				},
			},
			Instruction: getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Error:       "error setting pipette dispense speed: value 4.75 out of range 0.0225000 - 3.7500000",
		},
		{
			Name: "blow pipette speed too small, error",
			Rules: []*Rule{
				{
					Name: "soup",
					Conditions: []Condition{
						&CategoryCondition{
							Attribute: "LIQUIDCLASS",
							Value:     "soup",
						},
					},
					Policy: map[string]interface{}{
						"DSPSPEED":             0.01,
						"OVERRIDEPIPETTESPEED": false,
					},
				},
			},
			Instruction: getTestBlowout(getLVConfig(), 1, "Gilson20"),
			Error:       "error setting pipette dispense speed: value 0.01 out of range 0.0225000 - 3.7500000",
		},
	}

	for _, test := range tests {
		test.Run(t)
	}
}
