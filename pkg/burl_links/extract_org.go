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
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

var reHeading = regexp.MustCompile(`^(\*+)\s+(\S(?:.*\S)?)?$`)

// "mid": mail messages, absent in default Org configuration, see
// RFC 2392 - Content-ID and Message-ID Uniform Resource Locators
// https://datatracker.ietf.org/doc/html/rfc2392.html
var reSchemeStr = "doi|https?|mid"
var reScheme = regexp.MustCompile("^" + reSchemeStr + ":")
var reBracketStr = "\\[\\[((?:[^\\]\\[]|\\\\(?:\\\\\\\\)*[\\]\\[]|\\\\+[^\\]\\[])+)](?:\\[((?:.|\n)+?)\\])?\\]"
var reAngleSuffixStr = "[^>\n]*(?:\n[ \t]*[^> \t\n][^> \n]*)*"
var reAngleStr = "<(" + reSchemeStr + "):(" + reAngleSuffixStr + ")>"
var rePlainSuffixStr = "(?:[^][ \t\n(\\)<>]|\\((?:[^][ \t\n(\\)<>]|\\([^][ \t\n(\\)<>]*\\))*\\))+(?:[^[:punct:] \t\n]|/|\\((?:[^][ \t\n(\\)<>]|\\([^][ \t\n(\\)<>]*\\))*\\))"
var rePlainStr = "\\b(" + reSchemeStr + "):(" + rePlainSuffixStr + ")"

var reLink = regexp.MustCompile(reBracketStr + "|" + reAngleStr + "|" + rePlainStr)

// To validate linkSet query parameter
var reSetPrefix = regexp.MustCompile("^(?:(?i)[a-z]+(?:[-+a-z0-9]*[a-z0-9])?)(:[^\n]*)?$")

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

func MakeLinkSetBase(filters []string) (string, error) {
	if len(filters) == 0 {
		return "", fmt.Errorf("Empty filter list")
	}
	base := make([]string, 0, len(filters))
	for _, prefix := range filters {
		match := reSetPrefix.FindStringSubmatch(prefix)
		if match == nil {
			return "", fmt.Errorf("Invalid prefix")
		}
		withColon := regexp.QuoteMeta(prefix)
		if len(match[1]) == 0 {
			withColon += ":"
		}
		base = append(base, withColon)
	}
	return "(?:" + strings.Join(base, "|") + ")", nil
}

// TODO: buld filter regexp ones for several files.
func (_ OrgLinkSource) ExtractSet(file io.Reader, filters []string, result *map[string]bool) error {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	base, err := MakeLinkSetBase(filters)
	if err != nil {
		return err
	}
	reAngle := "<(" + base + reAngleSuffixStr + ")>"
	rePlain := "\\b(" + base + rePlainSuffixStr + ")"
	reSetLink, err := regexp.Compile(reBracketStr + "|" + reAngle + "|" + rePlain)
	if err != nil {
		return err
	}
	reBracketPrefix, err := regexp.Compile("^" + base)
	if err != nil {
		return err
	}
	for scanner.Scan() {
		line := scanner.Text()
		matchArray := reSetLink.FindAllStringSubmatch(line, -1)
		if matchArray == nil {
			continue
		}
		for _, match := range matchArray {
			if bracket := match[1]; len(bracket) > 0 {
				if reBracketPrefix.MatchString(bracket) {
					(*result)[bracket] = true
				}
			} else if angle := match[3]; len(angle) > 0 {
				(*result)[angle] = true
			} else if plain := match[4]; len(plain) > 0 {
				(*result)[plain] = true
			} else {
				log.Printf("burl_links.OrgLinkSource.ExtractSet: internal error '%v'", match[0])
			}
		}
	}
	return scanner.Err()
}
