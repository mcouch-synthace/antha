// /anthalib/driver/liquidhandling/types.go: Part of the Antha language
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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antha-lang/antha/antha/anthalib/material"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/antha/anthalib/wutil"
	"github.com/antha-lang/antha/inventory"
	"github.com/antha-lang/antha/microArch/logger"
	"github.com/antha-lang/antha/microArch/sampletracker"
)

// describes a liquid handler, its capabilities and current state
// probably needs splitting up to separate out the state information
// from the properties information
type LHProperties struct {
	ID                   string
	Nposns               int
	Positions            map[string]*wtype.LHPosition
	PlateLookup          map[string]interface{}
	PosLookup            map[string]string
	PlateIDLookup        map[string]string
	Plates               map[string]*wtype.LHPlate
	Tipboxes             map[string]*wtype.LHTipbox
	Tipwastes            map[string]*wtype.LHTipwaste
	Wastes               map[string]*wtype.LHPlate
	Washes               map[string]*wtype.LHPlate
	Devices              map[string]string
	Model                string
	Mnfr                 string
	LHType               string
	TipType              string
	Heads                []*wtype.LHHead
	HeadsLoaded          []*wtype.LHHead
	Adaptors             []*wtype.LHAdaptor
	Tips                 []*wtype.LHTip
	Tip_preferences      []string
	Input_preferences    []string
	Output_preferences   []string
	Tipwaste_preferences []string
	Waste_preferences    []string
	Wash_preferences     []string
	Driver               LiquidhandlingDriver `gotopb:"-"`
	CurrConf             *wtype.LHChannelParameter
	Cnfvol               []*wtype.LHChannelParameter
	Layout               map[string]wtype.Coordinates
	MaterialType         material.MaterialType
}

// utility print function

func (p LHProperties) OutputLayout() {
	fmt.Println("Layout for liquid handler ", p.ID, " type ", p.Mnfr, " ", p.Model)
	n := p.OrderedPositionNames()

	for _, pos := range n {
		plateID, ok := p.PosLookup[pos]

		fmt.Print("\tPosition ", pos, " ")

		if !ok {
			fmt.Println(" Empty")
		} else {
			lw := p.PlateLookup[plateID]

			switch lw.(type) {
			case *wtype.LHPlate:
				plt := lw.(*wtype.LHPlate)
				fmt.Println("Plate ", plt.PlateName, " type ", plt.Mnfr, " ", plt.Type, " Contents:")
				plt.OutputLayout()
			case *wtype.LHTipbox:
				tb := lw.(*wtype.LHTipbox)
				fmt.Println("Tip box ", tb.Mnfr, " ", tb.Type, " ", tb.Boxname, " ", tb.N_clean_tips())
			case *wtype.LHTipwaste:
				tw := lw.(*wtype.LHTipwaste)
				fmt.Println("Tip Waste ", tw.Mnfr, " ", tw.Type, " capacity ", tw.SpaceLeft())
			default:
				fmt.Println("Labware :", lw)
			}
		}
	}
}

func (p LHProperties) CanPrompt() bool {
	// presently true for all varieties of liquid handlers
	// TODO (when no longer the case): revise!
	return true
}

func (p LHProperties) OrderedPositionNames() []string {
	// canonical ordering

	s := make([]string, 0, len(p.Positions))
	for n, _ := range p.Positions {
		s = append(s, n)
	}

	sort.Strings(s)

	return s
}

// validator for LHProperties structure

func ValidateLHProperties(props *LHProperties) (bool, string) {
	bo := true
	so := "OK"

	be := false
	se := "LHProperties Error: No position"

	if props.Positions == nil || len(props.Positions) == 0 {
		return be, se + "s"
	}

	for k, p := range props.Positions {
		if p == nil || p.ID == "" {
			return be, se + " " + k + " not set"
		}
	}

	se = "LHProperties error: No position lookup"

	if props.PosLookup == nil || len(props.PosLookup) == 0 {
		return be, se
	}

	se = "LHProperties Error: No tip preference information"

	if props.Tip_preferences == nil || len(props.Tip_preferences) == 0 {
		return be, se
	}

	se = "LHProperties Error: No input preference information"

	if props.Input_preferences == nil || len(props.Input_preferences) == 0 {
		return be, se
	}

	se = "LHProperties Error: No output preference information"

	if props.Output_preferences == nil || len(props.Output_preferences) == 0 {
		return be, se
	}

	se = "LHProperties Error: No waste preference information"

	if props.Waste_preferences == nil || len(props.Waste_preferences) == 0 {
		return be, se
	}

	se = "LHProperties Error: No tipwaste preference information"

	if props.Tipwaste_preferences == nil || len(props.Tipwaste_preferences) == 0 {
		return be, se
	}
	se = "LHProperties Error: No wash preference information"

	if props.Wash_preferences == nil || len(props.Wash_preferences) == 0 {
		return be, se
	}
	se = "LHProperties Error: No Plate ID lookup information"

	if props.PlateIDLookup == nil {
		return be, se
	}

	se = "LHProperties Error: No tip defined"

	if props.Tips == nil {
		return be, se
	}

	se = "LHProperties Error: No headsloaded array"

	if props.HeadsLoaded == nil {
		return be, se
	}

	return bo, so
}

