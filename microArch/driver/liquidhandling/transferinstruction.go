// /anthalib/driver/liquidhandling/transferinstruction.go: Part of the Antha language
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

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/antha/anthalib/wutil"
	"github.com/antha-lang/antha/inventory"
)

func firstInArray(a []*wtype.LHPlate) *wtype.LHPlate {
	for _, v := range a {
		if v != nil {
			return v
		}
	}

	return nil
}

type TransferInstruction struct {
	GenericRobotInstruction
	Type      int
	Platform  string
	Transfers []MultiTransferParams
}

func (ti *TransferInstruction) ToString() string {
	s := fmt.Sprintf("%s ", Robotinstructionnames[ti.Type])
	for i := 0; i < len(ti.Transfers); i++ {
		s += ti.ParamSet(i).ToString()
		s += "\n"
	}

	return s
}

func (ti *TransferInstruction) ParamSet(n int) MultiTransferParams {
	return ti.Transfers[n]
}

func NewTransferInstruction(what, pltfrom, pltto, wellfrom, wellto, fplatetype, tplatetype []string, volume, fvolume, tvolume []wunit.Volume, FPlateWX, FPlateWY, TPlateWX, TPlateWY []int, Components []string) *TransferInstruction {
	var tfri TransferInstruction
	tfri.Type = TFR
	tfri.Transfers = make([]MultiTransferParams, 0, 1)

	/*
		v := MultiTransferParams{
			What:       what,
			PltFrom:    pltfrom,
			PltTo:      pltto,
			WellFrom:   wellfrom,
			WellTo:     wellto,
			Volume:     volume,
			FPlateType: fplatetype,
			TPlateType: tplatetype,
			FVolume:    fvolume,
			TVolume:    tvolume,
			FPlateWX:   FPlateWX,
			FPlateWY:   FPlateWY,
			TPlateWX:   TPlateWX,
			TPlateWY:   TPlateWY,
			Components: Components,
		}
	*/

	v := MTPFromArrays(what, pltfrom, pltto, wellfrom, wellto, fplatetype, tplatetype, volume, fvolume, tvolume, FPlateWX, FPlateWY, TPlateWX, TPlateWY, Components)

	tfri.Add(v)
	tfri.GenericRobotInstruction.Ins = RobotInstruction(&tfri)
	return &tfri
}

func (ins *TransferInstruction) OutputTo(drv LiquidhandlingDriver) error {
	hlld, ok := drv.(HighLevelLiquidhandlingDriver)

	if !ok {
		return fmt.Errorf("Driver type %T not compatible with TransferInstruction, need HighLevelLiquidhandlingDriver", drv)
	}

	// make sure we disable the RobotInstruction pointer
	ins.GenericRobotInstruction = GenericRobotInstruction{}

	volumes := make([]float64, len(SetOfMultiTransferParams(ins.Transfers).Volume()))
	for i, vol := range SetOfMultiTransferParams(ins.Transfers).Volume() {
		volumes[i] = vol.ConvertTo(wunit.ParsePrefixedUnit("ul"))
	}

	reply := hlld.Transfer(SetOfMultiTransferParams(ins.Transfers).What(), SetOfMultiTransferParams(ins.Transfers).PltFrom(), SetOfMultiTransferParams(ins.Transfers).WellFrom(), SetOfMultiTransferParams(ins.Transfers).PltTo(), SetOfMultiTransferParams(ins.Transfers).WellTo(), volumes)

	if !reply.OK {
		return fmt.Errorf(" %d : %s", reply.Errorcode, reply.Msg)
	}

	return nil
}

func (tfri *TransferInstruction) Add(tp MultiTransferParams) {
	tfri.Transfers = append(tfri.Transfers, tp)
}

func (ins *TransferInstruction) InstructionType() int {
	return ins.Type
}

//what, pltfrom, pltto, wellfrom, wellto, fplatetype, tplatetype []string, volume, fvolume, tvolume []wunit.Volume, FPlateWX, FPlateWY, TPlateWX, TPlateWY []int, Components []string
func (ins *TransferInstruction) Dup() *TransferInstruction {
	var tfri TransferInstruction
	tfri.Type = TFR
	tfri.Transfers = make([]MultiTransferParams, 0, 1)
	tfri.Platform = ins.Platform

	for i := 0; i < len(ins.Transfers); i++ {
		tfri.Add(ins.Transfers[i].Dup())
	}

	tfri.GenericRobotInstruction.Ins = RobotInstruction(&tfri)
	return &tfri
}

