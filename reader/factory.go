// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"io"
	"time"

	"github.com/araddon/dateparse"
	"github.com/arsham/logpipe/tools"
	jason "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

// The following constants are used for log levels.
const (
	InfoLevel  = "info"
	ErrorLevel = "error"
	WarnLevel  = "warning"
)

// GetReader tries to guess an appropriate reader from the input reader and
// returns it. It will fall back to Plain reader. It returns an error if there
// is no type or message are in the input or the message is empty.
func GetReader(r io.Reader, logger tools.FieldLogger) (io.Reader, error) {
	j, err := jason.NewFromReader(r)
	if err != nil {
		return nil, errors.Wrap(err, ErrCorruptedJSON.Error())
	}

	if m, err := j.Map(); err != nil {
		return nil, errors.Wrap(err, ErrCorruptedJSON.Error())
	} else if len(m) == 0 {
		return nil, ErrEmptyObject
	}

	kind, err := j.Get("type").String()
	if err != nil || kind == "" {
		kind = InfoLevel
	}

	message, err := j.Get("message").String()
	if err != nil {
		err = errors.Wrap(err, ErrEmptyMessage.Error())
		return nil, err
	}

	if message == "" {
		return nil, ErrEmptyMessage
	}

	timestamp, err := j.Get("timestamp").String()
	if err != nil || timestamp == "" {
		timestamp = time.Now().Format(TimestampFormat)
	}

	t, err := dateparse.ParseAny(timestamp)
	if err != nil {
		err = errors.Wrap(err, ErrTimestamp.Error())
		return nil, err
	}

	return &Plain{
		Message:   message,
		Kind:      kind,
		Timestamp: t,
		Logger:    logger,
	}, nil
}
