// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("GetReader", func() {
	var (
		hook   *test.Hook
		logger internal.FieldLogger
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()

	})
	AfterEach(func() {
		hook.Reset()
	})

	Describe("guessing a Plain reader", func() {
		DescribeTable("with bad json object", func(message string, err error) {
			r, er := reader.GetReader(bytes.NewReader([]byte(message)), logger)
			if err != nil {
				Expect(er.Error()).To(ContainSubstring(err.Error()))
				Expect(r).To(BeNil())
			} else {
				Expect(er).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
			}
		},
			Entry("no message", `{"type":"error","timestamp":"2017-01-01"}`, reader.ErrEmptyMessage),
			Entry("empty message string", `{"type":"error","timestamp":"2017-01-01","message":""}`, reader.ErrEmptyMessage),
			Entry("no type", `{"message":"error should not occur","timestamp":"2017-01-01"}`, nil),
			Entry("empty type type", `{"message":"error should not occur","timestamp":"2017-01-01","kind":""}`, nil),
			Entry("no timestamp", `{"message":"error should not occur","type":"info"}`, nil),
			Entry("bad timestamp", `{"message":"ok","timestamp":"01a"}`, reader.ErrTimestamp),
			Entry("empty timestamp string", `{"message":"error should not occur","type":"info","timestamp":""}`, nil),
			Entry("all right", `{"type":"error", "message":"Devil is the king!","timestamp":"2017-01-01"}`, nil),
			Entry("all right too", `{"message":"Devil is the king!"}`, nil),
			Entry("capital type", `{"type":"INFO", "message":"Devil is the king!"}`, nil),
			Entry("all right + more", `{"type":"error", "message":"Devil","timestamp":"2017-01-01", "king": true}`, nil),
			Entry("corrupted json object", `{"type":"error", ,}`, reader.ErrCorruptedJSON),
			Entry("only one string in json object", `"type"`, reader.ErrCorruptedJSON),
			Entry("empty map in json object", `{}`, reader.ErrEmptyObject),
		)

		Context("when there is no type defined", func() {
			It("should set it as info", func() {
				r, err := reader.GetReader(bytes.NewReader([]byte(`{"message":"blah"}`)), logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				p := r.(*reader.Plain)
				Expect(p.Kind).To(Equal(reader.INFO))
			})
		})

		Context("when there is no timestamp defined", func() {
			It("should set it as the time it receives the log", func() {
				r, err := reader.GetReader(bytes.NewReader([]byte(`{"message":"blah"}`)), logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
				p := r.(*reader.Plain)
				Expect(p.Timestamp).To(BeTemporally("~", time.Now(), time.Second))
			})
		})

		Context("given a one line log", func() {
			input := []byte(fmt.Sprintf(
				`{"type":"error","message":"%s","timestamp":"2017-01-01"}`,
				strings.Repeat("lhad ;adfadf adfaf", 100),
			))
			r, err := reader.GetReader(bytes.NewReader(input), logger)
			It("returns a Plain object", func() {
				Expect(r).To(BeAssignableToTypeOf(&reader.Plain{}))
			})
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("having a new line in the one line log message", func() {
			input := []byte(fmt.Sprintf(
				`{"type":"error","message":"sdsd sdsd \n sdddds \n\r sdkj sdds \r\n sdsd","timestamp":"2017-01-01"}`,
			))
			r, err := reader.GetReader(bytes.NewReader(input), logger)
			It("returns a Plain object", func() {
				Expect(r).To(BeAssignableToTypeOf(&reader.Plain{}))
			})
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