func (ins *TransferInstruction) MergeWith(ins2 *TransferInstruction) *TransferInstruction {
	ret := ins.Dup()

	for _, v := range ins2.Transfers {
		ins.Add(v)
	}

	return ret
}

func (ins *TransferInstruction) GetParameter(name string) interface{} {
	switch name {
	case "LIQUIDCLASS":
		return SetOfMultiTransferParams(ins.Transfers).What()
	case "VOLUME":
		return SetOfMultiTransferParams(ins.Transfers).Volume()
	case "FROMPLATETYPE":
		return SetOfMultiTransferParams(ins.Transfers).FPlateType()
	case "WELLFROMVOLUME":
		return SetOfMultiTransferParams(ins.Transfers).FVolume()
	case "POSFROM":
		return SetOfMultiTransferParams(ins.Transfers).PltFrom()
	case "POSTO":
		return SetOfMultiTransferParams(ins.Transfers).PltTo()
	case "WELLFROM":
		return SetOfMultiTransferParams(ins.Transfers).WellFrom()
	case "WELLTO":
		return SetOfMultiTransferParams(ins.Transfers).WellTo()
	case "WELLTOVOLUME":
		return SetOfMultiTransferParams(ins.Transfers).TVolume()
	case "TOPLATETYPE":
		return SetOfMultiTransferParams(ins.Transfers).TPlateType()
	case "FPLATEWX":
		return SetOfMultiTransferParams(ins.Transfers).FPlateWX()
	case "FPLATEWY":
		return SetOfMultiTransferParams(ins.Transfers).FPlateWY()
	case "TPLATEWX":
		return SetOfMultiTransferParams(ins.Transfers).TPlateWX()
	case "TPLATEWY":
		return SetOfMultiTransferParams(ins.Transfers).TPlateWY()
	case "INSTRUCTIONTYPE":
		return ins.InstructionType()
	case "PLATFORM":
		return ins.Platform
	case "COMPONENT":
		return SetOfMultiTransferParams(ins.Transfers).Component()
	}
	return nil
}

func (vs VolumeSet) MaxMultiTransferVolume(minLeave wunit.Volume) wunit.Volume {
	// the minimum volume in the set... ensuring that we what we leave is
	// either 0 or minLeave or greater

	ret := vs[0].Dup()

	for _, v := range vs {
		if v.LessThan(ret) && !v.IsZero() {
			ret = v.Dup()
		}
	}

	vs2 := vs.Dup().Sub(ret)

	if !vs2.NonZeros().Min().IsZero() && vs2.NonZeros().Min().LessThan(minLeave) {
		//slightly inefficient but we refuse to leave less than minleave
		ret.Subtract(minLeave)
	}

	// fail if ret is now < 0 or < the min possible

	if ret.LessThan(wunit.ZeroVolume()) || ret.LessThan(minLeave) {
		ret = wunit.ZeroVolume()
	}

	return ret
}

func (ins *TransferInstruction) CheckMultiPolicies(which int) bool {
	// first iteration: ensure all the WHAT prms are the same
	// later	  : actually check the policies per channel

	nwhat := wutil.NUniqueStringsInArray(ins.Transfers[which].What(), true)

	if nwhat != 1 {
		return false
	}

	return true
}

func plateTypeArray(ctx context.Context, types []string) ([]*wtype.LHPlate, error) {
	plates := make([]*wtype.LHPlate, len(types))
	for i, typ := range types {
		if typ == "" {
			continue
		}
		p, err := inventory.NewPlate(ctx, typ)
		if err != nil {
			return nil, err
		}
		plates[i] = p
	}
	return plates, nil
}

func (ins *TransferInstruction) GetParallelSetsFor(ctx context.Context, channel *wtype.LHChannelParameter) []int {
	// if the channel is not multi just return nil

	if channel.Multi == 1 {
		return nil
	}

	r := make([]int, 0, len(ins.Transfers))

	for i := 0; i < len(ins.Transfers); i++ {
		if ins.validateParallelSet(ctx, channel, i) {
			r = append(r, i)
		}
	}

	return r
}

