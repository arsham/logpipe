// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler

import "github.com/pkg/errors"

var (
	// ErrNoWriter is returned when no write is provided.
	ErrNoWriter = errors.New("no writers specified")

	// ErrNilLogger is returned when logger is nil.
	ErrNilLogger = errors.New("logger cannot be nil")

	// ErrDuplicateWriter is returned on duplicated writers.
	ErrDuplicateWriter = errors.New("duplicated writer")

	// ErrWritingEntry is for when there is a problem with writing the entry.
	ErrWritingEntry = errors.New("writing the entry")

	// ErrGettingReader is returned when the reader factory cannot return an
	// appropriate reader for the entry.
	ErrGettingReader = errors.New("getting reader")

	// ErrNoOptions is returned when no option is provided.
	ErrNoOptions = errors.New("no option provided")

	// ErrTimeout is returned when the timeout is zero.
	ErrTimeout = errors.New("timeout cannot be zero")
)
