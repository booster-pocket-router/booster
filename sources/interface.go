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

// Package sources provides implementations of entities, such as network
// interfaces, that are able to create network connections.
package sources

import (
	"context"
	"net"
	"sync"

	"upspin.io/log"
)

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

func (i *Interface) Add(val int) int {
	i.mux.Lock()
	defer i.mux.Unlock()

	i.N += val
	return i.N
}

func (i *Interface) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Implementations of the `dialContext` function can be found
	// in the {unix, darwin}_dial.go files.

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

func (i *Interface) ID() string {
	return i.Name
}

func (i *Interface) Metrics() map[string]interface{} {
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
