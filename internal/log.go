// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package internal

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// FieldLogger interface set by logrus
type FieldLogger logrus.FieldLogger

// Level type set by logrus
type Level logrus.Level

// Logger embeds logrus.Logger
type Logger struct{ *logrus.Logger }

// Entry embeds logrus.Entry
type Entry struct{ *logrus.Entry }

// StandardLogger returns an instance of Logger
func StandardLogger() *Logger { return &Logger{logrus.StandardLogger()} }

const (
	// InfoLevel for Info level
	InfoLevel = logrus.InfoLevel
	// WarnLevel for Warn level
	WarnLevel = logrus.WarnLevel
	// DebugLevel for Debug level
	DebugLevel = logrus.DebugLevel
	// ErrorLevel for Error level
	ErrorLevel = logrus.ErrorLevel
)

// GetLogger returns the default logger with the given log level.
func GetLogger(level string) *Logger {
	logrus.SetLevel(ErrorLevel)
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
	logrus.SetOutput(os.Stdout)
	customFormatter.FullTimestamp = true
	switch strings.ToLower(level) {
	case "debug":
		logrus.SetLevel(DebugLevel)
	case "info":
		logrus.SetLevel(InfoLevel)
	case "warn":
		logrus.SetLevel(WarnLevel)
	case "error":
		logrus.SetLevel(ErrorLevel)
	default:
		logrus.SetLevel(ErrorLevel)
	}

	return StandardLogger()
}

// DiscardLogger returns a dummy logger.
// This is useful for tests when you don't want to actually write to the Stdout.
func DiscardLogger() *Logger {
	log := logrus.New()
	log.Out = ioutil.Discard
	return &Logger{log}
}
