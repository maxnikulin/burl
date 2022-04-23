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

type UrlRecord struct {
	Url string
}

type TxtLinkSource string

var _ TextLinkSource = (*TxtLinkSource)(nil)

func (s TxtLinkSource) Name() string {
	return string(s)
}

func (_ TxtLinkSource) Flag() string {
	return "txt"
}

func (_ TxtLinkSource) Clone(src string) TextLinkSource {
	v := TxtLinkSource(src)
	return &v
}

func (_ TxtLinkSource) Extract(file io.Reader, filter Filter) (*TreeChildrenNode, error) {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	lineNo := 0
	tree := NewTreeChildrenNode(nil)
	re := regexp.MustCompile(UrlPatternFull)
	for scanner.Scan() {
		line := scanner.Bytes()
		lineNo++
		for len(line) != 0 {
			advance, token, err := bufio.ScanWords(line, true)
			if err != nil && err != bufio.ErrFinalToken {
				return &tree, err
			}
			if matchArray := re.FindAllStringSubmatch(string(token), -1); matchArray != nil {
				for _, match := range matchArray {
					if MatchIsUrl(match) {
						link := &Link{match[0], "", lineNo}
						if filter != nil && !filter(link) {
							continue
						}
						tree.AddLink(link)
					}
				}
			}
			if err == bufio.ErrFinalToken {
				break
			}
			line = line[advance:]
		}
	}
	return &tree, scanner.Err()
}

// cb return value is likely useless. It was conceived to break iterations earlier,
// but it is necessary to check all items to get best match.
func ExtractUrls(cb func(match []string) bool, file io.Reader) error {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	re := regexp.MustCompile(UrlPatternFull)

	for scanner.Scan() {
		word := scanner.Text()
		if matchArray := re.FindAllStringSubmatch(word, -1); matchArray != nil {
			for _, match := range matchArray {
				if MatchIsUrl(match) {
					if !cb(match) {
						break
					}
				}
			}
		}
	}
	return scanner.Err()
}

// TODO: replace Org regexp to something more general
func (_ TxtLinkSource) ExtractSet(file io.Reader, filters []string, result *map[string]bool) error {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	base, err := MakeLinkSetBase(filters)
	if err != nil {
		return err
	}
	rePlain := "\\b(" + base + rePlainSuffixStr + ")"
	reSetLink, err := regexp.Compile(rePlain)
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
				log.Printf("burl_links.TxtLinkSource.ExtractSet: internal error '%v'", match[0])
			}
		}
	}
	return scanner.Err()
}
