// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/internal/config"
	"github.com/arsham/logpipe/writer"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	ConfigFile string `short:"c" long:"config-file" env:"CONFIGFILE" description:"configuration file" required:"true"`
	LogLevel   string `short:"l" long:"log-level" default:"error" description:"application log level"`
	Port       int    `short:"p" long:"port" default:"8080" env:"PORT" description:"port to listen for incoming payload"`
}

// this main function is fully covered in the main_test.go file and is excluded
// from coveralls statistics.
func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		log.Fatal(err)
	}

	logger := internal.GetLogger(opts.LogLevel)
	s, err := config.Read(opts.ConfigFile)
	if err != nil {
		logger.Fatal(err, opts.ConfigFile)
	}

	var (
		fileLocation string
		ok           bool
		writers      []io.Writer
	)

LOOP:
	for _, conf := range s.Writers {

		switch mod := conf["type"]; mod {
		case "file":
			if fileLocation, ok = conf["location"]; !ok {
				logger.Warn(s.Writers)
				continue LOOP
			}
			w, err := writer.NewFile(
				writer.WithFileLoc(fileLocation),
				writer.WithLogger(logger),
			)
			if err != nil {
				logger.Fatal(err)
			}
			writers = append(writers, w)
		}
	}

	server := handler.Service{
		Writers: writers,
		Logger:  logger,
	}

	http.HandleFunc("/", server.RecieveHandler)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Fatal(http.ListenAndServe(":"+strconv.Itoa(opts.Port), nil))
}
