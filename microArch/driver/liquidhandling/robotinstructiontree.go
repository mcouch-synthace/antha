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

// ITreeNode a node within a RobotInstruction tree.
// The children in each node are the output of instruction.Generate
type ITreeNode struct {
	instruction RobotInstruction
	children    []*ITreeNode
}

// NewITreeNode create a new tree node from the given instruction.
// if ri is nill, then the node will be the root of a new tree
func NewITreeNode(ri RobotInstruction) *ITreeNode {
	return &ITreeNode{
		instruction: ri,
	}
}

// MakeTreeRoot construct the root node of the RobotInstruction tree
func MakeTreeRoot(ctx context.Context, ch *wtype.IChain, policies *wtype.LHPolicyRuleSet, robot *LHProperties) (*ITreeNode, error) {

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

			// Need to apply the effect of the split immediately as it may be required for later calls to MakeTransfers
			// (n.b. the effect is to update the ID of the non-moving component to its new value)
			splitBlock := NewSplitBlockInstruction(ch.Values)
			splitBlock.Generate(ctx, policies, robot)
			//ret.AddChild(splitBlock)
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

// Generate (re)generate all instructions below this node
func (ri *ITreeNode) Generate(ctx context.Context, lhpr *wtype.LHPolicyRuleSet, lhpm *LHProperties) ([]TerminalRobotInstruction, *LHProperties, error) {
	// make a copy because generate affects its input
	props := lhpm.DupKeepIDs()
	ins, err := ri.generate(ctx, lhpr, props, nil)
	return ins, props, err
}

func (ri *ITreeNode) generate(ctx context.Context, lhpr *wtype.LHPolicyRuleSet, lhpm *LHProperties, ret []TerminalRobotInstruction) ([]TerminalRobotInstruction, error) {

	// call generate on our own instruction
	if ri.instruction != nil {
		ri.children = nil
		if children, err := ri.instruction.Generate(ctx, lhpr, lhpm); err != nil {
			return nil, err
		} else if len(children) == 0 {
			// if the instruction doesn't generate anything then we've reached a leaf
			if tri, ok := ri.instruction.(TerminalRobotInstruction); ok {
				ret = append(ret, tri)
				return ret, nil
			}
		} else {
			for _, child := range children {
				ri.AddChild(child)
			}
		}
	}

	// call generate for each child in turn
	for _, ins := range ri.children {
		if r, err := ins.generate(ctx, lhpr, lhpm, ret); err != nil {
			return r, err
		} else {
			ret = r
		}
	}

	// add the initialize and finalize instructions if we're the root
	if ri.instruction == nil {
		r := make([]TerminalRobotInstruction, len(ret)+2)
		r[0] = NewInitializeInstruction()
		copy(r[1:], ret)
		r[len(r)-1] = NewFinalizeInstruction()
		ret = r
	}

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
