// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/arsham/logpipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Plain", func() {
	Describe("Read", func() {
		Describe("Errors", func() {
			var (
				r             *reader.Plain
				kind, message string
				timestamp     time.Time
				err           error
				p             []byte
			)

			JustBeforeEach(func() {
				r = &reader.Plain{
					Kind:      kind,
					Message:   message,
					Timestamp: timestamp,
				}
				p = make([]byte, 0)
				_, err = r.Read(p)
			})

			AfterEach(func() {
				kind = ""
				message = ""
				timestamp = time.Time{}
			})

			Context("nil Timestamp will cause an error", func() {

				BeforeEach(func() {
					kind = "error"
					message = "this is a message"
					timestamp = time.Time{}
				})

				It("returns error containing a warning about timestamp", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("timestamp"))
				})
				Specify("input buffer is empty", func() {
					Expect(p).To(BeEmpty())
				})
			})

			Context("empty message will cause an error", func() {
				BeforeEach(func() {
					kind = "error"
					timestamp = time.Now()
				})

				It("returns error containing a warning about message", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("message"))
				})
			})

			Context("empty kind will cause an error", func() {
				BeforeEach(func() {
					message = "this is a message"
					timestamp = time.Now()
				})

				It("returns error containing a warning about kind", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("kind"))
				})
			})

		})

		Describe("constructing log entries", func() {

			Describe("Read method", func() {
				var (
					now     = time.Now()
					nowStr  = now.Format(reader.TimestampFormat)
					p       *reader.Plain
					kind    string
					message = strings.Repeat("this is a long message", 150)
				)

				JustBeforeEach(func() {
					p = &reader.Plain{
						Kind:      kind,
						Message:   message,
						Timestamp: now,
					}
				})

				AfterEach(func() {
					kind = ""
				})

				Context("by calling Read method", func() {
					var (
						expected string
						b        []byte
					)

					BeforeEach(func() {
						expected = fmt.Sprintf("[%s] [%s] %s", nowStr, "ERROR", message)
						kind = "error"
						b = make([]byte, len(expected))
					})
					It("should not error or return io.EOF", func() {
						_, err := p.Read(b)
						Expect(err).To(Or(BeNil(), Equal(io.EOF)))
					})
					Specify("length and return value to be as expected", func() {
						n, _ := p.Read(b)
						Expect(n).To(Equal(len(expected)))
						Expect(b).To(BeEquivalentTo(expected))
					})
				})

				Context("with using io.Copy", func() {
					var (
						expected string
						buf      *bytes.Buffer
					)

					BeforeEach(func() {
						expected = fmt.Sprintf("[%s] [%s] %s", nowStr, "ERROR", message)
						kind = "error"
						buf = &bytes.Buffer{}
					})
					It("should not error or return io.EOF", func() {
						_, err := io.Copy(buf, p)
						Expect(err).NotTo(HaveOccurred())
						Expect(err).To(Or(BeNil(), Equal(io.EOF)))
					})
					Specify("length and return value to be as expected", func() {
						n, _ := io.Copy(buf, p)
						Expect(int(n)).To(Equal(len(expected)))
						Expect(buf.String()).To(BeEquivalentTo(expected))
					})
				})

				Context("with using ioutil.ReadAll", func() {
					var expected string

					BeforeEach(func() {
						expected = fmt.Sprintf("[%s] [%s] %s", nowStr, "ERROR", message)
						kind = "error"
					})
					It("should not error or return io.EOF", func() {
						_, err := ioutil.ReadAll(p)
						Expect(err).To(Or(BeNil(), Equal(io.EOF)))
					})
					Specify("length and return value to be as expected", func() {
						b, _ := ioutil.ReadAll(p)
						Expect(len(b)).To(Equal(len(expected)))
						Expect(b).To(BeEquivalentTo(expected))
					})
				})
			})
		})
	})
})
