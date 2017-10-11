// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"io"
	"time"

	"github.com/araddon/dateparse"
	"github.com/arsham/logpipe/internal"
	jason "github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
)

const (
	// INFO is a log level.
	INFO = "info"
	// ERROR is a log level.
	ERROR = "error"
	// WARN is a log level.
	WARN = "warning"
)

// GetReader tries to guess an appropriate reader from the input byte slice
// and returns it. It will fall back to Plain reader.
// It returns an error if there is no type or message are in the input or the
// message is empty.
func GetReader(input []byte, logger internal.FieldLogger) (io.Reader, error) {
	j, err := jason.NewJson(input)
	if err != nil {
		logger.Errorf("decoding json object: %s", err)
		return nil, errors.Wrap(err, ErrDecodeJSON.Error())
	}

	kind, err := j.Get("type").String()
	if err != nil || kind == "" {
		logger.Debugf("falling back to info: %s", input)
		kind = INFO
	}

	message, err := j.Get("message").String()
	if err != nil || message == "" {
		err = errors.Wrap(err, ErrEmptyMessage.Error())
		logger.Error(err)
		return nil, err
	}

	timestamp, err := j.Get("timestamp").String()
	if err != nil || timestamp == "" {
		logger.Debugf("no timestamp provided: %s", input)
		timestamp = time.Now().Format(TimestampFormat)
	}

	t, err := dateparse.ParseAny(timestamp)
	if err != nil {
		err = errors.Wrap(err, ErrTimestamp.Error())
		logger.Error(err)
		return nil, err
	}

	r := &Plain{
		Message:   message,
		Kind:      kind,
		Timestamp: t,
		Logger:    logger,
	}
	return r, nil
}
