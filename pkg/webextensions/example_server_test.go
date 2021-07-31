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

package webextensions_test

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	"github.com/maxnikulin/burl/pkg/webextensions"
)

type Backend struct{}

func (b *Backend) Ping(query *string, reply *string) error {
	*reply = *query + "pong"
	return nil
}

func Example() {
	rpc.RegisterName("Example", new(Backend))
	rpc.ServeCodec(webextensions.NewServerCodecSplit(os.Stdin, os.Stdout, jsonrpc.NewServerCodec))
}
