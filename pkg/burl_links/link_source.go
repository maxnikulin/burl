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
	"io"
	"os"
)

type Filter func(*Link) bool

type TextLinkSource interface {
	Name() string
	Extract(file io.Reader, filter Filter) (*TreeChildrenNode, error)
	// map is a set for poors.
	ExtractSet(file io.Reader, filters []string, result *map[string]bool) error
	Flag() string
	Clone(string) TextLinkSource
}

// Separate function to have proper scope for file.Close
func ExtractLinksFromFile(src TextLinkSource, filter Filter) (*TreeChildrenNode, error) {
	reader := os.Stdin
	name := src.Name()
	if name != "-" {
		file, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		reader = file
		defer file.Close()
	}
	tree, err := src.Extract(reader, filter)
	if tree != nil {
		if !tree.Empty() {
			tree.Props = &FileProps{name}
		} else {
			tree = nil
		}
	}
	return tree, err
}

func ExtractLinksFromFileGroup(list []TextLinkSource, filter Filter) (*TreeChildrenNode, error) {
	if len(list) == 0 {
		return nil, nil // TODO error
	}
	var err error
	var tree *TreeChildrenNode
	g := NewTreeChildrenNode(&FileGroupProps{})
	group := &g
	for _, src := range list {
		tree, err = ExtractLinksFromFile(src, filter)
		if tree != nil {
			group.AppendChild(tree)
		}
		if err != nil {
			break
		}
	}
	if group.Empty() {
		group = nil
	}
	return group, err
}

func ExtractLinkSetFromFileGroup(list []TextLinkSource, filters []string) (map[string]bool, error) {
	result := map[string]bool{}
	var err error
	for _, src := range list {
		var reader io.Reader
		name := src.Name()
		if name == "-" {
			reader = os.Stdin
		} else {
			var file io.ReadCloser
			file, err = os.Open(name)
			if err != nil {
				break
			}
			reader = file
			defer file.Close()
		}
		err = src.ExtractSet(reader, filters, &result)
		if err != nil {
			break
		}
	}
	return result, err
}
