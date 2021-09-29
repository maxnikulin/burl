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

package version

var (
	version = "v0.3" // To be overridden at the build time
)

var Description string = "Burl - LinkRemark interface to Emacs"
var ChromeExtension = "mgmcoaemjnaehlliifkgljdnbpedihoe"
var FirefoxExtension = "linkremark@maxnikulin.github.io"
var AppName string = "io.github.maxnikulin.burl"

// Version usually specified at the build time
func GetVersion() string {
	return version
}
