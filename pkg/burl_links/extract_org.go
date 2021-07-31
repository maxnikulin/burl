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
	"bufio"
	"io"
	"log"
	"regexp"
)

var reHeading = regexp.MustCompile(`^(\*+)\s+(\S(?:.*\S)?)?$`)
var reSchemeStr = "https?|doi"
var reScheme = regexp.MustCompile("^" + reSchemeStr + ":")
var reBracketStr = "\\[\\[((?:[^\\]\\[]|\\\\(?:\\\\\\\\)*[\\]\\[]|\\\\+[^\\]\\[])+)](?:\\[((?:.|\n)+?)\\])?\\]"
var reAngleStr = "<(" + reSchemeStr + "):([^>\n]*(?:\n[ \t]*[^> \t\n][^> \n]*)*)>"
var rePlainStr = "\\b(" + reSchemeStr + "):((?:[^][ \t\n(\\)<>]|\\((?:[^][ \t\n(\\)<>]|\\([^][ \t\n(\\)<>]*\\))*\\))+(?:[^[:punct:] \t\n]|/|\\((?:[^][ \t\n(\\)<>]|\\([^][ \t\n(\\)<>]*\\))*\\)))"

var reLink = regexp.MustCompile(reBracketStr + "|" + reAngleStr + "|" + rePlainStr)

type OrgLinkSource string

var _ TextLinkSource = (*OrgLinkSource)(nil)

func (s OrgLinkSource) Name() string {
	return string(s)
}

func (_ OrgLinkSource) Flag() string {
	return "org"
}

func (_ OrgLinkSource) Clone(src string) TextLinkSource {
	v := OrgLinkSource(src)
	return &v
}

type Heading struct {
	LineNo  int    `json:"lineNo"`
	RawText string `json:"rawText"`
}

var _ TreeNodeProps = (*Heading)(nil)

func (_ *Heading) BurlType() string {
	return "Heading"
}

func OrgLinkMatchIsUrl(match []string) *Link {
	if len(match[1]) > 0 {
		if reScheme.MatchString(match[1]) {
			return &Link{match[1], match[2], 0}
		} else {
			return nil
		}
	} else if len(match[3]) > 0 {
		return &Link{match[3] + ":" + match[4], "", 0}
	} else if len(match[5]) > 0 {
		return &Link{match[5] + ":" + match[6], "", 0}
	}
	log.Printf("burl_links.OrgLinkMatchIsUrl: something wrong '%v'", match[0])
	return nil
}

func (_ OrgLinkSource) Extract(file io.Reader, filter Filter) (*TreeChildrenNode, error) {
	headings := make([]*Heading, 0, 10)
	tree := NewTreeChildrenNode(nil)
	treeNodes := make([]*TreeChildrenNode, 1, cap(headings)+1)
	treeNodes[0] = &tree
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	lineNo := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNo++
		if matchHeading := reHeading.FindStringSubmatch(line); matchHeading != nil {
			level := len(matchHeading[1])
			if level > cap(headings) {
				level = cap(headings) - 1
			}
			headings = headings[0 : level-1]
			if len(treeNodes) >= level {
				treeNodes = treeNodes[0:level]
			}
			for i := len(headings); i < level-1; i++ {
				headings = append(headings, nil)
			}
			h := &Heading{lineNo, matchHeading[2]}
			headings = append(headings, h)
		}
		if matchArray := reLink.FindAllStringSubmatch(line, -1); matchArray != nil {
			for _, match := range matchArray {
				if link := OrgLinkMatchIsUrl(match); link != nil &&
					(filter == nil || filter(link)) {
					for i := len(treeNodes) - 1; i < len(headings); i++ {
						newNode := NewTreeChildrenNode(headings[i])
						treeNodes = append(treeNodes, &newNode)
						treeNodes[i].AddChild(&newNode)
					}
					tip := treeNodes[len(treeNodes)-1]
					link.LineNo = lineNo
					tip.AddLink(link)
				}
			}
		}
	}
	return &tree, scanner.Err()
}
