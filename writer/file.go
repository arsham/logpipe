// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/arsham/logpipe/internal"
	"github.com/pkg/errors"
)

// MinimumDelay is the minimum time set for flush delays.
var MinimumDelay = 10 * time.Millisecond

type writeCloseNamer interface {
	io.WriteCloser
	Name() string
}

// File writs records log entries to a file. It buffers the writes to obtain
// better performance. It flushes the buffer every 1 seconds.
// It implements io.WriteCloser interface.
type File struct {
	file writeCloseNamer

	sync.Mutex // guards against the buffer
	buf        *bufio.Writer

	closed uint32

	delay  time.Duration // the delay between flushes
	logger internal.FieldLogger
}

// NewFile returns error if the file can not be created.
// It starts a goroutine that flushes the logs in intervals.
func NewFile(conf ...func(*File) error) (*File, error) {
	fl := &File{}

	for _, f := range conf {
		if err := f(fl); err != nil {
			return nil, err
		}
	}

	if fl.delay == 0 {
		WithFlushDelay(time.Second)(fl)
	}

	go fl.sync()

	return fl, nil
}

// Close closes the File.
func (f *File) Close() error {
	if err := f.Flush(); err != nil {
		return errors.Wrap(err, "flushing on close")
	}

	atomic.StoreUint32(&f.closed, uint32(1))
	return f.file.Close()
}

// Name returns the file location on disk.
func (f *File) Name() string {
	return f.file.Name()
}

// Write writes the input bytes to the file.
// The write will not appear in the file unless the buffer is flushed. (see Flush())
func (f *File) Write(p []byte) (int, error) {
	f.Lock()
	defer f.Unlock()

	if atomic.LoadUint32(&f.closed) > 0 {
		return 0, errors.New("file closed")
	}

	n1, err := f.buf.Write(p)
	if err != nil {
		return n1, errors.Wrap(err, "writing the bytes")
	}

	if !bytes.HasSuffix(p, []byte("\n")) {
		err = f.buf.WriteByte('\n') // required for creating a new line
	}

	if err != nil {
		return 0, errors.Wrap(err, "writing new line")
	}
	return n1, nil
}

// Flush flushes the underlying buffer.
func (f *File) Flush() error {
	f.Lock()
	defer f.Unlock()
	return f.buf.Flush()
}

// flusher flushes the logs onto the file in intervals.
func (f *File) sync() {
	for {
		<-time.After(f.delay)
		f.Flush()
	}
}

// Logger returns the logger attached to the file
func (f *File) Logger() internal.FieldLogger {
	return f.logger
}

// WithFileLoc opens a new file at location, or creates one if not exists.
// It returns error if it could not create or have write permission to the file.
func WithFileLoc(location string) func(*File) error {
	return func(f *File) error {
		var (
			file *os.File
			err  error
		)
		if file, err = os.OpenFile(location, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
			if os.IsPermission(err) {
				return errors.Wrap(err, "opening file")
			}
		}
		return WithWriter(file)(f)
	}
}

// WithWriter sets the output as the given writer. It wraps it in a buffer
// for better performance.
func WithWriter(w writeCloseNamer) func(*File) error {
	return func(f *File) error {
		f.file = w
		f.buf = bufio.NewWriter(f.file)
		return nil
	}
}

// WithBufWriter sets the buffered writer.
func WithBufWriter(w *bufio.Writer) func(*File) error {
	return func(f *File) error {
		f.buf = w
		return nil
	}
}

// WithFlushDelay sets the delay time between flushes.
func WithFlushDelay(delay time.Duration) func(*File) error {
	return func(f *File) error {
		if delay < MinimumDelay {
			return fmt.Errorf("low (%d) delay", delay)
		}
		f.delay = delay
		return nil
	}
}

// WithLogger sets the delay time between flushes.
func WithLogger(logger internal.FieldLogger) func(*File) error {
	return func(f *File) error {
		if logger == nil {
			return errors.New("nil logger")
		}
		f.logger = logger
		return nil
	}
}
