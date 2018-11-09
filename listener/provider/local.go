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

type inspection struct {
	at     time.Time
	active bool
}

var ttl time.Duration = time.Second * 15

type Local struct {
	// known is the list of interfaces that passed a the high confidence tests.
	known map[string]inspection
}

func (l *Local) provide(ctx context.Context, level Confidence) ([]*source.Interface, error) {
	ift, err := net.Interfaces()
	if err != nil {
		return []*source.Interface{}, err
	}

	interfaces := make([]*source.Interface, 0, len(ift))
	for _, ifi := range ift {
		if s := l.filter(&source.Interface{Interface: ifi}, level); s != nil {
			interfaces = append(interfaces, s)
		}
	}

	// Update known interfaces
	return interfaces, nil
}

func (l *Local) filter(ifi *source.Interface, level Confidence) *source.Interface {
	if level == High {
		// If it is required to make a high confidence test, just do it.
		return l.makeChecks(ifi, level)
	}

	ifi = l.makeChecks(ifi, level) // perform the low level tests first
	if ifi == nil {
		return nil
	}

	// In case of a low confidence test, check if this interfaces is already in the
	// known list. In that case use it as a measure.
	// This way we avoid to keep on adding an interface that appers active to the
	// low confidence checks, but is actually not active to the eyes of the high
	// confidence ones.

	// Side effect: this way we're not able to detect that an interface became
	// active again after having failed just the high level tests once.
	if inspection, ok := l.known[ifi.Name]; ok {
		if inspection.active == false {
			// This means that the low confidence test returned ok, but
			// the high one tells us that the interface does not actually
			// provide internet connections.
			return nil
		}

		if time.Now().Sub(inspection.at) <= ttl {
			return ifi // we consider the last high confidence test stil valid
		} else {
			// Perform an high confidence test again, because the one that we've
			// performed is outdated
			return l.makeChecks(ifi, High)
		}
	}

	// If we reach this point, it means that we've encountered an interface up to
	// now not known. Perform the high confidence tests on it in any case.
	return l.makeChecks(ifi, High)
}

func (l *Local) makeChecks(ifi *source.Interface, level Confidence) *source.Interface {
	checks := []check{hasHardwareAddr, hasIP}
	if level == High {
		checks = append(checks, hasNetworkConn)
	}

	if l.known == nil {
		l.known = make(map[string]inspection)
	}

	_ifi, err := pipeline(ifi, checks...)
	if err != nil {
		log.Debug.Printf("Local provider: pipeline with confidence (%d): %v", level, err)
	}

	l.updateKnown(ifi, level, err != nil)
	return _ifi
}

func (l *Local) updateKnown(ifi *source.Interface, level Confidence, failed bool) {
	if level != High {
		return
	}

	l.known[ifi.Name] = inspection{
		active: !failed,
		at:     time.Now(),
	}
}

type check func(*source.Interface) error

func pipeline(ifi *source.Interface, checks ...check) (*source.Interface, error) {
	for _, f := range checks {
		if err := f(ifi); err != nil {
			return nil, err
		}
	}
	return ifi, nil
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
		return fmt.Errorf("unable to dial connection using interface %s: %v", ifi.Name, err)
	}
	conn.Close()
	return nil
}
