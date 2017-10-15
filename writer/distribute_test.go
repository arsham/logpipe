// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer_test

import (
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/arsham/logpipe/writer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

type writerStub struct {
	c     []byte
	delay time.Duration
	sync.RWMutex
	writeFunc func([]byte) (int, error)
}

func (w *writerStub) Write(p []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	if w.writeFunc != nil {
		return w.writeFunc(p)
	}
	time.Sleep(w.delay)
	return copy(w.c, p), nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

var _ = Describe("Distribute", func() {

	DescribeTable("having multiple writers", func(input []byte, expLen int, expErr error, writers ...io.Writer) {

		d := writer.NewDistribute(writers...)
		n, err := d.Write(input)

		By("error should be equal to what is expected")
		if err == nil {
			Expect(err).To(BeNil())
		} else {
			Expect(err.Error()).To(ContainSubstring(expErr.Error()))
			return
		}

		By("writing the same payload to all")
		for _, w := range writers {
			if w == nil || reflect.ValueOf(w).IsNil() {
				continue
			}
			wrote := w.(*writerStub)
			writtenC := wrote.c[:min(expLen, len(wrote.c))]
			Expect(string(input)).
				To(ContainSubstring(string(writtenC)), "expected to see (%+v) to be in (%+v)", writtenC, input)
		}

		By("returned length should be the same as the buffer lengths'")
		Expect(n).To(Equal(expLen))

	},
		Entry("one writer with enough space", []byte("aaaaa "), 6, nil, &writerStub{c: make([]byte, 6)}),
		Entry("one writer with space space", []byte("aaaaa "), 2, nil, &writerStub{c: make([]byte, 2)}),
		Entry("two identical writers with enough space",
			[]byte("aaaaa "), 6, nil,
			&writerStub{c: make([]byte, 6)},
			&writerStub{c: make([]byte, 6)},
		),
		Entry("two identical writers with less space",
			[]byte("aaaaa "), 2, nil,
			&writerStub{c: make([]byte, 2)},
			&writerStub{c: make([]byte, 2)},
		),
		Entry("two writers, different space with enough space",
			[]byte("aaaaa "), 6, nil,
			&writerStub{c: make([]byte, 6)},
			&writerStub{c: make([]byte, 8)},
		),
		Entry("two writers, different space with less space",
			[]byte("aaaaa "), 5, nil,
			&writerStub{c: make([]byte, 2)},
			&writerStub{c: make([]byte, 5)},
		),
		Entry("9 identical writers with enough space",
			[]byte("aaaaa "), 6, nil,
			&writerStub{c: make([]byte, 6)}, &writerStub{c: make([]byte, 6)}, &writerStub{c: make([]byte, 6)},
			&writerStub{c: make([]byte, 6)}, &writerStub{c: make([]byte, 6)}, &writerStub{c: make([]byte, 6)},
			&writerStub{c: make([]byte, 6)}, &writerStub{c: make([]byte, 6)}, &writerStub{c: make([]byte, 6)},
		),
		Entry("nil writer",
			[]byte("aaaaa "), 6, nil,
			&writerStub{c: make([]byte, 6)},
			(*writerStub)(nil),
		),
		Entry("panic during writes",
			[]byte("aaaaa "), 6, errors.New("this is the panic"),
			&writerStub{
				c: make([]byte, 6),
				writeFunc: func([]byte) (int, error) {
					panic("this is the panic")
					return 0, nil
				},
			},
		),
		Entry("panicing with error during writes",
			[]byte("aaaaa "), 6, errors.New("this is the panic"),
			&writerStub{
				c: make([]byte, 6),
				writeFunc: func([]byte) (int, error) {
					panic(errors.New("this is the panic"))
					return 0, nil
				},
			},
		),
	)

	Describe("slow writers", func() {
		Context("having a fast writer and a slow writer", func() {
			var (
				fast, slow *writerStub
				w          *writer.Distribute
				input      = []byte("this is the message")
				slowDelay  = time.Second
			)

			BeforeEach(func() {
				slow = &writerStub{c: make([]byte, len(input)), delay: slowDelay}
				fast = &writerStub{c: make([]byte, len(input))}
				w = writer.NewDistribute(slow, fast)
			})

			It("should write to the faster one as soon as it can", func() {
				go func() {
					_, err := w.Write(input)
					Expect(err).NotTo(HaveOccurred())
				}()
				Eventually(func() string {
					fast.RLock()
					defer fast.RUnlock()
					return string(fast.c)
				}, 0.1).Should(Equal(string(input)))
			})

			It("eventually should write to the slow one", func() {
				go func() {
					_, err := w.Write(input)
					Expect(err).NotTo(HaveOccurred())
				}()
				Eventually(func() string {
					slow.RLock()
					defer slow.RUnlock()
					return string(slow.c)
				}, slowDelay.Seconds()+0.01).Should(Equal(string(input)))
			})
		})
	})
})
