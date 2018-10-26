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
	"strings"

	"github.com/booster-proj/booster/sources"
	"github.com/booster-proj/core"
	"upspin.io/log"
)

type Booster struct {
	*core.Balancer
}

func (b *Booster) DialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	bl := make([]core.Source, 0, b.Len()) // blacklisted sources

	// If the dialing fails, keep on trying with the other sources until exaustion.
	for i := 0; len(bl) < b.Len(); i++ {
		var src core.Source
		src, err = b.Get(bl...)
		if err != nil {
			// Fail directly if the balancer returns an error, as
			// we do not have any source to use.
			return
		}

		log.Debug.Printf("DialContext: Attempt #%d to connect to %v (source %v)", i, address, src.ID())

		conn, err = src.DialContext(ctx, "tcp4", address)
		if err != nil {
			// Log this error, otherwise it will be silently skipped.
			log.Error.Printf("Unable to dial connection to %v using source %v. Error: %v", address, src.ID(), err)
			bl = append(bl, src)
			continue
		}

		// Connection dialed successfully.
		break
	}

	return
}

func GetFilteredInterfaces(s string) []*sources.Interface {
	ifs, err := net.Interfaces()
	if err != nil {
		log.Error.Printf("Unable to get interfaces: %v\n", err)
		return []*sources.Interface{}
	}

	l := make([]*sources.Interface, 0, len(ifs))

	for _, v := range ifs {
		log.Debug.Printf("Inspecting interface %+v\n", v)

		if len(v.HardwareAddr) == 0 {
			log.Debug.Printf("Empty hardware address. Skipping interface...")
			continue
		}

		if s != "" && !strings.Contains(v.Name, s) {
			log.Debug.Printf("Interface name does not satisfy name requirements: must contain \"%s\"", s)
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			// If the source does not contain an error
			log.Debug.Printf("Unable to get interface addresses: %v. Skipping interface...", err)
			continue
		}
		if len(addrs) == 0 {
			log.Debug.Printf("Empty unicast/multicast address list. Skipping interface...")
			continue
		}

		l = append(l, &sources.Interface{Interface: v})
	}

	return l
}
