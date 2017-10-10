// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/arsham/logpipe/internal"
)

// TimestampFormat is the default formatting defined for logs.
var TimestampFormat = time.RFC3339

// Plain implements the io.Reader interface and can read a json object and
// output a one line error message. For example:
//
//     {
//         "type": "error",
//         "timestamp": "2017-10-09 10:45:00",
//         "message": "something happened",
//     }
//
// Will become:
//     [2017-10-09 10:45:00] [ERROR] something happened
//
type Plain struct {
	Kind      string
	Message   string
	Timestamp time.Time
	Logger    internal.FieldLogger

	once     sync.Once
	compiled []byte
	current  int //current position on reading the message
}

func (p *Plain) Read(b []byte) (int, error) {
	var n int

	if p.Timestamp.Equal(time.Time{}) {
		p.Logger.Error(ErrNilTimestamp)
		return 0, ErrNilTimestamp
	}

	if p.Message == "" {
		p.Logger.Error(ErrEmptyMessage)
		return 0, ErrEmptyMessage
	}

	if p.Kind == "" {
		p.Logger.Debugf("falling back to info: %s", b)
		p.Kind = INFO
	}

	if len(p.compiled) > 0 && p.current >= len(p.compiled) {
		return 0, io.EOF
	}

	p.once.Do(func() {
		t := p.Timestamp.Format(TimestampFormat)
		l := 6 + len(p.Kind) + len(p.Message) + len(t)
		buf := bytes.NewBuffer(make([]byte, l))
		buf.Reset()

		inputs := []string{
			"[",
			t,
			"] [",
			strings.ToUpper(p.Kind),
			"] ",
			p.Message,
		}
		for _, in := range inputs {
			nn, _ := buf.WriteString(in)
			n += nn
		}
		p.compiled = buf.Bytes()
	})

	end := len(p.compiled)
	if end > len(b) {
		end = len(b) + p.current
		if end > len(p.compiled) {
			end = len(p.compiled)
		}
	}
	cnt := copy(b, p.compiled[p.current:end])
	p.current += cnt
	return cnt, nil
}
