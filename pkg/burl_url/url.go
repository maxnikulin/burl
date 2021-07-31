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

package burl_url

import "strings"

func UrlVariants(url string) []string {
	replacements := map[string]string{
		"http:":  "https:",
		"https:": "http:",
	}
	retval := RedundantTailVariants(url)
	for k, v := range replacements {
		if strings.HasPrefix(url, k) {
			retval = append(retval, RedundantTailVariants(v+url[len(k):])...)
		}
	}
	return retval
}

// Consider URLs ending with "/?&#" as equivalent.
func RedundantTailVariants(url string) []string {
	retval := make([]string, 0, 1)
	l := len(url)
	if l == 0 {
		return retval
	}
	slice := make([]rune, l)
	n := 0
	for _, r := range url {
		slice[n] = r
		n++
	}
	slice = slice[:n]
	i := n
	if i > 0 && slice[i-1] == '#' {
		i--
	}
	for i > 0 && slice[i-1] == '&' {
		i--
	}
	if i > 0 && slice[i-1] == '?' {
		i--
	}
	for i > 0 && slice[i-1] == '/' {
		i--
	}

	for ; i <= n; i++ {
		retval = append(retval, string(slice[:i]))
	}
	return retval
}
