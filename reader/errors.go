// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package reader

import "github.com/pkg/errors"

// Errors returned when reading.
var (
	ErrNilTimestamp  = errors.New("nil timestamp")
	ErrEmptyMessage  = errors.New("empty message")
	ErrTimestamp     = errors.New("invalid timestamp")
	ErrEmptyObject   = errors.New("empty object")
	ErrCorruptedJSON = errors.New("corrupted json")
)
