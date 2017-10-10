// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import "github.com/pkg/errors"

var (
	// ErrNilTimestamp is returned when timestamp is and empty time.Time
	ErrNilTimestamp = errors.New("nil timestamp")

	// ErrEmptyMessage is returned when the message body is an empty string.
	ErrEmptyMessage = errors.New("empty message")

	// ErrDecodeJSON is returned when the payload is not JSON unmarshallable.
	ErrDecodeJSON = errors.New("decoding json object")

	// ErrTimestamp is returned when the timestamp is not valid.
	ErrTimestamp = errors.New("invalid timestamp")
)
