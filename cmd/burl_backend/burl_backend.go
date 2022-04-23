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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strings"
	"sync"

	"github.com/maxnikulin/burl/pkg/burl_emacs"
	"github.com/maxnikulin/burl/pkg/burl_fuzzy"
	"github.com/maxnikulin/burl/pkg/burl_links"
	"github.com/maxnikulin/burl/pkg/burl_rpc"
	"github.com/maxnikulin/burl/pkg/burl_url"
	"github.com/maxnikulin/burl/pkg/version"
	"github.com/maxnikulin/burl/pkg/webextensions"
)

func Usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage: %s [-log LOG_FILE] [{-txt TEXT_FILE|-org ORG_FILE}...]\n", os.Args[0])
	fmt.Fprintf(out, "   or: %s [-force] -wrapper SCRIPT_FILE [BACKEND_OPTIONS...]\n", os.Args[0])
	fmt.Fprintf(out, "   or: %s [-force] [-backend NAME] {-manifest-chrome|-manifest-firefox} DIR/[NAME] [WRAPPER_OPTIONS...]\n", os.Args[0])
	fmt.Fprintf(out, "   or: %s {-h|--help|--version}\n", os.Args[0])
	fmt.Fprintf(out, "\nFirst form starts backend.\n")
	fmt.Fprintf(out, "Second form creates a wrapper that could be executed by a browser.\n")
	fmt.Fprintf(out, "Third variant creates a wrapper and a manifest file\n")
	fmt.Fprintf(out, "that allow browser extension to communicate with native application.\n\n")
	flag.PrintDefaults()
}

type generalFlagValues struct {
	Help    bool
	Version bool
	FlagSet *flag.FlagSet
}

func createGeneralFlags(flagset *flag.FlagSet) *generalFlagValues {
	v := generalFlagValues{}
	if flagset == nil {
		flagset = flag.CommandLine
	}
	v.FlagSet = flagset
	flagset.BoolVar(&v.Help, "h", false, "print help message")
	flagset.BoolVar(&v.Help, "help", false, "print help message")
	flagset.BoolVar(&v.Version, "version", false, "print application version and exit")
	flagset.BoolVar(&v.Version, "v", false, "print application version and exit")
	flagset.BoolVar(&v.Version, "V", false, "print application version and exit")
	return &v
}

func processGeneralFlags(v *generalFlagValues) (bool, error) {
	if v.Help {
		v.FlagSet.SetOutput(os.Stdout)
		Usage()
		return true, nil
	}
	if v.Version {
		fmt.Fprintf(os.Stdout,
			`burl_backend %s
Copyright (C) 2021 Max Nikulin
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
`,
			version.GetVersion())
		return true, nil
	}
	return false, nil
}

// JSON-RPC endpoint
type BurlBackend struct {
	srcFiles    []burl_links.TextLinkSource
	fileGroup   *burl_links.TreeChildrenNode
	err         error
	once        sync.Once
	initializer func()
	linkSetImpl func([]burl_links.TextLinkSource, []string, *burl_rpc.LinkSetResponse) error
}

func NewBurlBackendPtr(args *BurlBackendArgs) *BurlBackend {
	backend := BurlBackend{args.LinkSources, nil, nil, sync.Once{}, nil, nil}
	if !args.DisableLinkSet {
		backend.linkSetImpl = LinkSetReal
	}
	backend.initializer = func() {
		if len(backend.srcFiles) == 0 {
			backend.err = errors.New("No files specified for backend")
			return
		}
		backend.fileGroup, backend.err =
			burl_links.ExtractLinksFromFileGroup(backend.srcFiles, nil)
		if backend.err != nil {
			log.Println("Lazy read files:", backend.err)
		} else if backend.fileGroup == nil {
			backend.err = errors.New("No link in source files")
		}
	}
	return &backend
}

func (b *BurlBackend) Hello(query *burl_rpc.HelloQuery, reply *burl_rpc.HelloResponse) error {
	var version string
	format := "org-protocol"
	for _, descr := range query.Formats {
		if descr.Format != format {
			continue
		}
		version = descr.Version
		break
	}
	if version == "" {
		return errors.New("Extension is too old: 'org-protocol' format is not offered")
	}
	reply.Format = format
	reply.Version = version
	reply.Options = map[string]interface{}{"clipboardForBody": false}
	if len(b.srcFiles) > 0 {
		reply.Capabilities = append(reply.Capabilities, "visit", "urlMentions")
		if b.linkSetImpl != nil {
			reply.Capabilities = append(reply.Capabilities, "linkSet")
		}
	}
	return nil
}

func (_ *BurlBackend) Capture(query *burl_rpc.CaptureQuery, reply *burl_rpc.CaptureResponse) error {
	if query.Format != "org-protocol" {
		return errors.New("linkremark.capture: format is not 'org-protocol'")
	}
	if query.Error != nil {
		*reply = burl_rpc.CaptureResponse{Preview: true, Status: "preview"}
		return nil
	}
	url := query.Data.Url
	if err := burl_emacs.OrgProtocol(url); err != nil {
		return fmt.Errorf("linkremark.capture: %w", err)
	}
	*reply = burl_rpc.CaptureResponse{Preview: false, Status: "success"}
	return nil
}

