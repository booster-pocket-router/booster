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
	"fmt"
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

	conns struct {
		sync.Mutex
		val map[string]*conn
	}
}

func (i *Interface) ID() string {
	return i.Name
}

func (i *Interface) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Implementations of the `dialContext` function can be found
	// in the {unix, linux}_dial.go files.
	conn, err := i.dialContext(ctx, network, address)
	if err != nil {
		if f := i.ErrHook; f != nil {
			f(i.ID(), network, address, err)
		}
		return nil, err
	}

	// Follow the new connection if possible
	go func() {
		if err := i.Follow(conn); err != nil {
			log.Error.Println(err)
		}
	}()

	return conn, nil
}

// Follow adds conn to the list of connections that the source is handling.
// The connection is left intact even in case of error, in which case the
// connection is simply ignored by the interface.
func (i *Interface) Follow(c net.Conn) error {
	wc := newConn(c, i.ID()) // wrapped connection

	i.conns.Lock()
	defer i.conns.Unlock()

	if i.conns.val == nil {
		i.conns.val = make(map[string]*conn)
	}

	if _, ok := i.conns.val[wc.uuid()]; ok {
		// Another connection with the same identifier as this one is already in
		// process. The connection identifiers are supposed to be unique, so this
		// means that we'll not be able to follow this connection.
		return fmt.Errorf("DialContext: discarding connection (id: %s) because source (%s) has a connection in process with the same identifier", wc.uuid(), i.ID())
	}

	wc.onClose = func(id string) {
		// Be careful with race condition with the Close funtion here.
		i.conns.Lock()
		delete(i.conns.val, id)
		i.conns.Unlock()
	}
	i.conns.val[wc.uuid()] = wc

	return nil
}

func (i *Interface) Close() error {
	i.conns.Lock()
	for _, v := range i.conns.val {
		// Call close on each connection, but make the code run
		// after this loop as ended.
		defer v.Close()

	}
	i.conns.Unlock()

	return nil
}

func (i *Interface) String() string {
	return i.ID()
}

// Len returns the size
func (i *Interface) Len() int {
	i.conns.Lock()
	defer i.conns.Unlock()

	if i.conns.val == nil {
		return 0
	}

	return len(i.conns.val)
}
