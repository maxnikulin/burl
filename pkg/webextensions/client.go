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
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
)

type clientCodec struct {
	rpc.ClientCodec
	Framer FrameReadWriteCloser
}

var _ rpc.ClientCodec = (*clientCodec)(nil)

func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	if err := c.ClientCodec.WriteRequest(r, param); err != nil {
		return err
	}
	return c.Framer.WriteFrame()
}

func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	if err := c.Framer.ReadHeader(); err != nil {
		return err
	}
	return c.ClientCodec.ReadResponseHeader(r)
}

// The purpose of client codec is usage in test applications
// that allows to debug backend without a browser. RPC client created
// with such codec is a usual RPC client.
func NewClientCodecSplit(reader io.ReadCloser, writer io.WriteCloser,
	parentFactory func(io.ReadWriteCloser) rpc.ClientCodec,
) rpc.ClientCodec {
	framer := NewSplitFrameReadWriteCloser(reader, writer)
	codec := parentFactory(framer)
	return &clientCodec{codec, framer}
}

type ClientCommand struct {
	*rpc.Client
	Command *exec.Cmd
}

func (c *ClientCommand) Close() error {
	errClose := c.Client.Close()
	errWait := c.Command.Wait()
	if errClose != nil {
		return errClose
	}
	return errWait
}

var ErrNoCommand = errors.New("Command is not specified")

func NewClientCommand(args []string) (client *ClientCommand, err error) {
	if len(args) < 1 {
		err = ErrNoCommand
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmdStdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	cmdStdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
	}
	rpcClient := rpc.NewClientWithCodec(NewClientCodecSplit(cmdStdout, cmdStdin, jsonrpc.NewClientCodec))
	client = &ClientCommand{Client: rpcClient, Command: cmd}
	return
}