func (ins *TransferInstruction) validateParallelSet(ctx context.Context, channel *wtype.LHChannelParameter, which int) bool {

	if len(ins.Transfers[which].What()) > channel.Multi {
		return false
	}

	npositions := wutil.NUniqueStringsInArray(ins.Transfers[which].PltFrom(), true)

	if npositions != 1 {
		// fall back to single-channel
		// TODO -- find a subset we CAN do, if such exists
		return false
	}

	nplatetypes := wutil.NUniqueStringsInArray(ins.Transfers[which].FPlateType(), true)

	if nplatetypes != 1 {
		// fall back to single-channel
		// TODO -- find a subset we CAN do , if such exists
		return false
	}

	pa, err := plateTypeArray(ctx, ins.Transfers[which].FPlateType())

	if err != nil {
		panic(err)
	}

	// check source / tip alignment

	plate := firstInArray(pa)

	if plate == nil {
		panic("No from plates in instruction")
	}

	if !wtype.TipsWellsAligned(*channel, *plate, ins.Transfers[which].WellFrom()) {
		// fall back to single-channel
		// TODO -- find a subset we CAN do
		return false
	}

	pa, err = plateTypeArray(ctx, ins.Transfers[which].TPlateType())

	if err != nil {
		panic(err)
	}

	plate = firstInArray(pa)

	if plate == nil {
		panic("No to plates in instruction")
	}

	// for safety, check dest / tip alignment

	if !wtype.TipsWellsAligned(*channel, *plate, ins.Transfers[which].WellTo()) {
		// fall back to single-channel
		// TODO -- find a subset we CAN do
		return false
	}

	// check that we will not require different policies

	if !ins.CheckMultiPolicies(which) {
		// fall back to single-channel
		// TODO -- find a subset we CAN do
		return false
	}

	// looks OK

	return true
}

func GetMultiSet(a []string, channelmulti int, fromplatemulti int, toplatemulti int) [][]int {
	ret := make([][]int, 0, 2)
	var next []int
	for {
		next, a = GetNextSet(a, channelmulti, fromplatemulti, toplatemulti)
		if next == nil {
			break
		}

		ret = append(ret, next)
	}

	return ret
}

func GetNextSet(a []string, channelmulti int, fromplatemulti int, toplatemulti int) ([]int, []string) {
	if len(a) == 0 {
		return nil, nil
	}
	r := make([][]int, fromplatemulti)
	for i := 0; i < fromplatemulti; i++ {
		r[i] = make([]int, toplatemulti)
		for j := 0; j < toplatemulti; j++ {
			r[i][j] = -1
		}
	}

	// this is simply a greedy algorithm, it may miss things
	for _, s := range a {
		tx := strings.Split(s, ",")
		i, _ := strconv.Atoi(tx[0])
		j, _ := strconv.Atoi(tx[1])
		k, _ := strconv.Atoi(tx[2])
		r[i][j] = k
	}
	// now we just take the first one we find

	ret := getset(r, channelmulti)
	censa := censoredcopy(a, ret)

	return ret, censa
}

func getset(a [][]int, mx int) []int {
	r := make([]int, 0, mx)

	for i := 0; i < len(a); i++ {
		for j := 0; j < len(a[i]); j++ {
			if a[i][j] != -1 {
				r = append(r, a[i][j])
				// find a diagonal line
				for l := 1; l < mx; l++ {
					x := (i + l) % len(a)
					y := (j + l) % len(a[i])

					if a[x][y] != -1 {
						r = append(r, a[x][y])
					} else {
						r = make([]int, 0, mx)
					}
				}

				if len(r) == mx {
					break
				}
			}
		}
	}

	if len(r) == mx {
		sort.Ints(r)
		return r
	} else {
		return nil
	}
}

func censoredcopy(a []string, b []int) []string {
	if b == nil {
		return a
	}

	r := make([]string, 0, len(a)-len(b))

	for _, x := range a {
		tx := strings.Split(x, ",")
		i, _ := strconv.Atoi(tx[2])
		if IsIn(i, b) {
			continue
		}
		r = append(r, x)
	}

	return r
}

func IsIn(i int, a []int) bool {
	for _, x := range a {
		if i == x {
			return true
		}
	}

	return false
}

func (ins *TransferInstruction) ChooseChannels(prms *LHProperties) {
	for i, mtp := range ins.Transfers {
		// we need to remove leading blanks
		ins.Transfers[i] = mtp.RemoveInitialBlanks()
	}
}

