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

// https://gist.github.com/dperini/729294
const (
	UrlPatternSchemaSlashes = `(?:([[:alpha:]]+:(\/\/)?))`
	//patUserPassword = `(?:(\S+)(?::(\S*))?@)`
	patDomainComponent  = `(?:[\p{L}0-9][\p{L}0-9_-]{0,62})?[\p{L}0-9]`
	patTopLevelDomain   = `(?:\p{L}){2,64}\.?`
	UrlPatternDomainTLD = "((?:" + patDomainComponent + `\.){0,16}(` + patTopLevelDomain + `))`
	patPathComponent    = `(?:[\p{L}0-9.@,~_-]|(?:%[0-9a-fA-F]{2}))*`
	patPath             = "((?:/" + patPathComponent + ")*)"
	UrlPatternFull      = UrlPatternSchemaSlashes + "?" + UrlPatternDomainTLD + "?" + patPath + "?"
)

const (
	NPatFull = iota
	NPatSchema
	NPatSlashes
	NPatDomain
	NPatTopLevelDomain
	NPatPath
)

func MatchIsUrl(match []string) bool {
	if len(match) <= 0 {
		return false
	}
	if len(match[0]) == 0 {
		return false
	}
	if len(match) > NPatSchema && match[0] == match[NPatSchema] {
		return false
	}
	// FIXME
	if len(match) > NPatTopLevelDomain {
		if match[0] == match[NPatTopLevelDomain] {
			return false
		}
		discardRelativePath := match[NPatDomain] == match[NPatTopLevelDomain] && match[NPatSchema] == ""
		if discardRelativePath {
			// "localhost" might need special treatment
			return false
		}
	}
	if len(match) > NPatPath && match[0] == match[NPatPath] {
		return false
	}
	return true
}
