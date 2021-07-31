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
	"strings"

	"github.com/maxnikulin/burl/pkg/burl_fileutil"
	"github.com/maxnikulin/burl/pkg/version"
	"github.com/maxnikulin/burl/pkg/webextensions"
)

type installFlagValues struct {
	Name            string
	WrapperPath     string
	FirefoxManifest string
	ChromeManifest  string
	Force           bool
}

func createInstallFlags(flagset *flag.FlagSet) *installFlagValues {
	v := installFlagValues{}
	if flagset == nil {
		flagset = flag.CommandLine
	}
	flagset.StringVar(&v.Name, "backend", "",
		"Backend `NAME` to be specified in extension configuration")
	flagset.StringVar(&v.FirefoxManifest, "manifest-firefox", "",
		"Generate native application manifest in `FILE`."+
			" See https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Native_manifests"+
			" and https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Native_messaging#app_manifest"+
			" for expected directory.")
	flagset.StringVar(&v.ChromeManifest, "manifest-chrome", "",
		"Generate native application manifest in `FILE`."+
			" See https://developer.chrome.com/docs/apps/nativeMessaging/#native-messaging-host-location"+
			" for expected directory.")
	flagset.StringVar(&v.WrapperPath, "wrapper", "",
		"Create wrapper `SCRIPT` that serves as configuration file")
	flagset.BoolVar(&v.Force, "force", false,
		"When generating wrapper or manifest, overwrite existing files")
	return &v
}

// Returns true if generation of manifest or wrapper is requested.
func processInstallFlags(
	installOptions *installFlagValues, backendOptions *BurlBackendArgs,
) (retval bool, err error) {
	files := []string{installOptions.WrapperPath, installOptions.FirefoxManifest, installOptions.ChromeManifest}
	nonEmpty := 0
	toStdout := 0
	for _, f := range files {
		if f != "" {
			nonEmpty++
		}
		if f == "-" {
			toStdout++
		}
	}
	if nonEmpty == 0 {
		return false, nil
	}
	retval = true
	if toStdout > 1 {
		return retval, errors.New("Stdout is requested for more than one file")
	}
	if installOptions.WrapperPath == "-" &&
		(installOptions.FirefoxManifest != "" ||
			installOptions.ChromeManifest != "") {
		return retval, errors.New("Manifest requires wrapper path but stdout is requested")
	}
	custom := backendOptions.Customized()
	if installOptions.WrapperPath == "" && custom {
		return retval, errors.New("-wrapper PATH is required due to custom options")
	}
	if err := backendOptions.FixPaths(); err != nil {
		return retval, fmt.Errorf("make cwd independent args: %w", err)
	}
	manifestExe := ""
	if installOptions.WrapperPath != "" && installOptions.WrapperPath != "-" {
		if path, _, err := burl_fileutil.AsAbsFileName(
			installOptions.WrapperPath, burl_fileutil.AbsFileNameOptions{},
		); err != nil {
			return retval, fmt.Errorf("%w: failed to get absolute wrapper path", err)
		} else {
			manifestExe = path
			installOptions.WrapperPath = path
		}
	}
	if installOptions.WrapperPath != "" || custom {
		// Another way is to use os.SameFile however it may be more
		// reasonable to overwrite a hardlink.
		if installOptions.WrapperPath != backendOptions.Exe {
			escaped_args, err := backendOptions.AsShellCommand()
			if err != nil {
				return retval, err
			}
			content := fmt.Sprintf("#!/bin/sh -eu\nexec %s\n", strings.Join(escaped_args, " "))
			err = burl_fileutil.WriteFile(installOptions.WrapperPath, []byte(content), 0755, installOptions.Force)
			if err != nil {
				return retval, fmt.Errorf("%w: failed to write wrapper", err)
			}
		} else if custom {
			return retval, fmt.Errorf("-wrapper file must be distinct from backend path due to custom options")
		}
	}
	if manifestExe == "" {
		manifestExe = backendOptions.Exe
	}
	if installOptions.FirefoxManifest != "" {
		absNameParams := burl_fileutil.AbsFileNameOptions{
			Name:        installOptions.Name,
			DefaultName: version.AppName,
			Ext:         ".json",
		}
		path, name, err := burl_fileutil.AsAbsFileName(installOptions.FirefoxManifest, absNameParams)
		if err != nil {
			return retval, fmt.Errorf("Firefox manifest path: %w", err)
		}
		m := webextensions.Manifest{
			Name:              name,
			Path:              manifestExe,
			ManifestPath:      path,
			AllowedExtensions: []string{version.FirefoxExtension},
		}
		if err := writeManifest(&m, installOptions.Force); err != nil {
			return retval, fmt.Errorf("write Firefox manifest: %w", err)
		}
	}
	if installOptions.ChromeManifest != "" {
		absNameParams := burl_fileutil.AbsFileNameOptions{
			Name:        installOptions.Name,
			DefaultName: version.AppName,
			Ext:         ".json",
		}
		path, name, err := burl_fileutil.AsAbsFileName(installOptions.ChromeManifest, absNameParams)
		if err != nil {
			return retval, fmt.Errorf("Chrome manifest path: %w", err)
		}
		m := webextensions.Manifest{
			Name:         name,
			Path:         manifestExe,
			ManifestPath: path,
			AllowedOrigins: []string{
				fmt.Sprintf("chrome-extension://%s/", version.ChromeExtension),
			},
		}
		if err := writeManifest(&m, installOptions.Force); err != nil {
			return retval, fmt.Errorf("write Chrome manifest: %w", err)
		}
	}
	return retval, nil
}

func writeManifest(m *webextensions.Manifest, force bool) error {
	m.Description = version.Description
	if err := m.Init(); err != nil {
		return err
	}
	content, err := json.MarshalIndent(m, "", "  ")
	content = append(content, '\n')
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	err = burl_fileutil.WriteFile(m.ManifestPath, content, 0644, force)
	if err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	return nil
}
