// +build windows

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
