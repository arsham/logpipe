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
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/internal/config"
	"github.com/arsham/logpipe/reader"
	"github.com/arsham/logpipe/writer"
	jason "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

// Service listens to the incoming http requests and decides how to route
// the payload to be written.
type Service struct {
	// Writers is a slice of all writers.
	Writers []io.Writer

	// Logger is used for logging service's behaviours.
	Logger internal.FieldLogger

	// Timeout for shutting down the http server. Default is 5 seconds.
	Timeout time.Duration
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
		return nil, ErrNoLogger
	}
	s.Logger = logger

	if len(s.Writers) == 0 {
		return nil, ErrNoWriter
	}

	if s.Timeout == 0 {
		s.Timeout = 5 * time.Second
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
	for _, wr := range l.Writers {
		go func(wr io.Writer, b []byte) {
			_, err := wr.Write(b)
			if err != nil {
				l.Logger.Error(errors.Wrap(err, ErrWritingEntry.Error()))
			}
		}(wr, b)
	}
	w.WriteHeader(http.StatusOK)
}

// Serve starts a http.Server in a goroutine.
// It listens to the Interrupt signal and shuts down the server.
// It sends back the errors through the errChan channel.
// In case of graceful shut down, it will send an io.EOF error to the error
// channel to signal it has been stopped gracefully.
func (l *Service) Serve(stop chan os.Signal, errChan chan error, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", l.RecieveHandler)
	h := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	srvErr := make(chan error)
	quit := make(chan struct{})
	go func() {
		l.Logger.Infof("running on port: %d", port)
		select {
		case srvErr <- h.ListenAndServe():
		case <-quit:

		}
	}()

	select {
	case err := <-srvErr:
		errChan <- err
		return
	case <-stop:
		close(quit)
		ctx, _ := context.WithTimeout(context.Background(), l.Timeout)
		l.Logger.Infof("shutting down the server: %s", h.Shutdown(ctx))
		errChan <- io.EOF
	}
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
		s.Timeout = timeout
		return nil
	}
}
