// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/arsham/logpipe/reader"
)

func BenchmarkPlainRead(b *testing.B) {
	benchmarkPlainRead(b, 10)
}

func BenchmarkPlainReadLarge(b *testing.B) {
	benchmarkPlainRead(b, 200)
}

func BenchmarkPlainReadIoCopy(b *testing.B) {
	benchmarkPlainReadIoCopy(b, 10)
}

func BenchmarkPlainReadIoCopyLarge(b *testing.B) {
	benchmarkPlainReadIoCopy(b, 200)
}

func BenchmarkPlainReadIoutilReadAll(b *testing.B) {
	benchmarkPlainReadIoutilReadAll(b, 10)
}

func BenchmarkPlainReadIoutilReadAllLarge(b *testing.B) {
	benchmarkPlainReadIoutilReadAll(b, 200)
}

func benchmarkPlainRead(b *testing.B, count int) {
	message := strings.Repeat("this is a long message", count)
	for n := 0; n < b.N; n++ {
		p := &reader.Plain{
			Kind:      "error",
			Message:   message,
			Timestamp: time.Now(),
		}
		buf := make([]byte, len(message)+100)
		var _, _ = p.Read(buf)
	}
}

func benchmarkPlainReadIoCopy(b *testing.B, count int) {
	message := strings.Repeat("this is a long message", count)
	for n := 0; n < b.N; n++ {
		p := &reader.Plain{
			Kind:      "error",
			Message:   message,
			Timestamp: time.Now(),
		}
		buf := &bytes.Buffer{}
		var _, _ = io.Copy(buf, p)
	}
}

func benchmarkPlainReadIoutilReadAll(b *testing.B, count int) {
	message := strings.Repeat("this is a long message", count)
	for n := 0; n < b.N; n++ {
		p := &reader.Plain{
			Kind:      "error",
			Message:   message,
			Timestamp: time.Now(),
		}
		var _, _ = ioutil.ReadAll(p)
	}
}
