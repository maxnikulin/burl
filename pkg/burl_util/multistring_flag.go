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

import "errors"

const (
	initial = iota
	appended = iota
	reset = iota
)

type MultiStringFlag struct {
	slice      *[]string
	state      int
	defaultLen int
}

func NewMultiStringFlag(slice *[]string) *MultiStringFlag {
	return &MultiStringFlag{slice, initial, len(*slice)}
}

func (MultiStringFlag) String() string {
	return "FIXME: this is MultiStringFlag proxy, value should not be accessed directly"
}

func (f *MultiStringFlag) Set(value string) error {
	if value == "" {
		if f.state == appended {
			return errors.New("Attempt to discard earlier appended argument")
		}
		*f.slice = (*f.slice)[:0]
		f.defaultLen = 0
		f.state = reset
	} else {
		*f.slice = append(*f.slice, value)
		if f.state == initial {
			f.state = appended
		}
	}
	return nil
}

func (f *MultiStringFlag) IsModified() bool {
	return f.state != initial
}

func (f *MultiStringFlag) Values() []string {
	return (*f.slice)[:]
}

func (f *MultiStringFlag) ModifiedValues() []string {
	switch f.state {
	case initial:
		return nil
	case reset:
		retval := make([]string, len(*f.slice)+1)
		retval[0] = ""
		copy(retval[1:], *f.slice)
		return retval
	case appended:
		return (*f.slice)[f.defaultLen:]
	}
	return nil
}
