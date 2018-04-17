// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package tools_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/arsham/logpipe/tools"
)

func TestGetLoggerLevels(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		level    string
		expected tools.Level
	}{
		{"debug", tools.Level(tools.DebugLevel)},
		{"info", tools.Level(tools.InfoLevel)},
		{"warn", tools.Level(tools.WarnLevel)},
		{"error", tools.Level(tools.ErrorLevel)},
		{"DEBUG", tools.Level(tools.DebugLevel)},
		{"INFO", tools.Level(tools.InfoLevel)},
		{"WARN", tools.Level(tools.WarnLevel)},
		{"ERROR", tools.Level(tools.ErrorLevel)},
		{"dEbUG", tools.Level(tools.DebugLevel)},
		{"iNfO", tools.Level(tools.InfoLevel)},
		{"wArN", tools.Level(tools.WarnLevel)},
		{"eRrOR", tools.Level(tools.ErrorLevel)},
		{"panic", tools.Level(tools.PanicLevel)},
		{"PANIC", tools.Level(tools.PanicLevel)},
		{"PaniC", tools.Level(tools.PanicLevel)},
		{"", tools.Level(tools.ErrorLevel)},
		{"sdfsdf", tools.Level(tools.ErrorLevel)},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			logger := tools.GetLogger(tc.level)
			if tools.Level(logger.Level) != tc.expected {
				t.Errorf("want (%v), got (%v)", tc.expected, logger.Level)
			}
		})
	}
}

func TestGetDiscardLogger(t *testing.T) {
	logger := tools.DiscardLogger()
	if logger.Out != ioutil.Discard {
		t.Errorf("want (ioutil.Discard), got (%v)", logger.Out)
	}
}

func TestWithWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := tools.WithWriter(buf)

	if logger.Out != buf {
		t.Fatalf("want (bytes.Buffer), got (%v)", logger.Out)
	}

	message := "this is the message"
	logger.Info(message)

	if !strings.Contains(buf.String(), message) {
		t.Errorf("want (%s) in the logs, got (%v)", message, buf.String())

	}
}
