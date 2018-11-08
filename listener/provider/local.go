/* Copyright (C) 2018 KIM KeepInMind GmbH/srl

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

package provider

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/booster-proj/booster/source"
	"upspin.io/log"
)

type Local struct {
}

func (l *Local) provide(ctx context.Context, level Confidence) ([]*source.Interface, error) {
	ift, err := net.Interfaces()
	if err != nil {
		return []*source.Interface{}, err
	}

	interfaces := make([]*source.Interface, 0, len(ift))
	for _, ifi := range ift {
		s := &source.Interface{Interface: ifi}
		if level <= Low {
			if err := pipeline(s, hasHardwareAddr, hasIP); err != nil {
				log.Debug.Printf("Local provider: low confidence pipeline: %v", err)
				continue
			}
		}
		if level <= High {
			if err := pipeline(s, hasNetworkConn); err != nil {
				log.Debug.Printf("Local provider: high confidence pipeline %v", err)
				continue
			}
		}

		interfaces = append(interfaces, s)
	}
	return interfaces, nil
}

type filter func(*source.Interface) error

func pipeline(ifi *source.Interface, ff ...filter) error {
	for _, f := range ff {
		if err := f(ifi); err != nil {
			return err
		}
	}
	return nil
}

func hasHardwareAddr(ifi *source.Interface) error {
	if len(ifi.HardwareAddr) == 0 {
		return fmt.Errorf("interface %s does not have a valid hardware address", ifi.Name)
	}
	return nil
}

func hasIP(ifi *source.Interface) error {
	addrs, err := ifi.Addrs()
	if err != nil {
		return fmt.Errorf("unable to get addresses of interface %s: %v", ifi.Name, err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("interface %s does not have any valid multicast/unicast address", ifi.Name)
	}

	var ok bool
	for _, v := range addrs {
		ip, _, err := net.ParseCIDR(v.String())
		if err != nil {
			continue
		}

		if ip4 := ip.To4(); ip4 != nil {
			ok = true
			continue
		}
		if ip16 := ip.To16(); ip16 != nil {
			ok = true
			continue
		}
	}
	if !ok {
		return fmt.Errorf("neither a valid IPv4 nor IPv6 was found in interface %s", ifi.Name)
	}

	return nil
}

func hasNetworkConn(ifi *source.Interface) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	conn, err := ifi.DialContext(ctx, "tcp", "google.com:80")
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
