// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler

import "github.com/pkg/errors"

var (
	// ErrNoWriter is returned when no write is provided.
	ErrNoWriter = errors.New("no writers specified")

	// ErrNoLogger is returned when no logger specified.
	ErrNoLogger = errors.New("no logger specified")

	// ErrNilLogger is returned when nil logger specified.
	ErrNilLogger = errors.New("nil logger specified")

	// ErrDuplicateWriter is returned on duplicated writers.
	ErrDuplicateWriter = errors.New("duplicated writer")

	// ErrWritingEntry is for when there is a problem with writing the entry.
	ErrWritingEntry = errors.New("writing the entry")

	// ErrEmptyObject is retuned when the payload is an empty object.
	ErrEmptyObject = errors.New("empty object")

	// ErrGettingReader is returned when the reader factory cannot return an
	// appropriate reader for the entry.
	ErrGettingReader = errors.New("getting reader")

	// ErrGettingMap is returned when the payload does not contain a map.
	ErrGettingMap = errors.New("getting map")

	// ErrCorruptedJSON is returned the payload is corrupted.
	ErrCorruptedJSON = errors.New("corrupted json")

	// ErrNoOptions is returned when no option is provided.
	ErrNoOptions = errors.New("no option provided")

	// ErrTimeout is returned when the timeout is zero.
	ErrTimeout = errors.New("timeout cannot be zero")
)
