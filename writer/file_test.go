// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer_test

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"

	"github.com/arsham/logpipe/writer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating the log file", func() {

	Context("creating a File from scratch", func() {
		var f *os.File

		BeforeEach(func() {
			var err error
			cwd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			f, err = ioutil.TempFile(cwd, "test")
			if err != nil {
				panic(err)
			}
		})

		AfterEach(func() {
			f.Close()
			os.Remove(f.Name())
		})

		Context("starting the reader", func() {

			It("should create a new file", func() {
				name := f.Name()
				os.Remove(name)
				fl, err := writer.NewFile(name)
				Expect(err).NotTo(HaveOccurred())
				defer fl.Close()
				defer os.Remove(name)

				Expect(fl.Name()).To(BeAnExistingFile())
				Expect(f).NotTo(BeNil())
			})

			By("having a file in place")

			It("should not create a new file", func() {
				fl, err := writer.NewFile(f.Name())
				defer fl.Close()

				Expect(err).NotTo(HaveOccurred())
				Expect(fl.Name()).To(BeAnExistingFile())
			})
		})

		Describe("generation errors", func() {

			Context("obtaining a File in a non existence place", func() {
				fl, err := writer.NewFile("/does not exist")
				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(fl).To(BeNil())
				})
			})

			Context("obtaining a File in a non-writeable place", func() {
				It("should error", func() {
					fl, err := writer.NewFile(path.Join("/", "testfile"))
					Expect(err).To(HaveOccurred())
					Expect(fl).To(BeNil())
				})
			})

			Context("obtaining a File with a non-writeable file", func() {
				It("should error", func() {
					if err := f.Chmod(0000); err != nil {
						panic(err)
					}
					fl, err := writer.NewFile(f.Name())
					Expect(err).To(HaveOccurred())
					Expect(fl).To(BeNil())
				})
			})
		})
	})
})

var _ = Describe("Writing logs", func() {
	var (
		f  *os.File
		fl *writer.File
	)

	BeforeEach(func() {
		var err error
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		f, err = ioutil.TempFile(cwd, "test")
		if err != nil {
			panic(err)
		}

		fl, err = writer.NewFile(f.Name())
		if err != nil {
			panic(err)
		}
	})

	AfterEach(func() {
		fl.Close()
		os.Remove(f.Name())
	})

	Context("having a File object", func() {

		Context("writing one line", func() {
			line1 := []byte("line 1 contents")
			Specify("line should appear in the file", func() {
				n, err := fl.Write(line1)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(line1)))
				fl.Flush()
				content, _ := ioutil.ReadAll(f)
				Expect(content).To(ContainSubstring(string(line1)))
			})
		})

		Context("writing two lines", func() {
			line1 := []byte("line 1 contents")
			line2 := []byte("line 2 contents")
			Specify("both lines should appear in the file", func() {
				n, err := fl.Write(line1)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(line1)))

				n, err = fl.Write(line2)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(len(line2)))

				fl.Flush()
				content, _ := ioutil.ReadAll(f)
				Expect(content).To(ContainSubstring(string(line1)))
				Expect(content).To(ContainSubstring(string(line2)))
			})

			DescribeTable("should contain two lines", func(n int) {
				for range make([]struct{}, n) { //writing n lines
					fl.Write(line1)
				}
				fl.Flush()

				scanner := bufio.NewScanner(bufio.NewReader(f))
				counter := 0
				for scanner.Scan() {
					counter++
				}
				Expect(counter).To(Equal(n))
			},
				Entry("0", 0),
				Entry("1", 1),
				Entry("2", 2),
				Entry("10", 10),
				Entry("20", 20),
			)
		})
	})

	Context("appending to an existing log file", func() {
		line2 := []byte("line 2 contents")

		It("should retain the existing contents", func() {
			line1 := []byte("line 1 contents")
			fl, _ = writer.NewFile(f.Name())
			if _, err := fl.Write(line1); err != nil {
				panic(err)
			}
			fl.Flush()
			fl.Close()

			fl, _ := writer.NewFile(f.Name())
			n, err := fl.Write(line2)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(line2)))
			fl.Flush()

			content, err := ioutil.ReadAll(f)
			Expect(content).To(ContainSubstring(string(line1)))
			Expect(content).To(ContainSubstring(string(line2)))
		})
	})
})
