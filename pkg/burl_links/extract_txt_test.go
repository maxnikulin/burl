// Copyright (C) 2022 Max Nikulin
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

package burl_links

import (
	"reflect"
	"strings"
	"testing"
)

func TestTxtExtractSetCases(t *testing.T) {
	// Plain text regexp is general enough to recognize URLs
	// in angled or bracketed Org links.
	for _, props := range orgExtractSetCases {
		expect := map[string]bool{}
		for _, link := range props.set {
			expect[link] = true
		}
		t.Run(props.name, func(t *testing.T) {
			source := TxtLinkSource("test")
			actual := map[string]bool{}
			err := source.ExtractSet(strings.NewReader(props.input), props.filter, &actual)
			if err != nil {
				t.Errorf("unexpected error %v, actual %v", err, actual)
			}
			if !reflect.DeepEqual(expect, actual) {
				t.Errorf("%v != %v", expect, actual)
			}
		})
	}
}
