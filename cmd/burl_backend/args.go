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
	"flag"
	"fmt"
	"os"

	"github.com/maxnikulin/burl/pkg/burl_emacs"
	"github.com/maxnikulin/burl/pkg/burl_fileutil"
	"github.com/maxnikulin/burl/pkg/burl_links"
	"github.com/maxnikulin/burl/pkg/burl_util"
)

type BurlBackendArgs struct {
	DisableLinkSet bool
	Exe            string
	LogFile        string
	LinkSources    burl_links.MixedSrcTypeSlice
	Scheme         burl_util.MultiStringFlag
	EmacsArgs      burl_util.MultiStringFlag
}

var DefaultLogDestination string = "-"

func AddBackendFlags(flagset *flag.FlagSet) *BurlBackendArgs {
	if flagset == nil {
		flagset = flag.CommandLine
	}
	v := BurlBackendArgs{
		DisableLinkSet: false,
		LogFile:        DefaultLogDestination,
		LinkSources:    make(burl_links.MixedSrcTypeSlice, 0, 4),
		Scheme:         *burl_util.NewMultiStringFlag(&burl_links.SchemeVariants),
		EmacsArgs:      *burl_util.NewMultiStringFlag(&burl_emacs.UserArgs),
	}
	flagset.StringVar(&v.LogFile, "log", DefaultLogDestination,
		"`FILE` name for logging, \"\" to disable looging, \"-\" for stderr")
	flagset.Var(&v.Scheme, "scheme",
		"Add `SCHEME` to pattern for link extraction")
	flagset.BoolVar(&v.DisableLinkSet, "disable-link-set", false,
		"Do not allow linkSet method for extracting of all links by e.g. https: prefix")
	flagset.StringVar(&burl_emacs.Command, "emacsclient", burl_emacs.Command,
		"Use `EXE` command instead of emacsclient")
	flagset.Var(&v.EmacsArgs, "emacsarg", "Add `ARG` to emacsclient")
	burl_links.AddSourceFlags(&v.LinkSources, flagset)
	return &v
}

func (a *BurlBackendArgs) Customized() bool {
	return (a.LogFile != DefaultLogDestination ||
		len(a.LinkSources) != 0 ||
		a.DisableLinkSet ||
		a.Scheme.IsModified() ||
		a.EmacsArgs.IsModified() ||
		burl_emacs.Command != "emacsclient")
}

// Initialize Exe and ensure absolute paths
func (a *BurlBackendArgs) FixPaths() error {
	var err error
	if a.Exe, err = os.Executable(); err != nil {
		return fmt.Errorf("get current executable: %w", err)
	}
	if a.Exe, err = burl_fileutil.RealPath(a.Exe); err != nil {
		return fmt.Errorf("current executable: %w", err)
	}
	if a.LogFile != "" && a.LogFile != "-" {
		a.LogFile, _, err =
			burl_fileutil.AsAbsFileName(a.LogFile, burl_fileutil.AbsFileNameOptions{})
		if err != nil {
			return err
		}
	}
	for i, s := range a.LinkSources {
		if path, err := burl_fileutil.RealPath(s.Name()); err == nil {
			a.LinkSources[i] = s.Clone(path)
		} else {
			return err
		}
	}
	return nil
}

// "$@" (or "$1" and "$2") is not added, so arguments passed by the browser are unavailable
func (a *BurlBackendArgs) AsShellCommand() ([]string, error) {
	retval := make([]string, 0, 2+len(a.LinkSources))
	if logfile, err := burl_fileutil.EscapeShellArg(a.LogFile); err == nil {
		retval = append(retval, "--log", logfile)
	} else {
		return retval, err
	}
	if a.DisableLinkSet {
		retval = append(retval, "--disable-link-set=true")
	}
	if burl_emacs.Command != "emacsclient" {
		escaped, err := burl_fileutil.EscapeShellArg(burl_emacs.Command)
		if err != nil {
			return retval, err
		}
		retval = append(retval, "--emacsclient="+escaped)
	}
	if a.EmacsArgs.IsModified() {
		for _, arg := range a.EmacsArgs.Values() {
			escaped, err := burl_fileutil.EscapeShellArg(arg)
			if err != nil {
				return retval, err
			}
			retval = append(retval, "--emacsarg="+escaped)
		}
	}
	if a.Scheme.IsModified() {
		for _, arg := range a.Scheme.Values() {
			escaped, err := burl_fileutil.EscapeShellArg(arg)
			if err != nil {
				return retval, err
			}
			retval = append(retval, "--scheme="+escaped)
		}
	}
	for _, s := range a.LinkSources {
		value, err := burl_fileutil.EscapeShellArg(s.Name())
		if err != nil {
			return retval, err
		}
		retval = append(retval, "--"+s.Flag()+"="+value)
	}
	return retval, nil
}
