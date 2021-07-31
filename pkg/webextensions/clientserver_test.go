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
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"testing"
	"time"

	"github.com/maxnikulin/burl/pkg/webextensions"
)

type TestBackend struct{}

func (b *TestBackend) Twice(s string, reply *string) error {
	*reply = s + "-" + s
	return nil
}

func makeServer() (stdin io.ReadCloser, stdout io.WriteCloser, done <-chan int) {
	server := rpc.NewServer()
	server.RegisterName("test", new(TestBackend))
	stdin, serverStdout := io.Pipe()
	serverStdin, stdout := io.Pipe()
	finish := make(chan int)
	done = finish
	go func() {
		server.ServeCodec(webextensions.NewServerCodecSplit(
			serverStdin, serverStdout, jsonrpc.NewServerCodec,
		))
		stdin.Close()
		stdout.Close()
		close(finish)
	}()
	return
}

func makeClient(stdin io.ReadCloser, stdout io.WriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(webextensions.NewClientCodecSplit(
		stdin, stdout, jsonrpc.NewClientCodec,
	))
}

func TestSingleQuery(t *testing.T) {
	stdin, stdout, done := makeServer()
	client := makeClient(stdin, stdout)
	var reply string
	if err := client.Call("test.Twice", "Baden", &reply); err != nil {
		t.Errorf("error: %v", err)
	} else if reply != "Baden-Baden" {
		t.Errorf("expected: Baden-Baden, got %q", reply)
	}
	client.Close()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Errorf("server should finish")
	}
}

type caseTwice struct{ query, result string }

var casesTwice = []caseTwice{
	{"Baden", "Baden-Baden"},
	{"Go", "Go-Go"},
	{"go-go", "go-go-go-go"},
	{"Белый", "Белый-Белый"},
}

func TestSerialQueries(t *testing.T) {
	stdin, stdout, done := makeServer()
	client := makeClient(stdin, stdout)
	for i, test := range casesTwice {
		t.Run(fmt.Sprintf("query[%d]=%s", i, test.query), func(t *testing.T) {
			var reply string
			if err := client.Call("test.Twice", test.query, &reply); err != nil {
				t.Errorf("error: %v", err)
			} else if reply != test.result {
				t.Errorf("expected: %q, got %q", test.result, reply)
			}
		})
	}
	client.Close()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Errorf("server should finish")
	}
}

func TestParallelQueries(t *testing.T) {
	stdin, stdout, done := makeServer()
	client := makeClient(stdin, stdout)
	t.Run("parallel", func(t *testing.T) {
		for i, test := range casesTwice {
			i := i
			test := test
			t.Run(fmt.Sprintf("query[%d]=%s", i, test.query), func(t *testing.T) {
				t.Parallel()
				var reply string
				call := <-client.Go("test.Twice", test.query, &reply, nil).Done
				if call.Error != nil {
					t.Errorf("error: %v", call.Error)
				} else if reply != test.result {
					t.Errorf("expected: %q, got %q", test.result, reply)
				}
			})
		}
	})
	client.Close()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Errorf("server should finish")
	}
}
