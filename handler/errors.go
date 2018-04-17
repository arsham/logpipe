// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler

import "github.com/pkg/errors"

// Errors returned when something happens in the handler.
// ErrGettingReader is returned when the reader factory cannot return an
// appropriate reader for the entry.
var (
	ErrNoWriter        = errors.New("no writers specified")
	ErrNilLogger       = errors.New("logger cannot be nil")
	ErrDuplicateWriter = errors.New("duplicated writer")
	ErrWritingEntry    = errors.New("writing the entry")
	ErrGettingReader   = errors.New("getting reader")
	ErrNoOptions       = errors.New("no option provided")
	ErrTimeout         = errors.New("timeout cannot be zero")
)
