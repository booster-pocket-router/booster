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
	"net"
	"time"
)

// DataFlow collects data about a data tranmission.
type DataFlow struct {
	Type      string
	StartedAt time.Time // Start of the first data transmitted.
	EndedAt   time.Time // Time of the last byte read/written. May be overridden multiple times.
	N         int       // Number of bytes transmitted.
	Avg       float64   // Avg bytes/seconds.
}

// Begin sets the data flow start value.
func (f *DataFlow) Start() {
	f.StartedAt = time.Now()
}

// End computes the Avg, N, and End valus considering start as
// beginning of this data transmission.
func (f *DataFlow) Stop(n int) {
	end := time.Now()
	d := end.Sub(f.StartedAt)

	f.N += n
	f.EndedAt = end

	avg := float64(n) / d.Seconds() // avg transmission speed of this connection.
	f.Avg = avg
}

// Conn is a wrapper around net.Conn, with the addition of some functions
// useful to uniquely identify the connection and receive callbacks on
// close events.
type Conn struct {
	net.Conn

	closed     bool   // tells wether the connection was closed.
	OnClose    func() // Callback for close event.
	OnRead func(df *DataFlow)
	OnWrite   func(df *DataFlow)
}

// Read is the io.Reader implementation of Conn. It forwards the request
// to the underlying net.Conn, but it also records the number of bytes
// tranferred and the duration of the transmission. It then exposes the
// data using the OnRead callback.
func (c *Conn) Read(p []byte) (int, error) {
	dl := &DataFlow{Type: "read"}
	dl.Start()
	n, err := c.Conn.Read(p) // Transmit the data.

	go func() {
		if n == 0 {
			return
		}
		dl.Stop(n)
		if f := c.OnRead; f != nil {
			f(dl)
		}
	}()

	return n, err
}

// Write is the io.Writer implementation of Conn. It forwards the request
// to the underlying net.Conn, but it also records the number of bytes
// tranferred and the duration of the transmission. It then exposes the
// data using the OnWrite callback.
func (c *Conn) Write(p []byte) (int, error) {
	upl := &DataFlow{Type: "write"}
	upl.Start()
	n, err := c.Conn.Write(p) // Transmit the data.

	go func() {
		if n == 0 {
			return
		}
		upl.Stop(n)
		if f := c.OnWrite; f != nil {
			f(upl)
		}
	}()

	return n, err
}

// Close closes the underlying net.Conn, calling the OnClose callback
// afterwards.
func (c *Conn) Close() error {
	if c.closed {
		// Multiple parts of the code might try to close the connection. Better be sure
		// that the underlying connection gets closed at some point, leave that code and
		// avoid repetitions here.
		return nil
	}
	if f := c.OnClose; f != nil {
		defer f()
	}

	c.closed = true
	return c.Conn.Close()
}