// copy constructor

func (lhp *LHProperties) Dup() *LHProperties {
	lo := make(map[string]wtype.Coordinates, len(lhp.Layout))
	for k, v := range lhp.Layout {
		lo[k] = v
	}
	r := NewLHProperties(lhp.Nposns, lhp.Model, lhp.Mnfr, lhp.LHType, lhp.TipType, lo)

	for _, a := range lhp.Adaptors {
		r.Adaptors = append(r.Adaptors, a.Dup())
	}

	for _, h := range lhp.Heads {
		r.Heads = append(r.Heads, h.Dup())
	}

	for _, hl := range lhp.HeadsLoaded {
		r.HeadsLoaded = append(r.HeadsLoaded, hl.Dup())
	}

	/*
		for name, pl := range lhp.PosLookup {
			r.PosLookup[name] = pl
		}

		for name, pl := range lhp.PlateIDLookup {
			r.PlateIDLookup[name] = pl
		}
	*/

	// plate lookup can contain anything

	for name, pt := range lhp.PlateLookup {
		var pt2 interface{}
		var newid string
		var pos string
		switch pt.(type) {
		case *wtype.LHTipwaste:
			tmp := pt.(*wtype.LHTipwaste).Dup()
			pt2 = tmp
			newid = tmp.ID
			pos = lhp.PlateIDLookup[name]
			r.Tipwastes[pos] = tmp
		case *wtype.LHPlate:
			tmp := pt.(*wtype.LHPlate).Dup()
			pt2 = tmp
			newid = tmp.ID
			pos = lhp.PlateIDLookup[name]
			_, waste := lhp.Wastes[pos]
			_, wash := lhp.Washes[pos]

			if waste {
				r.Wastes[pos] = tmp
			} else if wash {
				r.Washes[pos] = tmp
			} else {
				r.Plates[pos] = tmp
			}

		case *wtype.LHTipbox:
			tmp := pt.(*wtype.LHTipbox).Dup()
			pt2 = tmp
			newid = tmp.ID
			pos = lhp.PlateIDLookup[name]
			r.Tipboxes[pos] = tmp
		}
		r.PlateLookup[newid] = pt2
		r.PlateIDLookup[newid] = pos
		r.PosLookup[pos] = newid
	}

	for name, dev := range lhp.Devices {
		r.Devices[name] = dev
	}

	for name, head := range lhp.Heads {
		r.Heads[name] = head.Dup()
	}

	for i, hl := range lhp.HeadsLoaded {
		r.HeadsLoaded[i] = hl.Dup()
	}

	for i, ad := range lhp.Adaptors {
		r.Adaptors[i] = ad.Dup()
	}

	for _, tip := range lhp.Tips {
		r.Tips = append(r.Tips, tip.Dup())
	}

	for _, pref := range lhp.Tip_preferences {
		r.Tip_preferences = append(r.Tip_preferences, pref)
	}
	for _, pref := range lhp.Input_preferences {
		r.Input_preferences = append(r.Input_preferences, pref)
	}
	for _, pref := range lhp.Output_preferences {
		r.Output_preferences = append(r.Output_preferences, pref)
	}
	for _, pref := range lhp.Waste_preferences {
		r.Waste_preferences = append(r.Waste_preferences, pref)
	}
	for _, pref := range lhp.Tipwaste_preferences {
		r.Tipwaste_preferences = append(r.Tipwaste_preferences, pref)
	}
	for _, pref := range lhp.Wash_preferences {
		r.Wash_preferences = append(r.Wash_preferences, pref)
	}

	if lhp.CurrConf != nil {
		r.CurrConf = lhp.CurrConf.Dup()
	}

	for i, v := range lhp.Cnfvol {
		r.Cnfvol[i] = v
	}

	for i, v := range lhp.Layout {
		r.Layout[i] = v
	}

	r.MaterialType = lhp.MaterialType

	// copy the driver

	r.Driver = lhp.Driver

	return r
}

