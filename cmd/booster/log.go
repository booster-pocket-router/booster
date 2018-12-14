/*
Copyright (C) 2018 KIM KeepInMind GmbH/srl

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	stdLog "log"
	"os"

	"upspin.io/log"
)

func setupLogger(verbose bool, external bool) {
	level := log.InfoLevel
	if verbose {
		log.SetLevel("debug")
		level = log.DebugLevel
	}
	if external {
		log.SetOutput(nil)                     // disable "local" logging
		log.Register(newExternalLogger(level)) // enable "remote" (snapcraft's daemon handled logger usually) logging
	}
}

type externalLogger struct {
	defaultLogger log.Logger
	level         log.Level
}

func newExternalLogger(level log.Level) *externalLogger {
	return &externalLogger{
		level:         level,
		defaultLogger: stdLog.New(os.Stderr, "", 0), // Do not add date/time information
	}
}

func (l *externalLogger) Log(level log.Level, msg string) {
	if level < l.level {
		return
	}

	l.defaultLogger.Println(msg)
}

func (l *externalLogger) Flush() {
}
