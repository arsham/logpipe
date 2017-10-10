// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/arsham/logpipe/handler"
	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/writer"
	"github.com/namsral/flag"
)

var (
	port     = flag.Int("port", 8080, "port to listen to")
	logfile  = flag.String("logfile", "", "log file to write to")
	logLevel = flag.String("log", "error", "log level")
)

// this main function is fully covered in the main_test.go file and is excluded
// from coveralls statistics.
func main() {
	flag.Parse()
	if *logfile == "" {
		log.Fatal("need log file")
	}
	logger := internal.GetLogger(*logLevel)

	w, err := writer.NewFile(
		writer.WithFileLoc(*logfile),
		writer.WithLogger(logger),
	)
	if err != nil {
		log.Fatal(err)
	}

	server := handler.Service{
		Writer: w,
		Logger: logger,
	}

	http.HandleFunc("/", server.RecieveHandler)
	if err != nil {
		logger.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
