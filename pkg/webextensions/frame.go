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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// Do not know if it may happen
var SizeMismatchError = errors.New("Written frame size differs from actual")

var TooLargeError = errors.New("Frame size is too large")
var ErrReadBody = errors.New("Frame body has not read till its end")
var ErrOutBufferNotEmpty = errors.New("Output buffer is not empty on close")

// Said to protect browser from buggy backends
var MaxInSize uint32 = 1024 * 1024 * 1024

// Split continuous data into packets using "size,data" protocol.
// Call ReadHeader() then consume packet data using Read.
// More internal interface than public one.
type FrameReader interface {
	io.Reader
	// Read packet size
	ReadHeader() error
}

// FrameReader with binary uint32 size header as in Webextensions native messaging
// protocol. Internal helper that might be helpful in other application.
type FrameLimitedReader struct {
	io.LimitedReader
}

var _ FrameReader = (*FrameLimitedReader)(nil)

func NewFrameLimitedReader(reader io.ReadCloser) FrameReader {
	return &FrameLimitedReader{io.LimitedReader{R: reader, N: 0}}
}

func (r *FrameLimitedReader) ReadHeader() error {
	if r.LimitedReader.N != 0 {
		// TODO maybe just discard
		return ErrReadBody
	}
	r.LimitedReader.N = 4
	var size uint32
	if err := binary.Read(r, NativeEndian, &size); err != nil {
		return err
	}
	if size > MaxInSize {
		return TooLargeError
	}
	r.LimitedReader.N = int64(size)
	return nil
}

type FrameWriter interface {
	io.Writer
	WriteFrame() error
	// Likely have an underlying buffer. Returns if it is empty.
	Empty() bool
	// Annoying item, unsure if it is worth to check for non-empty buffer error
	Discard()
}

type FrameBufferedWriter struct {
	bytes.Buffer
	W io.Writer
}

var _ FrameWriter = (*FrameBufferedWriter)(nil)

func NewFrameBufferedWriter(w io.WriteCloser) FrameWriter {
	return &FrameBufferedWriter{bytes.Buffer{}, w}
}

func (w *FrameBufferedWriter) WriteFrame() error {
	defer w.Buffer.Reset()
	size := uint32(w.Buffer.Len())
	if err := binary.Write(w.W, NativeEndian, &size); err != nil {
		return err
	}
	outSize, err := w.Buffer.WriteTo(w.W)
	if err != nil {
		return err
	}
	if outSize != int64(size) {
		return SizeMismatchError
	}
	return nil
}

func (w *FrameBufferedWriter) Empty() bool {
	return w.Buffer.Len() == 0
}

func (w *FrameBufferedWriter) Discard() {
	w.Buffer.Reset()
}

// Wrapper interface that behaves as ordinary connection for RPC codec
// and use "size,data" format for communication with other end.
// It is an internal interface, but there is no reason to hide it.
type FrameReadWriteCloser interface {
	FrameReader
	FrameWriter
	io.Closer
}

// A mediator that merges stdin and stdout into unified connection expected
// by a RPC codec. Handle packet size before data during data exchange over
// IO streams. Is not supposed to be used directly in ordinary cases.
type SplitFrameReadWriteCloser struct {
	FrameReader
	FrameWriter
	R io.Closer // might be obtained from FrameReader using type cast
	W io.Closer
}

var _ io.ReadWriteCloser = (*SplitFrameReadWriteCloser)(nil)

func NewSplitFrameReadWriteCloser(reader io.ReadCloser, writer io.WriteCloser) FrameReadWriteCloser {
	frameReader := NewFrameLimitedReader(reader)
	frameWriter := NewFrameBufferedWriter(writer)
	return &SplitFrameReadWriteCloser{frameReader, frameWriter, reader, writer}
}

func (s *SplitFrameReadWriteCloser) Close() error {
	defer s.FrameWriter.Discard()
	errReader := s.R.Close()
	errWriter := s.W.Close()
	if errReader != nil {
		return errReader
	}
	if errWriter != nil {
		return errWriter
	}
	if !s.FrameWriter.Empty() {
		return ErrOutBufferNotEmpty
	}
	return nil
}

func (s *SplitFrameReadWriteCloser) Discard() {
	s.FrameWriter.Discard()
}
