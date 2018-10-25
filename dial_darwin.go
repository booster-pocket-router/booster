// +build darwin

package booster

import (
	"context"
	"net"
	"syscall"
	"errors"

	"github.com/booster-proj/log"
	"golang.org/x/sys/unix"
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

		if ip.To4() != nil {
			// IPv4
			var buf [4]byte
			copy(buf[:], ip)

			addr = &unix.SockaddrInet4{
				Port: 0,
				Addr: buf,
			}
			break
		}
		// TODO(jecoz): Support ipv6
	}

	if addr == nil {
		return nil, errors.New("Unable to create a valid socket address from interface " + i.Name)
	}

	d := &net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if err := unix.Bind(int(fd), addr); err != nil {
					log.Error.Printf("Unable to bind to interface %v: %v", i.Name, err)
				}
			})
		},
	}

	return d.DialContext(ctx, network, address)
}

