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

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/tools"
	"github.com/arsham/logpipe/tools/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func getRandomPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	Expect(err).ShouldNot(HaveOccurred())

	tcpConn, err := net.ListenTCP("tcp", addr)
	Expect(err).ShouldNot(HaveOccurred())
	port := tcpConn.Addr().(*net.TCPAddr).Port
	Expect(tcpConn.Close()).NotTo(HaveOccurred())
	return port
}

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
		port = getRandomPort()

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
			serveFunc        func(s handler.Server, logger tools.FieldLogger, stop chan os.Signal, port int) error
			logger           tools.FieldLogger
			logWriter        *logLocker
			defaultServeFunc = handler.ServeHTTP
		)

		JustBeforeEach(func() {
			handler.ServeHTTP = serveFunc
			logWriter = &logLocker{
				new(bytes.Buffer),
				new(sync.Mutex),
			}

			logger = tools.WithWriter(logWriter)
		})

		AfterEach(func() {
			handler.ServeHTTP = defaultServeFunc
		})

		Context("when config file does not exist", func() {
			var (
				filename = "/no where to find"
				err      error
			)
			JustBeforeEach(func() {
				err = handler.Bootstrap(logger, filename, 8080)
			})

			It("should return ErrFileNotExist error", func() {
				Expect(errors.Cause(err)).To(Equal(config.ErrFileNotExist))
			})
			It("should mention the file name", func() {
				Expect(err.Error()).To(ContainSubstring(filename))
			})
		})

		Context("when passing a port number", func() {

			BeforeEach(func() {
				serveFunc = func(s handler.Server, logger tools.FieldLogger, stop chan os.Signal, port int) error {
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
			var thisLogger tools.FieldLogger
			BeforeEach(func() {
				serveFunc = func(s handler.Server, logger tools.FieldLogger, stop chan os.Signal, port int) error {
					thisLogger = logger
					return nil
				}
			})
			AfterEach(func() { thisLogger = nil })

			JustBeforeEach(func() {
				handler.Bootstrap(nil, configFile, port)
			})

			It("should get the default error logger", func() {
				Eventually(thisLogger).Should(Equal(tools.GetLogger("error")))
			})

			It("should start the server", func() {
				Eventually(logWriter.String).ShouldNot(ContainSubstring("error when serving"))
			})
		})

		Context("when no writer is passed", func() {
			var (
				originType  string
				expectedErr = errors.New("this should not happen")
				err         error
			)
			BeforeEach(func() {
				originType = writerType
				writerType = "does not apply"
				serveFunc = func(s handler.Server, logger tools.FieldLogger, stop chan os.Signal, port int) error {
					return expectedErr
				}
			})

			JustBeforeEach(func() {
				err = handler.Bootstrap(logger, configFile, port)
			})

			AfterEach(func() { writerType = originType })

			It("should return the ErrNoWriter error", func() {
				Expect(errors.Cause(err)).To(Equal(handler.ErrNoWriter))
			})
		})

		Context("when server has an error while serving", func() {

			var (
				expectedErr = errors.New("this should not happen")
				err         error
			)
			BeforeEach(func() {
				serveFunc = func(s handler.Server, logger tools.FieldLogger, stop chan os.Signal, port int) error {
					return expectedErr
				}
			})
			JustBeforeEach(func() {
				err = handler.Bootstrap(logger, configFile, port)
			})

			It("should return the error", func() {
				Expect(errors.Cause(err)).To(Equal(expectedErr))
			})
		})
	})
})

var _ = Describe("Serve", func() {
	var (
		port      int
		logger    tools.FieldLogger
		logWriter *logLocker
		timeout   = 500 * time.Millisecond
	)

	BeforeEach(func() {
		port = getRandomPort()

		logWriter = &logLocker{
			new(bytes.Buffer),
			new(sync.Mutex),
		}
		logger = tools.WithWriter(logWriter)
	})

	Describe("setting up the server", func() {
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
				errChan <- handler.ServeHTTP(service, logger, stop, port)
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
				e := handler.ServeHTTP(service, logger, stop, port)
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
						errChan <- handler.ServeHTTP(service, logger, stop, port)
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