// constructor for the above
func NewLHProperties(num_positions int, model, manufacturer, lhtype, tiptype string, layout map[string]wtype.Coordinates) *LHProperties {
	var lhp LHProperties

	lhp.ID = wtype.GetUUID()

	lhp.Nposns = num_positions

	lhp.Model = model
	lhp.Mnfr = manufacturer
	lhp.LHType = lhtype
	lhp.TipType = tiptype

	lhp.Adaptors = make([]*wtype.LHAdaptor, 0, 2)
	lhp.Heads = make([]*wtype.LHHead, 0, 2)
	lhp.HeadsLoaded = make([]*wtype.LHHead, 0, 2)

	positions := make(map[string]*wtype.LHPosition, num_positions)

	for i := 0; i < num_positions; i++ {
		// not overriding these defaults seems like a
		// bad idea --- TODO: Fix, e.g., MAXH here
		posname := fmt.Sprintf("position_%d", i+1)
		positions[posname] = wtype.NewLHPosition(i+1, "position_"+strconv.Itoa(i+1), 80.0)
	}

	lhp.Positions = positions
	lhp.PosLookup = make(map[string]string, lhp.Nposns)
	lhp.PlateLookup = make(map[string]interface{}, lhp.Nposns)
	lhp.PlateIDLookup = make(map[string]string, lhp.Nposns)
	lhp.Plates = make(map[string]*wtype.LHPlate, lhp.Nposns)
	lhp.Tipboxes = make(map[string]*wtype.LHTipbox, lhp.Nposns)
	lhp.Tipwastes = make(map[string]*wtype.LHTipwaste, lhp.Nposns)
	lhp.Wastes = make(map[string]*wtype.LHPlate, lhp.Nposns)
	lhp.Washes = make(map[string]*wtype.LHPlate, lhp.Nposns)
	lhp.Devices = make(map[string]string, lhp.Nposns)
	lhp.Heads = make([]*wtype.LHHead, 0, 2)
	lhp.Tips = make([]*wtype.LHTip, 0, 3)

	lhp.Layout = layout

	// lhp.Curcnf, lhp.Cmnvol etc. intentionally left blank

	lhp.MaterialType = material.DEVICE

	return &lhp
}

func (lhp *LHProperties) TipsLeftOfType(tiptype string) int {
	n := 0

	for _, pref := range lhp.Tip_preferences {
		tb := lhp.Tipboxes[pref]
		if tb != nil {
			n += tb.N_clean_tips()
		}
	}

	return n
}

func (lhp *LHProperties) AddTipBox(tipbox *wtype.LHTipbox) error {
	for _, pref := range lhp.Tip_preferences {
		if lhp.PosLookup[pref] != "" {
			continue
		}

		lhp.AddTipBoxTo(pref, tipbox)
		return nil
	}

	return wtype.LHError(wtype.LH_ERR_NO_DECK_SPACE, "Trying to add tip box")
}
func (lhp *LHProperties) AddTipBoxTo(pos string, tipbox *wtype.LHTipbox) bool {
	/*
		fmt.Println("Adding tip box of type, ", tipbox.Type, " To position ", pos)
		if lhp.PosLookup[pos] != "" {
			logger.Fatal("CAN'T ADD TIPBOX TO FULL POSITION")
			panic("CAN'T ADD TIPBOX TO FULL POSITION")
		}
	*/

	if lhp.PosLookup[pos] != "" {
		logger.Debug(fmt.Sprintf("Tried to add tipbox to full position: %s", pos))
		return false
	}
	lhp.Tipboxes[pos] = tipbox
	lhp.PlateLookup[tipbox.ID] = tipbox
	lhp.PosLookup[pos] = tipbox.ID
	lhp.PlateIDLookup[tipbox.ID] = pos

	return true
}

func (lhp *LHProperties) RemoveTipBoxes() {
	for pos, tbx := range lhp.Tipboxes {
		lhp.PlateLookup[tbx.ID] = nil
		lhp.PosLookup[pos] = ""
		lhp.PlateIDLookup[tbx.ID] = ""
	}

	lhp.Tipboxes = make(map[string]*wtype.LHTipbox)
}

func (lhp *LHProperties) TipWastesMounted() int {
	r := 0
	for _, pref := range lhp.Tipwaste_preferences {
		if lhp.PosLookup[pref] != "" {
			_, ok := lhp.Tipwastes[lhp.PosLookup[pref]]

			if !ok {
				logger.Debug(fmt.Sprintf("Position %s claims to have a tipbox but is empty", pref))
				continue
			}

			r += 1
		}
	}

	return r

}

func (lhp *LHProperties) TipSpacesLeft() int {
	r := 0
	for _, pref := range lhp.Tipwaste_preferences {
		if lhp.PosLookup[pref] != "" {
			bx, ok := lhp.Tipwastes[lhp.PosLookup[pref]]

			if !ok {
				logger.Debug(fmt.Sprintf("Position %s claims to have a tipbox but is empty", pref))
				continue
			}

			r += bx.SpaceLeft()
		}
	}

	return r
}

