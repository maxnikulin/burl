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
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"

	"github.com/maxnikulin/burl/pkg/webextensions"
)

func ExampleNewServerCodecSplit() {
	rpc.ServeCodec(webextensions.NewServerCodecSplit(os.Stdin, os.Stdout, jsonrpc.NewServerCodec))
}

func ExampleNewClientCodecSplit() {
	runClient := func() error {
		cmd := exec.Command("we_example", "--option", "argument")
		cmd.Stderr = os.Stderr
		cmdStdin, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("piping command stdin: %w", err)
		}
		cmdStdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("piping command stdout: %w", err)
		}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("run backend: %w", err)
		}
		rpcClient := rpc.NewClientWithCodec(webextensions.NewClientCodecSplit(cmdStdout, cmdStdin, jsonrpc.NewClientCodec))
		query := "ping"
		var response string
		if err := rpcClient.Call("Backend.Ping", &query, &response); err != nil {
			return err
		} else {
			fmt.Printf("response: %s\n", response)
		}
		if err := rpcClient.Close(); err != nil {
			return fmt.Errorf("client close: %w", err)
		}
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("exec wait: %w", err)
		}
		return nil
	}
	runIt := false // compile-only example
	if runIt {
		if err := runClient(); err != nil {
			fmt.Printf("error: %s", err)
		}
	}
	// Output:
}
