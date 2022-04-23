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

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/maxnikulin/burl/pkg/burl_links"
	"github.com/maxnikulin/burl/pkg/burl_url"
)

type MultiStringFlag []string

func (MultiStringFlag) String() string {
	return "FIXME: this is a MultiStringFlag proxy, value should be accessed directly"
}

func (f *MultiStringFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func Usage() {
	out := flag.CommandLine.Output()
	cmd := flag.CommandLine.Name()
	fmt.Fprintf(out, "Usage: %s [-org ORG_FILE]... URL ORG_FILE...\n", cmd)
	fmt.Fprintf(out, "   or: %s [-org ORG_FILE]... (-set PREFIX)... ORG_FILE...\n", cmd)
	fmt.Fprintf(out, "\nFilter heading with links to URL. An iternal test utility for bURL.\n")
	fmt.Fprintf(out, "Currently result is reported as JSON but it may be changed soon.\n")
}

var usageError error = errors.New("incorrect arguments")

func mainWithGracefulShutdown() error {
	countLimit := 8
	linkSources := make(burl_links.MixedSrcTypeSlice, 0, 4)
	burl_links.AddSourceFlags(&linkSources, nil)
	set := make(MultiStringFlag, 0, 4)
	flag.Var(&set, "set", "Link prefix to extract set of links")
	flag.Usage = Usage
	flag.Parse()
	if flag.NArg() < 1 {
		return fmt.Errorf("%w: no url to query", usageError)
	}

	nextArg := 0
	variants := make([]string, 0, 8)
	if len(set) == 0 {
		burl_url.UrlVariants(flag.Arg(0))
		nextArg = 1
	}
	queryIsEmpty := len(variants) == 0

	if flag.NArg() == nextArg && len(linkSources) == 0 {
		linkSources = append(linkSources, burl_links.TxtLinkSource("-"))
	}
	linkSources = burl_links.AddSourceArgs(linkSources, flag.Args()[nextArg:])

	if len(set) > 0 {
		result, err := burl_links.ExtractLinkSetFromFileGroup(linkSources, set)
		if err != nil {
			return err
		}
		for key, _ := range result {
			fmt.Println(key)
		}
		return nil
	}

	filterExact := func(link *burl_links.Link) bool {
		if queryIsEmpty {
			return true
		}
		for _, l := range variants {
			if l == link.URL {
				return true
			}
		}
		return false
	}

	// f,_ := os.Create("cpu.prof")
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	group, err := burl_links.ExtractLinksFromFileGroup(linkSources, filterExact)
	if group != nil {
		attrTree := burl_links.CountDescendants(group, filterExact)
		attrTree = burl_links.FilterChildrenCount(attrTree, countLimit)
		if attrTree != nil {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(attrTree)
			// enc.Encode(group)
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := mainWithGracefulShutdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		if errors.Is(err, usageError) {
			Usage()
		}
		os.Exit(1)
	}
}