func (lhp *LHProperties) AddTipWaste(tipwaste *wtype.LHTipwaste) error {
	for _, pref := range lhp.Tipwaste_preferences {
		if lhp.PosLookup[pref] != "" {
			continue
		}

		err := lhp.AddTipWasteTo(pref, tipwaste)
		return err
	}

	return wtype.LHError(wtype.LH_ERR_NO_DECK_SPACE, "Trying to add tip waste")
}

func (lhp *LHProperties) AddTipWasteTo(pos string, tipwaste *wtype.LHTipwaste) error {
	if lhp.PosLookup[pos] != "" {
		return wtype.LHError(wtype.LH_ERR_NO_DECK_SPACE, fmt.Sprintf("Trying to add tip waste to full position %s", pos))
	}
	lhp.Tipwastes[pos] = tipwaste
	lhp.PlateLookup[tipwaste.ID] = tipwaste
	lhp.PosLookup[pos] = tipwaste.ID
	lhp.PlateIDLookup[tipwaste.ID] = pos
	return nil
}

func (lhp *LHProperties) AddPlate(pos string, plate *wtype.LHPlate) error {
	if lhp.PosLookup[pos] != "" {
		return wtype.LHError(wtype.LH_ERR_NO_DECK_SPACE, fmt.Sprintf("Trying to add plate to full position %s", pos))
	}
	lhp.Plates[pos] = plate
	lhp.PlateLookup[plate.ID] = plate
	lhp.PosLookup[pos] = plate.ID
	lhp.PlateIDLookup[plate.ID] = pos
	return nil
}

// reverse the above

func (lhp *LHProperties) RemovePlateWithID(id string) {
	pos := lhp.PlateIDLookup[id]
	delete(lhp.PosLookup, pos)
	delete(lhp.PlateIDLookup, id)
	delete(lhp.PlateLookup, id)
	delete(lhp.Plates, pos)
}

func (lhp *LHProperties) RemovePlateAtPosition(pos string) {
	id := lhp.PosLookup[pos]
	delete(lhp.PosLookup, pos)
	delete(lhp.PlateIDLookup, id)
	delete(lhp.PlateLookup, id)
	delete(lhp.Plates, pos)
}

func (lhp *LHProperties) addWaste(waste *wtype.LHPlate) bool {
	for _, pref := range lhp.Waste_preferences {
		if lhp.PosLookup[pref] != "" {

			//fmt.Println(pref, " ", lhp.PlateLookup[lhp.PosLookup[pref]])

			continue
		}

		lhp.AddWasteTo(pref, waste)
		return true
	}

	logger.Debug("NO WASTE SPACES LEFT")
	return false
}

func (lhp *LHProperties) AddWasteTo(pos string, waste *wtype.LHPlate) bool {
	if lhp.PosLookup[pos] != "" {
		logger.Debug("CAN'T ADD WASTE TO FULL POSITION")
		return false
	}
	lhp.Wastes[pos] = waste
	lhp.PlateLookup[waste.ID] = waste
	lhp.PosLookup[pos] = waste.ID
	lhp.PlateIDLookup[waste.ID] = pos
	return true
}

func (lhp *LHProperties) AddWash(wash *wtype.LHPlate) bool {
	for _, pref := range lhp.Wash_preferences {
		if lhp.PosLookup[pref] != "" {
			//fmt.Println(pref, " ", lhp.PlateLookup[lhp.PosLookup[pref]])
			continue
		}

		lhp.AddWashTo(pref, wash)
		return true
	}

	logger.Debug("NO WASH SPACES LEFT")
	return false
}

func (lhp *LHProperties) AddWashTo(pos string, wash *wtype.LHPlate) bool {
	if lhp.PosLookup[pos] != "" {

		logger.Debug("CAN'T ADD WASH TO FULL POSITION")
		return false
	}
	lhp.Washes[pos] = wash
	lhp.PlateLookup[wash.ID] = wash
	lhp.PosLookup[pos] = wash.ID
	lhp.PlateIDLookup[wash.ID] = pos
	return true
}

// returns go: [transfer1][c1c2c3c4...], [transfer2][c1c2c3c4...]
func (lhp *LHProperties) GetComponents(cmps []*wtype.LHComponent, carryvol wunit.Volume, ori, multi int, independent, legacyVolume bool) (plateIDs, wellCoords [][]string, vols [][]wunit.Volume, err error) {
	plateIDs = make([][]string, len(cmps))
	wellCoords = make([][]string, len(cmps))
	vols = make([][]wunit.Volume, len(cmps))

	// might just do this this way
	if multi > 1 {
		//for _, ipref := range lhp.Input_preferences {
		for _, ipref := range lhp.OrderedMergedPlatePrefs() {
			p, ok := lhp.Plates[ipref]

			if ok {
				// find components multi can return anywhere between 1x Multi and Multi x 1
				// transfers as sets
				plateIDs, wellCoords, vols, err = p.FindComponentsMulti(cmps, ori, multi, independent)
				if err != nil {
					continue
				}

				return
			}
		}
	}
	// if we really can't get anywhere, try this badger
	return lhp.GetComponentsSingle(cmps, carryvol, legacyVolume)
}

