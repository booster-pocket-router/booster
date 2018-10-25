// +build linux

package booster

import (
	"context"
	"net"
	"syscall"

	"github.com/booster-proj/log"
	"golang.org/x/sys/unix"
)

func (i Interface) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				// Make the socket bind to the specified interface
				// before dialing the connection, which is created
				// through the device itself.
				if err := unix.BindToDevice(int(fd), i.Name); err != nil {
					log.Error.Printf("BindToDevice: %v", err)
				}
			})
		},
	}

	return d.DialContext(ctx, network, address)
}
