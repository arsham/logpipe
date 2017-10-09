// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package writer contains a series of writers that can write the log entries.
// A File can write logs to a given file. It will collapse the object if a json
// object is given, and uses them as the context of the log.
package writer