func GetLocTox(cmp *wtype.LHComponent) ([]string, error) {
	// try the cmp's own loc

	if cmp.Loc != "" {
		return strings.Split(cmp.Loc, ":"), nil
	} else {
		// try the ID of the thing

		tx, err := getSTLocTox(cmp.ID)

		if err == nil {
			return tx, err
		}

		// now try its parent

		tx, err = getSTLocTox(cmp.ParentID)

		if err == nil {
			return tx, err
		}
	}

	return []string{}, fmt.Errorf("No location found")
}

func getSTLocTox(ID string) ([]string, error) {
	st := sampletracker.GetSampleTracker()
	loc, ok := st.GetLocationOf(ID)

	if !ok {
		return []string{}, fmt.Errorf("No location found")
	}

	tx := strings.Split(loc, ":")

	if len(tx) == 2 {
		return tx, nil
	} else {
		return []string{}, fmt.Errorf("No location found")
	}
}

func (lhp *LHProperties) InputSearchPreferences() []string {
	// Definition 1: merge input plate preferences and output plate preferences
	return lhp.mergeInputOutputPreferences()
}

func (lhp *LHProperties) mergeInputOutputPreferences() []string {
	seen := make(map[string]bool, len(lhp.Positions))
	out := make([]string, 0, len(lhp.Positions))

	mergeToSet := func(in, out []string, seen map[string]bool) ([]string, map[string]bool) {
		// adds anything from in to out that isn't in seen, respecting input order
		for _, mem := range in {
			if seen[mem] {
				continue
			}
			seen[mem] = true
			out = append(out, mem)

		}

		return out, seen
	}

	out, seen = mergeToSet(lhp.Input_preferences, out, seen)
	out, seen = mergeToSet(lhp.Output_preferences, out, seen)

	return out
}

// logic of getting components:
// we look for things with the same ID
// the ID may or may not refer to an instance which is previously made
// but by this point we must have concrete locations for everything

func (lhp *LHProperties) GetComponentsSingle(cmps []*wtype.LHComponent, carryvol wunit.Volume, legacyVolume bool) ([][]string, [][]string, [][]wunit.Volume, error) {
	plateIDs := make([][]string, len(cmps))
	wellCoords := make([][]string, len(cmps))
	vols := make([][]wunit.Volume, len(cmps))

	// locally keep volumes straight

	localplates := make(map[string]*wtype.LHPlate, len(lhp.Plates))

	for k, v := range lhp.Plates {
		localplates[k] = v.DupKeepIDs()
	}

	// cmps are requests for components
	for i, cmp := range cmps {
		plateIDs[i] = make([]string, 0, 1)
		wellCoords[i] = make([]string, 0, 1)
		vols[i] = make([]wunit.Volume, 0, 1)
		foundIt := false

		cmpdup := cmp.Dup()

		// searches all plates: input and output
		for _, ipref := range lhp.InputSearchPreferences() {
			// check if the plate at position ipref has the
			// component we seek

			p, ok := localplates[ipref]
			if ok && !p.Empty() {
				// whaddya got?
				// nb this won't work if we need to split a volume across several plates
				wcarr, varr, ok := p.BetterGetComponent(cmpdup, lhp.MinPossibleVolume(), legacyVolume)

				if ok {
					foundIt = true
					for ix, _ := range wcarr {
						wc := wcarr[ix].FormatA1()
						vl := varr[ix].Dup()
						plateIDs[i] = append(plateIDs[i], p.ID)
						wellCoords[i] = append(wellCoords[i], wc)
						vols[i] = append(vols[i], vl)
						vl = vl.Dup()
						vl.Add(carryvol)
						//lhp.RemoveComponent(p.ID, wc, vl)
						p.RemoveComponent(wc, vl)
					}
					break
				}
			}
		}

		if !foundIt {
			err := wtype.LHError(wtype.LH_ERR_DIRE, fmt.Sprint("749: NO SOURCE FOR ", cmp.CName, " at volume ", cmp.Volume().ToString()))
			return plateIDs, wellCoords, vols, err
		}
	}

	return plateIDs, wellCoords, vols, nil
}

