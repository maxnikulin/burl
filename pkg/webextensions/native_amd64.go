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

import "encoding/binary"

// Conditional compilation hack to bypass attempts of Go authors to make
// world better by prohibiting usage of native encoding. Real world spec
// originating from Chrome developers states:
//   "Each message ... is preceded with a 32-bit value containing
//    the message length in native byte order."
//
// Portable alternative https://github.com/koneu/natend relies on unsafe.
// Refused request for a similar feature in the "binary/encoding" package:
// https://groups.google.com/d/topic/golang-nuts/3GEzwKfRRQw
var NativeEndian = binary.LittleEndian
