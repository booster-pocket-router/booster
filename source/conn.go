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
	Start time.Time // Start of the first data transmitted.
	End time.Time // Time of the last byte read/written. May be overridden multiple times.
	N int // Number of bytes transmitted.
	Avg float64 // Avg bytes/seconds.
}

// Begin computes the current time value, and sets it as the Start
// value of the data flow, if the field is still empty. It returns
// the time computed..
func (f *DataFlow) Begin() time.Time {
	start := time.Now()
	if f.Start.IsZero() {
		// These are the first bytes that we're collecing.
		f.Start = start
	}

	return start
}

// End computes the Avg, N, and End valus considering start as
// beginning of this data transmission.
func (f *DataFlow) Stop(start time.Time, n int) {
	if n == 0 {
		return
	}

	end := time.Now()
	d := end.Sub(start)

	f.N += n
	f.End = end

	avg := float64(n)/d.Seconds() // avg transmission speed of this connection.
	if f.Avg != 0 {
		// Make an average also with the last average computed
		avg = (f.Avg + avg) / 2
	}
	f.Avg = avg
}

// Conn is a wrapper around net.Conn, with the addition of some functions
// useful to uniquely identify the connection and receive callbacks on
// close events.
type Conn struct {
	net.Conn

	closed  bool // tells wether the connection was closed.
	OnClose func(download *DataFlow, upload *DataFlow) // Callback for close event.

	Download *DataFlow
	Upload *DataFlow
}

func WrapConn(c net.Conn) *Conn {
	return &Conn{
		Conn: c,
		Download: &DataFlow{},
		Upload: &DataFlow{},
	}
}

func (c *Conn) Read(p []byte) (int, error) {
	start := c.Download.Begin()
	n, err := c.Conn.Read(p) // Transmit the data.
	c.Download.Stop(start, n)

	return n, err
}

func (c *Conn) Write(p []byte) (int, error) {
	start := c.Upload.Begin()
	n, err := c.Conn.Write(p) // Transmit the data.
	c.Upload.Stop(start, n)

	return n, err
}

func (c *Conn) Close() error {
	if c.closed {
		// Multiple parts of the code might try to close the connection. Better be sure
		// that the underlying connection gets closed at some point, leave that code and
		// avoid repetitions here.
		return nil
	}
	if f := c.OnClose; f != nil {
		defer f(c.Download, c.Upload)
	}

	c.closed = true
	return c.Conn.Close()
}
