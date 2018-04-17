// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/arsham/logpipe/tools"
	"github.com/arsham/logpipe/tools/config"
	"github.com/pkg/errors"
)

// This file contains configuration required for bootstrapping the app.

// Server is an interface for the handler.Service
type Server interface {
	http.Handler
	Timeout() time.Duration
}

// ServeHTTP will set up the handlers and starts listening to the port. It sets
// up the server in a goroutine. It will shut down when it receives an Interrupt
// signal from the OS. It will hold on to the active connections until they
// finish their work, or a timeout occurs. It returns any errors occurred during
// the service.
var ServeHTTP func(s Server, logger tools.FieldLogger, stop chan os.Signal, port int) error

func init() {
	ServeHTTP = serveHTTP
}

// Bootstrap reads the command options and starts the server. It returns nil
// when the server finishes its work successfully, or else it will return the
// error.
func Bootstrap(logger tools.FieldLogger, configFile string, port int) error {
	if logger == nil {
		logger = tools.GetLogger("error")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	c, err := config.Read(configFile)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("reading config file: %s", configFile))
	}

	logger.Infof("config file: %s", configFile)

	s, err := New(
		WithLogger(logger),
		WithConfWriters(logger, c),
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("creating the service: %s", configFile))
	}

	logger.Infof("running on port: %d", port)
	return ServeHTTP(s, logger, stop, port)
}

// see ServeHTTP.
func serveHTTP(s Server, logger tools.FieldLogger, stop chan os.Signal, port int) error {
	mux := http.NewServeMux()
	mux.Handle("/", s)
	h := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	srvErr := make(chan error)
	quit := make(chan struct{})
	go func() {
		select {
		case srvErr <- h.ListenAndServe():
		case <-quit:
		}
	}()

	select {
	case err := <-srvErr:
		return err
	case <-stop:
		close(quit)
		ctx, cancel := context.WithTimeout(context.Background(), s.Timeout())
		defer cancel()
		logger.Infof("shutting down the server: %s", h.Shutdown(ctx))
	}
	return nil
}
