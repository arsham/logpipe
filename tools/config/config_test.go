// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/arsham/logpipe/tools/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Read", func() {

	Describe("loading configuration", func() {
		var (
			filename string
			input    []byte
			readErr  error
			setting  *config.Setting
		)

		JustBeforeEach(func() {
			f, err := ioutil.TempFile("", "test_config")
			Expect(err).NotTo(HaveOccurred())
			_, err = f.Write(input)
			Expect(err).NotTo(HaveOccurred())
			filename = f.Name()
			setting, readErr = config.Read(filename)
		})

		AfterEach(func() {
			err := os.Remove(filename)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("having non-existent file", func() {
			JustBeforeEach(func() {
				filename := "/does not exist"
				setting, readErr = config.Read(filename)
			})

			It("returns an error", func() {
				Expect(readErr).To(HaveOccurred())
				Expect(readErr).To(Equal(config.ErrFileNotExist))
			})
			Specify("Setting object is nil", func() {
				Expect(setting).To(BeNil())
			})
		})
		Context("having an invalid yaml file", func() {
			BeforeEach(func() {
				input = []byte(`
                :invalid
                `)
			})
			It("returns an error", func() {
				Expect(readErr).To(HaveOccurred())
			})
			Specify("Setting object is nil", func() {
				Expect(setting).To(BeNil())
			})
		})

		Context("having a yaml file without the log level", func() {
			BeforeEach(func() {
				input = []byte(`
writers:
  w1:
    type: elasticsearch
    url: http://localhost:9200
`)
			})

			It("sets the LogLevel to error", func() {
				Expect(readErr).NotTo(HaveOccurred())
				Expect(setting.LogLevel).To(Equal("error"))

			})
		})

		Context("having a yaml file", func() {
			url := "http://localhost:9200"
			indexName := "log-index"
			location := "/dev/null"
			BeforeEach(func() {
				input = []byte(fmt.Sprintf(`
app:
  log_level: info
writers:
  w1:
    type: elasticsearch
    url: %s
    index: %s
  w2:
    type: file
    location: %s
`, url, indexName, location))
			})
			It("loads the LogLevel", func() {
				Expect(readErr).NotTo(HaveOccurred())
				Expect(setting.LogLevel).To(Equal("info"))
			})
			It("loads the Writers", func() {
				Expect(readErr).NotTo(HaveOccurred())
				Expect(setting.Writers).NotTo(BeNil())

				Expect(setting.Writers).To(HaveKey("w1"))
				Expect(setting.Writers["w1"]).To(HaveKey("type"))
				Expect(setting.Writers["w1"]["type"]).To(Equal("elasticsearch"))
				Expect(setting.Writers["w1"]["url"]).To(Equal(url))
				Expect(setting.Writers["w1"]["index"]).To(Equal(indexName))

				Expect(setting.Writers["w2"]).NotTo(BeNil())
				Expect(setting.Writers["w2"]["type"]).To(Equal("file"))
				Expect(setting.Writers["w2"]["location"]).To(Equal(location))
			})
		})

		Context("having a yaml file without writers specified", func() {
			BeforeEach(func() {
				input = []byte(`
app:
  log_level: info
`)
			})
			It("returns an error", func() {
				Expect(readErr).To(HaveOccurred())
				Expect(readErr).To(Equal(config.ErrNoWriters))
				Expect(setting).To(BeNil())
			})
		})

		Context("having a yaml file with a list as log file name", func() {
			BeforeEach(func() {
				input = []byte(`
writers:
  w1:
    type: file
    location: [1,2]
`)
			})
			It("returns an error", func() {
				Expect(readErr).To(HaveOccurred())
				Expect(setting).To(BeNil())
			})
		})
	})
})
