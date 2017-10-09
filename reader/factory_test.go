// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"fmt"
	"strings"

	"github.com/arsham/logpipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetReader", func() {
	Describe("guessing a Plain reader", func() {
		DescribeTable("with bad json object", func(message, errMsg string) {
			r, err := reader.GetReader([]byte(message))
			if errMsg != "" {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(errMsg))
				Expect(r).To(BeNil())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
			}
		},
			Entry("no message", `{"type":"error","timestamp":"2017-01-01"}`, "empty message"),
			Entry("no type", `{"message":"error should occur","timestamp":"2017-01-01"}`, "empty type"),
			Entry("no timestamp", `{"message":"error should occur","type":"info"}`, "empty timestamp"),
			Entry("bad timestamp", `{"type":"error", "message":"ok","timestamp":"01a"}`, "parsing timestamp"),
			Entry("all right", `{"type":"error", "message":"Devil is the king!","timestamp":"2017-01-01"}`, ""),
			Entry("capital type", `{"type":"INFO", "message":"Devil is the king!","timestamp":"2017-01-01"}`, ""),
			Entry("all right + more", `{"type":"error", "message":"Devil","timestamp":"2017-01-01", "king": true}`, ""),
			Entry("decoding json object", `{"type":"error", ,}`, "decoding json object"),
		)

		Context("given a one line log", func() {
			input := []byte(fmt.Sprintf(
				`{"type":"error","message":"%s","timestamp":"2017-01-01"}`,
				strings.Repeat("lhad ;adfadf adfaf", 100),
			))
			r, err := reader.GetReader(input)
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
			r, err := reader.GetReader(input)
			It("returns a Plain object", func() {
				Expect(r).To(BeAssignableToTypeOf(&reader.Plain{}))
			})
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
