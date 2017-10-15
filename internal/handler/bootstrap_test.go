// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/internal/config"
	"github.com/arsham/logpipe/internal/handler"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Bootstrap", func() {
	var (
		port       int
		filename   string
		logLevel   string
		configFile string
		writerType = "file"
	)

	JustBeforeEach(func() {
		logLevel = "info"
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		Expect(err).ShouldNot(HaveOccurred())

		tcpConn, err := net.ListenTCP("tcp", addr)
		Expect(err).ShouldNot(HaveOccurred())
		port = tcpConn.Addr().(*net.TCPAddr).Port
		Expect(tcpConn.Close()).NotTo(HaveOccurred())

		file, err := ioutil.TempFile("", "bootstrap_test")
		Expect(err).NotTo(HaveOccurred())
		filename = file.Name()

		c, err := ioutil.TempFile("", "bootstrap_config_test")
		Expect(err).NotTo(HaveOccurred())
		_, err = c.WriteString(fmt.Sprintf(`
app:
  log_level: %s
writers:
  file1:
    type: %s
    location: %s
`, logLevel, writerType, filename))

		Expect(err).NotTo(HaveOccurred())
		configFile = c.Name()

	})

	AfterEach(func() {
		Expect(os.Remove(filename)).NotTo(HaveOccurred())
		Expect(os.Remove(configFile)).NotTo(HaveOccurred())
	})

	Context("when calling the function", func() {

		var (
			serveFunc func(s handler.ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error
			logger    internal.FieldLogger
			logWriter *logLocker
		)

		JustBeforeEach(func() {
			handler.ServeFunc = serveFunc
			logWriter = &logLocker{
				new(bytes.Buffer),
				new(sync.Mutex),
			}

			logger = internal.WithWriter(logWriter)
		})

		AfterEach(func() {
			serveFunc = handler.Serve
		})

		Context("when config file does not exist", func() {
			filename := "/no where to find"
			JustBeforeEach(func() {
				handler.Bootstrap(logger, filename, 8080)
			})

			It("should print an error", func() {
				Eventually(logWriter.String).Should(ContainSubstring(filename))
				Eventually(logWriter.String).Should(ContainSubstring(config.ErrFileNotExist.Error()))
			})

			It("should not start the server", func() {
				Eventually(logWriter.String).ShouldNot(ContainSubstring("running on port"))
			})
		})

		Context("when passing a port number", func() {

			BeforeEach(func() {
				serveFunc = func(s handler.ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error {
					return nil
				}
			})
			JustBeforeEach(func() {
				handler.Bootstrap(logger, configFile, port)
			})

			It("should print the port", func() {
				Eventually(logWriter.String, 2).Should(ContainSubstring(strconv.Itoa(port)))
			})
			It("should return without an error", func() {
				Eventually(logWriter.String).ShouldNot(ContainSubstring("error when serving"))
			})
		})

		Context("when a the logger is nil", func() {
			var thisLogger internal.FieldLogger
			BeforeEach(func() {
				serveFunc = func(s handler.ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error {
					thisLogger = logger
					return nil
				}
			})
			AfterEach(func() { thisLogger = nil })

			JustBeforeEach(func() {
				handler.Bootstrap(nil, configFile, port)
			})

			It("should get the default error logger", func() {
				Eventually(thisLogger).Should(Equal(internal.GetLogger("error")))
			})

			It("should start the server", func() {
				Eventually(logWriter.String).ShouldNot(ContainSubstring("error when serving"))
			})
		})

		Context("when no writer is passed", func() {
			var (
				originType string
				err        = errors.New("this should not happen")
			)
			BeforeEach(func() {
				originType = writerType
				writerType = "does not apply"
				serveFunc = func(s handler.ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error {
					return err
				}
			})

			JustBeforeEach(func() {
				handler.Bootstrap(logger, configFile, port)
			})

			AfterEach(func() { writerType = originType })

			It("should print the ErrNoWriter error", func() {
				Eventually(logWriter.String).Should(ContainSubstring(handler.ErrNoWriter.Error()))
			})
			It("should not start the server", func() {
				Eventually(logWriter.String).ShouldNot(ContainSubstring("running on port"))
			})
		})

		Context("when server has an error while serving", func() {

			var err = errors.New("this should not happen")
			BeforeEach(func() {
				serveFunc = func(s handler.ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error {
					return err
				}
			})
			JustBeforeEach(func() {
				handler.Bootstrap(logger, configFile, port)
			})

			It("should print the error", func() {
				Eventually(logWriter.String).Should(ContainSubstring(err.Error()))
			})
			It("should not start the server", func() {
				Eventually(logWriter.String).Should(ContainSubstring("error when serving"))
			})
		})
	})
})

var _ = Describe("Serve", func() {
	var (
		port      int
		logger    internal.FieldLogger
		logWriter *logLocker
		service   *handler.Service
		stop      chan os.Signal
		errChan   chan error
		timeout   = 500 * time.Millisecond
	)
	BeforeEach(func() {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		Expect(err).ShouldNot(HaveOccurred())
		tcpConn, err := net.ListenTCP("tcp", addr)
		Expect(err).ShouldNot(HaveOccurred())
		port = tcpConn.Addr().(*net.TCPAddr).Port
		Expect(tcpConn.Close()).NotTo(HaveOccurred())

		logWriter = &logLocker{
			new(bytes.Buffer),
			new(sync.Mutex),
		}
		logger = internal.WithWriter(logWriter)
	})

	Describe("setting up the server", func() {

		BeforeEach(func() {
			service = &handler.Service{Logger: logger}
			handler.WithTimeout(timeout)(service)
			stop = make(chan os.Signal)
			errChan = make(chan error)
			go func() {
				errChan <- handler.Serve(service, logger, stop, port)
			}()
		})

		AfterEach(func(done Done) {
			select {
			case e := <-errChan: // draining the errChan
				Expect(e).To(BeNil())
			case <-time.After(time.Second):
			}

			stop <- os.Interrupt
			Expect(<-errChan).To(BeNil())
			close(done)
		}, timeout.Seconds()+2)

		Context("calling serve twice", func() {
			It("should return an error saying the address is already in use", func(done Done) {
				// waiting for the first serve to finish starting up
				time.Sleep(100 * time.Millisecond)
				stop := make(chan os.Signal)
				e := handler.Serve(service, logger, stop, port)
				Expect(e).To(HaveOccurred())
				Expect(e).To(BeAssignableToTypeOf(&net.OpError{}))

				close(done)
			}, timeout.Seconds()+2)
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
					service = &handler.Service{Logger: logger}
					handler.WithTimeout(timeout)(service)
					stop = make(chan os.Signal)
					errChan = make(chan error)
					go func() {
						errChan <- handler.Serve(service, logger, stop, port)
					}()
					stop <- os.Interrupt
				})

				It("should shut down the server", func() {
					Expect(<-errChan).To(BeNil())
				})
				It("should log it has been shut down", func() {
					Eventually(logWriter.String).Should(ContainSubstring("shutting down"))
				})
			})
		})
	})
})
