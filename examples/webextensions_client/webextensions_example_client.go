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
	"flag"
	"fmt"
	"os"

	"github.com/maxnikulin/burl/pkg/webextensions"
)

func mainWithGracefulShutdown() (err error) {
	flag.Parse()
	client, err := webextensions.NewClientCommand(flag.Args())
	if err != nil {
		return err
	}

	defer func() {
		err = client.Close()
	}()

	sqrt := func(query float64) {
		var result float64
		errCall := client.Call("example.Sqrt", &query, &result)
		if errCall != nil {
			fmt.Println("sqrt: query", query, "error", errCall)
		} else {
			fmt.Println("sqrt: query", query, "result", result)
		}
	}

	sqrt(2)
	sqrt(-2)

	return
}

func main() {
	err := mainWithGracefulShutdown()
	if err == nil {
		return
	}
	name := os.Args[0]
	if err == webextensions.ErrNoCommand {
		fmt.Fprintf(os.Stderr, "Usage: %s BACKEND_COMMAND ARG...\n", name)
	} else {
		fmt.Fprintf(os.Stderr, "%s: backend error: %s\n", name, err)
	}
	os.Exit(1)
}