// GetComponents takes requests for components at particular volumes
// + a measure of carry volume
// returns lists of plate IDs + wells from which to get components or error

func (lhp *LHProperties) legacyGetComponentsSingle(cmps []*wtype.LHComponent, carryvol wunit.Volume, legacyVolume bool) ([][]string, [][]string, [][]wunit.Volume, error) {
	plateIDs := make([][]string, len(cmps))
	wellCoords := make([][]string, len(cmps))
	vols := make([][]wunit.Volume, len(cmps))

	// groan

	localplates := make(map[string]*wtype.LHPlate, len(lhp.Plates))

	for k, v := range lhp.Plates {
		localplates[k] = v.DupKeepIDs()
	}

	// need to disentangle some stuff here

	for i, v := range cmps {
		plateIDs[i] = make([]string, 0, 1)
		wellCoords[i] = make([]string, 0, 1)
		vols[i] = make([]wunit.Volume, 0, 1)
		foundIt := false

		vdup := v.Dup()
		if v.IsInstance() {
			/*
					tx := strings.Split(v.Loc, ":")

					if len(tx) < 2 || len(v.Loc) == 0 {
						st := sampletracker.GetSampleTracker()
						loc, _ := st.GetLocationOf(v.ID)
						tx = strings.Split(loc, ":")
					}


				if len(tx) < 2 {
					panic(fmt.Sprintf("NO LOCATION FOUND FOR COMPONENT: %s %s %s", v.ID, v.CName, v.Volume().ToString()))
				}
			*/

			tx, err := GetLocTox(v)
			if err != nil || len(tx) != 2 {
				panic(fmt.Sprintf("NO LOCATION FOUND FOR COMPONENT: %s %s %s", v.ID, v.CName, v.Volume().ToString()))
			}

			plateIDs[i] = append(plateIDs[i], tx[0])
			wellCoords[i] = append(wellCoords[i], tx[1])
			vols[i] = append(vols[i], v.Volume().Dup())

			//vol := v.Volume().Dup()
			//vol.Add(carryvol)
			/// XXX -- adding carry volumes is all very well but
			// assumes we have made more of this component than we really need!
			// -- this may just need to be removed pending a better fix
			//lhp.RemoveComponent(tx[0], tx[1], vol)

			foundIt = true

		} else {
			for _, ipref := range lhp.Input_preferences {
				// check if the plate at position ipref has the
				// component we seek

				p, ok := localplates[ipref]
				if ok {
					// whaddya got?
					// nb this won't work if we need to split a volume across several plates
					wcarr, varr, ok := p.BetterGetComponent(vdup, lhp.MinPossibleVolume(), legacyVolume)

					if ok {
						foundIt = true
						for ix, _ := range wcarr {
							wc := wcarr[ix].FormatA1()
							vl := varr[ix].Dup()
							plateIDs[i] = append(plateIDs[i], p.ID)
							wellCoords[i] = append(wellCoords[i], wc)
							vols[i] = append(vols[i], vl)
							vl = vl.Dup()
							vl.Add(carryvol)
							//lhp.RemoveComponent(p.ID, wc, vl)
							p.RemoveComponent(wc, vl)
						}
						break
					}
				}
			}

			if !foundIt {
				err := wtype.LHError(wtype.LH_ERR_DIRE, fmt.Sprint("NO SOURCE FOR ", v.CName, " at volume ", v.Volume().ToString()))
				return plateIDs, wellCoords, vols, err
			}

		}
	}

	return plateIDs, wellCoords, vols, nil
}

func (lhp *LHProperties) GetCleanTips(ctx context.Context, tiptype string, channel *wtype.LHChannelParameter, mirror bool, multi int, usetiptracking bool) (wells, positions, boxtypes []string, err error) {
	positions = make([]string, multi)
	boxtypes = make([]string, multi)

	// hack -- count tips left
	n_tips_left := 0
	for _, pos := range lhp.Tip_preferences {
		bx, ok := lhp.Tipboxes[pos]

		if !ok || bx.Tiptype.Type != tiptype {
			continue
		}

		n_tips_left += bx.N_clean_tips()
	}

	//	logger.Debug(fmt.Sprintf("There are %d clean tips of type %s left", n_tips_left, tiptype))

	foundit := false

	for _, pos := range lhp.Tip_preferences {
		//	for i := len(lhp.Tip_preferences) - 1; i >= 0; i-- {
		//		pos := lhp.Tip_preferences[i]
		bx, ok := lhp.Tipboxes[pos]
		if !ok || bx.Tiptype.Type != tiptype {
			continue
		}
		wells = bx.GetTips(mirror, multi, channel.Orientation)
		if wells != nil && len(wells) == multi {
			foundit = true
			for i := 0; i < multi; i++ {
				positions[i] = pos
				boxtypes[i] = bx.Boxname
			}
			break
		} else if usetiptracking && lhp.HasTipTracking() {
			bx.Refresh()
			return lhp.GetCleanTips(ctx, tiptype, channel, mirror, multi, usetiptracking)
		}
	}

	// if you don't find any suitable tips, why just make a
	// new box full of them!
	// nothing can possibly go wrong
	// surely

	if !foundit {
		// try adding a new tip box
		bx, err := inventory.NewTipbox(ctx, tiptype)

		if err != nil {
			return nil, nil, nil, wtype.LHError(wtype.LH_ERR_NO_TIPS, fmt.Sprintf("No tipbox of type ", tiptype, " found: %s", err))
		}

		r := lhp.AddTipBox(bx)

		if r != nil {
			err = r
			return nil, nil, nil, err
		}

		return lhp.GetCleanTips(ctx, tiptype, channel, mirror, multi, usetiptracking)
		//		return nil, nil, nil
	}

	return
}

