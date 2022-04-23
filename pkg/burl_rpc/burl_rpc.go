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

package burl_rpc

type SearchQuery struct {
	Query     string `json:"q"`
	Limit     *int   `json:"limit,omitempty"`
	Tolerance *int   `json:"tol,omitempty"`
}

type ReplyRecord struct {
	Url   string `json:"url"`
	Title string `json:"title,omitempty"`
}

type UrlMentionsQuery struct {
	Variants []string            `json:"variants"`
	Options  *UrlMentionsOptions `json:"options,omitempty"`
}

type UrlMentionsOptions struct {
	CountLimit int `json:"countLimit"`
}

type Location struct {
	FilePath string `json:"file"`
	LineNo   int    `json:"lineNo"`
}

type HelloFormat struct {
	Format  string                 `json:"format"`
	Version string                 `json:"version"`
	Options map[string]interface{} `json:"options,omitempty"` // string, bool, nested options
}

type HelloQuery struct {
	Formats []HelloFormat `json:"formats"`
	Version string        `json:"version"`
}

type HelloResponse struct {
	Format       string      `json:"format"`
	Version      string      `json:"version"`
	Capabilities []string    `json:"capabilities"`
	Options      interface{} `json:"options,omitempty"`
}

type CaptureData struct {
	Url   string      `json:"url"`
	Title string      `json:"title"`
	Body  interface{} `json:"body"`
}

type CaptureQuery struct {
	Format  string      `json:"format"`
	Version string      `json:"version"`
	Options interface{} `json:"options,omitempty"`
	Data    CaptureData `json:"data"`
	Error   interface{} `json:"error"`
}

type CaptureResponse struct {
	Preview bool   `json:"preview"`
	Status  string `json:"status"`
}

var LinkSetPrefixCountLimit = 16

type LinkSetQuery struct {
	Prefix []string `json:"prefix"`
}

type LinkSetResponse struct {
	Urls []string `json:"urls"`
}
