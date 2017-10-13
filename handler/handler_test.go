// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"io"
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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus/hooks/test"
)

// slowWriter writes really slow
type slowWriter struct {
	writer.File
	delay time.Duration
}

func (s *slowWriter) Write(p []byte) (int, error) {
	<-time.After(s.delay)
	return s.File.Write(p)
}

var _ = Describe("Handler", func() {
	Describe("New", func() {
		Context("without specifying any options", func() {
			s, err := handler.New()
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
			Specify("the service should be nil", func() {
				Expect(s).To(BeNil())
			})
		})

		Context("when no writer is passed", func() {
			logger := internal.DiscardLogger()
			s, err := handler.New(handler.WithLogger(logger))

			It("should return an error", func() {
				Expect(errors.Cause(err)).To(Equal(handler.ErrNoWriter))
			})
			Specify("service is nil", func() {
				Expect(s).To(BeNil())
			})
		})

		Context("when no logger is passed", func() {
			var (
				w   *writer.File
				s   *handler.Service
				err error
			)
			BeforeEach(func() {
				w, err = writer.NewFile(writer.WithFileLoc(os.DevNull))
				Expect(err).NotTo(HaveOccurred())
				s, err = handler.New(handler.WithWriters(w))
			})

			It("should return an error", func() {
				Expect(errors.Cause(err)).To(Equal(handler.ErrNoLogger))
			})
			Specify("service is nil", func() {
				Expect(s).To(BeNil())
			})
		})

		Describe("WithWriter", func() {
			var (
				w1, w2 *writer.File
				err    error
			)

			BeforeEach(func() {
				w1, err = writer.NewFile(writer.WithFileLoc(os.DevNull))
				Expect(err).NotTo(HaveOccurred())
				w2, err = writer.NewFile(writer.WithFileLoc(os.DevNull))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when adding one writer", func() {
				var s *handler.Service

				BeforeEach(func() {
					s, err = handler.New(
						handler.WithWriters(w1),
						handler.WithLogger(internal.DiscardLogger()),
					)
					Expect(err).NotTo(HaveOccurred())
				})
				Specify("the writer should contain that writer", func() {
					Expect(s.Writers[0]).To(Equal(w1))
				})
			})

			Context("when adding multiple writers", func() {
				var s *handler.Service

				BeforeEach(func() {
					s, err = handler.New(
						handler.WithWriters(w1),
						handler.WithWriters(w2),
						handler.WithLogger(internal.DiscardLogger()),
					)
					Expect(err).NotTo(HaveOccurred())
				})
				Specify("the writer should contain all those writers", func() {
					Expect(s.Writers).To(ContainElement(w1))
					Expect(s.Writers).To(ContainElement(w2))
				})
			})

			Context("when adding a writer after creation", func() {
				var s *handler.Service

				BeforeEach(func() {
					s, err = handler.New(
						handler.WithWriters(w1),
						handler.WithLogger(internal.DiscardLogger()),
					)
					Expect(err).NotTo(HaveOccurred())
				})
				JustBeforeEach(func() {
					Expect(handler.WithWriters(w2)(s)).NotTo(HaveOccurred())
				})

				Specify("the writer should contain all those writers", func() {
					Expect(s.Writers).To(ContainElement(w1))
					Expect(s.Writers).To(ContainElement(w2))
				})
			})

			Context("if one writer was added multiple times", func() {
				var (
					s   *handler.Service
					err error
				)

				BeforeEach(func() {
					s, err = handler.New(
						handler.WithWriters(w1),
						handler.WithWriters(w1),
						handler.WithLogger(internal.DiscardLogger()),
					)
				})

				It("should return an error", func() {
					Expect(errors.Cause(err)).To(Equal(handler.ErrDuplicateWriter))
				})
				Specify("the service should be nil", func() {
					Expect(s).To(BeNil())
				})
			})

		})

		Describe("WithLogger", func() {
			var (
				w      *writer.File
				logger = internal.DiscardLogger()
				err    error
			)

			BeforeEach(func() {
				w, err = writer.NewFile(writer.WithFileLoc(os.DevNull))
				Expect(err).NotTo(HaveOccurred())
			})

			Context("setting a nil logger", func() {
				It("should return an error", func() {
					err := handler.WithLogger(nil)(&handler.Service{})
					Expect(errors.Cause(err)).To(Equal(handler.ErrNilLogger))
				})
			})

			Context("having a writer and a given logger", func() {
				var s *handler.Service
				BeforeEach(func() {
					s, err = handler.New(
						handler.WithWriters(w),
						handler.WithLogger(logger),
					)
					Expect(err).NotTo(HaveOccurred())
				})

				Specify("the service should have the logger in its fields", func() {
					Expect(s.Logger).To(Equal(logger))
				})
			})
		})
	})

	Describe("ReceiveHandler", func() {

		Describe("post handling", func() {
			var (
				rec          *httptest.ResponseRecorder
				h            http.Handler
				logger, hook = test.NewNullLogger()
				file1, file2 *writer.File
				service      *handler.Service
				flushDelay   = writer.MinimumDelay + time.Millisecond
			)

			BeforeEach(func() {
				rec = httptest.NewRecorder()

				f1, err := ioutil.TempFile("", "handler_test")
				Expect(err).NotTo(HaveOccurred())

				f2, err := ioutil.TempFile("", "handler_test")
				Expect(err).NotTo(HaveOccurred())

				file1, err = writer.NewFile(
					writer.WithWriter(f1),
					writer.WithFlushDelay(flushDelay),
					writer.WithLogger(logger),
				)
				Expect(err).NotTo(HaveOccurred())

				file2, err = writer.NewFile(
					writer.WithWriter(f2),
					writer.WithFlushDelay(flushDelay),
					writer.WithLogger(logger),
				)
				Expect(err).NotTo(HaveOccurred())

				writers := []io.Writer{
					file1,
					file2,
				}
				service = &handler.Service{
					Writers: writers,
					Logger:  logger,
				}
				h = http.HandlerFunc(service.RecieveHandler)
			})

			AfterEach(func() {
				hook.Reset()
				file1.Close()
				file2.Close()
				os.Remove(file1.Name())
				os.Remove(file2.Name())
			})

			DescribeTable("with bad json object", func(message string, status int, errMsg error) {
				req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
				Expect(err).NotTo(HaveOccurred())

				req.Header.Set("Content-Type", "application/json")

				h.ServeHTTP(rec, req)
				Expect(rec.Code).To(Equal(status))
				Expect(hook.LastEntry().Message).To(ContainSubstring(errMsg.Error()))
			},
				Entry("empty", `{}`, http.StatusBadRequest, handler.ErrEmptyObject),
				Entry("bogus values", `{"something":"another thing"}`, http.StatusBadRequest, handler.ErrGettingReader),
				Entry("corrupted", `"something":"another thing"}`, http.StatusBadRequest, handler.ErrGettingMap),
				Entry("jason reader error", `{"s":[1,2 3]}`, http.StatusBadRequest, handler.ErrCorruptedJSON),
			)

			Context("handling copy to closed file", func() {
				var (
					req  *http.Request
					file *os.File
					err  error
				)

				BeforeEach(func() {
					file, err = ioutil.TempFile("", "handler_test_error")
					Expect(err).NotTo(HaveOccurred())

					err = file.Close()
					Expect(err).NotTo(HaveOccurred())

					s := &handler.Service{
						Writers: []io.Writer{file},
						Logger:  logger,
					}
					h = http.HandlerFunc(s.RecieveHandler)

					message := `{"type":"error","message":"blah","timestamp":"2017-01-01"}`
					req, err = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
					Expect(err).NotTo(HaveOccurred())

					req.Header.Set("Content-Type", "application/json")
					h.ServeHTTP(rec, req)
				})

				AfterEach(func() {
					os.Remove(file.Name())
				})

				It("should not error", func() {
					Expect(rec.Code).NotTo(Equal(http.StatusBadRequest))
				})

				It("eventually should log the error", func() {
					Eventually(func() string {
						if hook.LastEntry() == nil {
							return ""
						}
						return hook.LastEntry().Message
					}, flushDelay).Should(ContainSubstring(handler.ErrWritingEntry.Error()))
				})
			})

			Context("handling plain logs and writing to a file", func() {
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
				})

				It("should eventually write the log entry", func() {
					Eventually(func() string {

						Expect(file1.Flush()).NotTo(HaveOccurred())
						content, err := ioutil.ReadFile(file1.Name())
						Expect(err).NotTo(HaveOccurred())
						return string(content)

					}, flushDelay+5*time.Second, 0.2).Should(ContainSubstring(errMsg))
				})

				It("should eventually write the log level", func() {
					Eventually(func() string {

						Expect(file1.Flush()).NotTo(HaveOccurred())
						content, err := ioutil.ReadFile(file1.Name())
						Expect(err).NotTo(HaveOccurred())
						return string(content)

					}, flushDelay+5*time.Second, 0.2).Should(ContainSubstring(kind))
				})
			})

			Context("handling plain logs and writing to multiple files", func() {
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
				})

				Context("when both writers are available", func() {

					It("should eventually write the log entry to both files", func() {
						Eventually(func() string {

							Expect(file1.Flush()).NotTo(HaveOccurred())
							content1, err := ioutil.ReadFile(file1.Name())
							Expect(err).NotTo(HaveOccurred())
							return string(content1)

						}, flushDelay+time.Second*2, 0.2).Should(ContainSubstring(errMsg))

						Eventually(func() string {

							Expect(file2.Flush()).NotTo(HaveOccurred())
							content2, err := ioutil.ReadFile(file2.Name())
							Expect(err).NotTo(HaveOccurred())
							return string(content2)

						}, flushDelay+time.Second*2, 0.2).Should(ContainSubstring(errMsg))
					})
				})

				Context("when only one of the writers are available", func() {
					BeforeEach(func() {
						file2.Close()
					})

					It("should eventually write the log entry to the available file", func() {

						Eventually(func() string {
							content, err := ioutil.ReadFile(file1.Name())
							Expect(err).NotTo(HaveOccurred())
							return string(content)
						}, flushDelay+time.Second).Should(ContainSubstring(errMsg))
					})

					It("should eventually log there was an error on the other file", func() {
						Eventually(func() string {
							if hook.LastEntry() == nil {
								return ""
							}
							return hook.LastEntry().Message
						}, flushDelay+time.Second).Should(ContainSubstring(handler.ErrWritingEntry.Error()))
					})
				})
			})

			Describe("handling plain logs and writing to slow writers", func() {
				var (
					content1 []byte
					content2 []byte
					content3 []byte
					err      error
					w        *os.File
					file3    *slowWriter
					delay    = time.Second
				)
				errMsg := "this error has occurred"
				kind := "error"
				timestamp := "2017-01-14 19:10:10"

				BeforeEach(func() {
					w, err = ioutil.TempFile("", "handler_slow_test")
					Expect(err).NotTo(HaveOccurred())

					f, err := writer.NewFile(
						writer.WithWriter(w),
						writer.WithFlushDelay(flushDelay),
						writer.WithLogger(logger),
					)
					Expect(err).NotTo(HaveOccurred())
					file3 = &slowWriter{
						File:  *f,
						delay: delay,
					}

					Expect(handler.WithWriters(file3)(service)).NotTo(HaveOccurred())
					h = http.HandlerFunc(service.RecieveHandler)

				})

				JustBeforeEach(func(done Done) {
					message := fmt.Sprintf(`{"type":"%s","message":"%s","timestamp":"%s"}`,
						kind,
						errMsg,
						timestamp,
					)
					req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
					Expect(err).NotTo(HaveOccurred())

					req.Header.Set("Content-Type", "application/json")
					h.ServeHTTP(rec, req)

					close(done)

				}, flushDelay.Seconds()) // This request should not take as long as the slow writer

				AfterEach(func() {
					hook.Reset()
					os.Remove(file3.Name())
				})

				It("should write the log entry to the fast writers", func(done Done) {

					Eventually(func() string {
						Expect(file1.Flush()).NotTo(HaveOccurred())
						content1, err = ioutil.ReadFile(file1.Name())
						Expect(err).NotTo(HaveOccurred())
						return string(content1)
					}, flushDelay).Should(Not(BeEmpty()))

					Eventually(func() string {
						Expect(file2.Flush()).NotTo(HaveOccurred())
						content2, err = ioutil.ReadFile(file2.Name())
						Expect(err).NotTo(HaveOccurred())
						return string(content2)
					}, flushDelay).Should(Not(BeEmpty()))

					Expect(string(content1)).To(ContainSubstring(errMsg))
					Expect(string(content2)).To(ContainSubstring(errMsg))

					close(done)

				}, flushDelay.Seconds()+1)

				It("should eventually write the log entry to the slow one", func(done Done) {
					if hook.LastEntry() != nil {
						Expect(hook.LastEntry().Message).NotTo(ContainSubstring(handler.ErrWritingEntry.Error()))
					}

					Eventually(func() string {
						content3, err = ioutil.ReadFile(file3.Name())
						Expect(err).NotTo(HaveOccurred())
						return string(content3)
					}, delay*2).Should(ContainSubstring(errMsg))

					if hook.LastEntry() != nil {
						Expect(hook.LastEntry().Message).NotTo(ContainSubstring(handler.ErrWritingEntry.Error()))
					}

					close(done)

				}, delay.Seconds()+1)
			})
		})
	})
})
