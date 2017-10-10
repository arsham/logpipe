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
	"strings"
	"testing"
	"time"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/reader"
	"github.com/arsham/logpipe/writer"
)

func BenchmarkHandler(b *testing.B) {
	benchmarkHandler(b, 10)
}

func BenchmarkHandlerMedium(b *testing.B) {
	benchmarkHandler(b, 100)
}

func BenchmarkHandlerLarge(b *testing.B) {
	benchmarkHandler(b, 1000)
}

func benchmarkHandler(b *testing.B, msgLen int) {
	b.StopTimer()

	f, err := ioutil.TempFile("", "testing")
	if err != nil {
		b.Fatal(err)
	}
	file, err := writer.NewFile(
		writer.WithWriter(f),
	)
	if err != nil {
		b.Fatal(err)
	}
	s := &handler.Service{Writer: file}

	handler := http.HandlerFunc(s.RecieveHandler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	errMsg := strings.Repeat("afadf", msgLen)
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
}
