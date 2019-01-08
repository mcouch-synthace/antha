package wtype

import (
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/wunit"
)

func TestAddandGetComponent(t *testing.T) {
	newTestComponent := func(
		name string,
		typ LiquidType,
		smax float64,
		conc wunit.Concentration,
		vol wunit.Volume,
		componentList ComponentList,
	) *Liquid {
		c := NewLHComponent()
		c.SetName(name)
		c.Type = typ
		c.Smax = smax
		c.SetConcentration(conc)
		c.AddSubComponents(componentList)

		return c
	}
	someComponents := ComponentList{Components: map[string]wunit.Concentration{
		"glycerol": wunit.NewConcentration(0.25, "g/l"),
		"IPTG":     wunit.NewConcentration(0.25, "mM/l"),
		"water":    wunit.NewConcentration(0.25, "v/v"),
		"LB":       wunit.NewConcentration(0.25, "X"),
	},
	}
	mediaMixture := newTestComponent("LB",
		LTWater,
		9999,
		wunit.NewConcentration(1, "X"),
		wunit.NewVolume(2000.0, "ul"),
		ComponentList{})

	err := mediaMixture.AddSubComponents(someComponents)

	if err != nil {
		t.Error(err)
	}

	tests := ComponentList{Components: map[string]wunit.Concentration{
		"Glycerol":  wunit.NewConcentration(0.25, "g/l"),
		"GLYCEROL ": wunit.NewConcentration(0.25, "g/l"),
	},
	}

	err = mediaMixture.AddSubComponents(tests)

	if err == nil {
		t.Errorf("expected error adding equalfold sub components to liquid but no error reported. New Sub components: %v", mediaMixture.SubComponents.AllComponents())
	}

	for test, _ := range tests.Components {
		if !mediaMixture.HasSubComponent(test) {
			t.Errorf(
				"Expected sub component %s to be found but only found these: %+v",
				test,
				mediaMixture.SubComponents.AllComponents(),
			)
		}
	}
}
