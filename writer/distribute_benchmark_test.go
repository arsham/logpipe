// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/arsham/logpipe/writer"
)

func BenchmarkDistribute(b *testing.B) {
	tcs := []struct {
		name    string
		count   int
		message []byte
	}{
		{"One S", 1, []byte("this is a message!")},
		{"Ten S", 10, []byte("this is a message!")},
		{"Hundred S", 100, []byte("this is a message!")},

		{"One M", 1, []byte(strings.Repeat("this is a message!", 10))},
		{"Ten M", 10, []byte(strings.Repeat("this is a message!", 10))},
		{"Hundred M", 100, []byte(strings.Repeat("this is a message!", 10))},

		{"One L", 1, []byte(strings.Repeat("this is a message!", 100))},
		{"Ten L", 10, []byte(strings.Repeat("this is a message!", 100))},
		{"Hundred L", 100, []byte(strings.Repeat("this is a message!", 100))},

		{"One XL", 1, []byte(strings.Repeat("this is a message!", 1000))},
		{"Ten XL", 10, []byte(strings.Repeat("this is a message!", 1000))},
		{"Hundred XL", 100, []byte(strings.Repeat("this is a message!", 1000))},
	}

	for _, tc := range tcs {
		b.Run(tc.name+" normal", func(b *testing.B) {
			// This is used only for comparing how the internals work against
			// the written code.
			for i := 0; i < b.N; i++ {

				d := new(bytes.Buffer)
				r := bytes.NewBuffer(tc.message)

				io.Copy(d, r)
			}
		})

		b.Run(tc.name, func(b *testing.B) {

			for i := 0; i < b.N; i++ {

				writers := make([]io.Writer, tc.count)
				for i := 0; i < tc.count; i++ {
					writers[i] = new(bytes.Buffer)
				}

				d := writer.NewDistribute(writers...)
				r := bytes.NewBuffer(tc.message)

				io.Copy(d, r)
			}
		})
	}
}
