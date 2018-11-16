// +build darwin

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
	"errors"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
	"upspin.io/log"
)

func (i Interface) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// Find a suitable socket address from the interface
	var addr unix.Sockaddr

	addrs, err := i.Addrs()
	if err != nil {
		return nil, errors.New("Unable to retrieve interface addresses from interface " + i.Name + ": " + err.Error())
	}

	for _, v := range addrs {
		ip, _, err := net.ParseCIDR(v.String())
		if err != nil {
			return nil, errors.New("Unable to parse CIDR from interface " + i.Name + ": " + err.Error())
		}

		if ip4 := ip.To4(); ip4 != nil {
			// IPv4
			var buf [4]byte
			copy(buf[:], ip4[:4])
			addr = &unix.SockaddrInet4{
				Port: 0,
				Addr: buf,
			}

			break
		}
		// TODO(jecoz): Support ipv6
		if ip16 := ip.To16(); ip16 != nil {
		}
	}

	if addr == nil {
		return nil, errors.New("Unable to create a valid socket address from interface " + i.Name)
	}

	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if err := unix.Bind(int(fd), addr); err != nil {
					log.Debug.Printf("dialContext_unix error: unable to bind to interface %v: %v", i.Name, err)
				}
			})
		},
	}

	return d.DialContext(ctx, network, address)
}
