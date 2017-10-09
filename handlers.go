// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	jason "github.com/bitly/go-simplejson"
)

type LogService struct {
	Writer io.Writer
}

func writeError(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	fmt.Fprint(w, err.Error())
}

// RecieveHandler handles the logs coming from the endpoint
func (l *LogService) RecieveHandler(w http.ResponseWriter, r *http.Request) {
	j, err := jason.NewFromReader(r.Body)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	t, err := j.Get("type").String()
	if err != nil || t == "" {
		writeError(w, errors.New("empty type"), http.StatusBadRequest)
		return
	}

	t, err = j.Get("message").String()
	if err != nil || t == "" {
		writeError(w, errors.New("empty message"), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