func (lhp *LHProperties) DropDirtyTips(channel *wtype.LHChannelParameter, multi int) (wells, positions, boxtypes []string) {
	wells = make([]string, multi)
	positions = make([]string, multi)
	boxtypes = make([]string, multi)

	foundit := false

	for pos, bx := range lhp.Tipwastes {
		yes := bx.Dispose(multi)
		if yes {
			foundit = true
			for i := 0; i < multi; i++ {
				wells[i] = "A1"
				positions[i] = pos
				boxtypes[i] = bx.Type
			}

			break
		}
	}

	if !foundit {
		return nil, nil, nil
	}

	return
}

//GetMaterialType implement stockableMaterial
func (lhp *LHProperties) GetMaterialType() material.MaterialType {
	return lhp.MaterialType
}
func (lhp *LHProperties) GetTimer() *LHTimer {
	return GetTimerFor(lhp.Mnfr, lhp.Model)
}

func (lhp *LHProperties) GetChannelScoreFunc() ChannelScoreFunc {
	// this is to permit us to make this flexible

	sc := DefaultChannelScoreFunc{}

	return sc
}

// convenience method

func (lhp *LHProperties) RemoveComponent(plateID string, well string, volume wunit.Volume) bool {
	p := lhp.Plates[lhp.PlateIDLookup[plateID]]

	if p == nil {
		logger.Info(fmt.Sprint("RemoveComponent ", plateID, " ", well, " ", volume.ToString(), " can't find plate"))
		return false
	}

	r := p.RemoveComponent(well, volume)

	if r == nil {
		logger.Info(fmt.Sprint("CAN'T REMOVE COMPONENT ", plateID, " ", well, " ", volume.ToString()))
		return false
	}

	/*
		w := p.Wellcoords[well]

		if w == nil {
			logger.Info(fmt.Sprint("RemoveComponent ", plateID, " ", well, " ", volume.ToString(), " can't find well"))
			return false
		}

		c:=w.Remove(volume)

		if c==nil{
			logger.Info(fmt.Sprint("RemoveComponent ", plateID, " ", well, " ", volume.ToString(), " can't find well"))
			return false
		}
	*/

	return true
}

// RemoveUnusedAutoallocatedComponents removes any autoallocated component wells
// that didn't end up getting used
// In direct translation to component states that
// means any components that are temporary _and_ autoallocated.
func (lhp *LHProperties) RemoveUnusedAutoallocatedComponents() {
	ids := make([]string, 0, 1)
	for _, p := range lhp.Plates {
		if p.IsTemporary() && p.IsAutoallocated() {
			ids = append(ids, p.ID)
			continue
		}

		for _, w := range p.Wellcoords {
			if w.IsTemporary() && w.IsAutoallocated() {
				w.Clear()
			}
		}
	}

	for _, id := range ids {
		lhp.RemovePlateWithID(id)
	}

	// good
}
func (lhp *LHProperties) GetEnvironment() wtype.Environment {
	// static to start with

	return wtype.Environment{
		Temperature:         wunit.NewTemperature(25, "C"),
		Pressure:            wunit.NewPressure(100000, "Pa"),
		Humidity:            0.35,
		MeanAirFlowVelocity: wunit.NewVelocity(0, "m/s"),
	}
}

func (lhp *LHProperties) Evaporate(t time.Duration) []wtype.VolumeCorrection {
	// TODO: proper environmental calls
	env := lhp.GetEnvironment()
	ret := make([]wtype.VolumeCorrection, 0, 5)
	for _, v := range lhp.Plates {
		ret = append(ret, v.Evaporate(t, env)...)
	}

	return ret
}

