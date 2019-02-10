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

package dialer

import (
	"context"
	"net"
	"sync"

	"github.com/booster-proj/booster/core"
	"upspin.io/log"
)

// Balancer describes which functionalities must be provided in order
// to allow booster to get sources.
type Balancer interface {
	Get(ctx context.Context, target string, blacklisted ...core.Source) (core.Source, error)
	Len() int
}

// New returns an instance of a booster dialer.
func New(b Balancer) *Dialer {
	return &Dialer{b: b}
}

// MetricsExporter is an inteface around the IncSelectedSource function,
// which is used to collect a metric when a source is selected for use.
type MetricsExporter interface {
	IncSelectedSource(labels map[string]string)
}

// Dialer is a core.Dialer implementation, which uses a core.Balancer
// instance to to retrieve a source to use when it comes to dial a network
// connection.
type Dialer struct {
	b Balancer

	metrics struct {
		sync.Mutex
		exporter MetricsExporter
	}
}

// DialContext dials a connection using `network` to `address`. The connection returned
// is dialed through a specific network interface, which is chosen using the dialer's
// interal balancer provided. If it fails to create a connection using a source, it
// tries to dial it using another source, until source exhaustion. It that case,
// only the last error received is returned.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	bl := make([]core.Source, 0, d.Len()) // blacklisted sources

	// If the dialing fails, keep on trying with the other sources until exaustion.
	for i := 0; len(bl) < d.Len(); i++ {
		var src core.Source
		src, err = d.b.Get(ctx, address, bl...)
		if err != nil {
			// Fail directly if the balancer returns an error, as
			// we do not have any source to use.
			return
		}

		d.sendMetrics(src.ID(), address)

		log.Debug.Printf("DialContext: Attempt #%d to connect to %v (source %v)", i, address, src.ID())

		conn, err = src.DialContext(ctx, "tcp4", address)
		if err != nil {
			// Log this error, otherwise it will be silently skipped.
			log.Error.Printf("Unable to dial connection to %v using source %v. Error: %v", address, src.ID(), err)
			bl = append(bl, src)
			continue
		}

		// Connection dialed successfully.
		break
	}

	return
}

// Len returns the number of sources that the dialer as at it's disposal.
func (d *Dialer) Len() int {
	return d.b.Len()
}

// SetMetricsExporter makes the receiver use exp as metrics exporter.
func (d *Dialer) SetMetricsExporter(exp MetricsExporter) {
	d.metrics.Lock()
	defer d.metrics.Unlock()

	d.metrics.exporter = exp
}

func (d *Dialer) sendMetrics(name, target string) {
	if d.metrics.exporter == nil {
		return
	}

	d.metrics.Lock()
	defer d.metrics.Unlock()

	d.metrics.exporter.IncSelectedSource(map[string]string{
		"source": name,
		"target": target,
	})
}
