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

	"github.com/booster-proj/booster/core"
	"upspin.io/log"
)

// New returns an instance of a booster dialer.
func New(b *core.Balancer) core.Dialer {
	return &dialer{b}
}

type dialer struct {
	*core.Balancer
}

func (d *dialer) DialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	bl := make([]core.Source, 0, d.Len()) // blacklisted sources

	// If the dialing fails, keep on trying with the other sources until exaustion.
	for i := 0; len(bl) < d.Len(); i++ {
		var src core.Source
		src, err = d.Get(ctx, bl...)
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
