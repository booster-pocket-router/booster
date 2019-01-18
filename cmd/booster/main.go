// Copyright © 2019 KIM KeepInMind GmbH/srl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	stdLog "log"
	"os"
	"os/signal"
	"strings"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/booster/core"
	"github.com/booster-proj/booster/metrics"
	"github.com/booster-proj/booster/remote"
	"github.com/booster-proj/booster/source"
	"github.com/booster-proj/booster/store"
	"github.com/booster-proj/proxy"
	"github.com/grandcat/zeroconf"
	"golang.org/x/sync/errgroup"
	"upspin.io/log"
)

// Version and BuildTime are filled in during build by the Makefile
var (
	version   = "N/A"
	commit    = "N/A"
	buildTime = "N/A"
)

var (
	// Commands
	printVersion = flag.Bool("version", false, "Prints version")

	// Proxy configuration
	pPort    = flag.Int("proxy-port", 1080, "Proxy server listening port")
	rawProto = flag.String("proto", "socks5", "Proxy protocol used. Available protocols: http, socks5.")

	// API configuration
	apiPort  = flag.Int("api-port", 7764, "API server listening port")
	promPort = flag.Int("prometheus-port", 9090, "Port of a local prometheus server. Used for forwarding")

	// Log configuration
	verbose     = flag.Bool("verbose", false, "If set, makes the logger print also debug messages")
	scope       = flag.String("scope", "", "If set, enables debug logging only in the desired scope")
	externalLog = flag.Bool("external-log", false, "If set, assumes that the loggin is handled by a third party entity")
)

func main() {
	// Parse arguments
	flag.Parse()

	fmt.Printf("version: %s, commit: %s, built at: %s\n\n", version, commit, buildTime)
	if *printVersion {
		return
	}

	// Setup logger
	setupLogger(*verbose, *externalLog)

	// Find proxy from protocol
	p, err := proxyFromProto(*rawProto)
	if err != nil {
		log.Fatal(err)
	}

	remote.StaticConf = remote.Config{
		Version:    version,
		Commit:     commit,
		BuildTime:  buildTime,
		ProxyPort:  *pPort,
		ProxyProto: *rawProto,
		PromPort:   *promPort,
	}

	b := new(core.Balancer)
	rs := store.New(b)
	exp := new(metrics.Exporter)
	l := source.NewListener(source.Config{
		Store:           rs,
		MetricsExporter: exp,
	})
	d := booster.New(b)
	d.SetMetricsExporter(exp)

	router := remote.NewRouter()
	router.Store = rs
	router.MetricsProvider = exp
	router.SetupRoutes()
	r := remote.New(router)

	// Make the proxy use booster as dialer
	p.DialWith(d)

	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	captureSignals(cancel)

	// Expose out services as mDNS entries
	s, err := zeroconf.Register("booster api", "_http._tcp", "local.", *apiPort, []string{
		"Version=" + version,
		"Commit=" + commit,
	}, nil)
	defer s.Shutdown()

	s, err = zeroconf.Register("booster proxy", "_"+strings.ToLower(*rawProto)+"._tcp", "local.", *pPort, []string{
		"Protocol=" + *rawProto,
		"Version=" + version,
		"Commit=" + commit,
	}, nil)
	defer s.Shutdown()

	g.Go(func() error {
		log.Info.Printf("Listener started")
		defer log.Info.Printf("Listener stopped.")
		return l.Run(ctx)
	})
	g.Go(func() error {
		log.Info.Printf("Booster proxy (%v) listening on :%d", p.Protocol(), *pPort)
		defer log.Info.Print("Booster proxy stopped.")
		return p.ListenAndServe(ctx, *pPort)
	})
	g.Go(func() error {
		log.Info.Printf("Booster API listening on :%d", *apiPort)
		defer log.Info.Print("Booster API stopped.")
		return r.ListenAndServe(ctx, *apiPort)
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

func proxyFromProto(rawProto string) (proxy.Proxy, error) {
	if rawProto == "" {
		return nil, errors.New("\"proto\" flag is required. Run `--help` for more.")
	}

	proto, err := proxy.ParseProto(rawProto)
	if err != nil {
		return nil, err
	}

	var p proxy.Proxy
	switch proto {
	case proxy.HTTP:
		p, err = proxy.NewHTTP()
	case proxy.SOCKS5:
		p, err = proxy.NewSOCKS5()
	default:
		err = errors.New("protocol (" + rawProto + ") is not yet supported")
	}
	return p, err
}

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
