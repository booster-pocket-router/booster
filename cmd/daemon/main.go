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

// Package main starts a booster server that is adapted for
// snap's daemon usage.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/booster/listener"
	"github.com/booster-proj/core"
	"github.com/booster-proj/proxy"
	"golang.org/x/sync/errgroup"
	stdLog "log"
	"upspin.io/log"
)

// Version and BuildTime are filled in during build by the Makefile
var (
	version   = "N/A"
	commit    = "N/A"
	buildTime = "N/A"
)

var port = flag.Int("port", 1080, "Server listening port")
var verbose = flag.Bool("verbose", false, "Make logger print also debug messages")

func main() {
	// Parse arguments
	flag.Parse()

	fmt.Printf("version: %s, commit: %s, built at: %s\n\n", version, commit, buildTime)

	// Setup logger
	level := log.InfoLevel
	if *verbose {
		level = log.DebugLevel
	}
	log.SetOutput(nil)                     // disable "local" logging
	log.Register(newExternalLogger(level)) // enable "remote" (snapcraft's daemon handled logger usually) logging

	p, err := proxy.NewSOCKS5()
	if err != nil {
		log.Fatal(err)
	}

	b := new(core.Balancer)
	l := listener.New(b)
	d := booster.New(b)

	// Make the proxy use booster as dialer
	p.DialWith(d)

	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	captureSignals(cancel)

	g.Go(func() error {
		log.Info.Printf("Listener started")
		return l.Run(ctx)
	})
	g.Go(func() error {
		log.Info.Printf("Booster proxy (%v) listening on :%d", p.Protocol(), *port)
		return p.ListenAndServe(ctx, *port)
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func captureSignals(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		for range c {
			cancel()
		}
	}()
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
