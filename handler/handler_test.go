// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/writer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Handler", func() {
	var (
		rec    *httptest.ResponseRecorder
		h      http.Handler
		file   *writer.File
		hook   *test.Hook
		logger internal.FieldLogger
	)

	BeforeEach(func() {
		rec = httptest.NewRecorder()

		w, err := ioutil.TempFile("", "handler_test")
		Expect(err).NotTo(HaveOccurred())

		logger, hook = test.NewNullLogger()
		file, err = writer.NewFile(
			writer.WithWriter(w),
			writer.WithFlushDelay(writer.MinimumDelay+time.Millisecond),
			writer.WithLogger(logger),
		)
		Expect(err).NotTo(HaveOccurred())

		s := &handler.Service{
			Writer: file,
			Logger: logger,
		}
		h = http.HandlerFunc(s.RecieveHandler)
	})

	AfterEach(func() {
		hook.Reset()
		file.Close()
		os.Remove(file.Name())
	})

	Describe("post handling", func() {

		DescribeTable("with bad json object", func(message string, status int, errMsg string) {
			req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
			Expect(err).NotTo(HaveOccurred())

			req.Header.Set("Content-Type", "application/json")

			h.ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(status))
			Expect(hook.LastEntry().Message).To(ContainSubstring(errMsg))
		},
			Entry("empty", `{}`, http.StatusBadRequest, "empty object"),
			Entry("bogus values", `{"something":"another thing"}`, http.StatusBadRequest, "getting reader"),
			Entry("corrupted", `"something":"another thing"}`, http.StatusBadRequest, "getting map"),
			Entry("jason reader error", `{"s":[1,2 3]}`, http.StatusBadRequest, "corrupted json"),
		)

		Context("handling copy to closed file", func() {
			var req *http.Request

			JustBeforeEach(func() {
				file, err := ioutil.TempFile("", "handle_test_error")
				Expect(err).NotTo(HaveOccurred())

				err = file.Close()
				Expect(err).NotTo(HaveOccurred())

				s := &handler.Service{
					Writer: file,
					Logger: logger,
				}
				h = http.HandlerFunc(s.RecieveHandler)

				message := `{"type":"error","message":"blah","timestamp":"2017-01-01"}`
				req, err = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
				Expect(err).NotTo(HaveOccurred())

				req.Header.Set("Content-Type", "application/json")
				h.ServeHTTP(rec, req)
			})

			It("should error", func() {
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				Expect(rec.Body.String()).To(ContainSubstring("writing to file"))
				Expect(hook.LastEntry().Message).To(ContainSubstring("writing to file"))
			})

		})

		Context("handling plain logs and writing to a file", func() {
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
				Expect(err).NotTo(HaveOccurred())

				req.Header.Set("Content-Type", "application/json")
				h.ServeHTTP(rec, req)
				err = file.Flush()
				Expect(err).NotTo(HaveOccurred())

				content, err = ioutil.ReadFile(file.Name())
				Expect(err).NotTo(HaveOccurred())
			})

			It("should write the log entry", func() {
				Expect(content).To(ContainSubstring(errMsg))
			})
		})
	})
})
