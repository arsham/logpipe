// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/internal/config"
	"github.com/arsham/logpipe/writer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

// timedWriter writes after the delay.
type timedWriter struct {
	content string
	delay   time.Duration
	sync.Mutex
	closed bool
}

func (s *timedWriter) Write(p []byte) (int, error) {
	if s.closed {
		return 0, errors.New("file is already closed")
	}
	<-time.After(s.delay)
	s.Lock()
	defer s.Unlock()
	s.content = string(p)
	return len(p), nil
}

func (s *timedWriter) String() string {
	s.Lock()
	defer s.Unlock()
	return s.content
}

func (s *timedWriter) Close() error {
	s.closed = true
	return nil
}

type logLocker struct {
	*bytes.Buffer
	*sync.Mutex
}

func (l *logLocker) Write(p []byte) (int, error) {
	l.Lock()
	defer l.Unlock()
	return l.Buffer.Write(p)
}

var _ = Describe("Handler", func() {
	Describe("New", func() {

		Context("when no writer is passed", func() {
			logger := internal.DiscardLogger()
			s, err := handler.New(logger)

			It("should return an error", func() {
				Expect(errors.Cause(err)).To(Equal(handler.ErrNoWriter))
			})
			Specify("service is nil", func() {
				Expect(s).To(BeNil())
			})
		})

		Context("when no logger is passed", func() {
			var (
				s   *handler.Service
				err error
			)
			BeforeEach(func() {
				writers := handler.WithWriters(new(bytes.Buffer))
				s, err = handler.New(nil, writers)
			})

			It("should return an error", func() {
				Expect(errors.Cause(err)).To(Equal(handler.ErrNoLogger))
			})
			Specify("service is nil", func() {
				Expect(s).To(BeNil())
			})
		})

		Context("when no timeout is set", func() {
			var (
				s   *handler.Service
				err error
			)
			BeforeEach(func() {
				writers := handler.WithWriters(new(bytes.Buffer))
				s, err = handler.New(internal.DiscardLogger(), writers)
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			Specify("service is not nil", func() {
				Expect(s).NotTo(BeNil())
			})

			Specify("the timeout is 5 seconds by default", func() {
				Expect(s.Timeout).To(Equal(time.Second * 5))
			})
		})
	})

	Describe("WithWriters", func() {
		var (
			w1  = new(timedWriter)
			w2  = new(timedWriter)
			s   *handler.Service
			err error
		)

		Context("when adding one writer", func() {

			BeforeEach(func() {
				s, err = handler.New(
					internal.DiscardLogger(),
					handler.WithWriters(w1),
				)
				Expect(err).NotTo(HaveOccurred())
			})

			Specify("the writer should contain that writer", func() {
				Expect(s.Writers[0]).To(Equal(w1))
			})
		})

		Context("when adding multiple writers", func() {

			BeforeEach(func() {
				s, err = handler.New(
					internal.DiscardLogger(),
					handler.WithWriters(w1),
					handler.WithWriters(w2),
				)
				Expect(err).NotTo(HaveOccurred())
			})

			Specify("the writer should contain all those writers", func() {
				Expect(s.Writers).To(ContainElement(w1))
				Expect(s.Writers).To(ContainElement(w2))
			})
		})

		Context("when adding a writer after creation", func() {

			BeforeEach(func() {
				s, err = handler.New(
					internal.DiscardLogger(),
					handler.WithWriters(w1),
				)
				Expect(err).NotTo(HaveOccurred())
				Expect(handler.WithWriters(w2)(s)).NotTo(HaveOccurred())
			})

			Specify("the writer should contain all those writers", func() {
				Expect(s.Writers).To(ContainElement(w1))
				Expect(s.Writers).To(ContainElement(w2))
			})
		})

		Context("if one writer was added multiple times", func() {

			BeforeEach(func() {
				s, err = handler.New(
					internal.DiscardLogger(),
					handler.WithWriters(w1),
					handler.WithWriters(w1),
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

	Describe("WithConfWriters", func() {
		var (
			location string
			c        config.Setting
			logger   internal.FieldLogger
		)

		JustBeforeEach(func() {
			c = config.Setting{
				Writers: map[string]map[string]string{
					"file1": {
						"type":     "file",
						"location": location,
					},
				},
			}
		})

		Context("having a writer.File in the Setting object", func() {
			BeforeEach(func() {
				location = "/no where to find"
			})
			Context("when the writer.NewFile returns an error", func() {
				It("should return with an error", func() {
					Expect(handler.WithConfWriters(nil, &c)(&handler.Service{})).To(HaveOccurred())
				})
			})
		})

		Context("having a Setting object ", func() {
			var (
				f   *os.File
				err error
				s   *handler.Service
			)
			BeforeEach(func() {
				logger = internal.DiscardLogger()
				f, err = ioutil.TempFile("", "handler_with_config")
				Expect(err).NotTo(HaveOccurred())
				location = f.Name()
			})

			AfterEach(func() {
				os.Remove(f.Name())
			})
			Specify("the writers should be set in the Writers slice", func() {
				s = &handler.Service{}
				Expect(handler.WithConfWriters(logger, &c)(s)).NotTo(HaveOccurred())
				Expect(s.Writers).To(HaveLen(1))
				Expect(s.Writers[0]).To(BeAssignableToTypeOf(&writer.File{}))
			})
		})

		Context("when the file location is not set in the writer", func() {
			var (
				buf       *bytes.Buffer
				c         config.Setting
				s         *handler.Service
				warned    = "file_that_warned"
				notWarned = "file_that_should_not_warn"
				filename  = os.DevNull
			)
			BeforeEach(func() {
				buf = new(bytes.Buffer)
				logger = internal.WithWriter(buf)
				c = config.Setting{
					Writers: map[string]map[string]string{
						warned: {
							"type": "file",
						},
						//making sure it does not fail due to empty writers
						notWarned: {
							"type":     "file",
							"location": filename,
						},
					},
				}
				s = &handler.Service{}
				Expect(handler.WithConfWriters(logger, &c)(s)).NotTo(HaveOccurred())
			})

			It("should log the location", func() {
				Eventually(func() string {
					return buf.String()
				}).Should(ContainSubstring(warned))
				Expect(buf.String()).NotTo(ContainSubstring(notWarned))
			})

			It("should not add the writer", func() {
				Expect(s.Writers).To(HaveLen(1))
				w := s.Writers[0].(*writer.File)
				Expect(w.Name()).To(Equal(filename))
			})
		})
	})

	Describe("WithTimeout", func() {
		Context("when timeout is zero", func() {
			It("should error", func() {
				Expect(handler.WithTimeout(0)(&handler.Service{})).To(Equal(handler.ErrTimeout))
			})
		})
		Context("when there is a timeout set", func() {
			Specify("it should set the timeout in the service", func() {
				s := &handler.Service{}
				Expect(handler.WithTimeout(time.Second)(s)).NotTo(HaveOccurred())
				Expect(s.Timeout).To(Equal(time.Second))
			})
		})
	})

	Describe("Service", func() {

		Describe("ReceiveHandler", func() {

			Describe("post handling", func() {
				var (
					h            http.Handler
					rec          *httptest.ResponseRecorder
					logWriter    *logLocker
					logger       internal.FieldLogger
					file1, file2 *timedWriter
					service      *handler.Service
					flushDelay   = writer.MinimumDelay + time.Millisecond
				)

				BeforeEach(func() {
					logWriter = &logLocker{
						new(bytes.Buffer),
						new(sync.Mutex),
					}
					logger = internal.WithWriter(logWriter)
					rec = httptest.NewRecorder()

					file1 = &timedWriter{delay: flushDelay}
					file2 = &timedWriter{delay: flushDelay}
					writers := []io.Writer{file1, file2}

					service = &handler.Service{
						Writers: writers,
						Logger:  logger,
					}
					h = http.HandlerFunc(service.RecieveHandler)
				})

				DescribeTable("with bad json object", func(message string, status int, errMsg error) {
					req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
					Expect(err).NotTo(HaveOccurred())

					req.Header.Set("Content-Type", "application/json")

					h.ServeHTTP(rec, req)
					Expect(rec.Code).To(Equal(status))
					Eventually(logWriter.String()).Should(ContainSubstring(errMsg.Error()))
				},
					Entry("empty", `{}`, http.StatusBadRequest, handler.ErrEmptyObject),
					Entry("bogus values", `{"something":"another thing"}`, http.StatusBadRequest, handler.ErrGettingReader),
					Entry("corrupted", `"something":"another thing"}`, http.StatusBadRequest, handler.ErrGettingMap),
					Entry("jason reader error", `{"s":[1,2 3]}`, http.StatusBadRequest, handler.ErrCorruptedJSON),
				)

				Context("handling copy to closed file", func() {
					var (
						req *http.Request
						w   *timedWriter
						h   http.Handler
						err error
					)

					BeforeEach(func() {
						w = new(timedWriter)

						Expect(w.Close()).NotTo(HaveOccurred())

						s := &handler.Service{
							Writers: []io.Writer{w},
							Logger:  logger,
						}
						h = http.HandlerFunc(s.RecieveHandler)

						message := `{"type":"error","message":"blah","timestamp":"2017-01-01"}`
						req, err = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
						Expect(err).NotTo(HaveOccurred())

						req.Header.Set("Content-Type", "application/json")
						h.ServeHTTP(rec, req)
					})

					It("should not error", func() {
						Expect(rec.Code).NotTo(Equal(http.StatusBadRequest))
					})

					It("eventually should log the error", func() {
						Eventually(func() string {
							logWriter.Lock()
							defer logWriter.Unlock()
							return logWriter.String()
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
							return file1.String()
						}, flushDelay+5*time.Second, 0.2).Should(ContainSubstring(errMsg))
					})

					It("should eventually write the log level", func() {
						Eventually(func() string {
							return file1.String()
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
								return file1.String()
							}, flushDelay+time.Second*2, 0.2).Should(ContainSubstring(errMsg))

							Eventually(func() string {
								return file2.String()
							}, flushDelay+time.Second*2, 0.2).Should(ContainSubstring(errMsg))
						})
					})

					Context("when only one of the writers are available", func() {
						BeforeEach(func() {
							file2.Close()
						})

						It("should eventually write the log entry to the available file", func() {

							Eventually(func() string {
								return file1.String()
							}, flushDelay+time.Second).Should(ContainSubstring(errMsg))
						})

						It("should eventually log there was an error on the other file", func() {
							Eventually(func() string {
								logWriter.Lock()
								defer logWriter.Unlock()
								return logWriter.String()
							}, flushDelay+time.Second).Should(ContainSubstring(handler.ErrWritingEntry.Error()))
						})
					})
				})

				Context("handling plain logs and writing to slow writers", func() {
					var (
						file1     *timedWriter
						file2     *timedWriter
						file3     *timedWriter
						delay     = 500 * time.Millisecond
						fastDelay = time.Millisecond
						service   *handler.Service
					)
					errMsg := "this error has occurred"
					kind := "error"
					timestamp := "2017-01-14 19:10:10"

					BeforeEach(func(done Done) {
						file1 = &timedWriter{delay: fastDelay}
						file2 = &timedWriter{delay: fastDelay}
						file3 = &timedWriter{delay: delay}

						service = &handler.Service{
							Logger: logger,
						}

						Expect(handler.WithWriters(file1, file2, file3)(service)).NotTo(HaveOccurred())
						h := http.HandlerFunc(service.RecieveHandler)

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

					}, flushDelay.Seconds()*2) // This request should not take as long as the slow writer

					It("should write the log entry to the fast writers", func(done Done) {

						Eventually(func() string {
							return file1.String()
						}, flushDelay).Should(ContainSubstring(errMsg))

						Eventually(func() string {
							return file2.String()
						}, flushDelay).Should(ContainSubstring(errMsg))

						close(done)

					}, flushDelay.Seconds()+1)

					It("should eventually write the log entry to the slow one", func(done Done) {
						Eventually(func() string {
							return file3.String()
						}, delay*2).Should(ContainSubstring(errMsg))

						Eventually(func() string {
							return logWriter.String()
						}).ShouldNot(ContainSubstring(handler.ErrWritingEntry.Error()))

						close(done)

					}, delay.Seconds()+1)
				})
			})
		})

		Describe("Serve", func() {
			var (
				port       int
				logger     internal.FieldLogger
				loggerBuff *bytes.Buffer
				service    *handler.Service
				stop       chan os.Signal
				errChan    chan error
				timeout    = 500 * time.Millisecond
			)
			BeforeEach(func() {
				addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
				Expect(err).ShouldNot(HaveOccurred())
				tcpConn, err := net.ListenTCP("tcp", addr)
				Expect(err).ShouldNot(HaveOccurred())
				port = tcpConn.Addr().(*net.TCPAddr).Port
				Expect(tcpConn.Close()).NotTo(HaveOccurred())

				loggerBuff = new(bytes.Buffer)
				logger = internal.WithWriter(loggerBuff)
			})

			Describe("setting up the server", func() {

				BeforeEach(func() {
					service = &handler.Service{Logger: logger, Timeout: timeout}
					stop = make(chan os.Signal)
					errChan = make(chan error)
					go service.Serve(stop, errChan, port)
				})

				AfterEach(func(done Done) {
					select {
					case e := <-errChan: // draining the errChan
						Expect(e).To(BeNil())
					case <-time.After(time.Second):
					}

					stop <- os.Interrupt
					Expect(<-errChan).To(Equal(io.EOF))
					close(done)
				}, timeout.Seconds()*10)

				Context("calling serve twice", func() {
					It("should return an error saying the address is already in use", func(done Done) {
						// waiting for the first serve to finish starting up
						time.Sleep(100 * time.Millisecond)
						stop := make(chan os.Signal)
						errChan := make(chan error)
						go service.Serve(stop, errChan, port)
						e := <-errChan
						Expect(e).To(HaveOccurred())
						Expect(e).To(BeAssignableToTypeOf(&net.OpError{}))

						close(done)
					}, timeout.Seconds()*10)
				})

				Context("having set up on a port", func() {

					It("should log the port that it's listening on", func(done Done) {
						Eventually(func() string {
							return loggerBuff.String()
						}, 1).Should(ContainSubstring(strconv.Itoa(port)))
						close(done)
					}, timeout.Seconds()*10)
				})

			})

			Describe("shutting down", func() {

				Context("having called on a port", func() {

					Context("when sending the SIGINT", func() {
						var (
							service *handler.Service
							stop    chan os.Signal
							errChan chan error
						)

						BeforeEach(func() {
							service = &handler.Service{Logger: logger, Timeout: timeout}
							stop = make(chan os.Signal)
							errChan = make(chan error)
							go service.Serve(stop, errChan, port)
							stop <- os.Interrupt
						})

						It("should shut down the server", func() {
							Expect(<-errChan).To(Equal(io.EOF))
						})
						It("should log it has been shut down", func() {
							Eventually(func() string {
								return loggerBuff.String()
							}).Should(ContainSubstring("shutting down"))
						})
					})
				})
			})
		})
	})
})