func (ins *TransferInstruction) Generate(ctx context.Context, policy *wtype.LHPolicyRuleSet, prms *LHProperties) ([]RobotInstruction, error) {
	// if the liquid handler is of the high-level type we cut the tree here
	// after ensuring that the transfers are within limitations of the liquid handler

	if prms.LHType == HLLiquidHandler {
		err := ins.ReviseTransferVolumes(prms)

		if err != nil {
			return []RobotInstruction{}, err
		}
		return []RobotInstruction{}, nil
	}

	//  set the channel  choices first by cleaning out initial empties

	ins.ChooseChannels(prms)

	pol := GetPolicyFor(policy, ins)

	ret := make([]RobotInstruction, 0)

	// if we can multi we do this first
	if pol["CAN_MULTI"].(bool) {
		parallelsets := ins.GetParallelSetsFor(ctx, prms.HeadsLoaded[0].Params)

		mci := NewMultiChannelBlockInstruction()
		mci.Prms = prms.HeadsLoaded[0].Params // TODO Remove Hard code here

		// to do below
		//
		//	- divide up transfer into multi and single transfers
		//  	  in practice this means finding the maximum we can do
		//	  then doing that as a transfer and generating single channel transfers
		//	  to mop up the left
		//

		for _, set := range parallelsets {
			vols := VolumeSet(ins.Transfers[set].Volume())

			maxvol := vols.MaxMultiTransferVolume(prms.MinPossibleVolume())

			// if we can't do it, we can't do it
			if maxvol.IsZero() {
				continue
			}

			tp := ins.Transfers[set].Dup()

			for i := 0; i < len(tp.Transfers); i++ {
				tp.Transfers[i].Volume = maxvol.Dup()
			}

			// now set the vols for the transfer and remove this from the instruction's volume

			for i := range vols {
				vols[i] = wunit.CopyVolume(maxvol)
			}

			ins.Transfers[set].RemoveVolume(maxvol)

			// set the from and to volumes for the relevant part of the instruction
			ins.Transfers[set].RemoveFVolume(maxvol)
			ins.Transfers[set].AddTVolume(maxvol)

			mci.Multi = len(vols)
			mci.AddTransferParams(tp)
		}

		if len(parallelsets) > 0 && len(mci.Volume) > 0 {
			ret = append(ret, mci)
		}
	}

	// mop up all the single instructions which are left
	sci := NewSingleChannelBlockInstruction()
	sci.Prms = prms.HeadsLoaded[0].Params // TODO Fix Hard Code Here

	lastWhat := ""
	for _, t := range ins.Transfers {
		for _, tp := range t.Transfers {
			if tp.Volume.LessThanFloat(0.001) {
				continue
			}

			// TODO --> reorder instructions
			if lastWhat != "" && tp.What != lastWhat {
				if len(sci.Volume) > 0 {
					ret = append(ret, sci)
				}
				sci = NewSingleChannelBlockInstruction()
				sci.Prms = prms.HeadsLoaded[0].Params
			}

			sci.AddTransferParams(tp)
			lastWhat = tp.What
		}
	}
	if len(sci.Volume) > 0 {
		ret = append(ret, sci)
	}

	return ret, nil
}

func (ins *TransferInstruction) ReviseTransferVolumes(prms *LHProperties) error {
	newTransfers := make([]MultiTransferParams, len(ins.Transfers))

	for _, mtp := range ins.Transfers {
		//newMtp := make(MultiTransferParams, len(mtp))
		newMtp := NewMultiTransferParams(mtp.Multi)
		for _, tp := range mtp.Transfers {
			if tp.What == "" {
				continue
			}
			newTPs, err := safeTransfers(tp, prms)
			if err != nil {
				return err
			}
			newMtp.Transfers = append(newMtp.Transfers, newTPs...)
		}

		newTransfers = append(newTransfers, newMtp)
	}

	ins.Transfers = newTransfers

	return nil
}

func safeTransfers(tp TransferParams, prms *LHProperties) ([]TransferParams, error) {

	if tp.What == "" {
		return []TransferParams{tp}, nil
	}

	tvs, err := TransferVolumes(tp.Volume, prms.HeadsLoaded[0].Params.Minvol, prms.HeadsLoaded[0].Params.Maxvol)

	ret := []TransferParams{}

	if err != nil {
		return ret, err
	}

	fwv := tp.FVolume.Dup()
	twv := tp.TVolume.Dup()

	for _, v := range tvs {
		ntp := tp.Dup()
		ntp.Volume = v
		ntp.FVolume = fwv.Dup()
		ntp.TVolume = twv.Dup()
		fwv.Subtract(v)
		twv.Add(v)

		ret = append(ret, ntp)
	}

	return ret, nil
}
