// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/arsham/logpipe/reader"
	jason "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

// LogService listens to the incoming http requests and decides how to route
// the payload to be written.
type LogService struct {
	Writer io.Writer
}

func writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, err.Error())
}

// RecieveHandler handles the logs coming from the endpoint
func (l *LogService) RecieveHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	red := io.TeeReader(r.Body, &buf)
	j, err := jason.NewFromReader(red)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	if m, err := j.Map(); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	} else if len(m) == 0 {
		writeError(w, errors.New("empty object"), http.StatusBadRequest)
		return
	}

	rd, err := reader.GetReader(buf.Bytes())
	if err != nil {
		writeError(w, errors.New("getting reader"), http.StatusBadRequest)
		return
	}
	_, err = io.Copy(l.Writer, rd)
	if err != nil {
		writeError(w, errors.Wrap(err, "writing to file"), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
