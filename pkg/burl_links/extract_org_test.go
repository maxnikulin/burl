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

var orgExtractSetCases = []struct {
	name   string
	set    []string
	input  string
	filter []string
}{
	{
		"simpleSinglePlain",
		[]string{"mid:plain-match-1@de.fg"},
		"A mid:plain-match-1@de.fg link and http://anoth.er/plain-error-1.html link.",
		[]string{"mid"},
	},
	{
		"simpleSingleAngle",
		[]string{"http://anoth.er/angled-match-2.html"},
		"A <mid:angled-error-2@de.fg> link and <http://anoth.er/angled-match-2.html> link.",
		[]string{"http:"},
	},
	{
		"simpleSingleBracket",
		[]string{"mid:bracketed-match-3@de.fg"},
		"A [[mid:bracketed-match-3@de.fg]] link and [[http://anoth.er/mailnews:bracketed-error-3.html][error link]].",
		[]string{"mid", "news"},
	},
	{
		"multiplePlainAndAngled",
		[]string{"mid:plain-match-4@de.fg", "news://news.mozilla.org/angled-match-4@hi.jk"},
		`A mid:plain-match-4@de.fg link and http://anoth.er/plain-error-4.html link.
		<news://news.mozilla.org/angled-match-4@hi.jk>`,
		[]string{"news:", "mid", "nntp:"},
	},
	{
		"mixedPrefixes",
		[]string{"mid:plain-match-5@de.fg", "https://list.orgmode.org/bracketed-match-5@hi.jk"},
		`A mid:plain-match-5@de.fg link and https://anoth.er/error-1.html link.
		[[https://list.orgmode.org/bracketed-match-5@hi.jk]]`,
		[]string{"mid", "https://list.orgmode.org/"},
	},
}

func TestOrgExtractSetCases(t *testing.T) {
	for _, props := range orgExtractSetCases {
		expect := map[string]bool{}
		for _, link := range props.set {
			expect[link] = true
		}
		t.Run(props.name, func(t *testing.T) {
			source := OrgLinkSource("test")
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

var makeLinkSetBaseCases = []struct {
	expect  string
	filters []string
}{
	{"(?:mid:)", []string{"mid"}},
	{"(?:http:|https:)", []string{"http", "https:"}},
}

func TestMakeLinkSetBaseCases(t *testing.T) {
	for _, props := range makeLinkSetBaseCases {
		t.Run(props.expect, func(t *testing.T) {
			actual, err := MakeLinkSetBase(props.filters)
			if err != nil {
				t.Errorf("unexpected error %v, actual %v", err, actual)
			}
			if props.expect != actual {
				t.Errorf("%v != %v", props.expect, actual)
			}
		})
	}
}
