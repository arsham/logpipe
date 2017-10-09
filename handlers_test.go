// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

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

		w, err := ioutil.TempFile("", "handler_test")
		if err != nil {
			panic(err)
		}

		file, err = writer.NewFile(
			writer.WithWriter(w),
			writer.WithFlushDelay(writer.MinimumDelay+time.Millisecond),
		)
		if err != nil {
			panic(err)
		}

		s := &logpipe.LogService{Writer: file}
		handler = http.HandlerFunc(s.RecieveHandler)
	})

	AfterEach(func() {
		file.Close()
		os.Remove(file.Name())
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
		)

		Context("handling copy to file error", func() {
			var (
				req *http.Request
			)
			JustBeforeEach(func() {
				file, err := ioutil.TempFile("", "handle_test_error")
				if err != nil {
					panic(err)
				}
				err = file.Close()
				if err != nil {
					panic(err)
				}
				s := &logpipe.LogService{Writer: file}
				handler = http.HandlerFunc(s.RecieveHandler)

				message := `{"type":"error","message":"blah","timestamp":"2017-01-01"}`
				req, err = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				handler.ServeHTTP(rec, req)
			})

			It("should error", func() {
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				Expect(rec.Body.String()).To(ContainSubstring("writing"))
			})

		})

		Context("handling plain logs", func() {
			var content []byte
			errMsg := "this error has occurred"
			kind := "error"
			timestamp := "2017-01-14 19:10:10"

			JustBeforeEach(func() {
				message := fmt.Sprintf(`{"type":"%s","message":"%s","timestamp":"%s"}`,
					kind,
					errMsg,
					timestamp,
				)
				req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				handler.ServeHTTP(rec, req)
				err = file.Flush()
				if err != nil {
					panic(err)
				}
				content, err = ioutil.ReadFile(file.Name())
				if err != nil {
					panic(err)
				}
			})

			It("should write to the log file", func() {
				Expect(content).To(ContainSubstring(errMsg))
			})
		})
	})
})
