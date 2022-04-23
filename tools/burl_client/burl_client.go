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
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"

	"github.com/maxnikulin/burl/pkg/burl_rpc"
	"github.com/maxnikulin/burl/pkg/webextensions"
)

func Usage() {
	out := flag.CommandLine.Output()
	cmd := flag.CommandLine.Name()
	fmt.Fprintf(out, "Usage: %s BACKEND_LAUNCH_COMMAND... -- hello\n", cmd)
	fmt.Fprintf(out, "   or: %s BACKEND_LAUNCH_COMMAND... -- capture ORG_PROTOCOL_URI\n", cmd)
	fmt.Fprintf(out, "   or: %s BACKEND_LAUNCH_COMMAND... -- mentions URL...\n", cmd)
	fmt.Fprintf(out, "   or: %s BACKEND_LAUNCH_COMMAND... -- visit [--line LINE_NO] --file PATH\n", cmd)
	fmt.Fprintf(out, "   or: %s BACKEND_LAUNCH_COMMAND... -- set PREFIX...\n", cmd)
	fmt.Fprintf(out, "\nExecutes the following backend methods:\n")
	fmt.Fprintf(out, "linkremark.hello, linkremark.capture, linkremark.urlMentions, linkremark.visit\n")
	flag.PrintDefaults()
}

func callUrlMentions(args []string) (string, interface{}, error) {
	rpcMethod := "linkremark.urlMentions"
	query := &burl_rpc.UrlMentionsQuery{Variants: args[1:], Options: nil}
	return rpcMethod, query, nil
}

func callVisit(args []string) (string, interface{}, error) {
	set := flag.NewFlagSet(os.Args[0]+": "+args[0], flag.ContinueOnError)
	lineNo := set.Int("line", 0, "line number")
	path := set.String("file", "", "path to file")
	if err := set.Parse(args[1:]); err != nil {
		return "", nil, err
	}
	query := &burl_rpc.Location{FilePath: *path, LineNo: *lineNo}
	return "linkremark.visit", query, nil
}

func callCapture(args []string) (string, interface{}, error) {
	if len(args) != 2 {
		return "", nil, fmt.Errorf("Command takes exactly one argument")
	}
	query := burl_rpc.CaptureQuery{
		Data:    burl_rpc.CaptureData{Url: args[1]},
		Format:  "org-protocol",
		Version: "0.2",
	}
	return "linkremark.capture", query, nil
}

func callSet(args []string) (string, interface{}, error) {
	query := burl_rpc.LinkSetQuery{
		Prefix: args[1:],
	}
	return "linkremark.linkSet", query, nil
}

func callHello(args []string) (string, interface{}, error) {
	if len(args) > 2 {
		return "", nil, fmt.Errorf("Command takes one optional argument")
	}
	var query interface{}
	var err error
	if len(args) == 2 {
		err = json.Unmarshal([]byte(args[1]), &query)
	} else {
		err = json.Unmarshal([]byte(`{
	"version": "0.2",
	"formats": [
		{ "format": "org", "version": "0.2" },
		{ "format": "object", "version": "0.2" },
		{
			"format": "org-protocol",
			"version": "0.2",
			"options": { "format": "org", "version": "0.2", "clipboardForBody": false }
		}
	]
}`),
			&query)
	}
	if err != nil {
		return "", nil, err
	}
	return "linkremark.hello", query, nil
}

var usageError error = errors.New("incorrect arguments")

func mainWithGracefulShutdown() error {
	flag.Usage = Usage
	flag.Parse()
	separator := flag.NArg()
	for i, arg := range flag.Args() {
		if arg == "--" {
			separator = i
			break
		}
	}
	if flag.NArg() == 0 {
		return fmt.Errorf("%w: backend command is not specified", usageError)
	}
	backendCmd := flag.Args()[0]
	backendArgs := flag.Args()[1:separator]
	if separator >= flag.NArg()-1 {
		return fmt.Errorf("%w: no queries", usageError)
	}
	queries := flag.Args()[separator+1:]
	cmd := exec.Command(backendCmd, backendArgs...)
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
	var reply interface{}
	// rpcQuery := burl_rpc.SearchQuery{Query: query}
	// reply := make([]burl_rpc.ReplyRecord, 0)
	subcommands := map[string]func([]string) (string, interface{}, error){
		"visit":    callVisit,
		"mentions": callUrlMentions,
		"capture":  callCapture,
		"hello":    callHello,
		"set":      callSet,
	}
	subcommandName := flag.Arg(separator + 1)
	sub, ok := subcommands[subcommandName]
	if !ok {
		return fmt.Errorf("%s: subcommand not found", flag.Arg(separator+1))
	}
	method, data, err := sub(queries)
	if err != nil {
		return fmt.Errorf("%s: %w", subcommandName, err)
	}
	if err = rpcClient.Call(method, data, &reply); err == nil {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(reply)
	} else {
		return fmt.Errorf("%s response: %w", subcommandName, err)
	}
	if err := rpcClient.Close(); err != nil {
		return fmt.Errorf("%s client close: %w", subcommandName, err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%s exec wait: %w", subcommandName, err)
	}
	return nil
}

func main() {
	err := mainWithGracefulShutdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", os.Args[0], err)
		if errors.Is(err, usageError) {
			Usage()
		}
		os.Exit(1)
	}
}
