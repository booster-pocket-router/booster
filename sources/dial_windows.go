// +build windows

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

package sources

import (
	"context"
	"errors"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
	"upspin.io/log"
)

func (i Interface) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{
		// TODO: add windows implementation
		Control: func(network, address string, c syscall.RawConn) error {
			return errors.New("dialContext: Control not yet implemented on Windows")
		},
	}

	return d.DialContext(ctx, network, address)
}
