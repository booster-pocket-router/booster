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

type IfaceSource struct {
	Name         string    `json:"name"`
	HardwareAddr string    `json:"hardware_addr"`
	Addrs        []string  `json:"addrs"`
	MTU          int       `json:"mtu"`
	Flags        net.Flags `json:"flags"`

	mux sync.Mutex
	N            int       `json:"-"`
}

func (i *IfaceSource) Addr(network string) (*net.TCPAddr, error) {
	// Use only ipv4 address to connect to localhost
	var addr string
	for _, v := range i.Addrs {
		ip := net.ParseIP(v)
		if ip == nil {
			continue
		}
		if v4 := ip.To4(); v4 != nil {
			addr = v
			break
		}
	}

	return net.ResolveTCPAddr(network, addr)
}

type ifaceConn struct {
	net.Conn

	s *IfaceSource
	closed bool
}

func (c *ifaceConn) Close() error {
	if c.closed {
		// Multiple parts of the code might try to close the connection. Better be sure
		// that the underlying connection gets closed at some point, leave that code and
		// avoid repetitions here.
		return nil
	}

	c.s.mux.Lock()
	c.s.N--
	log.Debug.Printf("Closing connection on iface(%v), left(%d)", c.s.ID(), c.s.N)
	c.s.mux.Unlock()

	c.closed = true
	return c.Conn.Close()
}

var _ core.Source = &IfaceSource{}

func (i *IfaceSource) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	laddr, err := i.Addr(network)
	if err != nil {
		return nil, err
	}
	// TODO: Respond to ctx.Done()
	d := &net.Dialer {
		LocalAddr: laddr,
	}
	conn, err := d.DialContext(ctx, network, address)
	if err != nil {
		return conn, err
	}

	i.mux.Lock()
	i.N++
	log.Debug.Printf("Opening connection to(%v) on iface(%v), left(%d)", address, i.ID(), i.N)
	i.mux.Unlock()

	return &ifaceConn{
		Conn: conn,
		s: i,
	}, nil
}

func (i *IfaceSource) ID() string {
	return i.Name
}

func (i *IfaceSource) Metrics() map[string]interface{} {
	return make(map[string]interface{})
}

