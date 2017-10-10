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

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/reader"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Plain", func() {
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

	Describe("Read", func() {
		var (
			kind, message string
			timestamp     time.Time
			err           error
			bs            []byte
			rp            *reader.Plain
		)

		JustBeforeEach(func() {
			rp = &reader.Plain{
				Kind:      kind,
				Message:   message,
				Timestamp: timestamp,
				Logger:    logger,
			}
			bs = make([]byte, 0)
			_, err = rp.Read(bs)
		})

		AfterEach(func() {
			kind = ""
			message = ""
			timestamp = time.Time{}
		})

		Context("providing nil Timestamp", func() {

			BeforeEach(func() {
				message = "this is a message"
				timestamp = time.Time{}
			})

			It("returns an error with a warning about timestamp", func() {
				Expect(errors.Cause(err)).To(Equal(reader.ErrNilTimestamp))
				Expect(hook.LastEntry().Message).To(ContainSubstring(reader.ErrNilTimestamp.Error()))
			})
			Specify("input buffer should be empty", func() {
				Expect(bs).To(BeEmpty())
			})
		})

		Context("providing an empty message", func() {
			BeforeEach(func() {
				timestamp = time.Now()
			})

			It("returns error containing a warning about message", func() {
				Expect(errors.Cause(err)).To(Equal(reader.ErrEmptyMessage))
				Expect(hook.LastEntry().Message).To(ContainSubstring(reader.ErrEmptyMessage.Error()))
			})
		})

		Context("empty kind", func() {
			BeforeEach(func() {
				message = "this is a message"
				timestamp = time.Now()
			})

			It("should fall back to info", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(strings.ToLower(rp.Kind)).To(Equal("info"))
			})
		})

		Describe("constructing log entries", func() {

			Describe("Read method", func() {
				var (
					rp      *reader.Plain
					kind    string
					message = strings.Repeat("this is a long message", 150)
					now     = time.Now()
					nowStr  = now.Format(reader.TimestampFormat)
				)

				JustBeforeEach(func() {
					rp = &reader.Plain{
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
						_, err := rp.Read(b)
						Expect(err).To(Or(BeNil(), Equal(io.EOF)))
					})
					Specify("length and return value to be as expected", func() {
						n, _ := rp.Read(b)
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
						_, err := io.Copy(buf, rp)
						Expect(err).NotTo(HaveOccurred())
						Expect(err).To(Or(BeNil(), Equal(io.EOF)))
					})
					Specify("length and return value to be as expected", func() {
						n, _ := io.Copy(buf, rp)
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
						_, err := ioutil.ReadAll(rp)
						Expect(err).To(Or(BeNil(), Equal(io.EOF)))
					})
					Specify("length and return value to be as expected", func() {
						b, _ := ioutil.ReadAll(rp)
						Expect(len(b)).To(Equal(len(expected)))
						Expect(b).To(BeEquivalentTo(expected))
					})
				})
			})
		})
	})
})
