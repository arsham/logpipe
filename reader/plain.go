// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"github.com/arsham/logpipe/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// TimestampFormat is the default formatting defined for logs.
var TimestampFormat = time.RFC3339

// var TimestampFormat = "2006-01-02 15:04:05"

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

// TextFormatter is used for rendering a custom format. We need to put the time
// at the very beginning of the line.
type TextFormatter struct {
	logrus.TextFormatter
}

// Format will use the timestamp passed by the payload and injects it in the
// entry itself.
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	ts, ok := entry.Data["time"].(string)
	if !ok {
		return nil, errors.New("no time in the log entry")
	}
	t, err := dateparse.ParseAny(ts)
	if err != nil {
		return nil, errors.Wrap(err, "parsing datetime")
	}
	e := &logrus.Entry{
		Logger:  entry.Logger,
		Time:    t,
		Level:   entry.Level,
		Message: entry.Message,
		Buffer:  entry.Buffer,
	}
	delete(entry.Data, "time")
	e.Data = entry.Data
	return f.TextFormatter.Format(e)
}

func (p *Plain) Read(b []byte) (int, error) {
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
		logger := logrus.New()
		customFormatter := new(TextFormatter)
		customFormatter.DisableColors = true
		logger.Formatter = customFormatter

		buf := new(bytes.Buffer)
		logger.Out = buf
		ll := logger.WithField("time", p.Timestamp.Format(TimestampFormat))
		switch p.Kind {
		case INFO:
			ll.Info(p.Message)
		case WARN:
			ll.Warn(p.Message)
		case ERROR:
			ll.Error(p.Message)
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
