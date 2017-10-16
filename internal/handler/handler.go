// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package handler contains bootstrapping logic for the application, and all
// handlers for serving POST requests. The payload should have at least one
// key: message. And it contains the log entry for writing. If a "type" for
// the entry is not provided, it falls back to "info". If the "timestamp" is
// not provided, it uses the current time it receives the payload.
package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/internal/config"
	"github.com/arsham/logpipe/reader"
	"github.com/arsham/logpipe/writer"
	"github.com/pkg/errors"
)

// Service listens to the incoming http requests and decides how to route
// the payload to be written.
type Service struct {
	// Writers is a slice of all writers.
	Writers []io.Writer

	// Logger is used for logging service's behaviours.
	Logger internal.FieldLogger

	// timeout for shutting down the http server. Default is 5 seconds.
	timeout time.Duration
}

// New returns an error if there is no logger or no writer specified.
func New(logger internal.FieldLogger, opts ...func(*Service) error) (*Service, error) {
	s := &Service{}
	for _, f := range opts {
		err := f(s)
		if err != nil {
			return nil, err
		}
	}

	if logger == nil {
		return nil, ErrNilLogger
	}
	s.Logger = logger

	if len(s.Writers) == 0 {
		return nil, ErrNoWriter
	}

	if s.timeout == 0 {
		s.timeout = 5 * time.Second
	}

	return s, nil
}

// Timeout returns the timeout associated with this service.
func (l *Service) Timeout() time.Duration    { return l.timeout }
func (l *Service) Handler() http.HandlerFunc { return l.RecieveHandler }
func (l *Service) writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, err.Error())
	l.Logger.Error(err)
}

// RecieveHandler handles the logs coming from the endpoint.
// It handles the writes in a goroutine in order to avoid write loss.
// It will log any errors that might occur during writes.
// It returns a http.StatusBadRequest if the payload is not a valid JSON object
// or does not contain the required fields.
func (l *Service) RecieveHandler(w http.ResponseWriter, r *http.Request) {
	rd, err := reader.GetReader(r.Body, l.Logger)
	if errors.Cause(err) != nil {
		l.writeError(w, errors.Wrap(err, ErrGettingReader.Error()), http.StatusBadRequest)
		return
	}

	go func(l *Service) {
		concWriter := writer.NewDistribute(l.Writers...)
		_, err = io.Copy(concWriter, rd)
		if err != nil {
			l.Logger.Error(errors.Wrap(err, ErrWritingEntry.Error()))
		}
	}(l)

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

// WithConfWriters uses a config.Setting object to set up the writers.
// If any errors occurred during writer instantiation, it stops and
// returns that error.
func WithConfWriters(logger internal.FieldLogger, c *config.Setting) func(*Service) error {
	var (
		fileLocation string
		ok           bool
		writers      []io.Writer
	)

LOOP:
	for name, conf := range c.Writers {
		switch mod := conf["type"]; mod {
		case "file":
			if fileLocation, ok = conf["location"]; !ok {
				logger.Warnf("no location in settings: %s", name)
				continue LOOP
			}

			w, err := writer.NewFile(
				writer.WithLogger(logger),
				writer.WithFileLoc(fileLocation),
			)
			if err != nil {
				return func(*Service) error {
					return errors.Wrap(err, fileLocation)
				}
			}
			writers = append(writers, w)
		}
	}
	return WithWriters(writers...)
}

// WithTimeout sets the timeout on Service. It returns an error if the timeout
// is zero.
func WithTimeout(timeout time.Duration) func(*Service) error {
	return func(s *Service) error {
		if timeout == 0 {
			return ErrTimeout
		}
		s.timeout = timeout
		return nil
	}
}
