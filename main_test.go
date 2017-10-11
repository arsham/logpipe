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
		program    string
		args       []string
		port       int
		session    *gexec.Session
		envAdd     func([]string) []string
		filename1  string
		filename2  string
		logLevel   string
		configFile string
	)

	BeforeSuite(func() {
		var err error
		program, err = gexec.Build("github.com/arsham/logpipe")
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	JustBeforeEach(func() {
		logLevel = "info"
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		Expect(err).ShouldNot(HaveOccurred())

		tcpConn, err := net.ListenTCP("tcp", addr)
		Expect(err).ShouldNot(HaveOccurred())
		port = tcpConn.Addr().(*net.TCPAddr).Port
		err = tcpConn.Close()
		Expect(err).NotTo(HaveOccurred())

		f1, err := ioutil.TempFile("", "main_test")
		Expect(err).NotTo(HaveOccurred())
		filename1 = f1.Name()

		f2, err := ioutil.TempFile("", "main_test")
		Expect(err).NotTo(HaveOccurred())
		filename2 = f2.Name()

		c, err := ioutil.TempFile("", "main_config_test")
		Expect(err).NotTo(HaveOccurred())
		_, err = c.WriteString(fmt.Sprintf(`
app:
  log_level: %s
writers:
  file1:
    type: file
    location: %s
  file2:
    type: file
    location: %s
`, logLevel, filename1, filename2))

		Expect(err).NotTo(HaveOccurred())
		configFile = c.Name()

		command := exec.Command(program, args...)
		env := os.Environ()

		if envAdd != nil {
			env = envAdd(env)
		}

		command.Env = env
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.Remove(filename1)).NotTo(HaveOccurred())
		Expect(os.Remove(filename2)).NotTo(HaveOccurred())
		Expect(os.Remove(configFile)).NotTo(HaveOccurred())
	})

	Context("running the application without config file argument", func() {
		It("should complain", func() {
			Eventually(session.Err).Should(gbytes.Say("config-file"))
		})
	})

	Describe("loading environment variables", func() {

		Context("having port set in the environment", func() {
			BeforeEach(func() {
				envAdd = func(env []string) []string {
					env = append(env, fmt.Sprintf("PORT=%d", port))
					env = append(env, fmt.Sprintf("CONFIGFILE=%s", configFile))
					return env
				}
			})
			AfterEach(func() {
				envAdd = nil
			})

			It("should apply the port and logfile", func() {
				Eventually(session.Err).ShouldNot(gbytes.Say("config-file"))

				url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
				req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"message":"blah"}`)))
				Expect(err).NotTo(HaveOccurred())

				client := &http.Client{}
				Eventually(func() error {
					_, err := client.Do(req)
					return err
				}, 0.4, 0.1).ShouldNot(HaveOccurred())

			})
		})
	})

	Describe("config file", func() {

		Context("having a config file path set in the environment", func() {
			BeforeEach(func() {
				envAdd = func(env []string) []string {
					return append(env, fmt.Sprintf("CONFIGFILE=%s", configFile))
				}
			})

			AfterEach(func() {
				envAdd = nil
			})

			It("should not error", func() {
				Eventually(session.Err).ShouldNot(gbytes.Say("config-file"))
			})
		})
	})

	Describe("setting up the handlers", func() {

		BeforeEach(func() {
			envAdd = func(env []string) []string {
				env = append(env, fmt.Sprintf("PORT=%d", port))
				env = append(env, fmt.Sprintf("CONFIGFILE=%s", configFile))
				return env
			}
		})
		AfterEach(func() {
			envAdd = nil
		})

		Context("with given port number", func() {

			It("should set up a server on that port", func() {
				url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
				req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{"message":"blah"}`)))
				Expect(err).NotTo(HaveOccurred())

				client := &http.Client{}
				Eventually(func() error {
					_, err := client.Do(req)
					return err
				}, 0.4, 0.1).ShouldNot(HaveOccurred())
			})
		})
	})

	Describe("Loading writers", func() {
		message := "this message should be in the log file"

		BeforeEach(func() {
			envAdd = func(env []string) []string {
				env = append(env, fmt.Sprintf("PORT=%d", port))
				env = append(env, fmt.Sprintf("CONFIGFILE=%s", configFile))
				return env
			}
		})
		AfterEach(func() {
			envAdd = nil
		})

		Context("having two file loggers in the config file", func() {

			BeforeEach(func() {
				logLevel = "debug"
			})

			JustBeforeEach(func() {
				url := "http://127.0.0.1:" + strconv.Itoa(port) + "/"

				req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(fmt.Sprintf(`{"message":"%s"}`, message))))
				Expect(err).NotTo(HaveOccurred())

				client := &http.Client{}
				Eventually(func() error {
					_, err := client.Do(req)
					return err
				}, 0.4, 0.1).ShouldNot(HaveOccurred())
			})

			It("should write the line in both files", func() {

				Eventually(func() string {
					content, err := ioutil.ReadFile(filename1)
					Expect(err).NotTo(HaveOccurred())
					return string(content)
				}, 2, 0.1).Should(ContainSubstring(message))

				Eventually(func() string {
					content, err := ioutil.ReadFile(filename2)
					Expect(err).NotTo(HaveOccurred())
					return string(content)
				}, 2, 0.1).Should(ContainSubstring(message))

			})

			Specify("the log level is present in the log", func() {
				Eventually(func() string {
					content, err := ioutil.ReadFile(filename1)
					Expect(err).NotTo(HaveOccurred())
					return string(content)
				}, 2, 0.1).Should(ContainSubstring(logLevel))

				Eventually(func() string {
					content, err := ioutil.ReadFile(filename2)
					Expect(err).NotTo(HaveOccurred())
					return string(content)
				}, 2, 0.1).Should(ContainSubstring(logLevel))
			})
		})
	})
})
