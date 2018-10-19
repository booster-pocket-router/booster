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
	"io"
	"encoding/json"

	"github.com/booster-proj/proxy"
	"github.com/booster-proj/core"
	"github.com/booster-proj/log"
	"github.com/booster-proj/booster"
)

// Version and BuildTime are filled in during build by the Makefile
var (
	Version   = "N/A"
	BuildTime = "N/A"
)

var port = flag.Int("port", 1080, "Server listening port")
var rawProto = flag.String("proto", "", "Proxy protocol used. Available protocols: http, https, socks5.")
var ifacesFile = flag.String("ifaces", "", "Interfaces input file path. Usually generated with the `gen` tool.")

var verbose = flag.Bool("verbose", false, "Enable verbose mode")

func main() {
	flag.Parse()

	log.Info.Printf("Version: %s, BuildTime: %s\n\n", Version, BuildTime)
	if *verbose {
		log.Info.Printf("Running in verbose mode")
		log.SetLevel(log.DebugLevel)
	}

	if *rawProto == "" {
		log.Fatal("\"proto\" flag is required. Run `--help` for more.")
	}
	if *ifacesFile == "" {
		log.Fatal("\"ifaces\" flag is required. Run `--help` for more.")
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

	// Parse interfaces file

	file, err := os.Open(*ifacesFile)
	if err != nil {
		// ATM the presence of an interfaces file is mandatory
		log.Fatalf("Unable to open %v file. Generate it with the `gen` tool, use the `--help` flag for more", *ifacesFile)
	}

	ifaces, err := decodeIfacesFrom(file)
	if err != nil {
		log.Fatalf("Unable to decode %v file: %v", *ifacesFile, err)
	}
	if len(ifaces) == 0 {
		log.Fatal("At least one network interface with an active internet connection is needed. Aborting")
	}

	sources := make([]core.Source, len(ifaces))
	for i, v := range ifaces {
		sources[i] = v
	}

	// Create a booster instance that uses the colelcted interfaces as sources
	b := &booster.Booster{
		Balancer: &core.Balancer{},
	}
	b.Put(sources...)

	// Make the proxy use booster as dialer
	p.DialWith(b)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS signals
	captureSignals(cancel)

	log.Info.Printf("Booster proxy (%v) listening on :%d, using %d sources", p.Protocol(), *port, b.Len())
	if err := p.ListenAndServe(ctx, *port); err != nil {
		log.Fatal(err)
	}
}

func decodeIfacesFrom(r io.Reader) ([]*booster.IfaceSource, error) {
	var v struct {
		Version int `json:"version"`
		Body []*booster.IfaceSource `json:"body"`
	}
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return nil, err
	}

	return v.Body, nil
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
