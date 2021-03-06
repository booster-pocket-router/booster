// Copyright © 2019 KIM KeepInMind GmbH/srl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package source

import (
	"context"
	"fmt"
	"net"
	"time"

	"upspin.io/log"
)

type Local struct {
}

func (l *Local) Provide(ctx context.Context, level Confidence) ([]*Interface, error) {
	ift, err := net.Interfaces()
	if err != nil {
		return []*Interface{}, err
	}

	interfaces := make([]*Interface, 0, len(ift))
	for _, ifi := range ift {
		if s := l.filter(&Interface{ifi: ifi}, level); s != nil {
			interfaces = append(interfaces, s)
		}
	}

	return interfaces, nil
}

func (l *Local) Check(ctx context.Context, ifi *Interface, level Confidence) error {
	checks := []check{hasHardwareAddr, hasIP}
	if level == High {
		checks = append(checks, hasNetworkConnRetry)
	}

	return pipeline(ctx, ifi, checks...)
}

func (l *Local) filter(ifi *Interface, level Confidence) *Interface {
	if err := l.Check(context.Background(), ifi, level); err != nil {
		log.Debug.Printf("Local provider: pipeline with confidence (%d): %v", level, err)
		return nil
	}
	return ifi
}

type check func(context.Context, *Interface) error

func pipeline(ctx context.Context, ifi *Interface, checks ...check) error {
	for _, f := range checks {
		if err := f(ctx, ifi); err != nil {
			return err
		}
	}
	return nil
}

func hasHardwareAddr(ctx context.Context, ifi *Interface) error {
	if len(ifi.ifi.HardwareAddr) == 0 {
		return fmt.Errorf("interface %s does not have a valid hardware address", ifi.ID())
	}
	return nil
}

func hasIP(ctx context.Context, ifi *Interface) error {
	addrs, err := ifi.ifi.Addrs()
	if err != nil {
		return fmt.Errorf("unable to get addresses of interface %s: %v", ifi.ID(), err)
	}
	if len(addrs) == 0 {
		return fmt.Errorf("interface %s does not have any valid multicast/unicast address", ifi.ID())
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
		return fmt.Errorf("neither a valid IPv4 nor IPv6 was found in interface %s", ifi.ID())
	}

	return nil
}

func hasNetworkConn(ctx context.Context, ifi *Interface) error {
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()

	conn, err := ifi.DialContext(ctx, "tcp", "google.com:80")
	if err != nil {
		return fmt.Errorf("unable to dial connection using interface %s: %v", ifi.ID(), err)
	}
	conn.Close()
	return nil
}

func hasNetworkConnRetry(ctx context.Context, ifi *Interface) error {
	for i := 0; i < 3; i++ {
		if i == 2 {
			// last item
			return hasNetworkConn(ctx, ifi)
		}

		if err := hasNetworkConn(ctx, ifi); err == nil {
			return nil
		}

		select {
		case <-time.After(500 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil // will not be reached
}
