// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import (
	"io"

	jason "github.com/bitly/go-simplejson"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
)

// GetReader tries to guess an appropriate reader from the input byte slice
// and returns it. It will fall back to Plain reader.
// It returns an error if there is no type or message are in the input or the
// message is empty.
func GetReader(input []byte) (io.Reader, error) {
	j, err := jason.NewJson(input)
	if err != nil {
		return nil, errors.Wrap(err, "decoding json object")
	}

	kind, err := j.Get("type").String()
	if err != nil || kind == "" {
		return nil, errors.Wrap(err, "empty type")
	}

	message, err := j.Get("message").String()
	if err != nil || message == "" {
		return nil, errors.Wrap(err, "empty message")
	}

	timestamp, err := j.Get("timestamp").String()
	if err != nil || timestamp == "" {
		return nil, errors.Wrap(err, "empty timestamp")
	}
	t, err := now.Parse(timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "parsing timestamp")
	}

	r := &Plain{
		Message:   message,
		Kind:      kind,
		Timestamp: t,
	}
	return r, nil
}
