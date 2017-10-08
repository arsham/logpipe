// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package internal_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/arsham/logpipe/internal"
)

func TestGetLoggerLevels(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		level    string
		expected internal.Level
	}{
		{"debug", internal.Level(internal.DebugLevel)},
		{"info", internal.Level(internal.InfoLevel)},
		{"warn", internal.Level(internal.WarnLevel)},
		{"error", internal.Level(internal.ErrorLevel)},
		{"DEBUG", internal.Level(internal.DebugLevel)},
		{"INFO", internal.Level(internal.InfoLevel)},
		{"WARN", internal.Level(internal.WarnLevel)},
		{"ERROR", internal.Level(internal.ErrorLevel)},
		{"dEbUG", internal.Level(internal.DebugLevel)},
		{"iNfO", internal.Level(internal.InfoLevel)},
		{"wArN", internal.Level(internal.WarnLevel)},
		{"eRrOR", internal.Level(internal.ErrorLevel)},
		{"", internal.Level(internal.ErrorLevel)},
		{"sdfsdf", internal.Level(internal.ErrorLevel)},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("case_%d", i)
		t.Run(name, func(t *testing.T) {
			logger := internal.Get(tc.level)
			if internal.Level(logger.Level) != tc.expected {
				t.Errorf("want (%v), got (%v)", tc.expected, logger.Level)
			}
		})
	}
}

func TestGetDiscardLogger(t *testing.T) {
	logger := internal.DiscardLogger()
	if logger.Out != ioutil.Discard {
		t.Errorf("want (ioutil.Discard), got (%v)", logger.Out)
	}
}
