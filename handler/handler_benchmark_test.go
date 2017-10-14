// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/reader"
)

func BenchmarkHandler(b *testing.B) {
	tc := []struct {
		name         string
		wLen, msgLen int
	}{
		{"Small", 1, 10},
		{"Medium", 1, 100},
		{"Large", 1, 1000},
	}

	s := &handler.Service{
		Logger: internal.DiscardLogger(),
	}

	handler := http.HandlerFunc(s.RecieveHandler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	for _, t := range tc {
		b.Run(t.name, func(b *testing.B) {
			b.StopTimer()

			for i := 0; i < t.wLen; i++ {
				buf := new(timedWriter)
				s.Writers = append(s.Writers, buf)
			}

			errMsg := strings.Repeat("afadf ", t.msgLen)
			message := fmt.Sprintf(`{"type":"error","message":"%s","timestamp":"%s"}`,
				errMsg,
				time.Now().Format(reader.TimestampFormat),
			)

			req, err := http.NewRequest("POST", ts.URL, bytes.NewBuffer([]byte(message)))
			if err != nil {
				b.Fatal(err)
			}

			client := &http.Client{}
			b.StartTimer()

			for i := 0; i < b.N; i++ {
				_, err := client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
