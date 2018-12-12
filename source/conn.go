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
	"fmt"
	"math/rand"
	"net"
	"time"
)

// conn is a wrapper around net.Conn, with the addition of some functions
// useful to uniquely identify the connection and receive callbacks on
// close events.
type conn struct {
	net.Conn

	id      string
	ref     string // Reference identifier, usually the parent's source identifier.
	closed  bool
	onClose func(string) // Callback for close event.
}

func newConn(c net.Conn, ref string) *conn {
	r := rand.Int()
	uuid := fmt.Sprintf("%s-%d", ref, r)
	if c != nil {
		uuid = fmt.Sprintf("%v-%v@%d-%d", c.LocalAddr(), c.RemoteAddr(), time.Now().UnixNano(), r)
	}

	return &conn{
		Conn: c,
		id:   uuid,
		ref:  ref,
	}
}

func (c *conn) uuid() string {
	return c.id
}

func (c *conn) Close() error {
	if c.closed {
		// Multiple parts of the code might try to close the connection. Better be sure
		// that the underlying connection gets closed at some point, leave that code and
		// avoid repetitions here.
		return nil
	}
	if f := c.onClose; f != nil {
		defer f(c.uuid())
	}

	c.closed = true
	return c.Conn.Close()
}
