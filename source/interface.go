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

// Package source provides implementations of entities, such as network
// interfaces, that are able to create network connections, i.e. are
// "sources" of Internet.
package source

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

	// If ErrHook is not nil, it is called each time that the
	// dialer is not able to create a network connection.
	ErrHook func(ref, network, address string, err error)

	mux sync.Mutex
	// N is the number of network connections that
	// the interface is currently handling.
	N int
}

func (i *Interface) String() string {
	return i.ID()
}

func (i *Interface) Add(val int) int {
	i.mux.Lock()
	defer i.mux.Unlock()

	i.N += val
	return i.N
}

func (i *Interface) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Implementations of the `dialContext` function can be found
	// in the {unix, linux}_dial.go files.

	c, err := i.dialContext(ctx, network, address)
	if err != nil {
		if f := i.ErrHook; f != nil {
			f(i.ID(), network, address, err)
		}
		return nil, err
	}

	conn := &Conn{
		Conn: c,
		Add:  i.Add,
		Ref:  i.ID(),
	}

	n := conn.Add(1)
	log.Debug.Printf("Interface (ref: %v): Opening connection to(%v), left(%d)", conn.Ref, c.RemoteAddr(), n)

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
	log.Debug.Printf("Interface (ref: %v): Closing connection to(%v), left(%d)", c.Ref, c.RemoteAddr(), n)
	c.closed = true

	return c.Conn.Close()
}
