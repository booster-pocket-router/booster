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

package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/booster-proj/booster/core"
	"github.com/booster-proj/booster/dialer"
	"github.com/booster-proj/booster/metrics"
	"github.com/booster-proj/booster/remote"
	"github.com/booster-proj/booster/source"
	"github.com/booster-proj/booster/store"
	"github.com/booster-proj/proxy"
	"github.com/grandcat/zeroconf"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"upspin.io/log"
)

var (
	// Proxy configuration
	pPort int

	// API configuration
	apiPort  int
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		remote.Info = remote.BoosterInfo{
			Version:   Version,
			Commit:    Commit,
			BuildTime: BuildTime,
			ProxyPort: pPort,
		}

		p, err := proxy.NewSOCKS5()
		if err != nil {
			log.Fatal(err)
		}

		b := new(core.Balancer)
		rs := store.New(b)
		exp := new(metrics.Exporter)
		l := source.NewListener(source.Config{
			Store:           rs,
			MetricsExporter: exp,
		})
		d := dialer.New(rs)
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
		s, err := zeroconf.Register("booster api", "_http._tcp", "local.", apiPort, []string{
			"Version=" + Version,
			"Commit=" + Commit,
		}, nil)
		defer s.Shutdown()

		s, err = zeroconf.Register("booster proxy", "_SOCKS5_tcp", "local.", pPort, []string{
			"Version=" + Version,
			"Commit=" + Commit,
		}, nil)
		defer s.Shutdown()

		g.Go(func() error {
			log.Info.Printf("Listener started")
			defer log.Info.Printf("Listener stopped.")
			return l.Run(ctx)
		})
		g.Go(func() error {
			log.Info.Printf("Booster proxy (%v) listening on :%d", p.Protocol(), pPort)
			defer log.Info.Print("Booster proxy stopped.")
			return p.ListenAndServe(ctx, pPort)
		})
		g.Go(func() error {
			log.Info.Printf("Booster API listening on :%d", apiPort)
			defer log.Info.Print("Booster API stopped.")
			return r.ListenAndServe(ctx, apiPort)
		})

		if err := g.Wait(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Proxy configuration
	serverCmd.Flags().IntVar(&pPort, "proxy-port", 1080, "Proxy server listening port")

	// API configuration
	serverCmd.Flags().IntVar(&apiPort, "api-port", 7764, "API server listening port")
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
