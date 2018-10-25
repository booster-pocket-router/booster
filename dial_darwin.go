// +build darwin
package booster

import (
	"context"
	"net"
	"syscall"
)

func (i Interface) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			// TODO(jecoz): Implement control setting the correct property
			// to the socket.
			return nil
		},
	}

	return d.DialContext(ctx, network, address)
}
