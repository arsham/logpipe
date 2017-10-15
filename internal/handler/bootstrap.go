// Copyright 2017 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

package handler

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/arsham/logpipe/internal"
	"github.com/arsham/logpipe/internal/config"
)

// This file contains configuration required for bootstrapping the app.

// ServiceInt is an interface for the handler.Service
type ServiceInt interface {
	Handler() http.HandlerFunc
	Timeout() time.Duration
}

// ServeFunc is a function that is run for setting up the handlers.
var ServeFunc func(s ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error

func init() {
	ServeFunc = Serve
}

// Bootstrap reads the command options and starts the server.
func Bootstrap(logger internal.FieldLogger, configFile string, port int) {
	if logger == nil {
		logger = internal.GetLogger("error")
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	c, err := config.Read(configFile)
	if err != nil {
		logger.Errorf("%v: %s", err, configFile)
		return
	}

	logger.Infof("config file: %s", configFile)

	s, err := New(
		logger,
		WithConfWriters(logger, c),
	)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Infof("running on port: %d", port)
	err = ServeFunc(s, logger, stop, port)
	if err != nil {
		logger.Errorf("error when serving: %s", err)
	}
}

// Serve starts a http.Server in a goroutine.
// It listens to the Interrupt signal and shuts down the server.
// It returns any errors occurred during the service.
// In case of graceful shut down, it will nil.
func Serve(s ServiceInt, logger internal.FieldLogger, stop chan os.Signal, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.Handler())
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
		return nil
	}
}
