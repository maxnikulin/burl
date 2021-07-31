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

package burl_links

import (
	"regexp"
	"testing"
)

type casesT []struct {
	word   string
	result bool
}

// TODO use reflection to get symbol from module by its name.
// Unfortunately it will break compile-time checks
var caseArray = []struct {
	name, pattern string
	cases         casesT
}{
	{
		"schemaSlashes", UrlPatternSchemaSlashes, casesT{
			{`"http://"`, true},
			{`abc`, false},
			{`ftp`, false},
			{`https://`, true},
		},
	},
	{
		"domainTld", UrlPatternDomainTLD, casesT{
			{"www.geocities.com", true},
			{"12", false},
			{" почта.рф", true},
			{"abc", true},
		},
	},
	{
		"full", UrlPatternFull, casesT{
			{"http://www.geocities.com/~user/index.html", true},
			{"(http://www.geocities.com/~user/index.html)", true},
			{" absf asdf http://www.geocities.com/~user/index.html word end", true},
			{"src/FFmpeg.h", true},
		},
	},
}

func TestUrlPattern(t *testing.T) {
	for _, group := range caseArray {
		re := regexp.MustCompile(group.pattern)
		for _, c := range group.cases {
			t.Run(group.name+"="+c.word, func(t *testing.T) {
				m := re.FindAllString(c.word, -1)
				if c.result != (len(m) > 0) {
					t.Errorf("expected %v match %q", c.result, m)
				}
			})
		}
	}
}

var casesMatch = casesT{
	{"http://www.geocities.com/~user/index.html", true},
	{"(http://www.geocities.com/~user/index.html)", true},
	{"src/FFmpeg.h", false},
}

func TestMatchIsUrl(t *testing.T) {
	re := regexp.MustCompile(UrlPatternFull)
	for _, c := range casesMatch {
		t.Run("url="+c.word, func(t *testing.T) {
			m_dirty := re.FindAllStringSubmatch(c.word, -1)
			if len(m_dirty) > 0 {
				var m [][]string
				for _, match := range m_dirty {
					if match[0] != "" {
						m = append(m, match)
					}
				}
				if len(m) == 1 {
					actual := MatchIsUrl(m[0])
					if c.result != actual {
						t.Errorf("expected %v != actual %v match %q", c.result, actual, m[0])
					}
				} else {
					t.Errorf("expected exactly 1 match, actual %v: %q", len(m), m)
				}
			} else if c.result {
				t.Errorf("no url match")
			}
		})
	}
}