func (b *BurlBackend) UrlMentions(query *burl_rpc.UrlMentionsQuery, reply *burl_links.LimitCountNode) error {
	hasUrl := false
	var variants []string
	for _, u := range query.Variants {
		if u != "" {
			hasUrl = true
			variants = append(variants, burl_url.UrlVariants(u)...)
		}
	}
	if !hasUrl {
		return errors.New("No URLs in the query")
	}
	b.once.Do(b.initializer)
	if b.err != nil {
		return b.err
	}
	filter := func(link *burl_links.Link) bool {
		for _, l := range variants {
			if l == link.URL {
				return true
			}
		}
		return false
	}
	countLimit := 8
	if query.Options != nil {
		countLimit = query.Options.CountLimit
	}
	attrTree := burl_links.CountDescendants(b.fileGroup, filter)
	attrTree = burl_links.FilterChildrenCount(attrTree, countLimit)
	*reply = *attrTree
	return nil
}

func (b *BurlBackend) Visit(query *burl_rpc.Location, result *bool) error {
	fileAllowed := false
	for _, src := range b.srcFiles {
		if query.FilePath == src.Name() {
			fileAllowed = true
			break
		}
	}
	if !fileAllowed {
		return errors.New("Opening of arbitrary file is prohibited")
	}
	if err := burl_emacs.VisitFile(query.FilePath, query.LineNo); err != nil {
		return fmt.Errorf("linkremark.visit: %w", err)
	}
	*result = true
	return nil
}

func (b *BurlBackend) Search(query *burl_rpc.SearchQuery, reply *[]burl_rpc.ReplyRecord) error {
	b.once.Do(b.initializer)
	if b.err != nil {
		return b.err
	}

	source := func(cb func(string) bool) {
		filter := func(link *burl_links.Link) bool {
			cb(link.URL)
			return true
		}
		burl_links.ForEachLink(b.fileGroup, filter)
	}

	if query.Limit != nil {
		if *query.Limit <= 0 || *query.Limit >= 100 {
			// ProcessUrlSource should be rewritten to allow extraction of all urls.
			return errors.New("limit is out of range")
		}
	}
	tolerance := burl_fuzzy.EditorDistance(5)
	if query.Tolerance != nil {
		tolerance = burl_fuzzy.EditorDistance(*query.Tolerance)
	}
	if tolerance < 0 {
		return errors.New("Negative tolerance")
	}

	h := burl_fuzzy.ProcessUrlSource(strings.Fields(query.Query), source, query.Limit, &tolerance)
	*reply = make([]burl_rpc.ReplyRecord, len(h))
	for i, urlResult := range h {
		(*reply)[len(h)-i-1].Url = urlResult.Word
	}
	return nil
}

func LinkSetReal(srcFiles []burl_links.TextLinkSource, prefixes []string, reply *burl_rpc.LinkSetResponse) error {
	urls, err := burl_links.ExtractLinkSetFromFileGroup(srcFiles, prefixes)
	if err != nil {
		return err
	}
	reply.Urls = make([]string, len(urls))
	i := 0
	for key := range urls {
		reply.Urls[i] = key
		i++
	}
	return nil
}

func (b *BurlBackend) LinkSet(query *burl_rpc.LinkSetQuery, reply *burl_rpc.LinkSetResponse) error {
	if b.linkSetImpl == nil {
		return fmt.Errorf("Method is disabled")
	}
	if len(query.Prefix) > burl_rpc.LinkSetPrefixCountLimit {
		log.Printf("prefixes %v\n", query.Prefix)
		return errors.New("Too many prefix variants")
	}
	return b.linkSetImpl(b.srcFiles, query.Prefix, reply)
}

func mainWithGracefulShutdown() error {
	flag.Usage = Usage
	generalFlags := createGeneralFlags(nil)
	backendFlags := AddBackendFlags(nil)
	installFlags := createInstallFlags(nil)
	flag.Parse()
	if info, err := processGeneralFlags(generalFlags); info || err != nil {
		return err
	}
	backendFlags.LinkSources = burl_links.AddSourceArgs(backendFlags.LinkSources, flag.Args())
	if helpers, err := processInstallFlags(installFlags, backendFlags); helpers || err != nil {
		if err != nil {
			return fmt.Errorf("make helpers: %w", err)
		}
		return nil
	}

	switch backendFlags.LogFile {
	case "":
		// https://stackoverflow.com/questions/10571182/go-disable-a-log-logger
		log.SetOutput(ioutil.Discard)
		log.SetFlags(0)
	case "-":
	default:
		logFile, err := os.OpenFile(backendFlags.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("open log: %w", err)
		}
		defer logFile.Sync()
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	backend := NewBurlBackendPtr(backendFlags)
	err := rpc.RegisterName("Burl", backend)
	if err != nil {
		if backendFlags.LogFile != "-" {
			log.Println("register burl endpoint:", err)
		}
		return fmt.Errorf("register RPC: %w", err)
	}
	methodMap := map[string]string{
		"hello":                  "Burl.Hello",
		"linkremark.hello":       "Burl.Hello",
		"burl.hello":             "Burl.Hello",
		"capture":                "Burl.Capture",
		"linkremark.capture":     "Burl.Capture",
		"burl.capture":           "Burl.Capture",
		"burl.visit":             "Burl.Visit",
		"linkremark.visit":       "Burl.Visit",
		"burl.urlMentions":       "Burl.UrlMentions",
		"linkremark.urlMentions": "Burl.UrlMentions",
		"burl.linkSet":           "Burl.LinkSet",
		"linkremark.linkSet":     "Burl.LinkSet",
	}
	rpc.ServeCodec(webextensions.NewServerCodecSplit(
		os.Stdin, os.Stdout,
		webextensions.MappedServerCodecFactory(methodMap, jsonrpc.NewServerCodec)))
	return nil
}

func main() {
	err := mainWithGracefulShutdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
