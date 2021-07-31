// Copyright (C) 2021 Max Nikulin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package burl_fuzzy

import "testing"

var searchCases = []struct {
	id, query, text string
	distance        EditorDistance
	description     string
}{
	{"exact", "python", "python", 0, ""},
	{"exactSubstring", "dfg", "asdfghjkl", 0, ""},
	{"transposition_1", "pyhton", "python", TranspositionDistance, ""},
	{"insertionMiddle_1", "pyton", "python", InsertionDistance, ""},
	{"deletionStart_1", "apython", "python", DeletionDistance, ""},
	{"suffixMismatch_2", "pythonic", "pythons", SubstitutionDistance + DeletionDistance, ""},
	{"prefixSuffixMismatch_8", "python.org", "www.python", 4 * DeletionDistance, ""},
}

func TestApproximateSearch(t *testing.T) {
	for _, c := range searchCases {
		t.Run(c.id, func(t *testing.T) {
			d := SearchDistanceDL(c.query, c.text)
			if d != c.distance {
				t.Errorf("%v != %v = SearchDistanceDL(%q, %q) // %s",
					c.distance, d, c.query, c.text, c.description)
			}
		})
	}
}
