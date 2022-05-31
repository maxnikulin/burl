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

package burl_util

type MultiStringFlag struct {
	slice    *[]string
	modified bool
}

func NewMultiStringFlag(slice *[]string) *MultiStringFlag {
	return &MultiStringFlag{slice, false}
}

func (MultiStringFlag) String() string {
	return "FIXME: this is MultiStringFlag proxy, value should not be accessed directly"
}

func (f *MultiStringFlag) Set(value string) error {
	if value == "" {
		*f.slice = (*f.slice)[:0]
	} else {
		*f.slice = append(*f.slice, value)
	}
	f.modified = true
	return nil
}

func (f *MultiStringFlag) IsModified() bool {
	return f.modified
}

func (f *MultiStringFlag) Values() []string {
	return (*f.slice)[:]
}
