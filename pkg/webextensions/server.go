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

/*
	WebExtensions native messaging support based on net/rpc and
	net/rpc/jsonrpc facilities from the standard library.

	Most of the types and functions can be wiped from the public interface
	but such building blocks might be reused for similar protocols.
*/
package webextensions

import (
	"io"
	"net/rpc"
)

type serverCodec struct {
	rpc.ServerCodec
	Framer FrameReadWriteCloser
}

var _ rpc.ServerCodec = (*serverCodec)(nil)

func (c *serverCodec) ReadRequestHeader(r *rpc.Request) error {
	if err := c.Framer.ReadHeader(); err != nil {
		return err
	}
	return c.ServerCodec.ReadRequestHeader(r)
}

func (c *serverCodec) WriteResponse(r *rpc.Response, x interface{}) error {
	if err := c.ServerCodec.WriteResponse(r, x); err != nil {
		return err
	}
	if err := c.Framer.WriteFrame(); err != nil {
		return err
	}
	return nil
}

type CodecFactory func(io.ReadWriteCloser) rpc.ServerCodec

// The primary function in this library. Wrap server codec created
// by parentFactory (e.g. jsonrpc.NewServerCodec) for usage with split input
// and output streams and to handle preceding packet size.
func NewServerCodecSplit(reader io.ReadCloser, writer io.WriteCloser,
	parentFactory CodecFactory,
) rpc.ServerCodec {
	framer := NewSplitFrameReadWriteCloser(reader, writer)
	codec := parentFactory(framer)
	return &serverCodec{codec, framer}
}

// Mapping for arbitrary names of RPC methods.
// To overcome limitations that method names must start with a capital letter
// otherwise they are not exported.
type MappedServerCodec struct {
	rpc.ServerCodec
	m map[string]string
}

func (c *MappedServerCodec) ReadRequestHeader(r *rpc.Request) error {
	if err := c.ServerCodec.ReadRequestHeader(r); err != nil {
		return err
	}
	name, ok := c.m[r.ServiceMethod]
	if !ok {
		// Do not return error here since it makes failure reason
		// rather obscure at the client side.
		name = "unknown." + r.ServiceMethod
	}
	r.ServiceMethod = name
	return nil
}

// Usage:
//     methodMap := map[string]string{ "hello": "Addon.Hello" }
//     f := webextensions.MappedServerCodecFactory(methodMap, jsonrpc.NewServerCodec)))
// Result can be passed to NewServerCodecSplit.
func MappedServerCodecFactory(methodMap map[string]string, parentFactory CodecFactory) CodecFactory {
	return func(reader io.ReadWriteCloser) rpc.ServerCodec {
		return &MappedServerCodec{parentFactory(reader), methodMap}
	}
}
