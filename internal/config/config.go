// Copyright 2016 Arsham Shirvani <arshamshirvani@gmail.com>. All rights reserved.
// Use of this source code is governed by the Apache 2.0 license
// License that can be found in the LICENSE file.

// Package config loads the configurations from a yaml file.
// Configurations come in the following form:
//
//    app:
//      log_level: info
//    writers:
//      elasticsearch:
//         url: http://localhost:9200
//         index: logs
//      file:
//         path: /var/log/logpipe/logs.log
//
// The app part will be collapsed as the Setting properties.
package config

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	// ErrFileNotExist is returned when the file can not be read.
	ErrFileNotExist = errors.New("file does not exist")

	// ErrNoWriters is returned when there is no writes found on the settings.
	ErrNoWriters = errors.New("no writers found")
)

// Setting holds the required configuration settings for bootstrapping the
// application.
type Setting struct {

	// LogLevel is the level in which the logpipe application writes it's own
	// logs.
	LogLevel string

	// Writers has a map of "writer" name to its configuration.
	// Each writer decides its own configuration.
	// It goes as: [name:[type:file, location:foo, name:bar]],..
	Writers map[string]map[string]string
}

// Read loads the configurations from filename location.
// The filename should be a valid yaml file or it will return an error.
// It will return an error if there is no writers defined.
func Read(filename string) (*Setting, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, ErrFileNotExist
	}

	v := viper.New()
	v.SetConfigType("yaml")
	err = v.ReadConfig(f)
	if err != nil {
		return nil, err
	}

	s := &Setting{
		LogLevel: "error",
	}
	app := v.GetStringMap("app")
	if l, ok := app["log_level"]; ok {
		s.LogLevel = l.(string)
	}

	app = v.GetStringMap("writers")
	if len(app) == 0 {
		return nil, ErrNoWriters
	}

	var ok bool
	maps := make(map[string]map[string]string)
	// example of app: [file:[location:foo, name:bar]],...
	for moduleName, settings := range app {
		setMap := settings.(map[string]interface{}) // viper guarantees this

		configs := make(map[string]string)
		// setMap is: [location:foo, name:bar]
		for name, value := range setMap {
			var strVal string
			if strVal, ok = value.(string); !ok {
				return nil, errors.New("no string value")
			}

			configs[name] = strVal
		}
		maps[moduleName] = configs
	}
	s.Writers = maps

	return s, nil
}
