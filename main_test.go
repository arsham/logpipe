// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var (
		program string
		port    int
	)

	BeforeSuite(func() {
		var err error
		program, err = gexec.Build("github.com/arsham/logpipe")
		Expect(err).ShouldNot(HaveOccurred())

	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	Context("running the application without log file argument", func() {
		It("should complain", func() {
			command := exec.Command(program, "-port=8899")
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session.Err).Should(gbytes.Say("log file"))
		})
	})

	XContext("setting up the logger", func() {
	})

	Describe("setting up the handlers", func() {
		BeforeEach(func() {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
			Expect(err).ShouldNot(HaveOccurred())

			tcpConn, err := net.ListenTCP("tcp", addr)
			Expect(err).ShouldNot(HaveOccurred())
			defer tcpConn.Close()
			port = tcpConn.Addr().(*net.TCPAddr).Port
		})

		Context("with given port number", func() {
			It("should set up a server on that port", func() {
				command := exec.Command(program, "-port="+strconv.Itoa(port), "-logfile="+os.DevNull)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ShouldNot(HaveOccurred())
				Eventually(session.Err).ShouldNot(gbytes.Say("log file"))

				url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
				req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"message":"blah"}`)))
				Expect(err).NotTo(HaveOccurred())

				client := &http.Client{}
				Eventually(func() error {
					_, err := client.Do(req)
					return err
				}, 0.4, 0.1).ShouldNot(HaveOccurred())
			})

			Context("sending payload", func() {
				Specify("the log should be written on the file", func() {
					tmp, err := ioutil.TempFile("", "handler_test")
					Expect(err).NotTo(HaveOccurred())
					defer os.Remove(tmp.Name())

					command := exec.Command(program, "-port="+strconv.Itoa(port), "-logfile="+tmp.Name())
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Expect(err).ShouldNot(HaveOccurred())
					Eventually(session.Err).ShouldNot(gbytes.Say("log file"))

					url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
					message := "this message to be written"

					req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(fmt.Sprintf(`{"message":"%s"}`, message))))
					Expect(err).NotTo(HaveOccurred())

					client := &http.Client{}
					Eventually(func() error {
						_, err := client.Do(req)
						return err
					}, 0.4, 0.1).ShouldNot(HaveOccurred())

					Eventually(func() []byte {
						content, err := ioutil.ReadFile(tmp.Name())
						Expect(err).NotTo(HaveOccurred())
						return content
					}, 2, 0.1).Should(ContainSubstring(message))

				})
			})
		})
	})
})
