// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	logpipe "github.com/arsham/logpipe"
	"github.com/arsham/logpipe/writer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var (
		rec     *httptest.ResponseRecorder
		handler http.Handler
		file    *writer.File
	)

	BeforeEach(func() {
		rec = httptest.NewRecorder()
		w, err := ioutil.TempFile("", "handler")
		if err != nil {
			panic(err)
		}

		file, err = writer.NewFile(writer.WithWriter(w))
		if err != nil {
			panic(err)
		}

		s := &logpipe.LogService{Writer: file}
		handler = http.HandlerFunc(s.RecieveHandler)
	})

	Describe("post handling", func() {

		DescribeTable("with bad json object", func(message string, status int) {
			req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
			if err != nil {
				panic(err)
			}
			req.Header.Set("Content-Type", "application/json")

			handler.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(status))
		},
			Entry("empty", `{}`, http.StatusBadRequest),
			Entry("bogus values", `{"something":"another thing"}`, http.StatusBadRequest),
			Entry("corrupted", `"something":"another thing"}`, http.StatusBadRequest),
			Entry("jason reader error", `{"s":[1,2 3]}`, http.StatusBadRequest),
			Entry("only type", `{"type":"error"}`, http.StatusBadRequest),
			Entry("only message", `{"message":"error should occur"}`, http.StatusBadRequest),
			Entry("all right", `{"type":"error", "message":"Devil is the king!"}`, http.StatusOK),
			Entry("capital type", `{"type":"INFO", "message":"Devil is the king!"}`, http.StatusOK),
			Entry("all right + more", `{"type":"error", "message":"Devil", "king": true}`, http.StatusOK),
		)
	})
})
