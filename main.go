// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/tools"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	ConfigFile string `short:"c" long:"config-file" env:"CONFIGFILE" description:"configuration file" required:"true"`
	LogLevel   string `short:"l" long:"log-level" env:"LOGLEVEL" default:"error" description:"application log level"`
	Port       int    `short:"p" long:"port" default:"8080" env:"PORT" description:"port to listen for incoming payload"`
}

// this main function is fully covered in the main_test.go file and is excluded
// from coverage statistics.
func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	logger := tools.GetLogger(opts.LogLevel)
	err = handler.Bootstrap(logger, opts.ConfigFile, opts.Port)
	if err != nil {
		logger.Fatal(err)
	}
}
