// /anthalib/driver/liquidhandling/robotchildrenet.go: Part of the Antha language
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

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

// ITreeNode a node within a RobotInstruction tree
type ITreeNode struct {
	instruction RobotInstruction
	children    []*ITreeNode
}

// NewITreeNode create a new tree node from the given instruction.
// if ri is nill, then the node will be the root of a new tree
func NewITreeNode(ri RobotInstruction) *ITreeNode {
	var ret ITreeNode
	ret.children = make([]*ITreeNode, 0)
	ret.instruction = ri
	return &ret
}

// MakeTreeRoot construct the root node of the RobotInstruction tree
func MakeTreeRoot(ctx context.Context, ch *IChain, policies *wtype.LHPolicyRuleSet, robot *LHProperties) (*ITreeNode, error) {

	ret := NewITreeNode(nil)

	for {
		if ch == nil {
			break
		}

		if ch.Values[0].Type == wtype.LHIPRM {
			prm := NewMessageInstruction(ch.Values[0])
			ret.AddChild(prm)
		} else if hasSplit(ch.Values) {
			if !allSplits(ch.Values) {
				insTypes := func(inss []*wtype.LHInstruction) string {
					s := ""
					for _, ins := range inss {
						s += ins.InsType() + " "
					}

					return s
				}
				return nil, fmt.Errorf("Internal error: Failure in instruction sorting - got types %s in layer starting with split", insTypes(ch.Values))
			}

			splitBlock := NewSplitBlockInstruction(ch.Values)
			ret.AddChild(splitBlock)
		} else {
			if transfers, err := MakeTransfers(ctx, ch.Values, policies, robot); err != nil {
				return ret, err
			} else {
				for _, transfer := range transfers {
					ret.AddChild(transfer)
				}
			}
		}
		ch = ch.Child
	}

	return ret, nil
}

func allSplits(inss []*wtype.LHInstruction) bool {
	for _, ins := range inss {
		if ins.Type != wtype.LHISPL {
			return false
		}
	}
	return true
}

func hasSplit(inss []*wtype.LHInstruction) bool {
	for _, ins := range inss {
		if ins.Type == wtype.LHISPL {
			return true
		}
	}
	return false
}

// AddChild creates a new tree node containing ins and adds it to the children
// of this node
func (ri *ITreeNode) AddChild(ins RobotInstruction) {
	ris := NewITreeNode(ins)
	ri.children = append(ri.children, ris)
}

func (ri *ITreeNode) Generate(ctx context.Context, lhpr *wtype.LHPolicyRuleSet, lhpm *LHProperties) ([]RobotInstruction, error) {
	ret := make([]RobotInstruction, 0, 1)

	if ri.instruction != nil {
		arr, err := ri.instruction.Generate(ctx, lhpr, lhpm)

		if err != nil {
			return ret, err
		}

		// if the instruction doesn't generate anything then it is our return - bottom out here
		// assuming it's a Terminal
		if len(arr) == 0 {
			_, ok := ri.instruction.(TerminalRobotInstruction)
			if ok {
				ret = append(ret, ri.instruction)
				return ret, nil
			}
		} else {
			for _, ins := range arr {
				ri.AddChild(ins)
			}
		}
	}

	for _, ins := range ri.children {
		arr, err := ins.Generate(ctx, lhpr, lhpm)

		if err != nil {
			return arr, err
		}
		ret = append(ret, arr...)
	}

	if ri.instruction == nil {
		// add the initialize and finalize children
		ini := NewInitializeInstruction()
		newret := make([]RobotInstruction, 0, len(ret)+2)
		newret = append(newret, ini)
		newret = append(newret, ret...)
		fin := NewFinalizeInstruction()
		newret = append(newret, fin)
		ret = newret
	}

	// might need to do this instead of current version
	/*
		else if ri.instruction.Type == TFR {
			// update the vols
			prms.Evaporate()
		}
	*/

	return ret, nil
}

// String get a multi-line string representation of the tree below this node
func (ri *ITreeNode) String() string {
	return ri.toString(0)
}

func (ri *ITreeNode) toString(level int) string {

	name := ""

	if ri.instruction != nil {
		name = ri.instruction.Type().Name
	}
	s := ""
	for i := 0; i < level-1; i++ {
		s += fmt.Sprintf("\t")
	}
	s += fmt.Sprintf("%s\n", name)
	for i := 0; i < level; i++ {
		s += fmt.Sprintf("\t")
	}
	s += fmt.Sprintf("{\n")
	for _, ins := range ri.children {
		s += ins.toString(level + 1)
	}
	for i := 0; i < level; i++ {
		s += fmt.Sprintf("\t")
	}
	s += "}\n"
	return s
}
