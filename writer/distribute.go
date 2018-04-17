// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer

import (
	"fmt"
	"io"
	"reflect"
	"sync"
)

// Distribute is a concurrent writer.
type Distribute struct {
	sync.Mutex
	writers []io.Writer
}

// NewDistribute returns no errors. It dismissed the writers with nil values.
func NewDistribute(writers ...io.Writer) *Distribute {
	var wrs []io.Writer
	for _, w := range writers {
		if w != nil && !reflect.ValueOf(w).IsNil() {
			wrs = append(wrs, w)
		}
	}
	return &Distribute{
		writers: wrs,
	}
}

// used for sending the results back.
type result struct {
	n   int
	err error
}

// Write writes the input bytes into the writers concurrently. It returns an
// error if any of the writers fail to write.
func (c *Distribute) Write(p []byte) (int, error) {
	var (
		wg  sync.WaitGroup
		res = make(chan result, len(c.writers))
		err error
		n   int
	)

	c.Lock()
	defer c.Unlock()

	for _, w := range c.writers {
		wg.Add(1)
		go func(w io.Writer) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						res <- result{0, err}
						return
					}
					res <- result{0, fmt.Errorf("panic: %v", r)}
				}
			}()
			n, err := w.Write(p)
			res <- result{n, err}
		}(w)
	}
	wg.Wait()
	close(res)

	for r := range res {
		if r.err != nil {
			err = r.err
			break
		}
		n = max(n, r.n)
	}
	return n, err
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
