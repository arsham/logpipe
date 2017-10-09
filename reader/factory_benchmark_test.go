// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/logpipe/reader"
)

func BenchmarkGetReaderPlain(b *testing.B) {
	benchmarkGetReaderPlain(b, "aksh d", "2017-01-02", 10)
}

func BenchmarkGetReaderPlainMedium(b *testing.B) {
	benchmarkGetReaderPlain(b, "adkjh kjhasdkjh kjhkjahsd", "12:01:10", 100)
}

func BenchmarkGetReaderPlainLarge(b *testing.B) {
	benchmarkGetReaderPlain(b, "sdsd sddkkf j", "2017-01-02 19:00:22", 1000)
}

func benchmarkGetReaderPlain(b *testing.B, msg, timestamp string, length int) {
	b.StopTimer()
	input := []byte(
		fmt.Sprintf(
			`{"type":"error","message":"%s","timestamp":"%s"}`,
			strings.Repeat(msg, length),
			timestamp,
		),
	)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := reader.GetReader(input)
		if err != nil {
			b.Error(err)
		}
	}
}
