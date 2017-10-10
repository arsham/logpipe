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
	Writer io.Writer
	Logger internal.FieldLogger
}

func (l *Service) writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, err.Error())
	l.Logger.Error(err)
}

// RecieveHandler handles the logs coming from the endpoint
func (l *Service) RecieveHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	red := io.TeeReader(r.Body, &buf)
	j, err := jason.NewFromReader(red)
	if err != nil {
		l.writeError(w, errors.Wrap(err, "corrupted json"), http.StatusBadRequest)
		return
	}

	if m, err := j.Map(); err != nil {
		l.writeError(w, errors.Wrap(err, "getting map"), http.StatusBadRequest)
		return
	} else if len(m) == 0 {
		l.writeError(w, errors.New("empty object"), http.StatusBadRequest)
		return
	}

	rd, err := reader.GetReader(buf.Bytes(), l.Logger)
	if err != nil {
		l.writeError(w, errors.New("getting reader"), http.StatusBadRequest)
		return
	}

	_, err = io.Copy(l.Writer, rd)
	if err != nil {
		l.writeError(w, errors.Wrap(err, "writing to file"), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
