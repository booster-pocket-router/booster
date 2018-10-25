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

package booster

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/booster-proj/core"
	"github.com/booster-proj/log"
)

type Booster struct {
	*core.Balancer
}

func (b *Booster) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	src, err := b.Get()
	if err != nil {
		return nil, err
	}

	return src.DialContext(ctx, network, address)
}

func GetFilteredInterfaces(s string) []Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		log.Error.Printf("Unable to get interfaces: %v\n", err)
		return []Interface{}
	}

	l := make([]Interface, 0, len(ifs))

	for _, v := range ifs {
		log.Debug.Printf("Inspecting interface %+v\n", v)

		if len(v.HardwareAddr) == 0 {
			log.Debug.Printf("Empty hardware address. Skipping interface...")
			continue
		}

		if s != "" && !strings.Contains(v.Name, s) {
			log.Debug.Printf("Interface name does not satisfy name requirements: must contain \"%s\"", s)
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			// If the source does not contain an error
			log.Debug.Printf("Unable to get interface addresses: %v. Skipping interface...", err)
			continue
		}
		if len(addrs) == 0 {
			log.Debug.Printf("Empty unicast/multicast address list. Skipping interface...")
			continue
		}

		l = append(l, Interface{Interface: v})
	}

	return l
}

// Interface is a wrapper around net.Interface and
// implements the core.Source interface, i.e. is it
// capable of providing network connections through
// the device it is referring to.
type Interface struct {
	net.Interface

	mux sync.Mutex
	// N is the number of network connections that
	// the interface is currenlty handling.
	N int
}

func (i Interface) Add(val int) int {
	i.mux.Lock()
	defer i.mux.Unlock()

	i.N += val
	return i.N
}

func (i Interface) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Implementations of the `dialContext` function can be found
	// in the {unix, darwin}_dial.go files.

	// TODO(jecoz): add windows implementation
	c, err := i.dialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		Conn: c,
		Add:  i.Add,
		Ref:  i.ID(),
	}

	n := conn.Add(1)
	log.Debug.Printf("Opening connection (ref: %v) to(%v), left(%d)", conn.Ref, c.RemoteAddr(), n)

	return conn, nil
}

func (i Interface) ID() string {
	return i.Name
}

func (i Interface) Metrics() map[string]interface{} {
	return make(map[string]interface{})
}

type Conn struct {
	net.Conn
	Ref string // Reference identifier
	Add func(val int) int

	closed bool
}

func (c *Conn) Close() error {
	if c.closed {
		// Multiple parts of the code might try to close the connection. Better be sure
		// that the underlying connection gets closed at some point, leave that code and
		// avoid repetitions here.
		return nil
	}

	n := c.Add(-1)
	log.Debug.Printf("Closing connection (ref: %v) to(%v), left(%d)", c.Ref, c.RemoteAddr(), n)
	c.closed = true

	return c.Conn.Close()
}
