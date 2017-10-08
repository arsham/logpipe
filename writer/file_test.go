// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer_test

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/arsham/logpipe/writer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
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
				fl, err := writer.NewFile(
					writer.WithFileLoc(name),
				)
				Expect(err).NotTo(HaveOccurred())
				defer fl.Close()
				defer os.Remove(name)

				Expect(fl.Name()).To(BeAnExistingFile())
				Expect(f).NotTo(BeNil())
			})

			By("having a file in place")

			It("should not create a new file", func() {
				fl, err := writer.NewFile(
					writer.WithFileLoc(f.Name()),
				)
				defer fl.Close()

				Expect(err).NotTo(HaveOccurred())
				Expect(fl.Name()).To(BeAnExistingFile())
			})
		})

		Describe("generation errors", func() {

			Context("obtaining a File in a non existence place", func() {
				fl, err := writer.NewFile(
					writer.WithFileLoc("/does not exist"),
				)
				It("should error", func() {
					Expect(err).To(HaveOccurred())
					Expect(fl).To(BeNil())
				})
			})

			Context("obtaining a File in a non-writeable place", func() {
				It("should error", func() {
					fl, err := writer.NewFile(
						writer.WithFileLoc(path.Join("/", "testfile")),
					)
					Expect(err).To(HaveOccurred())
					Expect(fl).To(BeNil())
				})
			})

			Context("obtaining a File with a non-writeable file", func() {
				It("should error", func() {
					if err := f.Chmod(0000); err != nil {
						panic(err)
					}
					fl, err := writer.NewFile(
						writer.WithFileLoc(f.Name()),
					)
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

		fl, err = writer.NewFile(
			writer.WithFileLoc(f.Name()),
		)
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
			fl, _ = writer.NewFile(
				writer.WithFileLoc(f.Name()),
			)
			if _, err := fl.Write(line1); err != nil {
				panic(err)
			}
			fl.Flush()
			fl.Close()

			fl, _ := writer.NewFile(writer.WithFileLoc(
				f.Name()),
			)
			n, err := fl.Write(line2)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(len(line2)))
			fl.Flush()

			content, _ := ioutil.ReadAll(f)
			Expect(content).To(ContainSubstring(string(line1)))
			Expect(content).To(ContainSubstring(string(line2)))
		})
	})
})

func setup(t *testing.T) (*os.File, func()) {
	w, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	return w,
		func() {
			w.Close()
			os.Remove(w.Name())
		}
}

type writerMock struct {
	WriteFunc func([]byte) (int, error)
	CloseFunc func() error
	NameFunc  func() string
}

func (w *writerMock) Write(p []byte) (int, error) { return w.WriteFunc(p) }
func (w *writerMock) Close() error                { return w.CloseFunc() }
func (w *writerMock) Name() string                { return w.NameFunc() }

func TestCloseErrors(t *testing.T) {
	e := errors.New("blah")
	f := &writerMock{
		WriteFunc: func(p []byte) (int, error) {
			return -1, e
		},
	}

	fl, _ := writer.NewFile(writer.WithWriter(f))
	fl.Write([]byte("dummy"))
	if err := fl.Close(); errors.Cause(err) != e {
		t.Errorf("want (%s), got(%v)", e, err)
	}
}

func TestWriteErrors(t *testing.T) {
	e := errors.New("blah1")
	f := &writerMock{
		WriteFunc: func(p []byte) (int, error) {
			return 1, e
		},
	}

	buf := bufio.NewWriterSize(f, 2)
	fl, _ := writer.NewFile(writer.WithBufWriter(buf))

	if _, err := fl.Write([]byte("dd")); errors.Cause(err) != e {
		t.Errorf("want (%s), got(%v)", e, err)
	}
}

func TestWithDelay(t *testing.T) {
	fl := &writer.File{}
	if err := writer.WithFlushDelay(0)(fl); err == nil {
		t.Error("want error, got nil")
	}

	m := writer.MinimumDelay - time.Nanosecond
	if err := writer.WithFlushDelay(m)(fl); err == nil {
		t.Error("want error, got nil")
	}
}

func TestSync(t *testing.T) {
	t.Parallel()
	delay := writer.MinimumDelay + 100*time.Millisecond
	w, teardown := setup(t)
	defer teardown()

	fl, err := writer.NewFile(
		writer.WithFlushDelay(delay),
		writer.WithWriter(w),
	)
	if err != nil {
		t.Fatal(err)
	}
	message := []byte("this is the message")
	_, err = fl.Write(message)
	if err != nil {
		t.Fatal(err)
	}

	// should not yet contain the message
	content, err := ioutil.ReadFile(w.Name())
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(content), string(message)) {
		t.Errorf("want no contents, got (%s)", content)
	}

	<-time.After(delay) // waiting to sync
	content, err = ioutil.ReadFile(w.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), string(message)) {
		t.Errorf("want (%s) in contents, got (%s)", message, content)
	}
}
