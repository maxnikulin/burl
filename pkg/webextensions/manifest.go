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

package webextensions

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Due to excessive number of fields, there is no point in New function.
// Call Init to initialize Type field and validate.
type Manifest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	// Chrome
	AllowedOrigins []string `json:"allowed_origns,omitempty"`
	// Firefox
	AllowedExtensions []string `json:"allowed_extensions,omitempty"`
	ManifestPath      string   `json:"-"`
}

func (m *Manifest) Init() error {
	// "allowed_origins" field causes a error in Firefox.
	if (m.AllowedOrigins == nil || len(m.AllowedOrigins) == 0) ==
		(m.AllowedExtensions == nil || len(m.AllowedExtensions) == 0) {
		return errors.New("\"allowed_origins\" or \"allowed_extensions\" must be specified but not both.")
	}
	if m.Name == "" {
		return errors.New("\"name\" is not specified")
	}
	if m.Path == "" || !filepath.IsAbs(m.Path) {
		return errors.New("\"path\" is not specified")
	}
	if m.ManifestPath == "" {
		return errors.New("Path for manifest file is not specified")
	}
	if m.ManifestPath != "-" {
		base := filepath.Base(m.ManifestPath)
		ext := filepath.Ext(base)
		if strings.ToLower(ext) != ".json" {
			return fmt.Errorf("Extension of manifest file is not .json")
		}
		if base[:len(base)-len(ext)] != m.Name {
			return fmt.Errorf("Path of manifest file and name are inconsistent: %s %s",
				m.ManifestPath, m.Name)
		}
	}
	m.Type = "stdio"
	return nil
}
