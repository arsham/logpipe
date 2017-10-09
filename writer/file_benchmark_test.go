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

func BenchmarkSlimShortFileWrites(b *testing.B) {
	benchmarkFileWrites(b, 100, 10000)
}

func BenchmarkFatShortFileWrites(b *testing.B) {
	benchmarkFileWrites(b, 100*100, 10000)
}

func BenchmarkSlimTallFileWrites(b *testing.B) {
	benchmarkFileWrites(b, 100, 10000*10)
}

func BenchmarkFatTallFileWrites(b *testing.B) {
	benchmarkFileWrites(b, 100*100, 10000*10)
}

func BenchmarkFatVeryTallFileWrites(b *testing.B) {
	benchmarkFileWrites(b, 100*100, 10000*100)
}

func benchmarkFileWrites(b *testing.B, col, row int) {
	rowString := bytes.Repeat([]byte("a"), col)
	for n := 0; n < b.N; n++ {
		fl, err := writer.NewFile(writer.WithFileLoc(os.DevNull))
		defer fl.Flush()
		if err != nil {
			panic(err)
		}
		for i := 0; i < row; i++ {
			fl.Write(rowString)
		}
	}
}
