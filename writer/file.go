// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// File writs records log entries to a file. It buffers the writes to obtain
// better performance. It flushes the buffer every 1 seconds.
// It implements io.WriteCloser interface.
type File struct {
	name       string
	file       *os.File
	sync.Mutex // guards against the buffer
	buf        *bufio.Writer
}

// NewFile returns error if the file can not be created.
func NewFile(location string) (*File, error) {
	var (
		f   *os.File
		err error
	)

	if f, err = os.OpenFile(location, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		if os.IsPermission(err) {
			return nil, errors.Wrap(err, "opening file")
		}
	}
	buf := bufio.NewWriter(f)

	fl := &File{
		file: f,
		name: location,
		buf:  buf,
	}

	// this goroutine will flush the logs onto the file
	go func(buf *bufio.Writer) {
		ticker := time.NewTicker(time.Second)
		for range ticker.C {
			fl.Lock()
			buf.Flush()
			fl.Unlock()
		}
	}(buf)

	return fl, nil
}

// Close closes the File.
func (f *File) Close() error {
	if err := f.Flush(); err != nil {
		return errors.Wrap(err, "flushing on close")
	}
	return f.file.Close()
}

// Name returns the file location on disk.
func (f *File) Name() string {
	return f.name
}

// Write writes the input bytes to the file.
// The write will not appear in the file unless the buffer is flushed. (see Flush())
func (f *File) Write(p []byte) (int, error) {
	f.Lock()
	defer f.Unlock()

	n1, err := f.buf.Write(p)
	if err != nil {
		return n1, errors.Wrap(err, "writing the bytes")
	}

	n2, err := f.buf.Write([]byte("\n")) // required for creating a new line
	if err != nil {
		return n2, errors.Wrap(err, "writing new line")
	}
	return n1, nil
}

// Flush flushes the underlying buffer.
func (f *File) Flush() error {
	f.Lock()
	defer f.Unlock()
	return f.buf.Flush()
}
