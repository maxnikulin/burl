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
	"flag"
	"strings"
)

type MixedSrcTypeSlice []TextLinkSource

var MixedSrcNames MixedSrcTypeSlice

type MixedSrcTypeProxy struct {
	target  *MixedSrcTypeSlice
	factory func(value string) TextLinkSource
}

func NewMixedSrcTypeProxyPtr(target *MixedSrcTypeSlice, factory func(value string) TextLinkSource) *MixedSrcTypeProxy {
	return &MixedSrcTypeProxy{target, factory}
}

func (MixedSrcTypeProxy) String() string {
	return "FIXME: this is a proxy, value should be accessed directly"
}

// TODO check whether not empty and a readable regular file
func (p *MixedSrcTypeProxy) Set(value string) error {
	*p.target = append(*p.target, p.factory(value))
	return nil
}

func AddSourceFlags(slice *MixedSrcTypeSlice, flagSet *flag.FlagSet) {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	if slice == nil {
		if MixedSrcNames == nil {
			MixedSrcNames = make(MixedSrcTypeSlice, 0, 4)
		}
		slice = &MixedSrcNames
	}

	flag.Var(NewMixedSrcTypeProxyPtr(
		slice,
		func(value string) TextLinkSource { return TxtLinkSource(value) }),
		"txt", "Process `FILE` as plain text file (multiple)")
	flag.Var(NewMixedSrcTypeProxyPtr(
		slice,
		func(value string) TextLinkSource { return OrgLinkSource(value) }),
		OrgLinkSource("").Flag(), "Process `FILE` as Emacs Org Mode file (multiple)")
}

func AddSourceArgs(slice MixedSrcTypeSlice, args []string) MixedSrcTypeSlice {
	for _, arg := range args {
		var src TextLinkSource
		if strings.HasSuffix(arg, ".org") {
			src = OrgLinkSource(arg)
		} else {
			src = TxtLinkSource(arg)
		}
		slice = append(slice, src)
	}
	return slice
}
