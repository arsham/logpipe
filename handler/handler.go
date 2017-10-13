// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package handler contains all handlers for serving POST requests.
// The payload should have at least one key: message. And it contains the
// log entry for writing. If a "type" for the entry is not provided, it falls
// back to "info". If the "timestamp" is not provided, it uses the current time
// it receives the payload.
package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/reader"
	jason "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

// Service listens to the incoming http requests and decides how to route
// the payload to be written.
type Service struct {
	Writers []io.Writer
	Logger  internal.FieldLogger
}

// New returns an error if there is no logger or no writer specified.
func New(opts ...func(*Service) error) (*Service, error) {
	if opts == nil {
		return nil, ErrNoOptions
	}
	s := &Service{}
	for _, f := range opts {
		err := f(s)
		if err != nil {
			return nil, err
		}
	}

	if s.Logger == nil {
		return nil, ErrNoLogger
	}

	if len(s.Writers) == 0 {
		return nil, ErrNoWriter
	}

	return s, nil
}

func (l *Service) writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, err.Error())
	l.Logger.Error(err)
}

// RecieveHandler handles the logs coming from the endpoint.
// It handles all writes in their own goroutine in order to avoid write loss.
func (l *Service) RecieveHandler(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	red := io.TeeReader(r.Body, buf)
	j, err := jason.NewFromReader(red)
	if err != nil {
		l.writeError(w, errors.Wrap(err, ErrCorruptedJSON.Error()), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if m, err := j.Map(); err != nil {
		l.writeError(w, errors.Wrap(err, ErrGettingMap.Error()), http.StatusBadRequest)
		return
	} else if len(m) == 0 {
		l.writeError(w, ErrEmptyObject, http.StatusBadRequest)
		return
	}

	rd, err := reader.GetReader(buf.Bytes(), l.Logger)
	if err != nil {
		l.writeError(w, ErrGettingReader, http.StatusBadRequest)
		return
	}

	// allWriters := make([]io.Writer, len(l.Writers))
	// for i, theWriter := range l.Writers {
	// 	pipeReader, pipeWriter := io.Pipe()
	// 	allWriters[i] = pipeWriter
	// 	go func(theWriter io.Writer, pipeWriter io.WriteCloser, pipeReader io.ReadCloser) {
	// 		_, err = io.Copy(theWriter, pipeReader)
	// 		if err != nil {
	// 			l.Logger.Error(errors.Wrap(err, ErrWritingEntry.Error()))
	// 		}
	// 		pipeWriter.Close()
	// 		pipeReader.Close()
	// 	}(theWriter, pipeWriter, pipeReader)
	// }

	// go func() {
	// 	wr := io.MultiWriter(allWriters...)
	// 	_, err = io.Copy(wr, rd)
	// 	if err != nil {
	// 		l.Logger.Error(errors.Wrap(err, ErrWritingEntry.Error()))
	// 	}
	// }()

	buf = new(bytes.Buffer)
	buf.ReadFrom(rd)
	b := buf.Bytes()
	for _, wr := range l.Writers[:] {
		go func(wr io.Writer, b []byte) {
			_, err = wr.Write(b)
			if err != nil {
				l.Logger.Error(errors.Wrap(err, ErrWritingEntry.Error()))
			}
		}(wr, b[:])
	}

	w.WriteHeader(http.StatusOK)
}

// WithWriters will return an error if two identical writers are injected.
func WithWriters(ws ...io.Writer) func(*Service) error {
	return func(s *Service) error {
		for _, w := range ws {
			for _, ew := range s.Writers {
				if ew == w {
					return ErrDuplicateWriter
				}
			}
			s.Writers = append(s.Writers, w)
		}
		return nil
	}
}

// WithLogger will return an error if the logger is nil
func WithLogger(logger internal.FieldLogger) func(*Service) error {
	return func(s *Service) error {
		if logger == nil {
			return ErrNilLogger
		}
		s.Logger = logger
		return nil
	}
}