// TODO -- allow drivers to provide relevant constraint info... not all positions
// can be used for tip loading
func (lhp *LHProperties) CheckTipPrefCompatibility(prefs []string) bool {
	// no new tip preferences allowed for now
	if lhp.Mnfr == "CyBio" {
		if lhp.Model == "Felix" {
			for _, v := range prefs {
				if !wutil.StrInStrArray(v, lhp.Tip_preferences) {
					return false
				}
				return true
			}
		} else if lhp.Model == "GeneTheatre" {
			for _, v := range prefs {
				if !wutil.StrInStrArray(v, lhp.Tip_preferences) {
					return false
				}
			}
			return true
		}

	} else if lhp.Mnfr == "Tecan" {
		// fall through
		return lhp.CheckPreferenceCompatibility(prefs)
	}

	return true
}

// CheckPreferenceCompatibility returns if the device specific configuration
// positions are compatible with the current device.
func (lhp *LHProperties) CheckPreferenceCompatibility(prefs []string) bool {
	// TODO: Not the most portable or extensible way

	var checkFn func(string) bool

	if lhp.Mnfr == "Tecan" {
		checkFn = func(pos string) bool {
			return strings.HasPrefix(pos, "TecanPos_")
		}
	} else if lhp.Mnfr == "Gilson" || lhp.Mnfr == "CyBio" && lhp.Model == "Felix" {
		checkFn = func(pos string) bool {
			return strings.HasPrefix(pos, "position_")
		}
	} else if lhp.Mnfr == "CyBio" && lhp.Model == "GeneTheatre" {
		checkFn = func(pos string) bool {
			if len(pos) != 2 {
				return false
			}
			return 'A' <= pos[0] && pos[0] <= 'D' && '0' <= pos[1] && pos[1] <= '9'
		}
	}

	if checkFn == nil {
		return true
	}

	for _, p := range prefs {
		if !checkFn(p) {
			return false
		}
	}

	return true
}

type UserPlate struct {
	Plate    *wtype.LHPlate
	Position string
}
type UserPlates []UserPlate

func (p *LHProperties) SaveUserPlates() UserPlates {
	up := make(UserPlates, 0, len(p.Positions))

	for pos, plate := range p.Plates {
		if plate.IsUserAllocated() {
			up = append(up, UserPlate{Plate: plate.DupKeepIDs(), Position: pos})
		}
	}

	return up
}

func (p *LHProperties) RestoreUserPlates(up UserPlates) {
	for _, plate := range up {
		oldPlate := p.Plates[plate.Position]
		p.RemovePlateAtPosition(plate.Position)
		// merge these
		plate.Plate.MergeWith(oldPlate)

		p.AddPlate(plate.Position, plate.Plate)
	}
}

func (p *LHProperties) MinPossibleVolume() wunit.Volume {
	if len(p.HeadsLoaded) == 0 {
		return wunit.ZeroVolume()
	}
	minvol := p.HeadsLoaded[0].GetParams().Minvol
	for _, head := range p.HeadsLoaded {
		for _, tip := range p.Tips {
			lhcp := head.Params.MergeWithTip(tip)
			v := lhcp.Minvol
			if v.LessThan(minvol) {
				minvol = v
			}
		}

	}

	return minvol
}

func (p *LHProperties) CanPossiblyDo(v wunit.Volume) bool {
	return !p.MinPossibleVolume().LessThan(v)
}

func (p *LHProperties) IsAddressable(pos string, crd wtype.WellCoords, channel, reference int, offsetX, offsetY, offsetZ float64) bool {
	// can we reach well 'crd' at position 'pos' using channel 'channel' at reference 'reference' with
	// the given offsets?

	// yes (this will improve, honest!)
	return true
}

func dupStrArr(sa []string) []string {
	ret := make([]string, len(sa))

	for i, v := range sa {
		ret[i] = v
	}

	return ret
}

func inStrArr(s string, sa []string) bool {
	for _, v := range sa {
		if s == v {
			return true
		}
	}

	return false
}

func (p *LHProperties) OrderedMergedPlatePrefs() []string {
	r := dupStrArr(p.Input_preferences)

	for _, pr := range p.Output_preferences {
		if !inStrArr(pr, r) {
			r = append(r, pr)
		}
	}

	return r
}

func (p LHProperties) HasTipTracking() bool {
	// TODO --> improve this

	if p.Mnfr == "Gilson" && p.Model == "Pipetmax" {
		return true
	}

	return false
}

func (p *LHProperties) UpdateComponentIDs(updates map[string]*wtype.LHComponent) {
	for s, c := range updates {
		p.UpdateComponentID(s, c)
	}
}

func (p *LHProperties) UpdateComponentID(from string, to *wtype.LHComponent) bool {
	for _, p := range p.Plates {
		if p.FindAndUpdateID(from, to) {
			return true
		}
	}
	return false
}
