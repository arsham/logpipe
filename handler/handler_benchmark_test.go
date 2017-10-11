// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/reader"
	"github.com/arsham/logpipe/writer"
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
	for _, t := range tc {
		b.Run(t.name, func(b *testing.B) {
			b.StopTimer()

			s := &handler.Service{}

			for i := 0; i < t.wLen; i++ {
				f, err := ioutil.TempFile("", "testing_bechmark")
				if err != nil {
					b.Fatal(err)
				}
				defer func() {
					os.Remove(f.Name())
				}()
				file, err := writer.NewFile(
					writer.WithWriter(f),
				)
				if err != nil {
					b.Fatal(err)
				}
				s.Writers = append(s.Writers, file)
			}

			handler := http.HandlerFunc(s.RecieveHandler)
			ts := httptest.NewServer(handler)
			defer ts.Close()

			errMsg := strings.Repeat("afadf", t.msgLen)
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
				_, err = client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
