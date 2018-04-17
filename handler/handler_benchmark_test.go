// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/reader"
	"github.com/arsham/logpipe/tools"
)

func BenchmarkHandlerPostMode(b *testing.B) {

	makeWriters := func(count int, delay1, delay2 time.Duration) []io.Writer {
		writers := make([]io.Writer, count)
		for i := 0; i < count; i++ {
			d := delay2
			if math.Mod(float64(i), 2) == 0 {
				d = delay1
			}
			buf := &timedWriter{
				delay: d,
			}
			writers[i] = buf
		}
		return writers
	}

	tc := []struct {
		name    string
		msgLen  int
		writers []io.Writer
	}{
		{"Small M - 1", 10, makeWriters(1, time.Nanosecond, time.Millisecond*1900)},
		{"Medium M - 1", 100, makeWriters(1, time.Nanosecond, time.Millisecond*1900)},
		{"Large M - 1", 1000, makeWriters(1, time.Nanosecond, time.Millisecond*1900)},
		{"Small M - 10", 10, makeWriters(10, time.Nanosecond, time.Millisecond*1900)},
		{"Medium M - 100", 100, makeWriters(100, time.Nanosecond, time.Millisecond*1900)},
		{"Large M - 100", 1000, makeWriters(100, time.Nanosecond, time.Millisecond*1900)},
	}

	for _, t := range tc {
		b.Run(t.name, func(b *testing.B) {

			s := &handler.Service{
				Logger:  tools.DiscardLogger(),
				Writers: t.writers,
			}

			ts := httptest.NewServer(s)
			defer ts.Close()

			errMsg := string(randBytes(t.msgLen))
			message := fmt.Sprintf(`{"type":"error","message":"%s","timestamp":"%s"}`,
				errMsg,
				time.Now().Format(reader.TimestampFormat),
			)

			req, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer([]byte(message)))
			if err != nil {
				b.Fatal(err)
			}

			client := &http.Client{}

			for i := 0; i < b.N; i++ {
				_, err := client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkHandlerInternalMode(b *testing.B) {

	makeWriters := func(count int, delay1, delay2 time.Duration) []io.Writer {
		writers := make([]io.Writer, count)
		for i := 0; i < count; i++ {
			d := delay2
			if math.Mod(float64(i), 2) == 0 {
				d = delay1
			}
			buf := &timedWriter{
				delay: d,
			}
			writers[i] = buf
		}
		return writers
	}

	tc := []struct {
		name    string
		msgLen  int
		writers []io.Writer
	}{
		{"Small M - 1", 10, makeWriters(1, time.Nanosecond, time.Millisecond*1900)},
		{"Medium M - 1", 100, makeWriters(1, time.Nanosecond, time.Millisecond*1900)},
		{"Large M - 1", 1000, makeWriters(1, time.Nanosecond, time.Millisecond*1900)},
		{"Small M - 10", 10, makeWriters(10, time.Nanosecond, time.Millisecond*1900)},
		{"Medium M - 100", 100, makeWriters(100, time.Nanosecond, time.Millisecond*1900)},
		{"Large M - 100", 1000, makeWriters(100, time.Nanosecond, time.Millisecond*1900)},
	}

	for _, t := range tc {
		b.Run(t.name, func(b *testing.B) {

			s := &handler.Service{
				Logger:  tools.DiscardLogger(),
				Writers: t.writers,
			}

			errMsg := string(randBytes(t.msgLen))
			message := fmt.Sprintf(`{"type":"error","message":"%s","timestamp":"%s"}`,
				errMsg,
				time.Now().Format(reader.TimestampFormat),
			)

			req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(message)))
			if err != nil {
				b.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			for i := 0; i < b.N; i++ {
				s.ServeHTTP(rec, req)
				if rec.Code != http.StatusOK {
					b.Error("bad request")
				}
			}
		})
	}
}
