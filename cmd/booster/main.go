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
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/booster/listener"
	"github.com/booster-proj/core"
	"github.com/booster-proj/proxy"
	"upspin.io/log"
)

// Version and BuildTime are filled in during build by the Makefile
var (
	version   = "N/A"
	commit    = "N/A"
	buildTime = "N/A"
)

var port = flag.Int("port", 1080, "Server listening port")
var rawProto = flag.String("proto", "", "Proxy protocol used. Available protocols: http, https, socks5.")
var interfaceName = flag.String("iname", "", "Collect only interfaces which name contains \"name\"")

var verbose = flag.Bool("verbose", false, "Enable verbose mode")

func main() {
	// Parse arguments
	flag.Parse()

	log.Info.Printf("%s, commit: %s, built at: %s", version, commit, buildTime)
	if *verbose {
		log.Info.Printf("Running in verbose mode")
		log.SetLevel("debug")
	}

	if *rawProto == "" {
		log.Fatal("\"proto\" flag is required. Run `--help` for more.")
	}

	// Configure proxy server
	proto, err := proxy.ParseProto(*rawProto)
	if err != nil {
		log.Fatal(err)
	}

	var p proxy.Proxy
	switch proto {
	case proxy.HTTP:
		p, err = proxy.NewHTTP()
	case proxy.SOCKS5:
		p, err = proxy.NewSOCKS5()
	default:
		err = errors.New("protocol (" + *rawProto + ") is not yet supported")
	}
	if err != nil {
		log.Fatal(err)
	}

	// Find network interfaces
	log.Info.Println("Finding relevant network interfaces...")
	ifs := listener.GetFilteredInterfaces(*interfaceName)
	if len(ifs) == 0 {
		log.Fatal("At least one network interface with an active internet connection is needed. Aborting")
	}
	log.Info.Printf("Collected %d relevant interfaces:", len(ifs))
	for i, v := range ifs {
		log.Info.Printf("%d: %+v\n", i, v)
	}

	// Create a booster instance that uses the colelcted interfaces as sources
	b := new(core.Balancer)
	d := booster.New(b)

	srcs := make([]core.Source, len(ifs))
	for i, v := range ifs {
		srcs[i] = v
	}
	b.Put(srcs...)

	// Make the proxy use booster as dialer
	p.DialWith(d)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS signals
	captureSignals(cancel)

	log.Info.Printf("Booster proxy (%v) listening on :%d, using %d sources", p.Protocol(), *port, b.Len())
	if err := p.ListenAndServe(ctx, *port); err != nil {
		log.Fatal(err)
	}
}

func captureSignals(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() {
		for _ = range c {
			cancel()
		}
	}()
}
