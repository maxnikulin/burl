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

package burl_fileutil

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Argument of AsAbsFileName()
type AbsFileNameOptions struct {
	Name        string
	DefaultName string
	Ext         string
}

func RealPath(path string) (string, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return abspath, fmt.Errorf("realpath: to absolute: %w", err)
	}
	realpath, err := filepath.EvalSymlinks(abspath)
	if err != nil {
		return realpath, fmt.Errorf("realpath: expand symlinks: %w", err)
	}
	return realpath, nil
}

// Check file content or create file
//
// Umask has no effect on mode since ioutil.TempFile has no mode argument.
func WriteFile(path string, content []byte, mode os.FileMode, force bool) error {
	if path == "-" {
		_, err := os.Stdout.Write(content)
		return err
	}
	stat, err := os.Stat(path)
	if !errors.Is(err, os.ErrNotExist) {
		bufLen := 4096
		if len(content) > bufLen {
			bufLen += len(content)
		}
		data := make([]byte, bufLen)
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("%w: unable to check contents of existing helper", err)
		}
		defer file.Close()
		count, err := file.Read(data)
		if err != nil {
			return fmt.Errorf("%w: failed to read existing file", err)
		}
		if !force {
			if count == cap(data) {
				return fmt.Errorf("%s: helper file is too large, assume wrong path", path)
			}
			shebang := []byte("#!/")
			if bytes.HasPrefix(content, shebang) && !bytes.HasPrefix(data, shebang) {
				return fmt.Errorf("%s: no shebang in existing file. assume wrong path", path)
			}
		}
		if bytes.Equal(data[:count], content) {
			perm := stat.Mode().Perm()
			if (mode&0500) != 0 && (perm&0500) == 0 {
				if err := os.Chmod(path, perm|0500); err != nil {
					return fmt.Errorf("set exec permission: %w", err)
				}
			}
			return nil
		}
		if !force {
			return &os.PathError{Op: "Check contents", Path: path, Err: os.ErrExist}
		}
	}
	f, err := ioutil.TempFile(filepath.Dir(path), filepath.Base(path))
	defer f.Close()
	if err != nil {
		return err
	}
	if err := f.Chmod(mode); err != nil {
		os.Remove(f.Name()) // unlinkat(2) ?
		return err
	}
	if _, err := f.Write(content); err != nil {
		os.Remove(f.Name()) // unlinkat(2) ?
		return err
	}
	// renameat(2) ?
	if err := os.Rename(f.Name(), path); err != nil {
		os.Remove(f.Name()) // unlinkat(2) ?
		return err
	}
	return nil
}

// Convert filePath to absolute and check that it is either existing file
// or a file in existing directory. If directory name is passed,
// append either options.Name (if non-empty) or options.DefaultName and options.Ext.
// If file name is passed, check that its suffix is options.Ext and,
// if  options.Name is set, its base name is equal to it.
// If filePath is "-", return it back.
//
// Suitable to ensure e.g. that log file is requested in an existing directory.
// Another use case is that Native Application Manifest for browser must
// have ".json" extension and its base name must be the same as "name" field inside.
// If options.Name is passed, it is enforced, otherwise name is either derived
// from file part of path or options.DefaultName is used.
func AsAbsFileName(filePath string, options AbsFileNameOptions) (path string, name string, err error) {
	path = filePath
	if options.Name != "" {
		name = options.Name
	} else {
		name = options.DefaultName
	}
	if path == "-" {
		return
	}
	if path, err = filepath.Abs(path); err != nil {
		err = fmt.Errorf("%w: unable to derive absolute path", err)
		return
	}
	nameFromPath := func(file string) (string, error) {
		base := filepath.Base(file)
		if options.Name != "" && base != options.Name+options.Ext {
			return options.Name, fmt.Errorf("File name \"%s\" inconsistent with name \"%s\"",
				file, options.Name)
		}
		if options.Ext == "" {
			return base, nil
		}
		ext := filepath.Ext(base)
		if strings.ToLower(ext) != options.Ext {
			return base, fmt.Errorf("File name \"%s\" must have \"%s\" extension",
				path, options.Ext)
		}
		return base[:len(base)-len(ext)], nil
	}
	var stat os.FileInfo
	if stat, err = os.Stat(path); err == nil {
		if mode := stat.Mode(); mode.IsDir() {
			if name != "" {
				path = filepath.Join(path, name+options.Ext)
			} else {
				err = fmt.Errorf("%s: regular file required, directory specified", path)
			}
		} else {
			name, err = nameFromPath(path)
		}
		return
	} else if !errors.Is(err, os.ErrNotExist) {
		return
	}
	dir := filepath.Dir(path)
	dirStat, err := os.Stat(dir)
	if err == nil {
		if dirStat.Mode().IsDir() {
			name, err = nameFromPath(path)
		} else {
			err = fmt.Errorf("%s: not a directory", dir)
		}
		return
	}
	return
}
