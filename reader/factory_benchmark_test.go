// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/arsham/logpipe/reader"
	"github.com/arsham/logpipe/tools"
)

func BenchmarkGetReaderPlain(b *testing.B) {
	tc := []struct {
		name           string
		msg, timestamp string
		length         int
	}{
		{"Small", "aksh d", "2017-01-02", 10},
		{"Medium", "adkjh kjhasdkjh kjhkjahsd", "2013-01-10", 100},
		{"Large", "sdsd sddkkf j", "2017-01-02 19:00:22", 1000},
	}

	for _, t := range tc {
		b.Run(t.name, func(b *testing.B) {
			logger := tools.DiscardLogger()
			input := []byte(
				fmt.Sprintf(
					`{"type":"error","message":"%s","timestamp":"%s"}`,
					strings.Repeat(t.msg, t.length),
					t.timestamp,
				),
			)

			for i := 0; i < b.N; i++ {
				_, err := reader.GetReader(bytes.NewReader(input), logger)
				if err != nil {
					b.Error(err)
				}
			}
		})
	}
}
