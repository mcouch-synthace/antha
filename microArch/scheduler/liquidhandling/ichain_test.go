package liquidhandling

import (
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

func TestIChain(t *testing.T) {
	chain := NewIChain(nil)

	s := []string{"A", "B", "C", "D", "E", "F"}

	for _, k := range s {
		ins := wtype.NewLHMixInstruction()

		cmp := wtype.NewLHComponent()

		cmp.ID = k

		ins.AddInput(cmp)
		ins.AddOutput(wtype.NewLHComponent())
		chain.Add(ins)
	}
}
func TestIChain2(t *testing.T) {
	chain := NewIChain(nil)

	s := []string{"A", "B", "C", "D", "E", "F"}

	cmps := make([]*wtype.Liquid, 0, 1)

	for _, k := range s {

		cmp := wtype.NewLHComponent()

		cmp.ID = k
		cmps = append(cmps, cmp)
	}
	for i, cmp := range cmps {
		if i != 0 {
			cmp.AddParentComponent(cmps[i-1])
		}
		if i != len(s)-1 {
			cmp.AddDaughterComponent(cmps[i+1])
		}
	}

	for i, k := range cmps {
		ins := wtype.NewLHMixInstruction()
		ins.AddInput(k)
		if i != len(s)-1 {
			ins.AddOutput(cmps[i+1])
		} else {
			ins.AddOutput(wtype.NewLHComponent())
		}
		chain.Add(ins)
	}
}

func TestIChain3(t *testing.T) {
	chain := NewIChain(nil)

	s := []string{"A", "B", "C", "D", "E", "F"}

	cmps := make([]*wtype.Liquid, 0, 1)

	for _, k := range s {

		cmp := wtype.NewLHComponent()

		cmp.ID = k
		cmps = append(cmps, cmp)
	}
	for i, cmp := range cmps {
		if i != 0 {
			cmp.AddParentComponent(cmps[i-1])
		}
		if i != len(s)-1 {
			cmp.AddDaughterComponent(cmps[i+1])
		}
	}

	cmp := wtype.NewLHComponent()
	cmp.ID = "Z"
	cmp.AddParentComponent(cmps[2])
	cmps = append(cmps, cmp)

	cmp = wtype.NewLHComponent()
	cmp.ID = "Y"
	cmps = append(cmps, cmp)

	for i, k := range cmps {
		ins := wtype.NewLHMixInstruction()
		ins.AddInput(k)
		if i != len(s)-1 && cmp.ID != "Z" && cmp.ID != "Y" {
			ins.AddOutput(cmps[i+1])
		} else {
			ins.AddOutput(wtype.NewLHComponent())
		}
		chain.Add(ins)
	}
}
