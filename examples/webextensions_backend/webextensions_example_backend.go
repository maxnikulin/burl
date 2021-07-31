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
	"errors"
	"fmt"
	"math"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"time"

	"github.com/maxnikulin/burl/pkg/webextensions"
)

var ErrSqrtOfNegative = errors.New("square root of negative number")

type ExampleBackend struct{}

func (_ *ExampleBackend) Sqrt(x, result *float64) error {
	if *x < 0 {
		return ErrSqrtOfNegative
	}
	*result = math.Sqrt(*x)
	return nil
}

func (_ *ExampleBackend) Sleep(t, result *int) error {
	*result = *t
	if *t > 0 {
		time.Sleep(time.Duration(*t) * time.Millisecond)
	}
	return nil
}

func mainWithGracefulShutdown() error {
	rpc.RegisterName("example", &ExampleBackend{})
	rpc.ServeCodec(webextensions.NewServerCodecSplit(os.Stdin, os.Stdout, jsonrpc.NewServerCodec))
	return nil
}

func main() {
	if err := mainWithGracefulShutdown(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
		os.Exit(1)
	}
}
