// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package writer_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/arsham/logpipe/writer"
)

func BenchmarkFileWrites(b *testing.B) {
	tc := []struct {
		name     string
		col, row int
	}{
		{"Slim Short", 100, 10000},
		{"Fat Short", 100 * 100, 10000},
		{"Slim Tall", 100, 10000 * 10},
		{"Fat Tall", 100 * 100, 10000 * 10},
		{"Fat Very Tall", 100 * 100, 10000 * 100},
	}

	for _, t := range tc {
		b.Run(t.name, func(b *testing.B) {
			rowString := bytes.Repeat([]byte("a"), t.col)
			for n := 0; n < b.N; n++ {
				fl, err := writer.NewFile(writer.WithLocation(os.DevNull))
				defer fl.Flush()

				if err != nil {
					panic(err)
				}

				for i := 0; i < t.row; i++ {
					fl.Write(rowString)
				}
			}
		})
	}
}
