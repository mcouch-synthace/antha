// Part of the Antha language
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

// Package for facilitating DOE methodology in antha
package doe

import (
	"fmt"

	"github.com/antha-lang/antha/antha/anthalib/wtype"

	"github.com/tealeg/xlsx"
)

func JMPXLSXFilefromRuns(runs []Run, outputfilename string) (xlsxfile *xlsx.File) {

	// if output is a struct look for a sensible field to print

	//var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	xlsxfile = xlsx.NewFile()
	sheet, err = xlsxfile.AddSheet("Sheet1")
	if err != nil {
		panic(err.Error())
	}
	// new row
	row = sheet.AddRow()

	// then add subheadings and descriptors
	for _, descriptor := range runs[0].Factordescriptors {

		cell = row.AddCell()
		cell.Value = descriptor

	}
	for _, descriptor := range runs[0].Responsedescriptors {
		cell = row.AddCell()
		cell.Value = descriptor

	}
	for _, descriptor := range runs[0].AdditionalSubheaders {
		cell = row.AddCell()
		cell.Value = descriptor

	}
	//add data 1 row per run
	for _, run := range runs {

		row = sheet.AddRow()

		// factors
		for _, factor := range run.Setpoints {

			cell = row.AddCell()

			dna, amIdna := factor.(wtype.DNASequence)
			if amIdna {
				cell.SetValue(dna.Nm)
			} else {
				cell.SetValue(factor) //= factor.(string)
			}

		}

		// responses
		for _, response := range run.ResponseValues {
			cell = row.AddCell()
			cell.SetValue(response)
		}

		// additional
		for _, additional := range run.AdditionalValues {
			cell = row.AddCell()
			cell.SetValue(additional)
		}
	}
	err = xlsxfile.Save(outputfilename)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return
}

func XLSXFileFromRuns(runs []Run, outputfilename string, dxorjmp string) (xlsxfile *xlsx.File) {
	if dxorjmp == "DX" {
		xlsxfile = DXXLSXFilefromRuns(runs, outputfilename)
	} else if dxorjmp == "JMP" {
		xlsxfile = JMPXLSXFilefromRuns(runs, outputfilename)
	} else {
		panic(fmt.Sprintf("Unknown design file format %s when exporting design to XLSX file. Please specify File type as JMP or DX (Design Expert)", dxorjmp))
	}
	return
}