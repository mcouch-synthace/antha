// wunit/unitfromstring.go: Part of the Antha language
// Copyright (C) 2014 the Antha authors. All rights reserved.
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

package wunit

import (
	"github.com/pkg/errors"
	"strconv"
)

// extractFloat extract the longest valid float from the left hand side of
// the string and return it and the remainder of the string
// returns an error if no float is found
func extractFloat(s string) (float64, string, error) {
	var longest int
	var ret float64
	for i := range s {
		if f, err := strconv.ParseFloat(s[:i], 64); err == nil {
			ret = f
			longest = i
		}
	}

	if longest == 0 {
		return 0.0, s, errors.Errorf("unable to extract float from %q", s)
	}
	return ret, s[longest:], nil
}
