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
	"os/exec"
	"strings"
)

var Command string = "emacsclient"
var Args = []string{
	"--quiet",
	"--alternate-editor",
	`sh -c "echo 'bURL: Emacs server is not running'; exit 9" burl-report-no-emacs-server`,
}
var CheckOrgProtocolLisp = "(and (memq 'org-protocol features) 'org-protocol)"
var EnsureFrameLisp = `
(if (and (symbolp 'linkremark-ensure-frame) (fboundp 'linkremark-ensure-frame))
    (linkremark-ensure-frame)
  (or (memq 'x (mapcar #'framep (frame-list)))
      (select-frame (make-frame '((name . "bURL") (window-system . x))))))
`

func EnsureFrame() error {
	args := append(Args, "--eval", EnsureFrameLisp)
	cmd := exec.Command(Command, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s --eval '(linkremark-ensure-frame)': %w: %s",
			Command, err, out)
	}
	return nil
}

func CheckOrgProtocol() error {
	args := append(Args, "--eval", CheckOrgProtocolLisp)
	cmd := exec.Command(Command, args...)
	out, err := cmd.CombinedOutput()
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
	args := Args
	args = append(args, "--no-wait")
	if lineNo > 0 {
		args = append(args, fmt.Sprintf("+%d", lineNo))
	}
	args = append(args, path)
	cmd := exec.Command(Command, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
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
