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

package burl_emacs

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var Command string = "emacsclient"
var Args = []string{
	"--quiet",
	// Attempt to discriminate "can't find socket" from other errors
	// and to suppress excessively verbose error message.
	// --quiet and --suppress-output does not prevent the following
	// message if emacs server is not running:
	//
	//    emacsclient: No socket or alternate editor.  Please use:
	//
	//            --socket-name
	//            --server-file      (or environment variable EMACS_SERVER_FILE)
	//            --alternate-editor (or environment variable ALTERNATE_EDITOR)
	//
	// The trick with --alternate-editor allows to get minimal message
	//
	//     emacsclient: can't find socket; have you started the server?
	//     To start the server in Emacs, type "M-x server-start".
	`--alternate-editor=sh -c "exit 9"`,
}
var CheckOrgProtocolLisp = "(and (memq 'org-protocol features) 'org-protocol)"
var EnsureFrameLisp = `
(if (and (symbolp 'linkremark-ensure-frame) (fboundp 'linkremark-ensure-frame))
    (linkremark-ensure-frame)
  (or (memq 'x (mapcar #'framep (frame-list)))
      (select-frame (make-frame '((name . "bURL") (window-system . x))))))
`

var EmacsServerNotFoundError = errors.New("Emacs server is not running, please, start it")

func execEmacs(args... string) ([]byte, error) {
	cmd := exec.Command(Command, append(Args, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if cmd.ProcessState.ExitCode() == 9 {
			log.Println("burl_emacs.exec", string(out))
			return []byte{}, EmacsServerNotFoundError
		} else {
			return out, err
		}
	}
	return out, nil
}

func EnsureFrame() error {
	if out, err := execEmacs("--eval", EnsureFrameLisp); err != nil {
		return fmt.Errorf("%s --eval '(linkremark-ensure-frame)': %w: %s",
			Command, err, out)
	}
	return nil
}

func CheckOrgProtocol() error {
	out, err := execEmacs("--eval", CheckOrgProtocolLisp)
	if err != nil {
		return fmt.Errorf("%s check org-protocol: %w: %s",
			Command, err, out)
	}
	if !bytes.HasPrefix(out, []byte("org-protocol")) {
		return errors.New("org-protocol is not loaded")
	}
	return nil
}

func VisitFile(path string, lineNo int) error {
	if err := EnsureFrame(); err != nil {
		return err
	}
	args := []string{"--no-wait"}
	if lineNo > 0 {
		args = append(args, fmt.Sprintf("+%d", lineNo))
	}
	if out, err := execEmacs(args...); err != nil {
		return fmt.Errorf("%s: %w: %s", Command, err, out)
	}
	return nil
}

func OrgProtocol(uri string) error {
	if !strings.HasPrefix(uri, "org-protocol:") {
		return errors.New("scheme of URI is not \"org-protocol\"")
	}
	if err := CheckOrgProtocol(); err != nil {
		return err
	}
	if err := EnsureFrame(); err != nil {
		return err
	}
	args := Args
	args = append(args, uri)
	cmd := exec.Command(Command, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w: %s", Command, err, out)
	}
	return nil
}
