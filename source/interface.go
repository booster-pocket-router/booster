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

package source

import (
	"context"
	"net"
	"sync"
)

type DialHook func(ref, network, address string, err error)

// Interface is a wrapper around net.Interface and
// implements the core.Source interface, i.e. is it
// capable of providing network connections through
// the device it is referring to.
type Interface struct {
	ifi net.Interface

	// If OnDialErr is not nil, it is called each time that the
	// dialer is not able to create a network connection.
	OnDialErr DialHook

	conns *conns
}

func (i *Interface) Name() string {
	return i.ifi.Name
}

func (i *Interface) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Implementations of the `dialContext` function can be found
	// in the {darwin, linux, windows}_dial.go files.
	conn, err := i.dialContext(ctx, network, address)
	if err != nil {
		if f := i.OnDialErr; f != nil {
			f(i.Name(), network, address, err)
		}
		return nil, err
	}

	return i.Follow(conn), nil
}

func (i *Interface) Follow(conn net.Conn) net.Conn {
	wconn := WrapConn(conn)
	wconn.OnClose = func(download *DataFlow, upload *DataFlow) {
		// TODO: Do something with this data
		i.conns.Del(wconn)
	}
	if i.conns == nil {
		i.conns = &conns{}
	}

	i.conns.Add(wconn)

	return wconn
}

func (i *Interface) Close() error {
	i.conns.Close()

	return nil
}

func (i *Interface) Value(key interface{}) interface{} {
	return nil
}

func (i *Interface) String() string {
	return i.Name()
}

func (i *Interface) Len() int {
	if i.conns == nil {
		return 0
	}

	return i.conns.Len()
}

type conns struct {
	sync.Mutex
	val []*Conn
}

func (c *conns) Add(conn *Conn) {
	c.Lock()
	defer c.Unlock()

	if c.val == nil {
		c.val = make([]*Conn, 0, 10)
	}
	c.val = append(c.val, conn)
}

func (c *conns) Close() {
	c.Lock()
	for _, v := range c.val {
		// Call close on the connetion after Unlock to
		// avoid deadlocks.
		defer v.Close()
	}
	c.Unlock()
}

func (c *conns) Del(conn *Conn) {
	c.Lock()
	defer c.Unlock()

	var t int
	for i, v := range c.val {
		if v == conn {
			t = i
			break
		}
	}

	copy(c.val[t:], c.val[t+1:])
	c.val[len(c.val)-1] = nil
	c.val = c.val[:len(c.val)-1]
}
func (c *conns) Len() int {
	c.Lock()
	defer c.Unlock()

	return len(c.val)
}
