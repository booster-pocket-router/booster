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

// DialHook describes the function used to notify about
// dial errors.
type DialHook func(ref, network, address string, err error)

// MetricsBroker is the entity used to send data tranmission
// information to an entity that is supposed to persist or
// handle the data accordingly.
type MetricsBroker interface {
	SendDataFlow(labels map[string]string, data *DataFlow)
}

// Interface is a wrapper around net.Interface and
// implements the core.Source interface, i.e. is it
// capable of providing network connections through
// the device it is referring to.
type Interface struct {
	ifi net.Interface

	// If OnDialErr is not nil, it is called each time that the
	// dialer is not able to create a network connection.
	OnDialErr DialHook

	metrics struct {
		sync.Mutex
		broker MetricsBroker
	}

	conns *conns
}

// SetMetricsBroker sets br as the default MetricsBroker of interface
// `i`. It is safe to use by multiple goroutines.
func (i *Interface) SetMetricsBroker(br MetricsBroker) {
	i.metrics.Lock()
	defer i.metrics.Unlock()

	i.metrics.broker = br
}

// Name implements the core.Source interface.
func (i *Interface) Name() string {
	return i.ifi.Name
}

// DialContext dials a connection of type `network` to `address`. If an error is
// encoutered, it is both returned and logged using the OnDialErr function, if available.
// `Follow` is called is called on the net.Conn before returning it.
// This function dials the connection using the interface's actual device as mean.
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

// Follow wraps the net.Conn around a Conn type, and keeps track of its
// callbacks, sending the metrics collected with the OnRead and OnWrite
// hooks.
// The connection is added to a set of followed connections, allowing
// the interface to perform operations on the entire list of open
// connections. The connection is removed from such list when the conn's
// OnClose function is called.
func (i *Interface) Follow(conn net.Conn) net.Conn {
	wconn := &Conn{Conn: conn}
	wconn.OnClose = func() {
		i.conns.Del(wconn)
	}
	wconn.OnRead = func(data *DataFlow) {
		i.SendMetrics(map[string]string{
			"source": i.Name(),
			"target": conn.RemoteAddr().String(),
		}, data)
	}
	wconn.OnWrite = func(data *DataFlow) {
		i.SendMetrics(map[string]string{
			"source": i.Name(),
			"target": conn.RemoteAddr().String(),
		}, data)
	}
	if i.conns == nil {
		i.conns = &conns{}
	}

	i.conns.Add(wconn)

	return wconn
}

// SendMetrics sends the data using the Interface's MetricsBroker.
// It is safe to use by multiple goroutines.
func (i *Interface) SendMetrics(labels map[string]string, data *DataFlow) {
	if i.metrics.broker == nil {
		return
	}

	i.metrics.Lock()
	defer i.metrics.Unlock()

	i.metrics.broker.SendDataFlow(labels, data)
}

// Close closes all open connections.
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

// Len returns the number of open connections.
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
